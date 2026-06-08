package server

import (
	"strings"
	"testing"
)

func TestLeetCodeSlug(t *testing.T) {
	got := leetcodeSlug("https://leetcode.com/problems/valid-parentheses/")
	if got != "valid-parentheses" {
		t.Fatalf("expected valid-parentheses, got %q", got)
	}

	if got := leetcodeSlug("https://example.com/problems/two-sum/"); got != "" {
		t.Fatalf("expected non-LeetCode URL to be ignored, got %q", got)
	}
}

func TestParseLeetCodeProblemContentKeepsStatementAndExamplesOnly(t *testing.T) {
	content := `<p>Given an array of integers <code>nums</code>&nbsp;and an integer <code>target</code>, return indices.</p>
<p>You may assume exactly one solution.</p>
<p>&nbsp;</p>
<p><strong class="example">Example 1:</strong></p>
<pre>
<strong>Input:</strong> nums = [2,7,11,15], target = 9
<strong>Output:</strong> [0,1]
<strong>Explanation:</strong> Because nums[0] + nums[1] == 9, we return [0, 1].
</pre>
<p><strong class="example">Example 2:</strong></p>
<pre>
<strong>Input:</strong> nums = [3,2,4], target = 6
<strong>Output:</strong> [1,2]
</pre>
<p><strong>Constraints:</strong></p>
<ul><li><code>2 &lt;= nums.length &lt;= 10<sup>4</sup></code></li></ul>`

	parsed := parseLeetCodeProblemContent("Two Sum", "https://leetcode.com/problems/two-sum/", content)

	if !strings.Contains(parsed.Statement, "Given an array of integers nums and an integer target") {
		t.Fatalf("expected statement text, got %q", parsed.Statement)
	}
	if strings.Contains(parsed.Statement, "Constraints") || strings.Contains(parsed.Statement, "nums.length") {
		t.Fatalf("constraints leaked into statement: %q", parsed.Statement)
	}
	if len(parsed.Examples) != 2 {
		t.Fatalf("expected two examples, got %#v", parsed.Examples)
	}
	if parsed.Examples[0].Input != "nums = [2,7,11,15], target = 9" {
		t.Fatalf("unexpected input: %#v", parsed.Examples[0])
	}
	if parsed.Examples[0].Output != "[0,1]" {
		t.Fatalf("unexpected output: %#v", parsed.Examples[0])
	}
	if parsed.Examples[0].Explanation == "" {
		t.Fatalf("expected explanation: %#v", parsed.Examples[0])
	}
	if parsed.Examples[1].Input != "nums = [3,2,4], target = 6" || parsed.Examples[1].Output != "[1,2]" {
		t.Fatalf("unexpected second example: %#v", parsed.Examples[1])
	}
}
