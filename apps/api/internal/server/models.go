package server

type Problem struct {
	ID                     string     `json:"id"`
	Title                  string     `json:"title"`
	Difficulty             string     `json:"difficulty"`
	Pattern                string     `json:"pattern"`
	Tags                   []string   `json:"tags"`
	Status                 string     `json:"status"`
	MasteryStatus          string     `json:"mastery_status"`
	AttemptCount           int        `json:"attempt_count"`
	FollowUpCount          int        `json:"follow_up_count"`
	SolvedWithoutFollowUps bool       `json:"solved_without_followups"`
	LastAttemptedAt        *string    `json:"last_attempted_at"`
	Statement              string     `json:"statement"`
	Examples               []Example  `json:"examples"`
	Constraints            []string   `json:"constraints"`
	StarterCode            string     `json:"starter_code"`
	TestCases              []TestCase `json:"test_cases"`
	SolutionExplanation    string     `json:"solution_explanation,omitempty"`
	SourceURL              string     `json:"source_url,omitempty"`
	ListName               string     `json:"list_name,omitempty"`
	OrderIndex             int        `json:"order_index"`
	CreatedAt              string     `json:"created_at"`
}

type ProblemListItem struct {
	ID                     string   `json:"id"`
	Title                  string   `json:"title"`
	Difficulty             string   `json:"difficulty"`
	Pattern                string   `json:"pattern"`
	Tags                   []string `json:"tags"`
	Status                 string   `json:"status"`
	MasteryStatus          string   `json:"mastery_status"`
	AttemptCount           int      `json:"attempt_count"`
	FollowUpCount          int      `json:"follow_up_count"`
	SolvedWithoutFollowUps bool     `json:"solved_without_followups"`
	LastAttemptedAt        *string  `json:"last_attempted_at"`
	SourceURL              string   `json:"source_url,omitempty"`
	ListName               string   `json:"list_name,omitempty"`
	OrderIndex             int      `json:"order_index"`
}

type Example struct {
	Input       string `json:"input"`
	Output      string `json:"output"`
	Explanation string `json:"explanation,omitempty"`
}

type TestCase struct {
	Name     string                 `json:"name"`
	Input    map[string]interface{} `json:"input"`
	Expected interface{}            `json:"expected"`
}

type Attempt struct {
	ID                     string          `json:"id"`
	ProblemID              string          `json:"problem_id"`
	ProblemTitle           string          `json:"problem_title,omitempty"`
	Language               string          `json:"language"`
	Code                   string          `json:"code"`
	CodeFilePath           string          `json:"code_file_path,omitempty"`
	CodeFileUpdatedAt      string          `json:"code_file_updated_at,omitempty"`
	Status                 string          `json:"status"`
	Outcome                string          `json:"outcome"`
	FollowUpCount          int             `json:"follow_up_count"`
	SolvedWithoutFollowUps bool            `json:"solved_without_followups"`
	MistakeSummary         string          `json:"mistake_summary,omitempty"`
	TestResult             *RunResult      `json:"test_result,omitempty"`
	AIReview               *ReviewResponse `json:"ai_review,omitempty"`
	HintsUsed              int             `json:"hints_used"`
	TimeSpentSeconds       *int            `json:"time_spent_seconds,omitempty"`
	CompletedAt            *string         `json:"completed_at,omitempty"`
	CreatedAt              string          `json:"created_at"`
}

type ChatMessage struct {
	ID        string `json:"id"`
	AttemptID string `json:"attempt_id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

type LearningNote struct {
	ID             string  `json:"id"`
	ProblemID      string  `json:"problem_id"`
	ProblemTitle   string  `json:"problem_title,omitempty"`
	MistakeType    string  `json:"mistake_type"`
	Note           string  `json:"note"`
	NextReviewDate *string `json:"next_review_date,omitempty"`
	CreatedAt      string  `json:"created_at"`
}

type RunResult struct {
	Passed     bool             `json:"passed"`
	Status     string           `json:"status"`
	Results    []TestCaseResult `json:"results"`
	Stdout     string           `json:"stdout,omitempty"`
	Stderr     string           `json:"stderr,omitempty"`
	Error      string           `json:"error,omitempty"`
	DurationMS int64            `json:"duration_ms"`
}

type TestCaseResult struct {
	Name     string      `json:"name"`
	Passed   bool        `json:"passed"`
	Expected interface{} `json:"expected,omitempty"`
	Actual   interface{} `json:"actual,omitempty"`
	Error    string      `json:"error,omitempty"`
}

type ReviewResponse struct {
	IsCorrect                bool             `json:"is_correct"`
	Summary                  string           `json:"summary"`
	Bugs                     []string         `json:"bugs"`
	EdgeCases                []string         `json:"edge_cases"`
	Complexity               ReviewComplexity `json:"complexity"`
	ComplexityQuestions      []string         `json:"complexity_questions"`
	AlternativeApproaches    []string         `json:"alternative_approaches"`
	FollowUpQuestions        []string         `json:"follow_up_questions"`
	ReadabilityFeedback      []string         `json:"readability_feedback"`
	InterviewFeedback        []string         `json:"interview_feedback"`
	MistakesToRemember       []string         `json:"mistakes_to_remember"`
	GoogleReadiness          string           `json:"google_readiness"`
	DiscussionPlan           []string         `json:"discussion_plan"`
	NextReviewRecommendation string           `json:"next_review_recommendation"`
}

type ReviewComplexity struct {
	Time  string `json:"time"`
	Space string `json:"space"`
}

type DailyResponse struct {
	Problem        ProblemListItem `json:"problem"`
	Reason         string          `json:"reason"`
	Focus          []string        `json:"focus"`
	RecommendedURL string          `json:"recommended_url"`
}

type ReviewDashboard struct {
	RecentAttempts     []Attempt          `json:"recent_attempts"`
	RecentFollowUps    []FollowUpQuestion `json:"recent_follow_ups"`
	MistakesToRemember []LearningNote     `json:"mistakes_to_remember"`
	WeakPatterns       []WeakPattern      `json:"weak_patterns"`
	ProblemsForToday   []ProblemListItem  `json:"problems_for_today"`
}

type WeakPattern struct {
	Pattern string `json:"pattern"`
	Count   int    `json:"count"`
}

type CreateAttemptRequest struct {
	ProblemID string `json:"problem_id"`
	Language  string `json:"language"`
	Code      string `json:"code"`
}

type ChatRequest struct {
	Message     string `json:"message"`
	EnglishMode bool   `json:"english_mode"`
}

type RunRequest struct {
	Code     string `json:"code"`
	Language string `json:"language"`
}

type ReviewRequest struct {
	Code     string `json:"code"`
	Language string `json:"language"`
}

type CodeFileResponse struct {
	Path      string `json:"path"`
	Content   string `json:"content"`
	UpdatedAt string `json:"updated_at"`
	Size      int64  `json:"size"`
}

type CodeFileRequest struct {
	Code string `json:"code"`
}

type FollowUpQuestion struct {
	ID        string `json:"id"`
	AttemptID string `json:"attempt_id"`
	ProblemID string `json:"problem_id"`
	Question  string `json:"question"`
	Source    string `json:"source"`
	Answered  bool   `json:"answered"`
	CreatedAt string `json:"created_at"`
}

type AttemptMemory struct {
	ID                     string `json:"id"`
	Status                 string `json:"status"`
	Outcome                string `json:"outcome"`
	FollowUpCount          int    `json:"follow_up_count"`
	SolvedWithoutFollowUps bool   `json:"solved_without_followups"`
	MistakeSummary         string `json:"mistake_summary,omitempty"`
	GoogleReadiness        string `json:"google_readiness,omitempty"`
	Summary                string `json:"summary,omitempty"`
	CreatedAt              string `json:"created_at"`
}

type ProblemMemory struct {
	Attempts  []AttemptMemory    `json:"attempts"`
	FollowUps []FollowUpQuestion `json:"follow_ups"`
	Mistakes  []LearningNote     `json:"mistakes"`
}
