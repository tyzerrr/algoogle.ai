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

type CodexClient struct {
	path           string
	model          string
	workingDir     string
	timeoutSeconds int
}

func NewCodexClient(cfg Config) *CodexClient {
	return &CodexClient{
		path:           cfg.CodexCLIPath,
		model:          cfg.CodexModel,
		workingDir:     cfg.CodexWorkingDir,
		timeoutSeconds: cfg.CodexTimeoutSeconds,
	}
}

func (c *CodexClient) Generate(ctx context.Context, prompt string, schema *string) (string, error) {
	if strings.TrimSpace(c.path) == "" {
		return "", errors.New("Codex CLIが未設定です。CODEX_CLI_PATHを設定してください。")
	}

	runCtx, cancel := context.WithTimeout(ctx, time.Duration(c.timeoutSeconds)*time.Second)
	defer cancel()

	dir, err := os.MkdirTemp("", "algosensei-codex-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)

	outputPath := filepath.Join(dir, "last-message.txt")
	args := []string{
		"exec",
		"--skip-git-repo-check",
		"--ephemeral",
		"--ask-for-approval", "never",
		"--sandbox", "read-only",
		"--color", "never",
		"-C", c.workingDir,
		"--output-last-message", outputPath,
	}
	if c.model != "" {
		args = append(args, "--model", c.model)
	}
	if schema != nil {
		schemaPath := filepath.Join(dir, "schema.json")
		if err := os.WriteFile(schemaPath, []byte(*schema), 0600); err != nil {
			return "", err
		}
		args = append(args, "--output-schema", schemaPath)
	}
	args = append(args, "-")

	cmd := exec.CommandContext(runCtx, c.path, args...)
	cmd.Stdin = strings.NewReader(prompt)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "NO_COLOR=1")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
			return "", fmt.Errorf("Codex CLIの応答がタイムアウトしました。CODEX_CLI_TIMEOUT_SECONDSを長めにしてください")
		}
		if errors.Is(err, exec.ErrNotFound) {
			return "", fmt.Errorf("Codex CLIが見つかりません。CODEX_CLI_PATH=%q を確認してください", c.path)
		}
		detail := strings.TrimSpace(stderr.String())
		if detail == "" {
			detail = strings.TrimSpace(stdout.String())
		}
		if detail == "" {
			detail = err.Error()
		}
		return "", fmt.Errorf("Codex CLIの実行に失敗しました: %s", detail)
	}

	output, err := os.ReadFile(outputPath)
	if err == nil && strings.TrimSpace(string(output)) != "" {
		return strings.TrimSpace(string(output)), nil
	}
	return strings.TrimSpace(stdout.String()), nil
}

func (c *CodexClient) Chat(ctx context.Context, problem Problem, attempt Attempt, messages []ChatMessage, userMessage string, englishMode bool) (string, error) {
	prompt := buildChatPrompt(problem, attempt, messages, userMessage, englishMode)
	return c.Generate(ctx, prompt, nil)
}

func (c *CodexClient) Review(ctx context.Context, problem Problem, attempt Attempt, code string) (ReviewResponse, error) {
	schema := reviewJSONSchema()
	raw, err := c.Generate(ctx, buildReviewPrompt(problem, attempt, code), &schema)
	if err != nil {
		return fallbackReview(err.Error()), err
	}
	var review ReviewResponse
	if err := json.Unmarshal([]byte(extractJSONObject(raw)), &review); err != nil {
		return fallbackReview("AIレビューのJSONを解析できませんでした: " + raw), err
	}
	return review, nil
}

func fallbackReview(summary string) ReviewResponse {
	return ReviewResponse{
		IsCorrect: false,
		Summary:   summary,
		Bugs: []string{
			"AIレビューの生成に失敗しました。Codex CLIのログイン状態とCODEX_CLI_PATHを確認してください。",
		},
		EdgeCases:                []string{},
		Complexity:               ReviewComplexity{Time: "不明", Space: "不明"},
		ReadabilityFeedback:      []string{},
		InterviewFeedback:        []string{"テスト結果をもとに、自分で計算量と不変条件を説明してみてください。"},
		MistakesToRemember:       []string{"AIレビューが失敗した場合でも、失敗したテストケースを1つずつ手で追う。"},
		NextReviewRecommendation: "Codex CLIの設定を直したあと、もう一度レビューを実行してください。",
	}
}

func extractJSONObject(raw string) string {
	raw = strings.TrimSpace(raw)
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start >= 0 && end > start {
		return raw[start : end+1]
	}
	return raw
}
