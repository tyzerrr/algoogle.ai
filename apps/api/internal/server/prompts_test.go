package server

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestInterviewerTurnJSONSchemaIsValidJSON(t *testing.T) {
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(interviewerTurnJSONSchema()), &parsed); err != nil {
		t.Fatalf("interviewer turn schema is not valid JSON: %v", err)
	}
	props, ok := parsed["properties"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected properties object in schema")
	}
	for _, key := range []string{"reply", "phase", "artifact_request"} {
		if _, ok := props[key]; !ok {
			t.Fatalf("schema missing property %q", key)
		}
	}
}

func TestBuildChatPromptContainsPhaseContract(t *testing.T) {
	attempt := Attempt{CompanyPreset: "google", InterviewMode: "real", CurrentPhase: "clarify"}
	prompt := buildChatPrompt(Problem{Title: "Two Sum"}, attempt, nil, "hi", false, ProblemMemory{})
	if !strings.Contains(prompt, "Interview phases") {
		t.Fatalf("expected phase contract in prompt")
	}
	if !strings.Contains(prompt, "Structured response contract") {
		t.Fatalf("expected structured response contract in prompt")
	}
	if !strings.Contains(prompt, "Current phase: clarify") {
		t.Fatalf("expected interpolated current phase in prompt")
	}
	if !strings.Contains(prompt, "JSON object only") {
		t.Fatalf("expected JSON-only closing instruction")
	}
}

func TestBuildChatPromptNormalizesLegacyPhase(t *testing.T) {
	attempt := Attempt{CompanyPreset: "google", InterviewMode: "real", CurrentPhase: "planning"}
	prompt := buildChatPrompt(Problem{Title: "Two Sum"}, attempt, nil, "hi", false, ProblemMemory{})
	if strings.Contains(prompt, "Current phase: planning") {
		t.Fatalf("legacy phase must be normalized before interpolation")
	}
	if !strings.Contains(prompt, "Current phase: plan") {
		t.Fatalf("expected normalized phase in prompt")
	}
}

func TestChatPromptDoesNotFixateOnInvariants(t *testing.T) {
	attempt := Attempt{CompanyPreset: "google", InterviewMode: "real", CurrentPhase: "plan"}
	prompt := buildChatPrompt(Problem{Title: "Two Sum"}, attempt, nil, "hi", false, ProblemMemory{})
	if count := strings.Count(strings.ToLower(prompt), "invariant"); count > 1 {
		t.Fatalf("expected chat prompt to not fixate on invariants, got %d occurrences", count)
	}
}

func TestGoogleInstructionsDescribeRealisticInterview(t *testing.T) {
	instructions := companyInterviewInstructions("google")
	if !strings.Contains(instructions, "45-minute") {
		t.Fatalf("expected google instructions to mention 45-minute onsite")
	}
	if !strings.Contains(instructions, "Collaborate") {
		t.Fatalf("expected google instructions to mention Collaborate")
	}
	if strings.Contains(strings.ToLower(instructions), "invariant") {
		t.Fatalf("expected google instructions to not mention invariant")
	}
}

func TestBuildChatPromptFormatsArtifactMessages(t *testing.T) {
	messages := []ChatMessage{
		{Role: "assistant", Content: "疑似コードを書いてください。", Kind: "artifact_request", Payload: &ChatMessagePayload{
			ArtifactRequest: &ArtifactRequest{Kind: "pseudocode", Topic: "plan", Instructions: "全探索を書く"},
		}},
		{Role: "user", Content: "書きました。", Kind: "artifact", Payload: &ChatMessagePayload{
			Artifact: &ArtifactSubmission{Kind: "pseudocode", Topic: "plan", Content: "for x in nums:"},
		}},
	}
	prompt := buildChatPrompt(Problem{Title: "Two Sum"}, Attempt{CurrentPhase: "code"}, messages, "next", false, ProblemMemory{})
	if !strings.Contains(prompt, "[ARTIFACT REQUEST kind=pseudocode topic=plan] 全探索を書く") {
		t.Fatalf("expected formatted artifact request in prompt: %s", prompt)
	}
	if !strings.Contains(prompt, "[ARTIFACT kind=pseudocode topic=plan]\nfor x in nums:") {
		t.Fatalf("expected formatted artifact submission in prompt: %s", prompt)
	}
}
