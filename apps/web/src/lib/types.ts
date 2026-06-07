export type ProblemStatus = "not_started" | "in_progress" | "needs_review" | "solved";
export type MasteryStatus =
  | "not_started"
  | "in_progress"
  | "needs_review"
  | "first_try_clean"
  | "first_try_with_followups"
  | "solved_after_retry";
export type CompanyPreset = "google" | "meta" | "amazon" | "generic";
export type InterviewMode = "real" | "practice";
export type AIProvider = "codex" | "claude";
export type ThemeMode = "light" | "dark" | "netflix";

export type ProblemListItem = {
  id: string;
  title: string;
  difficulty: string;
  pattern: string;
  tags: string[];
  status: ProblemStatus;
  mastery_status: MasteryStatus;
  attempt_count: number;
  follow_up_count: number;
  solved_without_followups: boolean;
  last_attempted_at: string | null;
  source_url?: string;
  list_name?: string;
  order_index: number;
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
  source_url?: string;
  list_name?: string;
  order_index: number;
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
  complexity_questions: string[];
  alternative_approaches: string[];
  follow_up_questions: string[];
  readability_feedback: string[];
  interview_feedback: string[];
  mistakes_to_remember: string[];
  google_readiness: string;
  hire_recommendation: string;
  scorecard: ScorecardItem[];
  shadow_notes: string[];
  mini_rounds: MiniRound[];
  detected_weaknesses: WeaknessSignal[];
  discussion_plan: string[];
  next_review_recommendation: string;
};

export type ScorecardItem = {
  area: string;
  score: number;
  signal: string;
  evidence: string;
  action: string;
};

export type MiniRound = {
  kind: string;
  question: string;
  bar: string;
};

export type Attempt = {
  id: string;
  problem_id: string;
  problem_title?: string;
  language: string;
  code: string;
  code_file_path?: string;
  code_file_updated_at?: string;
  status: string;
  outcome: string;
  ai_provider: AIProvider;
  company_preset: CompanyPreset;
  interview_mode: InterviewMode;
  current_phase: string;
  time_limit_seconds: number;
  no_run: boolean;
  no_autocomplete: boolean;
  requires_plan: boolean;
  follow_up_count: number;
  solved_without_followups: boolean;
  mistake_summary?: string;
  test_result?: RunResult;
  ai_review?: ReviewResponse;
  hints_used: number;
  time_spent_seconds?: number;
  completed_at?: string;
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

export type FollowUpQuestion = {
  id: string;
  attempt_id: string;
  problem_id: string;
  question: string;
  source: string;
  answered: boolean;
  created_at: string;
};

export type WeakPattern = {
  pattern: string;
  count: number;
};

export type WeaknessSignal = {
  id?: string;
  attempt_id?: string;
  problem_id?: string;
  category: string;
  signal: string;
  severity: number;
  evidence: string;
  drill: string;
  created_at?: string;
};

export type WeaknessTrend = {
  category: string;
  count: number;
  average_severity: number;
  last_seen_at: string;
  suggested_drill: string;
};

export type ReviewDashboard = {
  recent_attempts: Attempt[];
  recent_follow_ups: FollowUpQuestion[];
  mistakes_to_remember: LearningNote[];
  weak_patterns: WeakPattern[];
  weakness_graph: WeaknessTrend[];
  problems_for_today: ProblemListItem[];
};

export type CodeFileResponse = {
  path: string;
  content: string;
  updated_at: string;
  size: number;
};
