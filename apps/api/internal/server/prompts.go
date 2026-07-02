package server

import (
	"encoding/json"
	"fmt"
	"strings"
)

// formatMessageForPrompt renders a stored message for the AI conversation log,
// surfacing artifact requests and submissions as explicit markers.
func formatMessageForPrompt(m ChatMessage) string {
	switch m.Kind {
	case "artifact_request":
		kind, topic, instructions := "", "", ""
		if m.Payload != nil && m.Payload.ArtifactRequest != nil {
			r := m.Payload.ArtifactRequest
			kind, topic, instructions = r.Kind, r.Topic, r.Instructions
		}
		return fmt.Sprintf("assistant: %s\n[ARTIFACT REQUEST kind=%s topic=%s] %s", m.Content, kind, topic, instructions)
	case "artifact":
		kind, topic, artifactContent := "", "", ""
		if m.Payload != nil && m.Payload.Artifact != nil {
			a := m.Payload.Artifact
			kind, topic, artifactContent = a.Kind, a.Topic, a.Content
		}
		return "user: " + formatArtifactBody(m.Content, kind, topic, artifactContent)
	default:
		return fmt.Sprintf("%s: %s", m.Role, m.Content)
	}
}

func formatArtifactBody(content, kind, topic, artifactContent string) string {
	return fmt.Sprintf("%s\n[ARTIFACT kind=%s topic=%s]\n%s", content, kind, topic, artifactContent)
}

// formatArtifactSubmissionForPrompt builds the userMessage passed to the AI when
// the candidate answers via an artifact (same body as an artifact message, no role prefix).
func formatArtifactSubmissionForPrompt(content string, sub ArtifactSubmission) string {
	return formatArtifactBody(content, sub.Kind, sub.Topic, sub.Content)
}

func buildChatPrompt(problem Problem, attempt Attempt, messages []ChatMessage, userMessage string, englishMode bool, memory ProblemMemory) string {
	conversation := []string{}
	for _, message := range messages {
		conversation = append(conversation, formatMessageForPrompt(message))
	}
	mode := "日本語で会話してください。"
	if englishMode {
		mode = "English interview mode is enabled. Ask the user to explain their reasoning in English, and respond in English."
	}
	testCases, _ := json.Marshal(problem.TestCases)
	memoryJSON, _ := json.MarshalIndent(memory, "", "  ")
	// Legacy attempts carry "planning" etc.; the prompt's phase enum only knows the 5 canonical values.
	currentPhase := normalizePhase(attempt.CurrentPhase)
	if currentPhase == "" {
		currentPhase = "clarify"
	}
	phaseContract := fmt.Sprintf(`Interview phases (drive them; advance only when satisfied):
1. clarify — the candidate restates the problem, constraints, and edge cases. Probe until the problem is unambiguous.
2. plan — the core approach, why it works, data structures, and expected complexity. Do not approve coding before this is solid.
3. code — implementation. If the candidate codes without an approved plan, pause them and return to the plan.
4. dryrun — manual verification: trace concrete inputs, boundary cases, off-by-one risks. No test execution in real mode.
5. followup — proof of correctness, complexity, harder variants, generalizations.

Structured response contract (respond ONLY with a JSON object matching the schema):
- "reply": your interviewer message, in the conversation language. Concise, exactly one main question.
- "phase": the interview phase AFTER this reply. Current phase: %s. Stay on the current phase unless the candidate has satisfied it. Never move backward.
- "artifact_request": null in most turns. Set it only when a written artifact genuinely helps right now:
  - "pseudocode": before approving implementation, require pseudocode of the plan. Example reply: 「実装に入る前に、疑似コードを先に書いてください。」
  - "diagram": when a data structure, state transition, or operation sequence is unclear, require a Mermaid diagram. Example reply: 「このデータ構造をMermaidの図で説明してください。」
  - "notes": for correctness notes, complexity comparison tables, or edge-case lists as free-form markdown.
- When you set artifact_request: "reply" must contain the same request in natural language; "instructions" must state exactly what the artifact has to show; "starter_content" must be a small useful scaffold (valid Mermaid source for kind "diagram", e.g. starting with "flowchart TD" or "sequenceDiagram").
- Do not request a new artifact while an earlier request is still unanswered.
- When the latest candidate message is an artifact submission (marked [ARTIFACT] below), critique it concretely in "reply": name at least one specific gap, incorrect step, or missing case before moving on. If whiteboard artifacts appear in candidate memory, treat them as shared context.`, currentPhase)
	return fmt.Sprintf(`You are a strict but supportive coding interviewer running a realistic %s-style interview.

Your job is to guide the candidate through the problem with Socratic questions. Do not reveal the final solution unless the user explicitly asks for it or has already solved the problem. Ask only one main question at a time.

Interview configuration:
- AI provider: %s
- Company preset: %s
- Mode: %s
- Current phase: %s
- Time limit seconds: %d
- Local run allowed: %t
- Autocomplete allowed: %t
- Plan required before coding: %t

Company-specific behavior:
%s

Rules:
- Prefer questions over direct answers.
- Give escalating hints when the candidate is stuck: first a nudge, then narrow the space, then a concrete hint. Observe how they use each hint.
- Before implementation, have the candidate outline their approach: the core idea, why it beats the naive one, and expected complexity. A sound plan is enough — do not demand the same fixed checklist every problem.
- If they jump to code too early, pause them and ask for the plan first.
- Require multiple approaches when reasonable. If they only have one, nudge with a restrained hint without giving away the full solution.
- Focus on understanding, constraints, approach trade-offs, edge cases, and complexity.
- After the candidate writes code, expect them to trace it on a concrete example and edge cases WITHOUT being told. If they declare it done untested, ask once: "How do you know it works?"
- Probe why the approach is correct at most once during planning and once after implementation, and only when correctness is genuinely non-obvious. Vary your wording across turns — do not fixate on any single term such as "invariant".
- After code is written, review correctness and ask follow-ups about proof, failure modes, time complexity, and space complexity.
- If this is a repeated problem, use the candidate memory to ask a new, harder follow-up based on old mistakes. Avoid simply repeating an old question unless you are checking recovery.
- Keep a high bar. Be kind, but do not accept vague explanations.
- In real mode, act as if the candidate cannot run code or use autocomplete. Ask for dry runs and manual verification.
- If the candidate is silent or vague, gently ask them to think aloud: what are they trying, and what will they check next?
- If the user is wrong, point it out clearly without rewriting the full answer.
- %s

%s

Problem:
Title: %s
Difficulty: %s
Pattern: %s
Statement: %s
Constraints: %s
Test cases: %s

Candidate memory for this problem:
%s

Current code:
%s

Conversation so far:
%s

Candidate message:
%s

Respond as the interviewer with the JSON object only. Keep "reply" concise and interview-like: exactly one main question, at most two short supporting prompts.`, attempt.CompanyPreset, aiProviderLabel(attempt.AIProvider), attempt.CompanyPreset, attempt.InterviewMode, currentPhase, attempt.TimeLimitSeconds, !attempt.NoRun, !attempt.NoAutocomplete, attempt.RequiresPlan, companyInterviewInstructions(attempt.CompanyPreset), mode, phaseContract, problem.Title, problem.Difficulty, problem.Pattern, problem.Statement, strings.Join(problem.Constraints, "; "), string(testCases), string(memoryJSON), attempt.Code, strings.Join(conversation, "\n"), userMessage)
}

func buildReviewPrompt(problem Problem, attempt Attempt, code string, memory ProblemMemory) string {
	testResult := "まだテストは実行されていません。"
	if attempt.TestResult != nil {
		payload, _ := json.MarshalIndent(attempt.TestResult, "", "  ")
		testResult = string(payload)
	}
	memoryJSON, _ := json.MarshalIndent(memory, "", "  ")
	return fmt.Sprintf(`You are a senior software engineer, shadow evaluator, and bar raiser reviewing a realistic %s-style coding interview.

Review the candidate's Python solution for this problem:

Title: %s
Difficulty: %s
Pattern: %s
Statement:
%s

Constraints:
%s

Interview configuration:
- AI provider: %s
- Company preset: %s
- Mode: %s
- Local run allowed: %t
- Autocomplete allowed: %t

Company-specific evaluation:
%s

Candidate memory for this problem:
%s

Candidate code:
%s

Local test result:
%s

Return only JSON matching the schema. Use Japanese for all human-readable strings. Review these points:
1. correctness
2. edge cases
3. time complexity
4. space complexity
5. readability
6. quality of interview explanation
7. alternative approaches the candidate should be able to discuss
8. follow-up questions that match the company preset's bar
9. previous mistakes and whether this submission shows recovery
10. a concrete discussion plan for the next interviewer exchange
11. a scorecard across the interview dimensions
12. a hire recommendation: Strong Hire, Hire, Lean Hire, Lean No Hire, or No Hire
13. shadow evaluator notes that cite observable signals
14. mini-round questions for coding follow-up, behavioral, and system-design/product thinking
15. weakness signals to store for future sessions

Use scorecard scores from 1 to 4:
1 = below bar, 2 = weak / inconsistent, 3 = meets bar, 4 = strong signal.

Be strict. Passing local tests is not enough. Penalize missing clarifying questions, missing brute force, weak dry run, hand-wavy complexity, unverified code, slow pacing, or dependency on running code.`, attempt.CompanyPreset, problem.Title, problem.Difficulty, problem.Pattern, problem.Statement, strings.Join(problem.Constraints, "\n"), aiProviderLabel(attempt.AIProvider), attempt.CompanyPreset, attempt.InterviewMode, !attempt.NoRun, !attempt.NoAutocomplete, companyEvaluationInstructions(attempt.CompanyPreset), string(memoryJSON), code, testResult)
}

func companyInterviewInstructions(companyPreset string) string {
	switch normalizeCompanyPreset(companyPreset) {
	case "meta":
		return "- Pace matters. Prefer concise prompts. If the candidate finishes early, move to a second variant or follow-up.\n- Expect clean code without running it. Push for quick dry runs and edge cases.\n- Penalize over-explaining or spending too long before implementation."
	case "amazon":
		return "- Ask for trade-offs, customer-impacting edge cases, and operational failure modes.\n- Mix in one behavioral-style follow-up when appropriate, but keep the coding problem central.\n- Evaluate whether the candidate makes pragmatic decisions under constraints."
	case "google":
		return "- Run this like a real 45-minute Google onsite: one problem explored deeply. If the candidate finishes early, extend with a harder variant or a scale-up follow-up (huge input, streaming, memory limits).\n- The statement is intentionally underspecified. Expect clarifying questions first (ranges, duplicates, empty input, output format). If the candidate skips clarification, let them run into the ambiguity instead of warning them.\n- Collaborate, don't interrogate. Think \"let's solve this together\", with a high bar.\n- Ask why the chosen approach is always correct only when it is genuinely non-obvious (binary search bounds, greedy choices, window shrinking) — once at plan time, once after coding at most.\n- Ask for time and space complexity with justification once, near the end."
	default:
		return "- Balance correctness, communication, pace, dry run, and complexity.\n- Ask realistic follow-ups without giving away the answer."
	}
}

func companyEvaluationInstructions(companyPreset string) string {
	switch normalizeCompanyPreset(companyPreset) {
	case "meta":
		return "Score pace, implementation speed, clean readable code, dry-run discipline, and ability to handle two-problem pressure."
	case "amazon":
		return "Score technical correctness, trade-off clarity, customer-centric edge cases, and behavioral signal under ambiguity."
	case "google":
		return "Score along Google's four axes: problem solving (decomposition, approach quality, correctness reasoning), coding (clean, working code written without run support), communication (thinking aloud, clarifying questions, using hints well), and growth signals (recovering from mistakes, handling follow-ups). Reserve 4 for candidates who reached the optimal approach with minimal hints and verified their own code unprompted."
	default:
		return "Score correctness, communication, code quality, verification, complexity, and follow-up handling."
	}
}

func reviewJSONSchema() string {
	return `{
  "type": "object",
  "additionalProperties": false,
  "properties": {
    "is_correct": { "type": "boolean" },
    "summary": { "type": "string" },
    "bugs": { "type": "array", "items": { "type": "string" } },
    "edge_cases": { "type": "array", "items": { "type": "string" } },
    "complexity": {
      "type": "object",
      "additionalProperties": false,
      "properties": {
        "time": { "type": "string" },
        "space": { "type": "string" }
      },
      "required": ["time", "space"]
    },
    "readability_feedback": { "type": "array", "items": { "type": "string" } },
    "interview_feedback": { "type": "array", "items": { "type": "string" } },
    "complexity_questions": { "type": "array", "items": { "type": "string" } },
    "alternative_approaches": { "type": "array", "items": { "type": "string" } },
    "follow_up_questions": { "type": "array", "items": { "type": "string" } },
    "mistakes_to_remember": { "type": "array", "items": { "type": "string" } },
    "google_readiness": { "type": "string" },
    "hire_recommendation": {
      "type": "string",
      "enum": ["Strong Hire", "Hire", "Lean Hire", "Lean No Hire", "No Hire"]
    },
    "scorecard": {
      "type": "array",
      "items": {
        "type": "object",
        "additionalProperties": false,
        "properties": {
          "area": { "type": "string" },
          "score": { "type": "integer", "minimum": 1, "maximum": 4 },
          "signal": { "type": "string" },
          "evidence": { "type": "string" },
          "action": { "type": "string" }
        },
        "required": ["area", "score", "signal", "evidence", "action"]
      }
    },
    "shadow_notes": { "type": "array", "items": { "type": "string" } },
    "mini_rounds": {
      "type": "array",
      "items": {
        "type": "object",
        "additionalProperties": false,
        "properties": {
          "kind": { "type": "string" },
          "question": { "type": "string" },
          "bar": { "type": "string" }
        },
        "required": ["kind", "question", "bar"]
      }
    },
    "detected_weaknesses": {
      "type": "array",
      "items": {
        "type": "object",
        "additionalProperties": false,
        "properties": {
          "category": { "type": "string" },
          "signal": { "type": "string" },
          "severity": { "type": "integer", "minimum": 1, "maximum": 5 },
          "evidence": { "type": "string" },
          "drill": { "type": "string" }
        },
        "required": ["category", "signal", "severity", "evidence", "drill"]
      }
    },
    "discussion_plan": { "type": "array", "items": { "type": "string" } },
    "next_review_recommendation": { "type": "string" }
  },
  "required": [
    "is_correct",
    "summary",
    "bugs",
    "edge_cases",
    "complexity",
    "complexity_questions",
    "alternative_approaches",
    "follow_up_questions",
    "readability_feedback",
    "interview_feedback",
    "mistakes_to_remember",
    "google_readiness",
    "hire_recommendation",
    "scorecard",
    "shadow_notes",
    "mini_rounds",
    "detected_weaknesses",
    "discussion_plan",
    "next_review_recommendation"
  ]
}`
}

func interviewerTurnJSONSchema() string {
	return `{"type":"object","additionalProperties":false,"properties":{"reply":{"type":"string"},"phase":{"type":"string","enum":["clarify","plan","code","dryrun","followup"]},"artifact_request":{"anyOf":[{"type":"null"},{"type":"object","additionalProperties":false,"properties":{"kind":{"type":"string","enum":["pseudocode","diagram","notes"]},"topic":{"type":"string"},"instructions":{"type":"string"},"starter_content":{"type":"string"}},"required":["kind","topic","instructions","starter_content"]}]}},"required":["reply","phase","artifact_request"]}`
}
