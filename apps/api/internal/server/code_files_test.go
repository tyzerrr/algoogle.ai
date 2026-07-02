package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestCodeFileManagerSyncsAttemptCode(t *testing.T) {
	dir := t.TempDir()
	manager := NewCodeFileManager(dir, "./workspace")
	attempt := Attempt{
		ID:        "attempt/with spaces",
		ProblemID: "two sum",
		Code:      "class Solution:\n    pass\n",
	}

	created, err := manager.Ensure(attempt)
	if err != nil {
		t.Fatalf("ensure code file: %v", err)
	}
	if created.Content != attempt.Code {
		t.Fatalf("expected starter content, got %q", created.Content)
	}
	if created.Path != "workspace/two-sum.py" {
		t.Fatalf("unexpected public path: %s", created.Path)
	}

	internalPath := filepath.Join(dir, "two-sum.py")
	if err := os.WriteFile(internalPath, []byte("print('from nvim')\n"), 0644); err != nil {
		t.Fatalf("external edit: %v", err)
	}
	read, err := manager.Read(attempt)
	if err != nil {
		t.Fatalf("read code file: %v", err)
	}
	if read.Content != "print('from nvim')\n" {
		t.Fatalf("expected external content, got %q", read.Content)
	}

	written, err := manager.Write(attempt, "print('from app')\n")
	if err != nil {
		t.Fatalf("write code file: %v", err)
	}
	if written.Content != "print('from app')\n" {
		t.Fatalf("expected app content, got %q", written.Content)
	}
}

func TestCodeFileNameIsPerProblemFixed(t *testing.T) {
	dir := t.TempDir()
	manager := NewCodeFileManager(dir, "./workspace")

	created, err := manager.Ensure(Attempt{ID: "1700000000000000000", ProblemID: "two-sum", Code: "x"})
	if err != nil {
		t.Fatalf("ensure: %v", err)
	}
	if created.Path != "workspace/two-sum.py" {
		t.Fatalf("expected fixed per-problem path, got %s", created.Path)
	}

	// A different attempt ID for the same problem resolves to the same file.
	other, err := manager.Read(Attempt{ID: "9999999999999999999", ProblemID: "two-sum"})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if other.Path != "workspace/two-sum.py" {
		t.Fatalf("expected same path for different attempt, got %s", other.Path)
	}

	empty, err := manager.Ensure(Attempt{ID: "abc", ProblemID: ""})
	if err != nil {
		t.Fatalf("ensure empty problem: %v", err)
	}
	if empty.Path != "workspace/problem.py" {
		t.Fatalf("expected problem.py fallback, got %s", empty.Path)
	}
}

func TestEnsurePreservesContentAcrossAttempts(t *testing.T) {
	dir := t.TempDir()
	manager := NewCodeFileManager(dir, "./workspace")

	attempt1 := Attempt{ID: "1", ProblemID: "two-sum", Code: "starter"}
	if _, err := manager.Ensure(attempt1); err != nil {
		t.Fatalf("ensure attempt1: %v", err)
	}
	if _, err := manager.Write(attempt1, "user code"); err != nil {
		t.Fatalf("write user code: %v", err)
	}

	attempt2 := Attempt{ID: "2", ProblemID: "two-sum", Code: "starter"}
	got, err := manager.Ensure(attempt2)
	if err != nil {
		t.Fatalf("ensure attempt2: %v", err)
	}
	if got.Content != "user code" {
		t.Fatalf("expected preserved content, got %q", got.Content)
	}
	if got.Path != "workspace/two-sum.py" {
		t.Fatalf("expected shared path, got %s", got.Path)
	}
}

func TestWriteOverwritesSharedFile(t *testing.T) {
	dir := t.TempDir()
	manager := NewCodeFileManager(dir, "./workspace")

	attempt := Attempt{ID: "1", ProblemID: "two-sum", Code: "starter"}
	if _, err := manager.Ensure(attempt); err != nil {
		t.Fatalf("ensure: %v", err)
	}
	if _, err := manager.Write(attempt, "edited"); err != nil {
		t.Fatalf("write: %v", err)
	}
	reset, err := manager.Write(attempt, "starter")
	if err != nil {
		t.Fatalf("reset write: %v", err)
	}
	if reset.Content != "starter" {
		t.Fatalf("expected overwrite to starter, got %q", reset.Content)
	}
}

func TestCreateAttemptResetSemanticsHandler(t *testing.T) {
	store, err := NewStore("sqlite:///:memory:")
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	seed, err := LoadSeedProblems()
	if err != nil {
		t.Fatalf("load seed: %v", err)
	}
	if err := store.Seed(seed); err != nil {
		t.Fatalf("seed: %v", err)
	}

	app := &App{
		store:     store,
		codeFiles: NewCodeFileManager(t.TempDir(), "./workspace"),
	}
	app.router = app.routes()

	firstAttempt := postAttempt(t, app, map[string]any{"problem_id": "two-sum", "code": "starter"})

	// User edits the shared file.
	putCode(t, app, firstAttempt.ID, "my solution")

	// A new attempt without reset keeps the on-disk user code.
	secondAttempt := postAttempt(t, app, map[string]any{"problem_id": "two-sum", "code": "starter"})
	if got := getCode(t, app, secondAttempt.ID); got != "my solution" {
		t.Fatalf("expected preserved user code without reset, got %q", got)
	}

	// Reset restores the starter code.
	thirdAttempt := postAttempt(t, app, map[string]any{"problem_id": "two-sum", "code": "starter", "reset": true})
	if got := getCode(t, app, thirdAttempt.ID); got != "starter" {
		t.Fatalf("expected reset to starter, got %q", got)
	}
}

func postAttempt(t *testing.T, app *App, body map[string]any) Attempt {
	t.Helper()
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/attempts", bytes.NewReader(payload))
	rec := httptest.NewRecorder()
	app.router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("create attempt status %d: %s", rec.Code, rec.Body.String())
	}
	var attempt Attempt
	if err := json.Unmarshal(rec.Body.Bytes(), &attempt); err != nil {
		t.Fatalf("decode attempt: %v", err)
	}
	return attempt
}

func putCode(t *testing.T, app *App, attemptID, code string) {
	t.Helper()
	payload, _ := json.Marshal(map[string]string{"code": code})
	req := httptest.NewRequest(http.MethodPut, "/attempts/"+attemptID+"/code-file", bytes.NewReader(payload))
	rec := httptest.NewRecorder()
	app.router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("put code status %d: %s", rec.Code, rec.Body.String())
	}
}

func getCode(t *testing.T, app *App, attemptID string) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/attempts/"+attemptID+"/code-file", nil)
	rec := httptest.NewRecorder()
	app.router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get code status %d: %s", rec.Code, rec.Body.String())
	}
	var info CodeFileResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &info); err != nil {
		t.Fatalf("decode code file: %v", err)
	}
	return info.Content
}
