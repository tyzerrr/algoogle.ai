package server

import "strings"

func normalizeWhiteboardKind(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "pseudocode", "pseudo_code", "pseudo-code":
		return "pseudocode"
	case "mermaid_sequence", "sequence", "sequence_diagram", "mermaid":
		return "mermaid_sequence"
	case "data_structure", "data-structure", "ds":
		return "data_structure"
	case "invariants", "invariant":
		return "invariants"
	case "state_transition", "state-machine", "state_machine":
		return "state_transition"
	case "complexity_table", "complexity":
		return "complexity_table"
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
	case "mermaid_sequence":
		return "sequenceDiagram\n  participant Candidate\n  participant Interviewer\n  Candidate->>Interviewer: Explain the approach\n  Interviewer-->>Candidate: Probe invariants and edge cases"
	case "data_structure":
		return "Data structure:\n- \n\nOperations:\n- lookup:\n- update:\n\nInvariant:\n- "
	case "invariants":
		return "Invariant:\n- \n\nWhy it holds:\n- Initialization:\n- Maintenance:\n- Termination:"
	case "state_transition":
		return "States:\n- \n\nTransitions:\n- \n\nInvalid transitions:\n- "
	case "complexity_table":
		return "| Approach | Time | Space | Trade-off |\n| --- | --- | --- | --- |\n| Brute force |  |  |  |\n| Optimized |  |  |  |"
	default:
		if strings.TrimSpace(topic) == "" {
			topic = "approach"
		}
		return "function solve(input):\n    # " + topic + "\n    clarify constraints\n    outline brute force\n    derive optimized invariant\n    handle edge cases\n    return answer"
	}
}

func fallbackWhiteboardSuggestion(summary string) WhiteboardSuggestion {
	return WhiteboardSuggestion{
		UseWhiteboard:  true,
		Kind:           "pseudocode",
		Topic:          "解法方針の疑似コード",
		Prompt:         "テスト実行に頼らず、全探索から最適化までの流れを疑似コードで整理してください。",
		StarterContent: defaultWhiteboardContent("pseudocode", "解法方針"),
		Reason:         "AI CLIからWhiteBoard判断を取得できなかったため、面接練習として疑似コード整理にフォールバックします。詳細: " + strings.TrimSpace(summary),
	}
}

func normalizeWhiteboardSuggestion(suggestion *WhiteboardSuggestion) {
	if suggestion == nil {
		return
	}
	suggestion.Kind = normalizeWhiteboardKind(suggestion.Kind)
	suggestion.Topic = trimOrDefault(suggestion.Topic, "WhiteBoard discussion")
	suggestion.Prompt = strings.TrimSpace(suggestion.Prompt)
	if suggestion.Prompt == "" {
		suggestion.Prompt = "このWhiteBoardを使って、解法の前提、不変条件、edge caseを面接官に説明してください。"
	}
	suggestion.StarterContent = strings.TrimSpace(suggestion.StarterContent)
	if suggestion.UseWhiteboard && suggestion.StarterContent == "" {
		suggestion.StarterContent = defaultWhiteboardContent(suggestion.Kind, suggestion.Topic)
	}
	suggestion.Reason = strings.TrimSpace(suggestion.Reason)
}
