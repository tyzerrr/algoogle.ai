package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func LoadSeedProblems() ([]Problem, error) {
	data, err := readFirstExisting(seedCandidates())
	if err != nil {
		return nil, err
	}
	var problems []Problem
	if err := json.Unmarshal(data, &problems); err != nil {
		return nil, err
	}
	return ensureArai60Problems(problems), nil
}

func seedCandidates() []string {
	candidates := []string{}
	if configured := os.Getenv("SEED_FILE"); configured != "" {
		candidates = append(candidates, configured)
	}
	candidates = append(candidates, "seed/problems.json")
	if _, file, _, ok := runtime.Caller(0); ok {
		candidates = append(candidates, filepath.Join(filepath.Dir(file), "../../seed/problems.json"))
	}
	return candidates
}

func readFirstExisting(paths []string) ([]byte, error) {
	var lastErr error
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err == nil {
			return data, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

type araiProblem struct {
	ID         string
	Title      string
	Difficulty string
	Pattern    string
	Tags       []string
	Method     string
	Signature  string
	Kind       string
	Starter    string
}

func ensureArai60Problems(seed []Problem) []Problem {
	byID := map[string]Problem{}
	byTitle := map[string]Problem{}
	for _, problem := range seed {
		byID[problem.ID] = problem
		byTitle[strings.ToLower(problem.Title)] = problem
	}

	problems := make([]Problem, 0, len(arai60Problems))
	for index, meta := range arai60Problems {
		problem, ok := byID[meta.ID]
		if !ok {
			problem, ok = byTitle[strings.ToLower(meta.Title)]
		}
		if !ok {
			problem = generatedAraiProblem(meta)
		}
		problem.ID = meta.ID
		problem.Title = meta.Title
		problem.Difficulty = meta.Difficulty
		problem.Pattern = meta.Pattern
		problem.Tags = meta.Tags
		problem.SourceURL = "https://leetcode.com/problems/" + meta.ID + "/"
		problem.ListName = "Arai60"
		problem.OrderIndex = index + 1
		if problem.CreatedAt == "" {
			problem.CreatedAt = fmt.Sprintf("2026-06-08T00:%02d:00Z", index)
		}
		problems = append(problems, problem)
	}
	return problems
}

func generatedAraiProblem(meta araiProblem) Problem {
	return Problem{
		ID:         meta.ID,
		Title:      meta.Title,
		Difficulty: meta.Difficulty,
		Pattern:    meta.Pattern,
		Tags:       meta.Tags,
		Statement: fmt.Sprintf(
			"Arai60収録の「%s」です。公式問題の入出力仕様を確認し、まず全探索、次に面接で説明できる最適化方針、エッジケース、計算量を言語化してから実装してください。LeetCode本文の丸写しではなく、このアプリでは面接練習用カードとして扱います。",
			meta.Title,
		),
		Examples: []Example{
			{
				Input:       "公式問題の例を参照",
				Output:      "期待値を説明",
				Explanation: "実装前に、代表例と境界例を自分の言葉でAI面接官へ説明してください。",
			},
		},
		Constraints: []string{
			"まずナイーブ解を説明する",
			"制約から最適化が必要な箇所を特定する",
			"境界条件を3つ以上挙げる",
			"時間計算量と空間計算量を面接官に説明する",
		},
		StarterCode:         starterFor(meta),
		TestCases:           []TestCase{},
		SolutionExplanation: "Arai60カードです。Codex面接官との会話で複数解法、証明、計算量、フォローアップを詰めてください。",
		SourceURL:           "https://leetcode.com/problems/" + meta.ID + "/",
		ListName:            "Arai60",
		CreatedAt:           "2026-06-08T00:00:00Z",
	}
}

func starterFor(meta araiProblem) string {
	if meta.Starter != "" {
		return meta.Starter
	}
	imports := "from typing import List, Optional\n\n\n"
	prefix := ""
	if meta.Kind == "list" {
		prefix = "class ListNode:\n    def __init__(self, val: int = 0, next: Optional['ListNode'] = None):\n        self.val = val\n        self.next = next\n\n\n"
	}
	if meta.Kind == "tree" {
		prefix = "class TreeNode:\n    def __init__(self, val: int = 0, left: Optional['TreeNode'] = None, right: Optional['TreeNode'] = None):\n        self.val = val\n        self.left = left\n        self.right = right\n\n\n"
	}
	signature := meta.Signature
	if signature == "" {
		signature = fmt.Sprintf("def %s(self, *args):", meta.Method)
	}
	return imports + prefix + "class Solution:\n    " + signature + "\n        # Explain the brute-force idea, optimized idea, invariants, and edge cases before coding.\n        pass\n"
}

var arai60Problems = []araiProblem{
	{ID: "linked-list-cycle", Title: "Linked List Cycle", Difficulty: "Easy", Pattern: "LinkedList", Tags: []string{"linked-list", "two-pointers"}, Method: "hasCycle", Signature: "def hasCycle(self, head: Optional[ListNode]) -> bool:", Kind: "list"},
	{ID: "linked-list-cycle-ii", Title: "Linked List Cycle II", Difficulty: "Medium", Pattern: "LinkedList", Tags: []string{"linked-list", "two-pointers"}, Method: "detectCycle", Signature: "def detectCycle(self, head: Optional[ListNode]) -> Optional[ListNode]:", Kind: "list"},
	{ID: "remove-duplicates-from-sorted-list", Title: "Remove Duplicates from Sorted List", Difficulty: "Easy", Pattern: "LinkedList", Tags: []string{"linked-list"}, Method: "deleteDuplicates", Signature: "def deleteDuplicates(self, head: Optional[ListNode]) -> Optional[ListNode]:", Kind: "list"},
	{ID: "remove-duplicates-from-sorted-list-ii", Title: "Remove Duplicates from Sorted List II", Difficulty: "Medium", Pattern: "LinkedList", Tags: []string{"linked-list"}, Method: "deleteDuplicates", Signature: "def deleteDuplicates(self, head: Optional[ListNode]) -> Optional[ListNode]:", Kind: "list"},
	{ID: "add-two-numbers", Title: "Add Two Numbers", Difficulty: "Medium", Pattern: "LinkedList", Tags: []string{"linked-list", "math"}, Method: "addTwoNumbers", Signature: "def addTwoNumbers(self, l1: Optional[ListNode], l2: Optional[ListNode]) -> Optional[ListNode]:", Kind: "list"},
	{ID: "valid-parentheses", Title: "Valid Parentheses", Difficulty: "Easy", Pattern: "Stack", Tags: []string{"string", "stack"}, Method: "isValid", Signature: "def isValid(self, s: str) -> bool:"},
	{ID: "reverse-linked-list", Title: "Reverse Linked List", Difficulty: "Easy", Pattern: "Stack", Tags: []string{"linked-list", "pointers"}, Method: "reverseList", Signature: "def reverseList(self, head: Optional[ListNode]) -> Optional[ListNode]:", Kind: "list"},
	{ID: "kth-largest-element-in-a-stream", Title: "Kth Largest Element in a Stream", Difficulty: "Easy", Pattern: "Heap, PriorityQueue", Tags: []string{"heap", "design"}, Starter: "from typing import List\n\n\nclass KthLargest:\n    def __init__(self, k: int, nums: List[int]):\n        # Maintain enough state to return the kth largest after each add.\n        pass\n\n    def add(self, val: int) -> int:\n        pass\n"},
	{ID: "top-k-frequent-elements", Title: "Top K Frequent Elements", Difficulty: "Medium", Pattern: "Heap, PriorityQueue", Tags: []string{"array", "hash-map", "heap"}, Method: "topKFrequent", Signature: "def topKFrequent(self, nums: List[int], k: int) -> List[int]:"},
	{ID: "find-k-pairs-with-smallest-sums", Title: "Find K Pairs with Smallest Sums", Difficulty: "Medium", Pattern: "Heap, PriorityQueue", Tags: []string{"array", "heap"}, Method: "kSmallestPairs", Signature: "def kSmallestPairs(self, nums1: List[int], nums2: List[int], k: int) -> List[List[int]]:"},
	{ID: "two-sum", Title: "Two Sum", Difficulty: "Easy", Pattern: "HashMap", Tags: []string{"array", "hash-map"}, Method: "twoSum", Signature: "def twoSum(self, nums: List[int], target: int) -> List[int]:"},
	{ID: "group-anagrams", Title: "Group Anagrams", Difficulty: "Medium", Pattern: "HashMap", Tags: []string{"string", "hash-map"}, Method: "groupAnagrams", Signature: "def groupAnagrams(self, strs: List[str]) -> List[List[str]]:"},
	{ID: "intersection-of-two-arrays", Title: "Intersection of Two Arrays", Difficulty: "Easy", Pattern: "HashMap", Tags: []string{"array", "set"}, Method: "intersection", Signature: "def intersection(self, nums1: List[int], nums2: List[int]) -> List[int]:"},
	{ID: "unique-email-addresses", Title: "Unique Email Addresses", Difficulty: "Easy", Pattern: "HashMap", Tags: []string{"string", "set"}, Method: "numUniqueEmails", Signature: "def numUniqueEmails(self, emails: List[str]) -> int:"},
	{ID: "first-unique-character-in-a-string", Title: "First Unique Character in a String", Difficulty: "Easy", Pattern: "HashMap", Tags: []string{"string", "counting"}, Method: "firstUniqChar", Signature: "def firstUniqChar(self, s: str) -> int:"},
	{ID: "subarray-sum-equals-k", Title: "Subarray Sum Equals K", Difficulty: "Medium", Pattern: "HashMap", Tags: []string{"array", "prefix-sum"}, Method: "subarraySum", Signature: "def subarraySum(self, nums: List[int], k: int) -> int:"},
	{ID: "number-of-islands", Title: "Number of Islands", Difficulty: "Medium", Pattern: "Graph, BFS, DFS", Tags: []string{"matrix", "dfs", "bfs"}, Method: "numIslands", Signature: "def numIslands(self, grid: List[List[str]]) -> int:"},
	{ID: "max-area-of-island", Title: "Max Area of Island", Difficulty: "Medium", Pattern: "Graph, BFS, DFS", Tags: []string{"matrix", "dfs", "bfs"}, Method: "maxAreaOfIsland", Signature: "def maxAreaOfIsland(self, grid: List[List[int]]) -> int:"},
	{ID: "number-of-connected-components-in-an-undirected-graph", Title: "Number of Connected Components in an Undirected Graph", Difficulty: "Medium", Pattern: "Graph, BFS, DFS", Tags: []string{"graph", "union-find"}, Method: "countComponents", Signature: "def countComponents(self, n: int, edges: List[List[int]]) -> int:"},
	{ID: "word-ladder", Title: "Word Ladder", Difficulty: "Hard", Pattern: "Graph, BFS, DFS", Tags: []string{"bfs", "string"}, Method: "ladderLength", Signature: "def ladderLength(self, beginWord: str, endWord: str, wordList: List[str]) -> int:"},
	{ID: "maximum-depth-of-binary-tree", Title: "Maximum Depth of Binary Tree", Difficulty: "Easy", Pattern: "Tree, BT, BST", Tags: []string{"tree", "dfs"}, Method: "maxDepth", Signature: "def maxDepth(self, root: Optional[TreeNode]) -> int:", Kind: "tree"},
	{ID: "minimum-depth-of-binary-tree", Title: "Minimum Depth of Binary Tree", Difficulty: "Easy", Pattern: "Tree, BT, BST", Tags: []string{"tree", "bfs"}, Method: "minDepth", Signature: "def minDepth(self, root: Optional[TreeNode]) -> int:", Kind: "tree"},
	{ID: "merge-two-binary-trees", Title: "Merge Two Binary Trees", Difficulty: "Easy", Pattern: "Tree, BT, BST", Tags: []string{"tree", "recursion"}, Method: "mergeTrees", Signature: "def mergeTrees(self, root1: Optional[TreeNode], root2: Optional[TreeNode]) -> Optional[TreeNode]:", Kind: "tree"},
	{ID: "convert-sorted-array-to-binary-search-tree", Title: "Convert Sorted Array to Binary Search Tree", Difficulty: "Easy", Pattern: "Tree, BT, BST", Tags: []string{"tree", "binary-search"}, Method: "sortedArrayToBST", Signature: "def sortedArrayToBST(self, nums: List[int]) -> Optional[TreeNode]:", Kind: "tree"},
	{ID: "path-sum", Title: "Path Sum", Difficulty: "Easy", Pattern: "Tree, BT, BST", Tags: []string{"tree", "dfs"}, Method: "hasPathSum", Signature: "def hasPathSum(self, root: Optional[TreeNode], targetSum: int) -> bool:", Kind: "tree"},
	{ID: "binary-tree-level-order-traversal", Title: "Binary Tree Level Order Traversal", Difficulty: "Medium", Pattern: "Tree, BT, BST", Tags: []string{"tree", "bfs"}, Method: "levelOrder", Signature: "def levelOrder(self, root: Optional[TreeNode]) -> List[List[int]]:", Kind: "tree"},
	{ID: "binary-tree-zigzag-level-order-traversal", Title: "Binary Tree Zigzag Level Order Traversal", Difficulty: "Medium", Pattern: "Tree, BT, BST", Tags: []string{"tree", "bfs"}, Method: "zigzagLevelOrder", Signature: "def zigzagLevelOrder(self, root: Optional[TreeNode]) -> List[List[int]]:", Kind: "tree"},
	{ID: "validate-binary-search-tree", Title: "Validate Binary Search Tree", Difficulty: "Medium", Pattern: "Tree, BT, BST", Tags: []string{"tree", "bst"}, Method: "isValidBST", Signature: "def isValidBST(self, root: Optional[TreeNode]) -> bool:", Kind: "tree"},
	{ID: "construct-binary-tree-from-preorder-and-inorder-traversal", Title: "Construct Binary Tree from Preorder and Inorder Traversal", Difficulty: "Medium", Pattern: "Tree, BT, BST", Tags: []string{"tree", "recursion"}, Method: "buildTree", Signature: "def buildTree(self, preorder: List[int], inorder: List[int]) -> Optional[TreeNode]:", Kind: "tree"},
	{ID: "paint-fence", Title: "Paint Fence", Difficulty: "Medium", Pattern: "Dynamic Programming", Tags: []string{"dp"}, Method: "numWays", Signature: "def numWays(self, n: int, k: int) -> int:"},
	{ID: "longest-increasing-subsequence", Title: "Longest Increasing Subsequence", Difficulty: "Medium", Pattern: "Dynamic Programming", Tags: []string{"array", "dp", "binary-search"}, Method: "lengthOfLIS", Signature: "def lengthOfLIS(self, nums: List[int]) -> int:"},
	{ID: "maximum-subarray", Title: "Maximum Subarray", Difficulty: "Easy", Pattern: "Dynamic Programming", Tags: []string{"array", "dp"}, Method: "maxSubArray", Signature: "def maxSubArray(self, nums: List[int]) -> int:"},
	{ID: "unique-paths", Title: "Unique Paths", Difficulty: "Medium", Pattern: "Dynamic Programming", Tags: []string{"dp", "grid"}, Method: "uniquePaths", Signature: "def uniquePaths(self, m: int, n: int) -> int:"},
	{ID: "unique-paths-ii", Title: "Unique Paths II", Difficulty: "Medium", Pattern: "Dynamic Programming", Tags: []string{"dp", "grid"}, Method: "uniquePathsWithObstacles", Signature: "def uniquePathsWithObstacles(self, obstacleGrid: List[List[int]]) -> int:"},
	{ID: "house-robber", Title: "House Robber", Difficulty: "Medium", Pattern: "Dynamic Programming", Tags: []string{"array", "dp"}, Method: "rob", Signature: "def rob(self, nums: List[int]) -> int:"},
	{ID: "house-robber-ii", Title: "House Robber II", Difficulty: "Medium", Pattern: "Dynamic Programming", Tags: []string{"array", "dp"}, Method: "rob", Signature: "def rob(self, nums: List[int]) -> int:"},
	{ID: "best-time-to-buy-and-sell-stock", Title: "Best Time to Buy and Sell Stock", Difficulty: "Easy", Pattern: "Dynamic Programming", Tags: []string{"array", "greedy"}, Method: "maxProfit", Signature: "def maxProfit(self, prices: List[int]) -> int:"},
	{ID: "best-time-to-buy-and-sell-stock-ii", Title: "Best Time to Buy and Sell Stock II", Difficulty: "Easy", Pattern: "Dynamic Programming", Tags: []string{"array", "greedy"}, Method: "maxProfit", Signature: "def maxProfit(self, prices: List[int]) -> int:"},
	{ID: "word-break", Title: "Word Break", Difficulty: "Medium", Pattern: "Dynamic Programming", Tags: []string{"string", "dp"}, Method: "wordBreak", Signature: "def wordBreak(self, s: str, wordDict: List[str]) -> bool:"},
	{ID: "coin-change", Title: "Coin Change", Difficulty: "Medium", Pattern: "Dynamic Programming", Tags: []string{"array", "dp"}, Method: "coinChange", Signature: "def coinChange(self, coins: List[int], amount: int) -> int:"},
	{ID: "search-insert-position", Title: "Search Insert Position", Difficulty: "Easy", Pattern: "Binary search", Tags: []string{"array", "binary-search"}, Method: "searchInsert", Signature: "def searchInsert(self, nums: List[int], target: int) -> int:"},
	{ID: "find-minimum-in-rotated-sorted-array", Title: "Find Minimum in Rotated Sorted Array", Difficulty: "Medium", Pattern: "Binary search", Tags: []string{"array", "binary-search"}, Method: "findMin", Signature: "def findMin(self, nums: List[int]) -> int:"},
	{ID: "search-in-rotated-sorted-array", Title: "Search in Rotated Sorted Array", Difficulty: "Medium", Pattern: "Binary search", Tags: []string{"array", "binary-search"}, Method: "search", Signature: "def search(self, nums: List[int], target: int) -> int:"},
	{ID: "capacity-to-ship-packages-within-d-days", Title: "Capacity To Ship Packages Within D Days", Difficulty: "Medium", Pattern: "Binary search", Tags: []string{"array", "binary-search"}, Method: "shipWithinDays", Signature: "def shipWithinDays(self, weights: List[int], days: int) -> int:"},
	{ID: "powx-n", Title: "Pow(x, n)", Difficulty: "Medium", Pattern: "Recursion", Tags: []string{"math", "recursion"}, Method: "myPow", Signature: "def myPow(self, x: float, n: int) -> float:"},
	{ID: "k-th-symbol-in-grammar", Title: "K-th Symbol in Grammar", Difficulty: "Medium", Pattern: "Recursion", Tags: []string{"recursion"}, Method: "kthGrammar", Signature: "def kthGrammar(self, n: int, k: int) -> int:"},
	{ID: "split-bst", Title: "Split BST", Difficulty: "Medium", Pattern: "Recursion", Tags: []string{"tree", "bst", "recursion"}, Method: "splitBST", Signature: "def splitBST(self, root: Optional[TreeNode], target: int) -> List[Optional[TreeNode]]:", Kind: "tree"},
	{ID: "longest-substring-without-repeating-characters", Title: "Longest Substring Without Repeating Characters", Difficulty: "Medium", Pattern: "Sliding window", Tags: []string{"string", "sliding-window"}, Method: "lengthOfLongestSubstring", Signature: "def lengthOfLongestSubstring(self, s: str) -> int:"},
	{ID: "minimum-size-subarray-sum", Title: "Minimum Size Subarray Sum", Difficulty: "Medium", Pattern: "Sliding window", Tags: []string{"array", "sliding-window"}, Method: "minSubArrayLen", Signature: "def minSubArrayLen(self, target: int, nums: List[int]) -> int:"},
	{ID: "permutations", Title: "Permutations", Difficulty: "Medium", Pattern: "Greedy + Backtracking", Tags: []string{"array", "backtracking"}, Method: "permute", Signature: "def permute(self, nums: List[int]) -> List[List[int]]:"},
	{ID: "subsets", Title: "Subsets", Difficulty: "Medium", Pattern: "Greedy + Backtracking", Tags: []string{"array", "backtracking"}, Method: "subsets", Signature: "def subsets(self, nums: List[int]) -> List[List[int]]:"},
	{ID: "combination-sum", Title: "Combination Sum", Difficulty: "Medium", Pattern: "Greedy + Backtracking", Tags: []string{"array", "backtracking"}, Method: "combinationSum", Signature: "def combinationSum(self, candidates: List[int], target: int) -> List[List[int]]:"},
	{ID: "generate-parentheses", Title: "Generate Parentheses", Difficulty: "Medium", Pattern: "Greedy + Backtracking", Tags: []string{"string", "backtracking"}, Method: "generateParenthesis", Signature: "def generateParenthesis(self, n: int) -> List[str]:"},
	{ID: "move-zeroes", Title: "Move Zeroes", Difficulty: "Easy", Pattern: "Mixed", Tags: []string{"array", "two-pointers"}, Method: "moveZeroes", Signature: "def moveZeroes(self, nums: List[int]) -> None:"},
	{ID: "meeting-rooms", Title: "Meeting Rooms", Difficulty: "Easy", Pattern: "Mixed", Tags: []string{"intervals", "sorting"}, Method: "canAttendMeetings", Signature: "def canAttendMeetings(self, intervals: List[List[int]]) -> bool:"},
	{ID: "meeting-rooms-ii", Title: "Meeting Rooms II", Difficulty: "Medium", Pattern: "Mixed", Tags: []string{"intervals", "heap"}, Method: "minMeetingRooms", Signature: "def minMeetingRooms(self, intervals: List[List[int]]) -> int:"},
	{ID: "is-subsequence", Title: "Is Subsequence", Difficulty: "Easy", Pattern: "Mixed", Tags: []string{"string", "two-pointers"}, Method: "isSubsequence", Signature: "def isSubsequence(self, s: str, t: str) -> bool:"},
	{ID: "next-permutation", Title: "Next Permutation", Difficulty: "Medium", Pattern: "Mixed", Tags: []string{"array"}, Method: "nextPermutation", Signature: "def nextPermutation(self, nums: List[int]) -> None:"},
	{ID: "string-to-integer-atoi", Title: "String to Integer (atoi)", Difficulty: "Medium", Pattern: "Mixed", Tags: []string{"string", "parsing"}, Method: "myAtoi", Signature: "def myAtoi(self, s: str) -> int:"},
	{ID: "zigzag-conversion", Title: "ZigZag Conversion", Difficulty: "Medium", Pattern: "Mixed", Tags: []string{"string", "simulation"}, Method: "convert", Signature: "def convert(self, s: str, numRows: int) -> str:"},
}
