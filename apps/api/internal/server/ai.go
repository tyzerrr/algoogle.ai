package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type InterviewAI interface {
	Chat(ctx context.Context, problem Problem, attempt Attempt, messages []ChatMessage, userMessage string, englishMode bool, memory ProblemMemory) (InterviewerTurn, error)
	Review(ctx context.Context, problem Problem, attempt Attempt, code string, memory ProblemMemory) (ReviewResponse, error)
}

type AIService struct {
	codex  *CodexClient
	claude *ClaudeClient
}

func NewAIService(cfg Config) *AIService {
	return &AIService{
		codex:  NewCodexClient(cfg),
		claude: NewClaudeClient(cfg),
	}
}

func (s *AIService) Chat(ctx context.Context, problem Problem, attempt Attempt, messages []ChatMessage, userMessage string, englishMode bool, memory ProblemMemory) (InterviewerTurn, error) {
	return s.client(attempt.AIProvider).Chat(ctx, problem, attempt, messages, userMessage, englishMode, memory)
}

func (s *AIService) Review(ctx context.Context, problem Problem, attempt Attempt, code string, memory ProblemMemory) (ReviewResponse, error) {
	return s.client(attempt.AIProvider).Review(ctx, problem, attempt, code, memory)
}

func (s *AIService) client(provider string) InterviewAI {
	if normalizeAIProvider(provider) == "claude" {
		return s.claude
	}
	return s.codex
}

func normalizeAIProvider(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "claude", "claude-code", "claudecode":
		return "claude"
	default:
		return "codex"
	}
}

func aiProviderLabel(provider string) string {
	if normalizeAIProvider(provider) == "claude" {
		return "Claude Code"
	}
	return "Codex"
}

type ClaudeClient struct {
	path           string
	model          string
	workingDir     string
	timeoutSeconds int
}

func NewClaudeClient(cfg Config) *ClaudeClient {
	return &ClaudeClient{
		path:           cfg.ClaudeCLIPath,
		model:          cfg.ClaudeModel,
		workingDir:     cfg.ClaudeWorkingDir,
		timeoutSeconds: cfg.ClaudeTimeoutSeconds,
	}
}

// claudeConfigFiles are the only files copied into an isolated Claude config
// directory. Everything else (history, logs) is intentionally left behind.
var claudeConfigFiles = []string{".credentials.json", "settings.json", ".claude.json"}

func (c *ClaudeClient) Generate(ctx context.Context, prompt string, schema *string) (string, error) {
	if strings.TrimSpace(c.path) == "" {
		return "", errors.New("Claude Code CLIが未設定です。CLAUDE_CLI_PATHを設定してください。")
	}

	runCtx, cancel := context.WithTimeout(ctx, time.Duration(c.timeoutSeconds)*time.Second)
	defer cancel()

	args := buildClaudeArgs(c.model, schema)

	cmd := exec.CommandContext(runCtx, c.path, args...)
	cmd.Stdin = strings.NewReader(prompt)

	// Best-effort: run against an isolated CLAUDE_CONFIG_DIR so concurrent runs
	// do not share session state. A preparation failure must not block the run.
	claudeHome, cleanupClaudeHome, prepErr := prepareClaudeHome()
	if prepErr == nil {
		defer cleanupClaudeHome()
		cmd.Env = claudeEnv(os.Environ(), claudeHome)
	} else {
		cmd.Env = append(os.Environ(), "NO_COLOR=1")
	}
	if strings.TrimSpace(c.workingDir) != "" {
		cmd.Dir = c.workingDir
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
			return "", fmt.Errorf("Claude Code CLIの応答がタイムアウトしました。CLAUDE_CLI_TIMEOUT_SECONDSを長めにしてください")
		}
		if errors.Is(err, exec.ErrNotFound) {
			return "", fmt.Errorf("Claude Code CLIが見つかりません。CLAUDE_CLI_PATH=%q を確認してください", c.path)
		}
		detail := strings.TrimSpace(stderr.String())
		if detail == "" {
			detail = strings.TrimSpace(stdout.String())
		}
		if detail == "" {
			detail = err.Error()
		}
		return "", errors.New(claudeFailureMessage(detail))
	}
	return strings.TrimSpace(stdout.String()), nil
}

func buildClaudeArgs(model string, schema *string) []string {
	args := []string{
		"-p",
		"--output-format", "text",
		"--no-session-persistence",
		"--permission-mode", "dontAsk",
		"--tools", "",
	}
	if model != "" {
		args = append(args, "--model", model)
	}
	if schema != nil {
		args = append(args, "--json-schema", *schema)
	}
	return args
}

// claudeEnv appends NO_COLOR and CLAUDE_CONFIG_DIR. os/exec keeps the LAST value
// for duplicate keys, so appending overrides any pre-existing CLAUDE_CONFIG_DIR.
func claudeEnv(base []string, claudeHome string) []string {
	env := append([]string{}, base...)
	env = append(env, "NO_COLOR=1", "CLAUDE_CONFIG_DIR="+claudeHome)
	return env
}

func claudeFailureMessage(detail string) string {
	lower := strings.ToLower(detail)
	for _, marker := range []string{"not logged in", "/login", "invalid api key", "authentication", "oauth"} {
		if strings.Contains(lower, marker) {
			return "Claude Code CLIが認証されていません。ホストで `claude setup-token` を実行し、表示されたトークンを .env に `CLAUDE_CODE_OAUTH_TOKEN=...` として設定してから `make up` し直してください。\n\n詳細: " + detail
		}
	}
	return fmt.Sprintf("Claude Code CLIの実行に失敗しました: %s", detail)
}

func prepareClaudeHome() (string, func(), error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", func() {}, err
	}
	source := strings.TrimSpace(os.Getenv("CLAUDE_CONFIG_DIR"))
	if source == "" {
		source = filepath.Join(home, ".claude")
	}
	homeClaudeJSON := filepath.Join(home, ".claude.json")
	base := filepath.Join(home, ".cache", "algoogle-claude")
	return prepareClaudeHomeFrom(source, homeClaudeJSON, base)
}

func prepareClaudeHomeFrom(sourceDir, homeClaudeJSON, base string) (string, func(), error) {
	if err := os.MkdirAll(base, 0700); err != nil {
		return "", func() {}, err
	}
	claudeHome, err := os.MkdirTemp(base, "home-*")
	if err != nil {
		return "", func() {}, err
	}
	cleanup := func() {
		_ = os.RemoveAll(claudeHome)
	}
	for _, name := range claudeConfigFiles {
		if err := copyClaudeConfigFile(filepath.Join(sourceDir, name), filepath.Join(claudeHome, name)); err != nil {
			cleanup()
			return "", func() {}, err
		}
	}
	if _, err := os.Stat(filepath.Join(claudeHome, ".claude.json")); os.IsNotExist(err) {
		if err := copyClaudeConfigFile(homeClaudeJSON, filepath.Join(claudeHome, ".claude.json")); err != nil {
			cleanup()
			return "", func() {}, err
		}
	}
	return claudeHome, cleanup, nil
}

func copyClaudeConfigFile(sourcePath, targetPath string) error {
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	return os.WriteFile(targetPath, data, 0600)
}

func (c *ClaudeClient) Chat(ctx context.Context, problem Problem, attempt Attempt, messages []ChatMessage, userMessage string, englishMode bool, memory ProblemMemory) (InterviewerTurn, error) {
	schema := interviewerTurnJSONSchema()
	raw, err := c.Generate(ctx, buildChatPrompt(problem, attempt, messages, userMessage, englishMode, memory), &schema)
	if err != nil {
		return InterviewerTurn{}, err
	}
	return parseInterviewerTurn(raw, attempt.CurrentPhase), nil
}

func (c *ClaudeClient) Review(ctx context.Context, problem Problem, attempt Attempt, code string, memory ProblemMemory) (ReviewResponse, error) {
	schema := reviewJSONSchema()
	raw, err := c.Generate(ctx, buildReviewPrompt(problem, attempt, code, memory), &schema)
	if err != nil {
		return fallbackReview(err.Error()), err
	}
	var review ReviewResponse
	if err := json.Unmarshal([]byte(extractJSONObject(raw)), &review); err != nil {
		return fallbackReview("AIレビューのJSONを解析できませんでした: " + raw), err
	}
	return review, nil
}
