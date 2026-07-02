package server

import (
	"encoding/json"
	"strings"
)

// whiteboardKindLabel gives a human-readable topic fallback per canonical kind.
func whiteboardKindLabel(kind string) string {
	switch normalizeWhiteboardKind(kind) {
	case "diagram":
		return "図"
	case "notes":
		return "メモ"
	default:
		return "疑似コード"
	}
}

// parseInterviewerTurn converts a raw AI response into a structured turn. A
// malformed payload or an empty reply must never break the interview, so it
// degrades to a plain-text turn that keeps the current phase.
func parseInterviewerTurn(raw, currentPhase string) InterviewerTurn {
	fallbackPhase := normalizePhase(currentPhase)
	if fallbackPhase == "" {
		fallbackPhase = currentPhase
	}
	fallback := InterviewerTurn{Reply: strings.TrimSpace(raw), Phase: fallbackPhase, ArtifactRequest: nil}

	var turn InterviewerTurn
	if err := json.Unmarshal([]byte(extractJSONObject(raw)), &turn); err != nil {
		return fallback
	}
	turn.Reply = strings.TrimSpace(turn.Reply)
	if turn.Reply == "" {
		return fallback
	}
	turn.Phase = advancePhase(currentPhase, turn.Phase)
	if turn.ArtifactRequest != nil {
		req := turn.ArtifactRequest
		req.Kind = normalizeWhiteboardKind(req.Kind)
		req.Topic = trimOrDefault(req.Topic, whiteboardKindLabel(req.Kind))
		req.Instructions = strings.TrimSpace(req.Instructions)
		if req.Instructions == "" {
			// An artifact request without concrete instructions is not actionable.
			turn.ArtifactRequest = nil
		} else {
			req.StarterContent = trimOrDefault(req.StarterContent, defaultWhiteboardContent(req.Kind, req.Topic))
		}
	}
	return turn
}
