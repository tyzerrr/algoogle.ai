package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeInterviewAI struct {
	turn            InterviewerTurn
	err             error
	lastUserMessage string
	called          int
}

func (f *fakeInterviewAI) Chat(_ context.Context, _ Problem, _ Attempt, _ []ChatMessage, userMessage string, _ bool, _ ProblemMemory) (InterviewerTurn, error) {
	f.called++
	f.lastUserMessage = userMessage
	if f.err != nil {
		return InterviewerTurn{}, f.err
	}
	return f.turn, nil
}

func (f *fakeInterviewAI) Review(_ context.Context, _ Problem, _ Attempt, _ string, _ ProblemMemory) (ReviewResponse, error) {
	return ReviewResponse{}, nil
}

func newChatApp(t *testing.T, ai InterviewAI) (*App, *Store) {
	t.Helper()
	store := newSeededStore(t)
	app := &App{store: store, ai: ai, codeFiles: NewCodeFileManager(t.TempDir(), "./workspace")}
	app.router = app.routes()
	return app, store
}

func postJSON(t *testing.T, app *App, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(payload))
	rec := httptest.NewRecorder()
	app.router.ServeHTTP(rec, req)
	return rec
}

func TestChatPersistsStructuredTurnAndPhase(t *testing.T) {
	fake := &fakeInterviewAI{turn: InterviewerTurn{
		Reply: "制約を確認しましょう。",
		Phase: "plan",
		ArtifactRequest: &ArtifactRequest{
			Kind:         "pseudocode",
			Topic:        "plan",
			Instructions: "全探索を疑似コードで",
		},
	}}
	app, store := newChatApp(t, fake)
	attempt, err := store.CreateAttempt(CreateAttemptRequest{ProblemID: "two-sum"})
	if err != nil {
		t.Fatalf("create attempt: %v", err)
	}

	rec := postJSON(t, app, "/attempts/"+attempt.ID+"/chat", ChatRequest{Message: "解きます"})
	if rec.Code != http.StatusOK {
		t.Fatalf("chat status %d: %s", rec.Code, rec.Body.String())
	}
	var resp ChatResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.CurrentPhase != "plan" {
		t.Fatalf("expected phase plan, got %q", resp.CurrentPhase)
	}
	if resp.Message.Kind != "artifact_request" || resp.Message.Payload == nil || resp.Message.Payload.ArtifactRequest == nil {
		t.Fatalf("expected artifact_request message, got %#v", resp.Message)
	}
	if resp.Message.Payload.ArtifactRequest.Instructions != "全探索を疑似コードで" {
		t.Fatalf("unexpected instructions: %#v", resp.Message.Payload.ArtifactRequest)
	}

	reloaded, err := store.GetAttempt(attempt.ID)
	if err != nil {
		t.Fatalf("get attempt: %v", err)
	}
	if reloaded.CurrentPhase != "plan" {
		t.Fatalf("expected persisted phase plan, got %q", reloaded.CurrentPhase)
	}
}

func TestChatIgnoresBackwardPhase(t *testing.T) {
	fake := &fakeInterviewAI{turn: InterviewerTurn{Reply: "戻りません。", Phase: "clarify"}}
	app, store := newChatApp(t, fake)
	attempt, err := store.CreateAttempt(CreateAttemptRequest{ProblemID: "two-sum"})
	if err != nil {
		t.Fatalf("create attempt: %v", err)
	}
	if _, err := store.UpdateAttemptPhase(attempt.ID, "code"); err != nil {
		t.Fatalf("update phase: %v", err)
	}

	rec := postJSON(t, app, "/attempts/"+attempt.ID+"/chat", ChatRequest{Message: "続き"})
	if rec.Code != http.StatusOK {
		t.Fatalf("chat status %d: %s", rec.Code, rec.Body.String())
	}
	var resp ChatResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.CurrentPhase != "code" {
		t.Fatalf("expected phase to stay code, got %q", resp.CurrentPhase)
	}
}

func TestChatFallsBackWhenAIFails(t *testing.T) {
	fake := &fakeInterviewAI{err: errors.New("cli exploded")}
	app, store := newChatApp(t, fake)
	attempt, err := store.CreateAttempt(CreateAttemptRequest{ProblemID: "two-sum"})
	if err != nil {
		t.Fatalf("create attempt: %v", err)
	}

	rec := postJSON(t, app, "/attempts/"+attempt.ID+"/chat", ChatRequest{Message: "解きます"})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 fallback, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp ChatResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Message.Kind != "text" {
		t.Fatalf("expected text fallback message, got %q", resp.Message.Kind)
	}
	if !strings.Contains(resp.Message.Content, "詳細: cli exploded") {
		t.Fatalf("expected CLI failure copy, got %q", resp.Message.Content)
	}
	if resp.CurrentPhase != "clarify" {
		t.Fatalf("expected phase unchanged clarify, got %q", resp.CurrentPhase)
	}
}
