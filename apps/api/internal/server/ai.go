package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type InterviewAI interface {
	Chat(ctx context.Context, problem Problem, attempt Attempt, messages []ChatMessage, userMessage string, englishMode bool, memory ProblemMemory) (string, error)
	Review(ctx context.Context, problem Problem, attempt Attempt, code string, memory ProblemMemory) (ReviewResponse, error)
	SuggestWhiteboard(ctx context.Context, problem Problem, attempt Attempt, messages []ChatMessage, memory ProblemMemory, userMessage string) (WhiteboardSuggestion, error)
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

func (s *AIService) Chat(ctx context.Context, problem Problem, attempt Attempt, messages []ChatMessage, userMessage string, englishMode bool, memory ProblemMemory) (string, error) {
	return s.client(attempt.AIProvider).Chat(ctx, problem, attempt, messages, userMessage, englishMode, memory)
}

func (s *AIService) Review(ctx context.Context, problem Problem, attempt Attempt, code string, memory ProblemMemory) (ReviewResponse, error) {
	return s.client(attempt.AIProvider).Review(ctx, problem, attempt, code, memory)
}

func (s *AIService) SuggestWhiteboard(ctx context.Context, problem Problem, attempt Attempt, messages []ChatMessage, memory ProblemMemory, userMessage string) (WhiteboardSuggestion, error) {
	return s.client(attempt.AIProvider).SuggestWhiteboard(ctx, problem, attempt, messages, memory, userMessage)
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

func (c *ClaudeClient) Generate(ctx context.Context, prompt string, schema *string) (string, error) {
	if strings.TrimSpace(c.path) == "" {
		return "", errors.New("Claude Code CLIが未設定です。CLAUDE_CLI_PATHを設定してください。")
	}

	runCtx, cancel := context.WithTimeout(ctx, time.Duration(c.timeoutSeconds)*time.Second)
	defer cancel()

	args := []string{
		"-p",
		"--output-format", "text",
		"--no-session-persistence",
		"--permission-mode", "dontAsk",
		"--tools", "",
	}
	if c.model != "" {
		args = append(args, "--model", c.model)
	}
	if schema != nil {
		args = append(args, "--json-schema", *schema)
	}

	cmd := exec.CommandContext(runCtx, c.path, args...)
	cmd.Stdin = strings.NewReader(prompt)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "NO_COLOR=1")
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
		return "", fmt.Errorf("Claude Code CLIの実行に失敗しました: %s", detail)
	}
	return strings.TrimSpace(stdout.String()), nil
}

func (c *ClaudeClient) Chat(ctx context.Context, problem Problem, attempt Attempt, messages []ChatMessage, userMessage string, englishMode bool, memory ProblemMemory) (string, error) {
	prompt := buildChatPrompt(problem, attempt, messages, userMessage, englishMode, memory)
	return c.Generate(ctx, prompt, nil)
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

func (c *ClaudeClient) SuggestWhiteboard(ctx context.Context, problem Problem, attempt Attempt, messages []ChatMessage, memory ProblemMemory, userMessage string) (WhiteboardSuggestion, error) {
	schema := whiteboardSuggestionJSONSchema()
	raw, err := c.Generate(ctx, buildWhiteboardSuggestionPrompt(problem, attempt, messages, memory, userMessage), &schema)
	if err != nil {
		return fallbackWhiteboardSuggestion(err.Error()), err
	}
	var suggestion WhiteboardSuggestion
	if err := json.Unmarshal([]byte(extractJSONObject(raw)), &suggestion); err != nil {
		return fallbackWhiteboardSuggestion("WhiteBoard提案のJSONを解析できませんでした: " + raw), err
	}
	normalizeWhiteboardSuggestion(&suggestion)
	return suggestion, nil
}
