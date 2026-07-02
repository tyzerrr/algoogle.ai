package server

import "testing"

func TestSeedMarksPlaceholderStatements(t *testing.T) {
	problems, err := LoadSeedProblems()
	if err != nil {
		t.Fatalf("load seed: %v", err)
	}
	byID := map[string]Problem{}
	for _, problem := range problems {
		byID[problem.ID] = problem
	}

	seedMatched, ok := byID["two-sum"]
	if !ok {
		t.Fatalf("expected two-sum in seed")
	}
	if seedMatched.StatementIsPlaceholder {
		t.Fatalf("expected seed-JSON-matched problem to not be a placeholder")
	}

	generated, ok := byID["add-two-numbers"]
	if !ok {
		t.Fatalf("expected add-two-numbers in seed")
	}
	if !generated.StatementIsPlaceholder {
		t.Fatalf("expected generated Arai problem to be a placeholder")
	}
}
