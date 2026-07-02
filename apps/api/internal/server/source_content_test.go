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
<p><img alt="Two Sum diagram" src="/uploads/two-sum.png" /></p>
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
	if len(parsed.Images) != 1 {
		t.Fatalf("expected one image, got %#v", parsed.Images)
	}
	if parsed.Images[0].URL != "https://leetcode.com/uploads/two-sum.png" || parsed.Images[0].Alt != "Two Sum diagram" {
		t.Fatalf("unexpected image: %#v", parsed.Images[0])
	}
}

func TestParseLeetCodeConstraintsWithoutHeadingReturnsEmptyNotNil(t *testing.T) {
	parsed := parseLeetCodeProblemContent("Two Sum", "https://leetcode.com/problems/two-sum/", "<p>Statement only.</p>")
	if parsed.Constraints == nil {
		t.Fatal("constraints must be an empty slice, not nil, so JSON encodes [] instead of null")
	}
	if len(parsed.Constraints) != 0 {
		t.Fatalf("expected no constraints, got %#v", parsed.Constraints)
	}
}

func TestParseLeetCodeConstraintsExtractsItemsAndExcludesFollowUp(t *testing.T) {
	content := `<p>Given an array, return indices.</p>
<p><strong class="example">Example 1:</strong></p>
<pre>
<strong>Input:</strong> nums = [2,7], target = 9
<strong>Output:</strong> [0,1]
</pre>
<p><strong>Constraints:</strong></p>
<ul><li>2 &lt;= nums.length &lt;= 10^4</li><li>-10^9 &lt;= nums[i] &lt;= 10^9</li></ul>
<p><strong class="example">Follow up:</strong> Can you do it in O(n)?</p>`

	parsed := parseLeetCodeProblemContent("Two Sum", "https://leetcode.com/problems/two-sum/", content)

	if len(parsed.Constraints) != 2 {
		t.Fatalf("expected 2 constraints, got %#v", parsed.Constraints)
	}
	if parsed.Constraints[0] != "2 <= nums.length <= 10^4" {
		t.Fatalf("unexpected first constraint: %q", parsed.Constraints[0])
	}
	if parsed.Constraints[1] != "-10^9 <= nums[i] <= 10^9" {
		t.Fatalf("unexpected second constraint: %q", parsed.Constraints[1])
	}
	for _, item := range parsed.Constraints {
		if strings.Contains(item, "Follow") || strings.Contains(item, "O(n)") {
			t.Fatalf("follow-up leaked into constraints: %#v", parsed.Constraints)
		}
	}
	if strings.Contains(parsed.Statement, "nums.length") || strings.Contains(parsed.Statement, "Constraints") {
		t.Fatalf("constraints leaked into statement: %q", parsed.Statement)
	}
	if len(parsed.Examples) != 1 {
		t.Fatalf("expected one example unaffected, got %#v", parsed.Examples)
	}
}

func TestParseLeetCodeImagesDeduplicatesAndIgnoresDataURLs(t *testing.T) {
	content := `<p>
<img src="//assets.leetcode.com/uploads/tree.jpg" alt="Tree">
<img src="//assets.leetcode.com/uploads/tree.jpg" alt="Duplicate">
<img src="data:image/png;base64,abc" alt="inline">
</p>`

	images := parseLeetCodeImages(content, "https://leetcode.com/problems/maximum-depth-of-binary-tree/")

	if len(images) != 1 {
		t.Fatalf("expected one image, got %#v", images)
	}
	if images[0].URL != "https://assets.leetcode.com/uploads/tree.jpg" || images[0].Alt != "Tree" {
		t.Fatalf("unexpected image: %#v", images[0])
	}
}
