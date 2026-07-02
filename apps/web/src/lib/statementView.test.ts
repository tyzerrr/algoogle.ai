import { describe, expect, it } from "vitest";
import { resolveStatementView } from "./statementView";
import type { OfficialProblemContent, Problem } from "./types";

function makeProblem(overrides: Partial<Problem> = {}): Problem {
  return {
    id: "two-sum",
    title: "Two Sum",
    difficulty: "Easy",
    pattern: "Hash Map",
    tags: [],
    status: "not_started",
    mastery_status: "not_started",
    attempt_count: 0,
    follow_up_count: 0,
    solved_without_followups: false,
    last_attempted_at: null,
    order_index: 1,
    statement: "配列から2つの数を選び、和がtargetになる添字を返してください。",
    statement_is_placeholder: false,
    examples: [{ input: "nums=[2,7]", output: "[0,1]" }],
    constraints: ["2 <= nums.length"],
    starter_code: "def solve():\n    pass",
    test_cases: [],
    created_at: "2024-01-01T00:00:00Z",
    source_url: "https://leetcode.com/problems/two-sum/",
    ...overrides,
  } as Problem;
}

function makeOfficial(overrides: Partial<OfficialProblemContent> = {}): OfficialProblemContent {
  return {
    source: "LeetCode",
    source_url: "https://leetcode.com/problems/two-sum/",
    title: "Two Sum",
    statement: "Given an array of integers nums and an integer target...",
    constraints: ["2 <= nums.length <= 10^4"],
    examples: [{ input: "nums = [2,7], target = 9", output: "[0,1]" }],
    images: [{ url: "https://example.com/fig.png", alt: "figure" }],
    fetched_at: "2024-01-01T00:00:00Z",
    ...overrides,
  };
}

describe("resolveStatementView", () => {
  it("real-JA problem defaults to local ja content with toggle when official is available", () => {
    const problem = makeProblem();
    const official = makeOfficial();
    const view = resolveStatementView({
      problem,
      official,
      officialLoading: false,
      officialError: "",
      language: "ja",
    });
    expect(view.mode).toBe("local");
    expect(view.statement).toBe(problem.statement);
    expect(view.examples).toEqual(problem.examples);
    expect(view.constraints).toEqual(problem.constraints);
    expect(view.showLanguageToggle).toBe(true);
    expect(view.guidance).toEqual([]);
    expect(view.images).toEqual(official.images);
  });

  it("tolerates official constraints serialized as null by the API", () => {
    const problem = makeProblem({ statement_is_placeholder: true });
    const official = makeOfficial({ constraints: null as unknown as string[] });
    const view = resolveStatementView({
      problem,
      official,
      officialLoading: false,
      officialError: "",
      language: "ja",
    });
    expect(view.mode).toBe("official");
    expect(view.constraints).toEqual([]);
  });

  it("real-JA problem with en toggle and null official constraints keeps local constraints", () => {
    const problem = makeProblem();
    const official = makeOfficial({ constraints: null as unknown as string[] });
    const view = resolveStatementView({
      problem,
      official,
      officialLoading: false,
      officialError: "",
      language: "en",
    });
    expect(view.constraints).toEqual(problem.constraints);
  });

  it("real-JA problem with en toggle uses official statement and examples", () => {
    const problem = makeProblem();
    const official = makeOfficial();
    const view = resolveStatementView({
      problem,
      official,
      officialLoading: false,
      officialError: "",
      language: "en",
    });
    expect(view.mode).toBe("official");
    expect(view.statement).toBe(official.statement);
    expect(view.examples).toEqual(official.examples);
    expect(view.images).toEqual(official.images);
    expect(view.showLanguageToggle).toBe(true);
  });

  it("placeholder problem shows official content by default with a source badge and no toggle", () => {
    const problem = makeProblem({
      statement_is_placeholder: true,
      constraints: ["まず制約を質問する", "全探索から話す"],
      examples: [{ input: "公式問題の例を参照", output: "公式問題の例を参照" }],
    });
    const official = makeOfficial();
    const view = resolveStatementView({
      problem,
      official,
      officialLoading: false,
      officialError: "",
      language: "ja",
    });
    expect(view.mode).toBe("official");
    expect(view.statement).toBe(official.statement);
    expect(view.examples).toEqual(official.examples);
    expect(view.constraints).toEqual(official.constraints);
    expect(view.showLanguageToggle).toBe(false);
    expect(view.sourceBadge).toBe("LeetCode 英語原文");
    expect(view.guidance).toEqual(problem.constraints);
  });

  it("placeholder problem while official is loading reports official-loading", () => {
    const problem = makeProblem({ statement_is_placeholder: true });
    const view = resolveStatementView({
      problem,
      official: null,
      officialLoading: true,
      officialError: "",
      language: "ja",
    });
    expect(view.mode).toBe("official-loading");
    expect(view.showLanguageToggle).toBe(false);
  });

  it("placeholder problem with official error falls back to local text with retry", () => {
    const problem = makeProblem({ statement_is_placeholder: true });
    const view = resolveStatementView({
      problem,
      official: null,
      officialLoading: false,
      officialError: "network down",
      language: "ja",
    });
    expect(view.mode).toBe("official-fallback");
    expect(view.statement).toBe(problem.statement);
    expect(view.guidance).toEqual(problem.constraints);
  });
});
