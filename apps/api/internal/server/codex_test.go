package server

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildCodexExecArgsPutsApprovalBeforeExec(t *testing.T) {
	args := buildCodexExecArgs("/work", "gpt-5", "/tmp/out.txt", "/tmp/schema.json")

	approvalIndex := indexOf(args, "--ask-for-approval")
	execIndex := indexOf(args, "exec")
	if approvalIndex != 0 || execIndex <= approvalIndex {
		t.Fatalf("expected approval policy before exec, got %#v", args)
	}
	if args[approvalIndex+1] != "never" {
		t.Fatalf("expected never approval policy, got %#v", args)
	}
	if indexOf(args[execIndex+1:], "--ask-for-approval") >= 0 {
		t.Fatalf("approval policy must not be passed to codex exec subcommand: %#v", args)
	}
	if args[len(args)-1] != "-" {
		t.Fatalf("expected stdin prompt marker at the end, got %#v", args)
	}
}

func TestPrepareCodexHomeFromCopiesOnlyRuntimeFiles(t *testing.T) {
	source := t.TempDir()
	base := filepath.Join(t.TempDir(), "codex-cache")
	for name, content := range map[string]string{
		"auth.json":       `{"token":"test"}`,
		"config.toml":     "model = \"gpt-5\"\n",
		"installation_id": "install-test",
		"history.jsonl":   "should not be copied",
	} {
		if err := os.WriteFile(filepath.Join(source, name), []byte(content), 0600); err != nil {
			t.Fatalf("write source file %s: %v", name, err)
		}
	}

	codexHome, cleanup, err := prepareCodexHomeFrom(source, base)
	if err != nil {
		t.Fatalf("prepare codex home: %v", err)
	}
	defer cleanup()

	for _, name := range []string{"auth.json", "config.toml", "installation_id"} {
		if _, err := os.Stat(filepath.Join(codexHome, name)); err != nil {
			t.Fatalf("expected copied %s: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(codexHome, "history.jsonl")); !os.IsNotExist(err) {
		t.Fatalf("history should not be copied, stat err=%v", err)
	}
	cleanup()
	if _, err := os.Stat(codexHome); !os.IsNotExist(err) {
		t.Fatalf("cleanup should remove codex home, stat err=%v", err)
	}
}

func indexOf(values []string, needle string) int {
	for i, value := range values {
		if value == needle {
			return i
		}
	}
	return -1
}
