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

func (c *CodexClient) Chat(ctx context.Context, problem Problem, attempt Attempt, messages []ChatMessage, userMessage string, englishMode bool, memory ProblemMemory) (string, error) {
	prompt := buildChatPrompt(problem, attempt, messages, userMessage, englishMode, memory)
	return c.Generate(ctx, prompt, nil)
}

func (c *CodexClient) Review(ctx context.Context, problem Problem, attempt Attempt, code string, memory ProblemMemory) (ReviewResponse, error) {
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

func (c *CodexClient) SuggestWhiteboard(ctx context.Context, problem Problem, attempt Attempt, messages []ChatMessage, memory ProblemMemory, userMessage string) (WhiteboardSuggestion, error) {
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

func fallbackReview(summary string) ReviewResponse {
	return ReviewResponse{
		IsCorrect: false,
		Summary:   summary,
		Bugs: []string{
			"AIレビューの生成に失敗しました。選択中のAI CLIのログイン状態とCLI pathを確認してください。",
		},
		EdgeCases:             []string{},
		Complexity:            ReviewComplexity{Time: "不明", Space: "不明"},
		ComplexityQuestions:   []string{"この解法の最悪ケースの時間計算量と、その根拠を説明してください。"},
		AlternativeApproaches: []string{"ローカルでレビューできないため、まず全探索と最適化案を自分で比較してください。"},
		FollowUpQuestions:     []string{"同じバグを面接中にどう検知しますか？"},
		ReadabilityFeedback:   []string{},
		InterviewFeedback:     []string{"テスト結果をもとに、自分で計算量と不変条件を説明してみてください。"},
		MistakesToRemember:    []string{"AIレビューが失敗した場合でも、失敗したテストケースを1つずつ手で追う。"},
		GoogleReadiness:       "判定不能",
		HireRecommendation:    "No Hire",
		Scorecard: []ScorecardItem{
			{Area: "Correctness", Score: 1, Signal: "AIレビューが失敗", Evidence: "AI CLIから構造化レビューを取得できませんでした。", Action: "設定復旧後に再レビューする。"},
			{Area: "Communication", Score: 2, Signal: "自己説明が必要", Evidence: "外部レビューなしで計算量と不変条件を説明する必要があります。", Action: "dry runと計算量説明をチャットで行う。"},
		},
		ShadowNotes: []string{"AI CLIの失敗により、bar raiser判定は保守的にNo Hire扱いです。"},
		MiniRounds: []MiniRound{
			{Kind: "coding_followup", Question: "この実装をテストなしでどう検証しますか？", Bar: "主要な境界条件を手でdry runできる。"},
			{Kind: "behavioral", Question: "面接中にツールが使えない状況で、どう品質を担保しますか？", Bar: "制約下での検証戦略を説明できる。"},
		},
		DetectedWeaknesses: []WeaknessSignal{
			{Category: "tooling_resilience", Signal: "AIレビュー失敗時の自己検証が必要", Severity: 3, Evidence: "AI CLIからレビューを取得できなかった。", Drill: "テストなしで3ケースdry runする。"},
		},
		DiscussionPlan:           []string{"AI CLIの設定復旧後に、解法説明、計算量、代替案の順に再レビューする。"},
		NextReviewRecommendation: "AI CLIの設定を直したあと、もう一度レビューを実行してください。",
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
