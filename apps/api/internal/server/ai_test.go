package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrepareClaudeHomeFromCopiesOnlyAuthFiles(t *testing.T) {
	source := t.TempDir()
	base := filepath.Join(t.TempDir(), "claude-cache")
	homeClaudeJSON := filepath.Join(t.TempDir(), ".claude.json")
	for name, content := range map[string]string{
		".credentials.json": `{"token":"test"}`,
		"settings.json":     `{"model":"claude"}`,
		".claude.json":      `{"projects":{}}`,
		"history.jsonl":     "should not be copied",
	} {
		if err := os.WriteFile(filepath.Join(source, name), []byte(content), 0600); err != nil {
			t.Fatalf("write source file %s: %v", name, err)
		}
	}

	claudeHome, cleanup, err := prepareClaudeHomeFrom(source, homeClaudeJSON, base)
	if err != nil {
		t.Fatalf("prepare claude home: %v", err)
	}
	defer cleanup()

	for _, name := range []string{".credentials.json", "settings.json", ".claude.json"} {
		if _, err := os.Stat(filepath.Join(claudeHome, name)); err != nil {
			t.Fatalf("expected copied %s: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(claudeHome, "history.jsonl")); !os.IsNotExist(err) {
		t.Fatalf("history should not be copied, stat err=%v", err)
	}
	cleanup()
	if _, err := os.Stat(claudeHome); !os.IsNotExist(err) {
		t.Fatalf("cleanup should remove claude home, stat err=%v", err)
	}
}

func TestPrepareClaudeHomeFromMissingSourceIsUsable(t *testing.T) {
	source := filepath.Join(t.TempDir(), "does-not-exist")
	base := filepath.Join(t.TempDir(), "claude-cache")
	homeClaudeJSON := filepath.Join(t.TempDir(), "missing-.claude.json")

	claudeHome, cleanup, err := prepareClaudeHomeFrom(source, homeClaudeJSON, base)
	if err != nil {
		t.Fatalf("expected no error for missing source, got %v", err)
	}
	defer cleanup()

	info, err := os.Stat(claudeHome)
	if err != nil || !info.IsDir() {
		t.Fatalf("expected writable dir, stat err=%v", err)
	}
	if err := os.WriteFile(filepath.Join(claudeHome, "probe"), []byte("x"), 0600); err != nil {
		t.Fatalf("expected writable claude home: %v", err)
	}
}

func TestPrepareClaudeHomeFromFallsBackToHomeClaudeJSON(t *testing.T) {
	source := t.TempDir()
	base := filepath.Join(t.TempDir(), "claude-cache")
	homeClaudeJSON := filepath.Join(t.TempDir(), ".claude.json")
	if err := os.WriteFile(filepath.Join(source, ".credentials.json"), []byte(`{"token":"t"}`), 0600); err != nil {
		t.Fatalf("write credentials: %v", err)
	}
	if err := os.WriteFile(homeClaudeJSON, []byte(`{"from":"home"}`), 0600); err != nil {
		t.Fatalf("write home claude json: %v", err)
	}

	claudeHome, cleanup, err := prepareClaudeHomeFrom(source, homeClaudeJSON, base)
	if err != nil {
		t.Fatalf("prepare claude home: %v", err)
	}
	defer cleanup()

	data, err := os.ReadFile(filepath.Join(claudeHome, ".claude.json"))
	if err != nil {
		t.Fatalf("expected fallback .claude.json: %v", err)
	}
	if string(data) != `{"from":"home"}` {
		t.Fatalf("unexpected .claude.json content: %s", data)
	}
}

func TestBuildClaudeArgs(t *testing.T) {
	base := buildClaudeArgs("", nil)
	want := []string{"-p", "--output-format", "text", "--no-session-persistence", "--permission-mode", "dontAsk", "--tools", ""}
	if len(base) != len(want) {
		t.Fatalf("unexpected base args: %#v", base)
	}
	for i := range want {
		if base[i] != want[i] {
			t.Fatalf("arg %d: got %q want %q (%#v)", i, base[i], want[i], base)
		}
	}

	schema := `{"type":"object"}`
	withExtras := buildClaudeArgs("claude-sonnet", &schema)
	if indexOf(withExtras, "--model") < 0 || withExtras[indexOf(withExtras, "--model")+1] != "claude-sonnet" {
		t.Fatalf("expected --model claude-sonnet, got %#v", withExtras)
	}
	if indexOf(withExtras, "--json-schema") < 0 || withExtras[indexOf(withExtras, "--json-schema")+1] != schema {
		t.Fatalf("expected --json-schema, got %#v", withExtras)
	}
}

func TestClaudeEnvAppendsConfigDirLast(t *testing.T) {
	env := claudeEnv([]string{"CLAUDE_CONFIG_DIR=/old", "PATH=/bin"}, "/tmp/x")
	if env[len(env)-1] != "CLAUDE_CONFIG_DIR=/tmp/x" {
		t.Fatalf("expected CLAUDE_CONFIG_DIR override last, got %#v", env)
	}
	foundNoColor := false
	for _, kv := range env {
		if kv == "NO_COLOR=1" {
			foundNoColor = true
		}
	}
	if !foundNoColor {
		t.Fatalf("expected NO_COLOR=1 in env: %#v", env)
	}
}

func TestClaudeFailureMessage(t *testing.T) {
	auth := claudeFailureMessage("Not logged in · Please run /login")
	if !strings.Contains(auth, "claude setup-token") || !strings.Contains(auth, "CLAUDE_CODE_OAUTH_TOKEN") {
		t.Fatalf("expected auth guidance, got %q", auth)
	}

	generic := claudeFailureMessage("boom")
	if generic != "Claude Code CLIの実行に失敗しました: boom" {
		t.Fatalf("unexpected generic message: %q", generic)
	}
}
