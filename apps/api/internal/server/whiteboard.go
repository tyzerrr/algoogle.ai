package server

import "strings"

func normalizeWhiteboardKind(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "pseudocode", "pseudo_code", "pseudo-code":
		return "pseudocode"
	case "diagram", "mermaid", "mermaid_sequence", "sequence", "sequence_diagram",
		"data_structure", "data-structure", "ds", "state_transition", "state-machine", "state_machine":
		return "diagram"
	case "notes", "note", "markdown", "invariants", "invariant", "complexity_table", "complexity":
		return "notes"
	default:
		return "pseudocode"
	}
}

func trimOrDefault(value, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

func defaultWhiteboardContent(kind, topic string) string {
	switch normalizeWhiteboardKind(kind) {
	case "diagram":
		return "flowchart TD\n  A[入力] --> B{判定}\n  B -->|yes| C[更新]\n  B -->|no| D[次へ]"
	case "notes":
		if strings.TrimSpace(topic) == "" {
			topic = "approach"
		}
		return "## " + topic + "\n\n- 不変条件: \n- エッジケース: \n\n| Approach | Time | Space |\n| --- | --- | --- |\n| Brute force |  |  |\n| Optimized |  |  |"
	default:
		if strings.TrimSpace(topic) == "" {
			topic = "approach"
		}
		return "function solve(input):\n    # " + topic + "\n    clarify constraints\n    outline brute force\n    derive optimized invariant\n    handle edge cases\n    return answer"
	}
}
