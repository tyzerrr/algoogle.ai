package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type CodeRunner struct {
	pythonBin string
}

func NewCodeRunner(pythonBin string) *CodeRunner {
	return &CodeRunner{pythonBin: pythonBin}
}

func (r *CodeRunner) Run(ctx context.Context, problemID, code string, testCases []TestCase) RunResult {
	start := time.Now()
	result := RunResult{Status: "failed"}

	if strings.TrimSpace(code) == "" {
		result.Error = "コードが空です。"
		result.DurationMS = elapsed(start)
		return result
	}
	if len(testCases) == 0 {
		result.Status = "not_configured"
		result.Error = "この問題のローカルテストはまだ登録されていません。実装後はAIレビューで方針、計算量、エッジケースを詰めてください。"
		result.DurationMS = elapsed(start)
		return result
	}

	dir, err := os.MkdirTemp("", "algosensei-run-*")
	if err != nil {
		result.Error = err.Error()
		result.DurationMS = elapsed(start)
		return result
	}
	defer os.RemoveAll(dir)

	filePath := filepath.Join(dir, "main.py")
	source, err := buildPythonHarness(problemID, code, testCases)
	if err != nil {
		result.Error = err.Error()
		result.DurationMS = elapsed(start)
		return result
	}
	if err := os.WriteFile(filePath, []byte(source), 0600); err != nil {
		result.Error = err.Error()
		result.DurationMS = elapsed(start)
		return result
	}

	runCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	cmd := exec.CommandContext(runCtx, r.pythonBin, filePath)
	cmd.Dir = dir
	cmd.Env = []string{"PYTHONIOENCODING=utf-8", "PATH=/usr/bin:/bin:/usr/local/bin:/opt/homebrew/bin"}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()

	result.Stdout = trimOutput(stdout.String())
	result.Stderr = trimOutput(stderr.String())
	result.DurationMS = elapsed(start)
	if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
		result.Status = "runtime_error"
		result.Error = "実行がタイムアウトしました。無限ループや重すぎる処理を確認してください。"
		return result
	}
	if err != nil {
		result.Status = "runtime_error"
		result.Error = err.Error()
		return result
	}

	parsed, err := parseRunnerOutput(stdout.String())
	if err != nil {
		result.Status = "runtime_error"
		result.Error = err.Error()
		return result
	}
	parsed.Stdout = result.Stdout
	parsed.Stderr = result.Stderr
	parsed.DurationMS = result.DurationMS
	if parsed.Passed {
		parsed.Status = "passed"
	} else {
		parsed.Status = "failed"
	}
	return parsed
}

func buildPythonHarness(problemID, userCode string, testCases []TestCase) (string, error) {
	testsJSON, err := json.Marshal(testCases)
	if err != nil {
		return "", err
	}
	harness := `

import json as __algosensei_json
import traceback as __algosensei_traceback

__ALGOSENSEI_PROBLEM_ID = ` + strconv.Quote(problemID) + `
__ALGOSENSEI_TESTS = __algosensei_json.loads(` + strconv.Quote(string(testsJSON)) + `)

def __algosensei_build_list(values):
    head = None
    tail = None
    for value in values:
        node = ListNode(value)
        if head is None:
            head = node
            tail = node
        else:
            tail.next = node
            tail = node
    return head

def __algosensei_list_to_values(head):
    values = []
    seen = 0
    while head is not None:
        values.append(head.val)
        head = head.next
        seen += 1
        if seen > 10000:
            raise RuntimeError("linked list appears to contain a cycle")
    return values

def __algosensei_run_case(case):
    data = case.get("input", {})
    sol = Solution()
    if __ALGOSENSEI_PROBLEM_ID == "two-sum":
        return sol.two_sum(data["nums"], data["target"])
    if __ALGOSENSEI_PROBLEM_ID == "valid-parentheses":
        return sol.is_valid(data["s"])
    if __ALGOSENSEI_PROBLEM_ID == "longest-substring-without-repeating" or __ALGOSENSEI_PROBLEM_ID == "longest-substring-without-repeating-characters":
        return sol.length_of_longest_substring(data["s"])
    if __ALGOSENSEI_PROBLEM_ID == "binary-search":
        return sol.binary_search(data["nums"], data["target"])
    if __ALGOSENSEI_PROBLEM_ID == "reverse-linked-list":
        head = __algosensei_build_list(data["values"])
        return __algosensei_list_to_values(sol.reverseList(head))
    raise RuntimeError("unknown problem id: " + __ALGOSENSEI_PROBLEM_ID)

def __algosensei_equal(actual, expected):
    if __ALGOSENSEI_PROBLEM_ID == "two-sum" and isinstance(actual, list) and isinstance(expected, list):
        return sorted(actual) == sorted(expected)
    return actual == expected

__algosensei_results = []
for __algosensei_case in __ALGOSENSEI_TESTS:
    try:
        __algosensei_actual = __algosensei_run_case(__algosensei_case)
        __algosensei_expected = __algosensei_case.get("expected")
        __algosensei_results.append({
            "name": __algosensei_case.get("name", "case"),
            "passed": __algosensei_equal(__algosensei_actual, __algosensei_expected),
            "actual": __algosensei_actual,
            "expected": __algosensei_expected,
        })
    except Exception:
        __algosensei_results.append({
            "name": __algosensei_case.get("name", "case"),
            "passed": False,
            "error": __algosensei_traceback.format_exc(limit=4),
            "expected": __algosensei_case.get("expected"),
        })

__algosensei_payload = {
    "passed": all(item["passed"] for item in __algosensei_results),
    "status": "passed" if all(item["passed"] for item in __algosensei_results) else "failed",
    "results": __algosensei_results,
}
print("__ALGOSENSEI_RESULT__" + __algosensei_json.dumps(__algosensei_payload, ensure_ascii=False))
`
	return userCode + harness, nil
}

func parseRunnerOutput(stdout string) (RunResult, error) {
	lines := strings.Split(stdout, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.HasPrefix(lines[i], "__ALGOSENSEI_RESULT__") {
			raw := strings.TrimPrefix(lines[i], "__ALGOSENSEI_RESULT__")
			var result RunResult
			if err := json.Unmarshal([]byte(raw), &result); err != nil {
				return result, err
			}
			return result, nil
		}
	}
	return RunResult{}, errors.New("テスト結果を読み取れませんでした。トップレベルでプロセスを終了していないか確認してください。")
}

func elapsed(start time.Time) int64 {
	return time.Since(start).Milliseconds()
}

func trimOutput(value string) string {
	value = strings.TrimSpace(value)
	if len(value) > 4000 {
		return value[:4000] + "\n... truncated ..."
	}
	return value
}
