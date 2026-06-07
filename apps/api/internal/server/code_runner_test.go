package server

import (
	"context"
	"os/exec"
	"testing"
)

func TestCodeRunnerRunsPassingTwoSum(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 is not available")
	}

	code := `from typing import List

class Solution:
    def two_sum(self, nums: List[int], target: int) -> List[int]:
        seen = {}
        for i, value in enumerate(nums):
            want = target - value
            if want in seen:
                return [seen[want], i]
            seen[value] = i
        return []
`

	result := NewCodeRunner("python3").Run(context.Background(), "two-sum", code, []TestCase{
		{Name: "basic", Input: map[string]interface{}{"nums": []int{2, 7, 11, 15}, "target": 9}, Expected: []int{0, 1}},
		{Name: "unordered answer is accepted", Input: map[string]interface{}{"nums": []int{3, 2, 4}, "target": 6}, Expected: []int{1, 2}},
	})

	if !result.Passed {
		t.Fatalf("expected tests to pass, got status=%s error=%q results=%v", result.Status, result.Error, result.Results)
	}
	if result.Status != "passed" {
		t.Fatalf("expected status passed, got %s", result.Status)
	}
}

func TestCodeRunnerReportsFailedCase(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 is not available")
	}

	code := `class Solution:
    def binary_search(self, nums, target):
        return 0
`

	result := NewCodeRunner("python3").Run(context.Background(), "binary-search", code, []TestCase{
		{Name: "not found", Input: map[string]interface{}{"nums": []int{1, 3, 5}, "target": 2}, Expected: -1},
	})

	if result.Passed {
		t.Fatal("expected failing test case")
	}
	if result.Status != "failed" {
		t.Fatalf("expected status failed, got %s", result.Status)
	}
	if len(result.Results) != 1 || result.Results[0].Passed {
		t.Fatalf("expected one failed result, got %#v", result.Results)
	}
}

func TestCodeRunnerReportsSyntaxErrors(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 is not available")
	}

	result := NewCodeRunner("python3").Run(context.Background(), "valid-parentheses", "class Solution:\n    def is_valid(", []TestCase{
		{Name: "simple", Input: map[string]interface{}{"s": "()"}, Expected: true},
	})

	if result.Status != "runtime_error" {
		t.Fatalf("expected runtime_error, got %s", result.Status)
	}
	if result.Error == "" {
		t.Fatal("expected syntax error details")
	}
}
