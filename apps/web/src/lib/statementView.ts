import type { Example, OfficialProblemContent, Problem, ProblemImage } from "./types";

export type StatementViewMode = "local" | "official" | "official-loading" | "official-fallback";

export type StatementView = {
  mode: StatementViewMode;
  statement: string;
  examples: Example[];
  constraints: string[];
  guidance: string[];
  images: ProblemImage[];
  sourceBadge: string;
  showLanguageToggle: boolean;
};

export type StatementViewInput = {
  problem: Problem;
  official?: OfficialProblemContent | null;
  officialLoading?: boolean;
  officialError?: string;
  language: "ja" | "en";
};

const OFFICIAL_BADGE = "LeetCode 英語原文";

export function resolveStatementView({
  problem,
  official,
  officialLoading = false,
  officialError = "",
  language,
}: StatementViewInput): StatementView {
  const images = official?.images ?? [];
  // The API serializes an absent constraints section as null, not [].
  const officialConstraints = official?.constraints ?? [];

  if (!problem.statement_is_placeholder) {
    // Real Japanese statement: keep today's behaviour where English is opt-in.
    const showLanguageToggle = Boolean(official?.statement);
    const useEnglish = language === "en" && showLanguageToggle;
    if (useEnglish && official) {
      return {
        mode: "official",
        statement: official.statement,
        examples: official.examples.length ? official.examples : problem.examples,
        constraints: officialConstraints.length ? officialConstraints : problem.constraints,
        guidance: [],
        images,
        sourceBadge: OFFICIAL_BADGE,
        showLanguageToggle,
      };
    }
    return {
      mode: "local",
      statement: problem.statement,
      examples: problem.examples,
      constraints: problem.constraints,
      guidance: [],
      // Official figures stay visible in the Japanese view; the local statement has none.
      images,
      sourceBadge: "日本語",
      showLanguageToggle,
    };
  }

  // Placeholder problem: the local statement/examples are dummy, so official content is
  // the real source. `problem.constraints` here holds interview-coaching bullets (進め方).
  const guidance = problem.constraints;

  if (official?.statement) {
    return {
      mode: "official",
      statement: official.statement,
      examples: official.examples,
      constraints: officialConstraints,
      guidance,
      images,
      sourceBadge: OFFICIAL_BADGE,
      showLanguageToggle: false,
    };
  }

  if (officialLoading) {
    return {
      mode: "official-loading",
      statement: "",
      examples: [],
      constraints: [],
      guidance,
      images: [],
      sourceBadge: OFFICIAL_BADGE,
      showLanguageToggle: false,
    };
  }

  // Error (or nothing loaded): show the local placeholder text with a retry affordance.
  return {
    mode: "official-fallback",
    statement: problem.statement,
    examples: problem.examples,
    constraints: [],
    guidance,
    images: [],
    sourceBadge: OFFICIAL_BADGE,
    showLanguageToggle: false,
  };
}
