package server

import (
	"encoding/json"
	"fmt"
	"strings"
)

func buildChatPrompt(problem Problem, attempt Attempt, messages []ChatMessage, userMessage string, englishMode bool, memory ProblemMemory) string {
	conversation := []string{}
	for _, message := range messages {
		conversation = append(conversation, fmt.Sprintf("%s: %s", message.Role, message.Content))
	}
	mode := "日本語で会話してください。"
	if englishMode {
		mode = "English interview mode is enabled. Ask the user to explain their reasoning in English, and respond in English."
	}
	testCases, _ := json.Marshal(problem.TestCases)
	memoryJSON, _ := json.MarshalIndent(memory, "", "  ")
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
- Give gradual hints when the user is stuck.
- Before implementation, ask the candidate how they intend to solve it. Require a brute-force baseline, an optimized direction, data structures, invariants, and edge cases before you approve coding.
- If they jump to code too early, pause them and ask for the plan first.
- Require multiple approaches when reasonable. If they only have one, nudge with a restrained hint without giving away the full solution.
- Focus on understanding, constraints, brute force, optimization, edge cases, invariants, complexity, and implementation details.
- After code is written, review correctness and ask follow-ups about proof, failure modes, time complexity, and space complexity.
- If this is a repeated problem, use the candidate memory to ask a new, harder follow-up based on old mistakes. Avoid simply repeating an old question unless you are checking recovery.
- Keep a high bar. Be kind, but do not accept vague explanations.
- In real mode, act as if the candidate cannot run code or use autocomplete. Ask for dry runs and manual verification.
- If the candidate is silent or vague, ask them to verbalize the exact invariant, next branch, or proof gap.
- If the user is wrong, point it out clearly without rewriting the full answer.
- %s

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

Respond as the interviewer. Keep it concise and interview-like. Ask exactly one main question, with at most two short supporting prompts.`, attempt.CompanyPreset, aiProviderLabel(attempt.AIProvider), attempt.CompanyPreset, attempt.InterviewMode, attempt.CurrentPhase, attempt.TimeLimitSeconds, !attempt.NoRun, !attempt.NoAutocomplete, attempt.RequiresPlan, companyInterviewInstructions(attempt.CompanyPreset), mode, problem.Title, problem.Difficulty, problem.Pattern, problem.Statement, strings.Join(problem.Constraints, "; "), string(testCases), string(memoryJSON), attempt.Code, strings.Join(conversation, "\n"), userMessage)
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
8. follow-up questions that stress proof, complexity, and Google-level rigor
9. previous mistakes and whether this submission shows recovery
10. a concrete discussion plan for the next interviewer exchange
11. a scorecard across the interview dimensions
12. a hire recommendation: Strong Hire, Hire, Lean Hire, Lean No Hire, or No Hire
13. shadow evaluator notes that cite observable signals
14. mini-round questions for coding follow-up, behavioral, and system-design/product thinking
15. weakness signals to store for future sessions

Use scorecard scores from 1 to 4:
1 = below bar, 2 = weak / inconsistent, 3 = meets bar, 4 = strong signal.

Be strict. Passing local tests is not enough. Penalize missing clarifying questions, missing brute force, weak dry run, hand-wavy complexity, no proof, slow pacing, or dependency on running code.`, attempt.CompanyPreset, problem.Title, problem.Difficulty, problem.Pattern, problem.Statement, strings.Join(problem.Constraints, "\n"), aiProviderLabel(attempt.AIProvider), attempt.CompanyPreset, attempt.InterviewMode, !attempt.NoRun, !attempt.NoAutocomplete, companyEvaluationInstructions(attempt.CompanyPreset), string(memoryJSON), code, testResult)
}

func companyInterviewInstructions(companyPreset string) string {
	switch normalizeCompanyPreset(companyPreset) {
	case "meta":
		return "- Pace matters. Prefer concise prompts. If the candidate finishes early, move to a second variant or follow-up.\n- Expect clean code without running it. Push for quick dry runs and edge cases.\n- Penalize over-explaining or spending too long before implementation."
	case "amazon":
		return "- Ask for trade-offs, customer-impacting edge cases, and operational failure modes.\n- Mix in one behavioral-style follow-up when appropriate, but keep the coding problem central.\n- Evaluate whether the candidate makes pragmatic decisions under constraints."
	case "google":
		return "- Go deep on ambiguity, invariants, proof of correctness, and generalized follow-ups.\n- Prefer one problem explored thoroughly over rushing.\n- Push the candidate to justify why the optimized approach is actually correct."
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
		return "Score problem decomposition, proof, invariants, optimality, edge cases, communication, and ability to handle deeper follow-ups."
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
