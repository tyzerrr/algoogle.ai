package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func newSeededStore(t *testing.T) *Store {
	t.Helper()
	store, err := NewStore("sqlite:///:memory:")
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	seed, err := LoadSeedProblems()
	if err != nil {
		t.Fatalf("load seed: %v", err)
	}
	if err := store.Seed(seed); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return store
}

func TestOfficialContentRoundTrip(t *testing.T) {
	store := newSeededStore(t)

	if _, err := store.GetOfficialContent("two-sum"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for missing official content, got %v", err)
	}

	content := OfficialProblemContent{
		Source:      "LeetCode",
		SourceURL:   "https://leetcode.com/problems/two-sum/",
		Title:       "Two Sum",
		Statement:   "Given an array...",
		Constraints: []string{"2 <= nums.length"},
		Examples:    []Example{{Input: "a", Output: "b"}},
		FetchedAt:   "2026-01-01T00:00:00Z",
	}
	if err := store.SaveOfficialContent("two-sum", content); err != nil {
		t.Fatalf("save official content: %v", err)
	}

	got, err := store.GetOfficialContent("two-sum")
	if err != nil {
		t.Fatalf("get official content: %v", err)
	}
	if got.Title != content.Title || len(got.Constraints) != 1 || got.Constraints[0] != "2 <= nums.length" {
		t.Fatalf("round-trip mismatch: %#v", got)
	}
	if len(got.Examples) != 1 || got.Examples[0].Input != "a" {
		t.Fatalf("examples round-trip mismatch: %#v", got.Examples)
	}
}

type countingSourceProvider struct {
	mu    sync.Mutex
	count int
}

func (c *countingSourceProvider) Fetch(_ context.Context, problem Problem) (*OfficialProblemContent, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count++
	return &OfficialProblemContent{
		Source:    "LeetCode",
		SourceURL: problem.SourceURL,
		Title:     problem.Title,
		Statement: "official statement",
	}, nil
}

func TestOfficialHandlerCachesInStore(t *testing.T) {
	store := newSeededStore(t)
	fetcher := &countingSourceProvider{}
	app := &App{store: store, sourceFetcher: fetcher, codeFiles: NewCodeFileManager(t.TempDir(), "./workspace")}
	app.router = app.routes()

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/problems/two-sum/official", nil)
		rec := httptest.NewRecorder()
		app.router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("official status %d: %s", rec.Code, rec.Body.String())
		}
	}
	if fetcher.count != 1 {
		t.Fatalf("expected fetch to happen once, got %d", fetcher.count)
	}

	req := httptest.NewRequest(http.MethodGet, "/problems/add-two-numbers", nil)
	rec := httptest.NewRecorder()
	app.router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("problem status %d: %s", rec.Code, rec.Body.String())
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode problem: %v", err)
	}
	if _, ok := payload["statement_is_placeholder"]; !ok {
		t.Fatalf("expected statement_is_placeholder in problem JSON: %s", rec.Body.String())
	}
}
