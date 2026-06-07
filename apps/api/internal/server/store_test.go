package server

import "testing"

func TestStoreSeedsProblemsAndTracksAttemptState(t *testing.T) {
	store, err := NewStore("sqlite:///:memory:")
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	problems, err := LoadSeedProblems()
	if err != nil {
		t.Fatalf("load seed: %v", err)
	}
	if err := store.Seed(problems); err != nil {
		t.Fatalf("seed: %v", err)
	}

	list, err := store.ListProblems()
	if err != nil {
		t.Fatalf("list problems: %v", err)
	}
	if len(list) != 60 {
		t.Fatalf("expected 60 ARAI60 seed problems, got %d", len(list))
	}
	if list[0].Status != "not_started" {
		t.Fatalf("expected first problem to be not_started, got %s", list[0].Status)
	}
	if list[0].Title != "Linked List Cycle" || list[0].OrderIndex != 1 {
		t.Fatalf("expected ARAI60 ordering, got %#v", list[0])
	}

	attempt, err := store.CreateAttempt(CreateAttemptRequest{ProblemID: "two-sum", Language: "python", Code: "code", AIProvider: "claude-code", CompanyPreset: "meta", InterviewMode: "real"})
	if err != nil {
		t.Fatalf("create attempt: %v", err)
	}
	if attempt.AIProvider != "claude" || attempt.CompanyPreset != "meta" || attempt.InterviewMode != "real" || !attempt.NoRun || !attempt.NoAutocomplete || !attempt.RequiresPlan {
		t.Fatalf("unexpected interview defaults: %#v", attempt)
	}

	updated, err := store.UpdateAttemptRun(attempt.ID, "code", "passed", RunResult{
		Passed: true,
		Status: "passed",
		Results: []TestCaseResult{
			{Name: "case", Passed: true},
		},
	})
	if err != nil {
		t.Fatalf("update run: %v", err)
	}
	if updated.Status != "passed" || updated.TestResult == nil || !updated.TestResult.Passed {
		t.Fatalf("unexpected updated attempt: %#v", updated)
	}

	after, err := store.ListProblems()
	if err != nil {
		t.Fatalf("list problems after attempt: %v", err)
	}
	twoSum := findProblem(after, "two-sum")
	if twoSum == nil || twoSum.Status != "solved" {
		t.Fatalf("expected two-sum solved status, got %#v", twoSum)
	}
}

func TestStorePersistsChatAndReviewNotes(t *testing.T) {
	store, err := NewStore("sqlite:///:memory:")
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	problems, err := LoadSeedProblems()
	if err != nil {
		t.Fatalf("load seed: %v", err)
	}
	if err := store.Seed(problems); err != nil {
		t.Fatalf("seed: %v", err)
	}

	attempt, err := store.CreateAttempt(CreateAttemptRequest{ProblemID: "valid-parentheses"})
	if err != nil {
		t.Fatalf("create attempt: %v", err)
	}
	if _, err := store.AddMessage(attempt.ID, "user", "stackで解きます"); err != nil {
		t.Fatalf("add user message: %v", err)
	}
	if _, err := store.AddMessage(attempt.ID, "assistant", "不変条件は何ですか？"); err != nil {
		t.Fatalf("add assistant message: %v", err)
	}
	if _, err := store.EnsureInitialMessage(attempt.ID, "最初の質問"); err != nil {
		t.Fatalf("ensure initial message: %v", err)
	}
	messages, err := store.ListMessages(attempt.ID)
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if len(messages) != 2 || messages[0].Role != "user" || messages[1].Role != "assistant" {
		t.Fatalf("unexpected messages: %#v", messages)
	}
	if err := store.RecordFollowUps(attempt.ID, "chat", []string{"不変条件は何ですか？", "計算量は？"}); err != nil {
		t.Fatalf("record followups: %v", err)
	}

	_, err = store.UpdateAttemptReview(attempt.ID, attempt.Code, ReviewResponse{
		IsCorrect:           false,
		Summary:             "境界条件を確認してください。",
		Complexity:          ReviewComplexity{Time: "O(n)", Space: "O(n)"},
		ComplexityQuestions: []string{"なぜO(n)ですか？"},
		FollowUpQuestions:   []string{"空文字はどう扱いますか？"},
		MistakesToRemember:  []string{"空文字を有効として扱う。"},
		HireRecommendation:  "Lean No Hire",
		DetectedWeaknesses: []WeaknessSignal{
			{Category: "edge_cases", Signal: "空文字の扱いが曖昧", Severity: 4, Evidence: "説明に空文字がなかった", Drill: "空入力からdry runする"},
		},
	})
	if err != nil {
		t.Fatalf("update review: %v", err)
	}

	dashboard, err := store.ReviewDashboard()
	if err != nil {
		t.Fatalf("review dashboard: %v", err)
	}
	if len(dashboard.MistakesToRemember) != 1 {
		t.Fatalf("expected one learning note, got %#v", dashboard.MistakesToRemember)
	}
	if len(dashboard.RecentFollowUps) != 4 {
		t.Fatalf("expected four followups, got %#v", dashboard.RecentFollowUps)
	}
	if len(dashboard.WeaknessGraph) != 1 || dashboard.WeaknessGraph[0].Category != "edge_cases" {
		t.Fatalf("expected edge case weakness graph, got %#v", dashboard.WeaknessGraph)
	}

	memory, err := store.ProblemMemory("valid-parentheses")
	if err != nil {
		t.Fatalf("problem memory: %v", err)
	}
	if len(memory.Attempts) != 1 || len(memory.FollowUps) != 4 || len(memory.Mistakes) != 1 || len(memory.Weaknesses) != 1 {
		t.Fatalf("unexpected memory: %#v", memory)
	}
	if memory.Attempts[0].HireRecommendation != "Lean No Hire" {
		t.Fatalf("expected hire recommendation in memory, got %#v", memory.Attempts[0])
	}
	if memory.Attempts[0].AIProvider != "codex" {
		t.Fatalf("expected default AI provider in memory, got %#v", memory.Attempts[0])
	}
}

func TestStorePersistsWhiteboardsInAttemptAndMemory(t *testing.T) {
	store, err := NewStore("sqlite:///:memory:")
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	problems, err := LoadSeedProblems()
	if err != nil {
		t.Fatalf("load seed: %v", err)
	}
	if err := store.Seed(problems); err != nil {
		t.Fatalf("seed: %v", err)
	}

	attempt, err := store.CreateAttempt(CreateAttemptRequest{ProblemID: "valid-parentheses"})
	if err != nil {
		t.Fatalf("create attempt: %v", err)
	}
	whiteboard, err := store.CreateWhiteboard(attempt.ID, WhiteboardRequest{
		Kind:    "mermaid",
		Topic:   "stack transitions",
		Prompt:  "sequence diagramでpush/popの流れを説明してください。",
		Content: "sequenceDiagram\n  Candidate->>Interviewer: explain stack",
	})
	if err != nil {
		t.Fatalf("create whiteboard: %v", err)
	}
	if whiteboard.Kind != "mermaid_sequence" || whiteboard.Version != 1 || whiteboard.ProblemID != "valid-parentheses" {
		t.Fatalf("unexpected whiteboard: %#v", whiteboard)
	}

	updated, err := store.UpdateWhiteboard(attempt.ID, whiteboard.ID, WhiteboardRequest{
		Kind:    "ds",
		Topic:   "stack invariant",
		Prompt:  "保持する不変条件を書いてください。",
		Content: "Invariant:\n- stack contains unmatched opening brackets",
	})
	if err != nil {
		t.Fatalf("update whiteboard: %v", err)
	}
	if updated.Kind != "data_structure" || updated.Version != 2 || updated.Topic != "stack invariant" {
		t.Fatalf("unexpected updated whiteboard: %#v", updated)
	}

	items, err := store.ListWhiteboardsForAttempt(attempt.ID)
	if err != nil {
		t.Fatalf("list whiteboards: %v", err)
	}
	if len(items) != 1 || items[0].ID != whiteboard.ID || items[0].Content != updated.Content {
		t.Fatalf("unexpected whiteboard list: %#v", items)
	}

	memory, err := store.ProblemMemory("valid-parentheses")
	if err != nil {
		t.Fatalf("problem memory: %v", err)
	}
	if len(memory.Whiteboards) != 1 || memory.Whiteboards[0].ID != whiteboard.ID {
		t.Fatalf("expected whiteboard in problem memory, got %#v", memory.Whiteboards)
	}
}

func findProblem(problems []ProblemListItem, id string) *ProblemListItem {
	for i := range problems {
		if problems[i].ID == id {
			return &problems[i]
		}
	}
	return nil
}
