export type ProblemStatus = "not_started" | "in_progress" | "needs_review" | "solved";

export type ProblemListItem = {
  id: string;
  title: string;
  difficulty: string;
  pattern: string;
  tags: string[];
  status: ProblemStatus;
  last_attempted_at: string | null;
};

export type Example = {
  input: string;
  output: string;
  explanation?: string;
};

export type TestCase = {
  name: string;
  input: Record<string, unknown>;
  expected: unknown;
};

export type Problem = ProblemListItem & {
  statement: string;
  examples: Example[];
  constraints: string[];
  starter_code: string;
  test_cases: TestCase[];
  solution_explanation?: string;
  created_at: string;
};

export type TestCaseResult = {
  name: string;
  passed: boolean;
  expected?: unknown;
  actual?: unknown;
  error?: string;
};

export type RunResult = {
  passed: boolean;
  status: string;
  results: TestCaseResult[];
  stdout?: string;
  stderr?: string;
  error?: string;
  duration_ms: number;
};

export type ReviewResponse = {
  is_correct: boolean;
  summary: string;
  bugs: string[];
  edge_cases: string[];
  complexity: {
    time: string;
    space: string;
  };
  readability_feedback: string[];
  interview_feedback: string[];
  mistakes_to_remember: string[];
  next_review_recommendation: string;
};

export type Attempt = {
  id: string;
  problem_id: string;
  problem_title?: string;
  language: string;
  code: string;
  status: string;
  test_result?: RunResult;
  ai_review?: ReviewResponse;
  hints_used: number;
  time_spent_seconds?: number;
  created_at: string;
};

export type ChatMessage = {
  id: string;
  attempt_id: string;
  role: "user" | "assistant";
  content: string;
  created_at: string;
};

export type DailyResponse = {
  problem: ProblemListItem;
  reason: string;
  focus: string[];
  recommended_url: string;
};

export type LearningNote = {
  id: string;
  problem_id: string;
  problem_title?: string;
  mistake_type: string;
  note: string;
  next_review_date?: string;
  created_at: string;
};

export type WeakPattern = {
  pattern: string;
  count: number;
};

export type ReviewDashboard = {
  recent_attempts: Attempt[];
  mistakes_to_remember: LearningNote[];
  weak_patterns: WeakPattern[];
  problems_for_today: ProblemListItem[];
};
