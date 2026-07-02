package server

import "strings"

var interviewPhases = []string{"clarify", "plan", "code", "dryrun", "followup"}

func normalizePhase(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "clarify", "clarification", "understanding", "confirm":
		return "clarify"
	case "plan", "planning":
		return "plan"
	case "code", "coding", "implement", "implementation":
		return "code"
	case "dryrun", "dry_run", "dry-run", "verify", "verification", "test", "testing":
		return "dryrun"
	case "followup", "follow_up", "follow-up", "review", "deep_dive":
		return "followup"
	default:
		return ""
	}
}

func phaseIndex(phase string) int {
	for i, p := range interviewPhases {
		if p == phase {
			return i
		}
	}
	return -1
}

// advancePhase enforces forward-only movement: the interview never rewinds to an
// earlier phase, and an unrecognized proposal simply keeps the current phase.
func advancePhase(current, proposed string) string {
	normCurrent := normalizePhase(current)
	if normCurrent == "" {
		normCurrent = "clarify"
	}
	normProposed := normalizePhase(proposed)
	if normProposed == "" {
		return normCurrent
	}
	if phaseIndex(normProposed) > phaseIndex(normCurrent) {
		return normProposed
	}
	return normCurrent
}
