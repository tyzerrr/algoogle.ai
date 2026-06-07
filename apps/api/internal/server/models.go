package server

type Problem struct {
	ID                  string     `json:"id"`
	Title               string     `json:"title"`
	Difficulty          string     `json:"difficulty"`
	Pattern             string     `json:"pattern"`
	Tags                []string   `json:"tags"`
	Status              string     `json:"status"`
	LastAttemptedAt     *string    `json:"last_attempted_at"`
	Statement           string     `json:"statement"`
	Examples            []Example  `json:"examples"`
	Constraints         []string   `json:"constraints"`
	StarterCode         string     `json:"starter_code"`
	TestCases           []TestCase `json:"test_cases"`
	SolutionExplanation string     `json:"solution_explanation,omitempty"`
	CreatedAt           string     `json:"created_at"`
}

type ProblemListItem struct {
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	Difficulty      string   `json:"difficulty"`
	Pattern         string   `json:"pattern"`
	Tags            []string `json:"tags"`
	Status          string   `json:"status"`
	LastAttemptedAt *string  `json:"last_attempted_at"`
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
	ID               string          `json:"id"`
	ProblemID        string          `json:"problem_id"`
	ProblemTitle     string          `json:"problem_title,omitempty"`
	Language         string          `json:"language"`
	Code             string          `json:"code"`
	Status           string          `json:"status"`
	TestResult       *RunResult      `json:"test_result,omitempty"`
	AIReview         *ReviewResponse `json:"ai_review,omitempty"`
	HintsUsed        int             `json:"hints_used"`
	TimeSpentSeconds *int            `json:"time_spent_seconds,omitempty"`
	CreatedAt        string          `json:"created_at"`
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
	ReadabilityFeedback      []string         `json:"readability_feedback"`
	InterviewFeedback        []string         `json:"interview_feedback"`
	MistakesToRemember       []string         `json:"mistakes_to_remember"`
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
	RecentAttempts     []Attempt         `json:"recent_attempts"`
	MistakesToRemember []LearningNote    `json:"mistakes_to_remember"`
	WeakPatterns       []WeakPattern     `json:"weak_patterns"`
	ProblemsForToday   []ProblemListItem `json:"problems_for_today"`
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
