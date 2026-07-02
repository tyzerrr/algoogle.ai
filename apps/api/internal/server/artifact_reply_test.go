package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestArtifactReplyCreatesWhiteboardMessagesAndReaction(t *testing.T) {
	fake := &fakeInterviewAI{turn: InterviewerTurn{Reply: "不変条件が抜けています。", Phase: "dryrun"}}
	app, store := newChatApp(t, fake)
	attempt, err := store.CreateAttempt(CreateAttemptRequest{ProblemID: "two-sum"})
	if err != nil {
		t.Fatalf("create attempt: %v", err)
	}

	rec := postJSON(t, app, "/attempts/"+attempt.ID+"/artifact-reply", ArtifactReplyRequest{
		Kind:    "diagram",
		Topic:   "stack",
		Content: "flowchart TD\n  A --> B",
		Message: "図を書きました。",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("artifact reply status %d: %s", rec.Code, rec.Body.String())
	}
	var resp ArtifactReplyResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Artifact == nil || resp.Artifact.Kind != "diagram" || resp.Artifact.Content == "" {
		t.Fatalf("expected persisted whiteboard, got %#v", resp.Artifact)
	}
	if resp.UserMessage.Kind != "artifact" || resp.UserMessage.Payload == nil || resp.UserMessage.Payload.Artifact == nil {
		t.Fatalf("expected artifact user message, got %#v", resp.UserMessage)
	}
	if resp.UserMessage.Payload.Artifact.WhiteboardID != resp.Artifact.ID {
		t.Fatalf("expected artifact submission to link whiteboard id")
	}
	if resp.Message.Content != "不変条件が抜けています。" {
		t.Fatalf("unexpected interviewer reaction: %q", resp.Message.Content)
	}
	if resp.CurrentPhase != "dryrun" {
		t.Fatalf("expected phase dryrun, got %q", resp.CurrentPhase)
	}
	// The AI must receive the artifact body as its userMessage.
	if !strings.Contains(fake.lastUserMessage, "[ARTIFACT") {
		t.Fatalf("expected [ARTIFACT marker in AI userMessage, got %q", fake.lastUserMessage)
	}

	// Whiteboard is retrievable and the conversation has both messages.
	boards, err := store.ListWhiteboardsForAttempt(attempt.ID)
	if err != nil || len(boards) != 1 {
		t.Fatalf("expected one whiteboard, got %#v err=%v", boards, err)
	}
	messages, err := store.ListMessages(attempt.ID)
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("expected user + assistant message, got %d", len(messages))
	}
}

func TestArtifactReplyLinksRequestMessage(t *testing.T) {
	fake := &fakeInterviewAI{turn: InterviewerTurn{Reply: "確認しました。", Phase: "code"}}
	app, store := newChatApp(t, fake)
	attempt, err := store.CreateAttempt(CreateAttemptRequest{ProblemID: "two-sum"})
	if err != nil {
		t.Fatalf("create attempt: %v", err)
	}
	// Persist an interviewer artifact request so the reply can resolve its instructions.
	request, err := store.AddStructuredMessage(attempt.ID, "assistant", "疑似コードを書いてください。", "artifact_request", &ChatMessagePayload{
		ArtifactRequest: &ArtifactRequest{Kind: "pseudocode", Topic: "plan", Instructions: "全探索から最適化まで"},
	})
	if err != nil {
		t.Fatalf("add request message: %v", err)
	}

	rec := postJSON(t, app, "/attempts/"+attempt.ID+"/artifact-reply", ArtifactReplyRequest{
		Content:          "function solve(): ...",
		RequestMessageID: request.ID,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("artifact reply status %d: %s", rec.Code, rec.Body.String())
	}
	var resp ArtifactReplyResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	// Kind/topic/prompt derived from the linked request.
	if resp.Artifact.Kind != "pseudocode" || resp.Artifact.Topic != "plan" {
		t.Fatalf("expected kind/topic from linked request, got %#v", resp.Artifact)
	}
	if resp.Artifact.Prompt != "全探索から最適化まで" {
		t.Fatalf("expected instructions used as prompt, got %q", resp.Artifact.Prompt)
	}
	if resp.UserMessage.Payload.Artifact.RequestMessageID != request.ID {
		t.Fatalf("expected linkage to request message id")
	}
}

func TestArtifactReplyRejectsEmptyContent(t *testing.T) {
	fake := &fakeInterviewAI{turn: InterviewerTurn{Reply: "x", Phase: "code"}}
	app, store := newChatApp(t, fake)
	attempt, err := store.CreateAttempt(CreateAttemptRequest{ProblemID: "two-sum"})
	if err != nil {
		t.Fatalf("create attempt: %v", err)
	}
	rec := postJSON(t, app, "/attempts/"+attempt.ID+"/artifact-reply", ArtifactReplyRequest{Content: "   "})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty content, got %d: %s", rec.Code, rec.Body.String())
	}
	if fake.called != 0 {
		t.Fatalf("AI must not be called on validation failure")
	}
}

func TestArtifactReplyFeedsProblemMemory(t *testing.T) {
	fake := &fakeInterviewAI{turn: InterviewerTurn{Reply: "確認。", Phase: "code"}}
	app, store := newChatApp(t, fake)
	attempt, err := store.CreateAttempt(CreateAttemptRequest{ProblemID: "two-sum"})
	if err != nil {
		t.Fatalf("create attempt: %v", err)
	}
	rec := postJSON(t, app, "/attempts/"+attempt.ID+"/artifact-reply", ArtifactReplyRequest{
		Kind:    "notes",
		Topic:   "invariants",
		Content: "## invariants\n- x",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("artifact reply status %d: %s", rec.Code, rec.Body.String())
	}
	memory, err := store.ProblemMemory("two-sum")
	if err != nil {
		t.Fatalf("problem memory: %v", err)
	}
	if len(memory.Whiteboards) != 1 || memory.Whiteboards[0].Kind != "notes" {
		t.Fatalf("expected whiteboard fed into problem memory, got %#v", memory.Whiteboards)
	}
}
