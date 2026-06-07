package server

import (
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
	if created.Path != "workspace/two-sum-attempt-with-spaces.py" {
		t.Fatalf("unexpected public path: %s", created.Path)
	}

	internalPath := filepath.Join(dir, "two-sum-attempt-with-spaces.py")
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
