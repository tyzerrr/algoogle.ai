package server

import (
	"strings"
	"testing"
)

func TestNormalizeWhiteboardKindConsolidatesLegacyKinds(t *testing.T) {
	cases := map[string]string{
		// Legacy 6 kinds collapse into 3 canonical kinds.
		"pseudocode":       "pseudocode",
		"mermaid_sequence": "diagram",
		"data_structure":   "diagram",
		"invariants":       "notes",
		"state_transition": "diagram",
		"complexity_table": "notes",
		// Aliases.
		"pseudo_code":   "pseudocode",
		"pseudo-code":   "pseudocode",
		"mermaid":       "diagram",
		"sequence":      "diagram",
		"ds":            "diagram",
		"state-machine": "diagram",
		"note":          "notes",
		"markdown":      "notes",
		"invariant":     "notes",
		"complexity":    "notes",
		// Default.
		"":        "pseudocode",
		"unknown": "pseudocode",
	}
	for input, want := range cases {
		if got := normalizeWhiteboardKind(input); got != want {
			t.Fatalf("normalizeWhiteboardKind(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestDefaultWhiteboardContentPerKind(t *testing.T) {
	diagram := defaultWhiteboardContent("diagram", "anything")
	if !strings.HasPrefix(diagram, "flowchart TD") {
		t.Fatalf("expected diagram flowchart scaffold, got %q", diagram)
	}

	notes := defaultWhiteboardContent("notes", "Two Sum")
	if !strings.Contains(notes, "## Two Sum") || !strings.Contains(notes, "不変条件") || !strings.Contains(notes, "| Approach | Time | Space |") {
		t.Fatalf("expected notes scaffold with topic and table, got %q", notes)
	}

	pseudocode := defaultWhiteboardContent("pseudocode", "solve it")
	if !strings.Contains(pseudocode, "function solve(input):") || !strings.Contains(pseudocode, "solve it") {
		t.Fatalf("expected pseudocode scaffold, got %q", pseudocode)
	}
}
