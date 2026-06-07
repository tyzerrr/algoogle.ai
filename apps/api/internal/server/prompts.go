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
	return fmt.Sprintf(`You are a strict but supportive Google-style technical interviewer.

Your job is to guide the candidate through the problem with Socratic questions. Do not reveal the final solution unless the user explicitly asks for it or has already solved the problem. Ask only one main question at a time.

Rules:
- Prefer questions over direct answers.
- Give gradual hints when the user is stuck.
- Before implementation, ask the candidate how they intend to solve it. Require a brute-force baseline, an optimized direction, data structures, invariants, and edge cases before you approve coding.
- If they jump to code too early, pause them and ask for the plan first.
- Require multiple approaches when reasonable. If they only have one, nudge with a restrained hint without giving away the full solution.
- Focus on understanding, constraints, brute force, optimization, edge cases, invariants, complexity, and implementation details.
- After code is written, review correctness and ask follow-ups about proof, failure modes, time complexity, and space complexity.
- If this is a repeated problem, use the candidate memory to ask a new, harder follow-up based on old mistakes. Avoid simply repeating an old question unless you are checking recovery.
- Keep a high Google interview bar. Be kind, but do not accept vague explanations.
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

Respond as the interviewer. Keep it concise and interview-like. Ask exactly one main question, with at most two short supporting prompts.`, mode, problem.Title, problem.Difficulty, problem.Pattern, problem.Statement, strings.Join(problem.Constraints, "; "), string(testCases), string(memoryJSON), attempt.Code, strings.Join(conversation, "\n"), userMessage)
}

func buildReviewPrompt(problem Problem, attempt Attempt, code string, memory ProblemMemory) string {
	testResult := "まだテストは実行されていません。"
	if attempt.TestResult != nil {
		payload, _ := json.MarshalIndent(attempt.TestResult, "", "  ")
		testResult = string(payload)
	}
	memoryJSON, _ := json.MarshalIndent(memory, "", "  ")
	return fmt.Sprintf(`You are a senior software engineer reviewing a coding interview answer.

Review the candidate's Python solution for this problem:

Title: %s
Difficulty: %s
Pattern: %s
Statement:
%s

Constraints:
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

Be strict. If the solution passes simple tests but has unproven assumptions, mark the risk clearly and ask follow-ups.`, problem.Title, problem.Difficulty, problem.Pattern, problem.Statement, strings.Join(problem.Constraints, "\n"), string(memoryJSON), code, testResult)
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
    "discussion_plan",
    "next_review_recommendation"
  ]
}`
}
