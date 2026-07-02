package server

import (
	"strings"
	"testing"
)

func TestParseInterviewerTurnStructured(t *testing.T) {
	raw := `noise before {"reply":"  制約を確認しましょう。  ","phase":"plan","artifact_request":{"kind":"mermaid","topic":"","instructions":"  データ構造を図示  ","starter_content":""}} trailing`
	turn := parseInterviewerTurn(raw, "clarify")
	if turn.Reply != "制約を確認しましょう。" {
		t.Fatalf("unexpected reply: %q", turn.Reply)
	}
	if turn.Phase != "plan" {
		t.Fatalf("expected advanced phase plan, got %q", turn.Phase)
	}
	if turn.ArtifactRequest == nil {
		t.Fatalf("expected artifact request")
	}
	if turn.ArtifactRequest.Kind != "diagram" {
		t.Fatalf("expected normalized kind diagram, got %q", turn.ArtifactRequest.Kind)
	}
	if strings.TrimSpace(turn.ArtifactRequest.Topic) == "" {
		t.Fatalf("expected topic default, got empty")
	}
	if turn.ArtifactRequest.Instructions != "データ構造を図示" {
		t.Fatalf("unexpected instructions: %q", turn.ArtifactRequest.Instructions)
	}
	if !strings.HasPrefix(turn.ArtifactRequest.StarterContent, "flowchart TD") {
		t.Fatalf("expected diagram starter default, got %q", turn.ArtifactRequest.StarterContent)
	}
}

func TestParseInterviewerTurnFallbackOnMalformedJSON(t *testing.T) {
	raw := "  これはJSONではありません  "
	turn := parseInterviewerTurn(raw, "code")
	if turn.Reply != "これはJSONではありません" {
		t.Fatalf("expected raw trimmed reply, got %q", turn.Reply)
	}
	if turn.Phase != "code" {
		t.Fatalf("expected phase unchanged, got %q", turn.Phase)
	}
	if turn.ArtifactRequest != nil {
		t.Fatalf("expected nil artifact request on fallback")
	}
}

func TestParseInterviewerTurnDropsEmptyInstructionRequest(t *testing.T) {
	raw := `{"reply":"続けてください。","phase":"code","artifact_request":{"kind":"pseudocode","topic":"plan","instructions":"   ","starter_content":"x"}}`
	turn := parseInterviewerTurn(raw, "code")
	if turn.ArtifactRequest != nil {
		t.Fatalf("expected empty-instruction request dropped to nil, got %#v", turn.ArtifactRequest)
	}
	if turn.Reply != "続けてください。" {
		t.Fatalf("unexpected reply: %q", turn.Reply)
	}
}

func TestParseInterviewerTurnEmptyReplyFallsBack(t *testing.T) {
	raw := `{"reply":"   ","phase":"plan","artifact_request":null}`
	turn := parseInterviewerTurn(raw, "clarify")
	// Empty reply triggers fallback: raw as reply, current phase preserved.
	if turn.Reply != raw {
		t.Fatalf("expected raw as reply on empty, got %q", turn.Reply)
	}
	if turn.Phase != "clarify" {
		t.Fatalf("expected current phase on fallback, got %q", turn.Phase)
	}
}
