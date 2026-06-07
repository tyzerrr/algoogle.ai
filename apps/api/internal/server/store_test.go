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

	attempt, err := store.CreateAttempt(CreateAttemptRequest{ProblemID: "two-sum", Language: "python", Code: "code"})
	if err != nil {
		t.Fatalf("create attempt: %v", err)
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

	memory, err := store.ProblemMemory("valid-parentheses")
	if err != nil {
		t.Fatalf("problem memory: %v", err)
	}
	if len(memory.Attempts) != 1 || len(memory.FollowUps) != 4 || len(memory.Mistakes) != 1 {
		t.Fatalf("unexpected memory: %#v", memory)
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
