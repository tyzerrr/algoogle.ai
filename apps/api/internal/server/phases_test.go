package server

import "testing"

func TestNormalizePhaseAliases(t *testing.T) {
	cases := map[string]string{
		"clarify":        "clarify",
		"clarification":  "clarify",
		"understanding":  "clarify",
		"confirm":        "clarify",
		"plan":           "plan",
		"planning":       "plan",
		"code":           "code",
		"coding":         "code",
		"implement":      "code",
		"implementation": "code",
		"dryrun":         "dryrun",
		"dry_run":        "dryrun",
		"dry-run":        "dryrun",
		"verify":         "dryrun",
		"verification":   "dryrun",
		"test":           "dryrun",
		"testing":        "dryrun",
		"followup":       "followup",
		"follow_up":      "followup",
		"follow-up":      "followup",
		"review":         "followup",
		"deep_dive":      "followup",
		"  PLAN  ":       "plan",
		"nonsense":       "",
		"":               "",
	}
	for input, want := range cases {
		if got := normalizePhase(input); got != want {
			t.Fatalf("normalizePhase(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestAdvancePhaseForwardOnly(t *testing.T) {
	// Forward advancement.
	if got := advancePhase("clarify", "plan"); got != "plan" {
		t.Fatalf("expected forward to plan, got %q", got)
	}
	// Backward proposals are ignored (stay on current).
	if got := advancePhase("code", "clarify"); got != "code" {
		t.Fatalf("expected stay on code, got %q", got)
	}
	// Same phase stays.
	if got := advancePhase("plan", "plan"); got != "plan" {
		t.Fatalf("expected stay on plan, got %q", got)
	}
	// Unknown proposed falls back to normalized current.
	if got := advancePhase("plan", "garbage"); got != "plan" {
		t.Fatalf("expected fallback to plan, got %q", got)
	}
	// Unknown current and unknown proposed fall back to clarify.
	if got := advancePhase("garbage", "garbage"); got != "clarify" {
		t.Fatalf("expected fallback to clarify, got %q", got)
	}
	// Legacy planning alias normalizes on both sides.
	if got := advancePhase("planning", "code"); got != "code" {
		t.Fatalf("expected planning->code, got %q", got)
	}
	if idx := phaseIndex("unknown"); idx != -1 {
		t.Fatalf("expected -1 for unknown phase index, got %d", idx)
	}
}
