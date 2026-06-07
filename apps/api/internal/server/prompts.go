package server

import (
	"encoding/json"
	"fmt"
	"strings"
)

func buildChatPrompt(problem Problem, attempt Attempt, messages []ChatMessage, userMessage string, englishMode bool) string {
	conversation := []string{}
	for _, message := range messages {
		conversation = append(conversation, fmt.Sprintf("%s: %s", message.Role, message.Content))
	}
	mode := "日本語で会話してください。"
	if englishMode {
		mode = "English interview mode is enabled. Ask the user to explain their reasoning in English, and respond in English."
	}
	testCases, _ := json.Marshal(problem.TestCases)
	return fmt.Sprintf(`You are a strict but supportive Google-style technical interviewer.

Your job is to guide the candidate through the problem with Socratic questions. Do not reveal the final solution unless the user explicitly asks for it or has already solved the problem. Ask only one main question at a time.

Rules:
- Prefer questions over direct answers.
- Give gradual hints when the user is stuck.
- Focus on understanding, constraints, brute force, optimization, edge cases, invariants, complexity, and implementation details.
- After code is written, review correctness and ask the user to explain the algorithm.
- Always check time and space complexity.
- If the user is wrong, point it out clearly without rewriting the full answer.
- %s

Problem:
Title: %s
Difficulty: %s
Pattern: %s
Statement: %s
Constraints: %s
Test cases: %s

Current code:
%s

Conversation so far:
%s

Candidate message:
%s

Respond as the interviewer. Keep it concise and interview-like.`, mode, problem.Title, problem.Difficulty, problem.Pattern, problem.Statement, strings.Join(problem.Constraints, "; "), string(testCases), attempt.Code, strings.Join(conversation, "\n"), userMessage)
}

func buildReviewPrompt(problem Problem, attempt Attempt, code string) string {
	testResult := "まだテストは実行されていません。"
	if attempt.TestResult != nil {
		payload, _ := json.MarshalIndent(attempt.TestResult, "", "  ")
		testResult = string(payload)
	}
	return fmt.Sprintf(`You are a senior software engineer reviewing a coding interview answer.

Review the candidate's Python solution for this problem:

Title: %s
Difficulty: %s
Pattern: %s
Statement:
%s

Constraints:
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
7. alternative approaches
8. mistakes to remember`, problem.Title, problem.Difficulty, problem.Pattern, problem.Statement, strings.Join(problem.Constraints, "\n"), code, testResult)
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
    "mistakes_to_remember": { "type": "array", "items": { "type": "string" } },
    "next_review_recommendation": { "type": "string" }
  },
  "required": [
    "is_correct",
    "summary",
    "bugs",
    "edge_cases",
    "complexity",
    "readability_feedback",
    "interview_feedback",
    "mistakes_to_remember",
    "next_review_recommendation"
  ]
}`
}
