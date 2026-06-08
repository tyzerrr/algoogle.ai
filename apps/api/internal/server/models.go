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

type OfficialProblemContent struct {
	Source    string    `json:"source"`
	SourceURL string    `json:"source_url"`
	Title     string    `json:"title"`
	Statement string    `json:"statement"`
	Examples  []Example `json:"examples"`
	FetchedAt string    `json:"fetched_at"`
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
	AIProvider             string          `json:"ai_provider"`
	CompanyPreset          string          `json:"company_preset"`
	InterviewMode          string          `json:"interview_mode"`
	CurrentPhase           string          `json:"current_phase"`
	TimeLimitSeconds       int             `json:"time_limit_seconds"`
	NoRun                  bool            `json:"no_run"`
	NoAutocomplete         bool            `json:"no_autocomplete"`
	RequiresPlan           bool            `json:"requires_plan"`
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
	HireRecommendation       string           `json:"hire_recommendation"`
	Scorecard                []ScorecardItem  `json:"scorecard"`
	ShadowNotes              []string         `json:"shadow_notes"`
	MiniRounds               []MiniRound      `json:"mini_rounds"`
	DetectedWeaknesses       []WeaknessSignal `json:"detected_weaknesses"`
	DiscussionPlan           []string         `json:"discussion_plan"`
	NextReviewRecommendation string           `json:"next_review_recommendation"`
}

type ReviewComplexity struct {
	Time  string `json:"time"`
	Space string `json:"space"`
}

type ScorecardItem struct {
	Area     string `json:"area"`
	Score    int    `json:"score"`
	Signal   string `json:"signal"`
	Evidence string `json:"evidence"`
	Action   string `json:"action"`
}

type MiniRound struct {
	Kind     string `json:"kind"`
	Question string `json:"question"`
	Bar      string `json:"bar"`
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
	WeaknessGraph      []WeaknessTrend    `json:"weakness_graph"`
	ProblemsForToday   []ProblemListItem  `json:"problems_for_today"`
}

type WeakPattern struct {
	Pattern string `json:"pattern"`
	Count   int    `json:"count"`
}

type CreateAttemptRequest struct {
	ProblemID     string `json:"problem_id"`
	Language      string `json:"language"`
	Code          string `json:"code"`
	AIProvider    string `json:"ai_provider"`
	CompanyPreset string `json:"company_preset"`
	InterviewMode string `json:"interview_mode"`
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

type NudgeRequest struct {
	Reason string `json:"reason"`
}

type WhiteboardArtifact struct {
	ID        string `json:"id"`
	AttemptID string `json:"attempt_id"`
	ProblemID string `json:"problem_id"`
	Kind      string `json:"kind"`
	Topic     string `json:"topic"`
	Prompt    string `json:"prompt"`
	Content   string `json:"content"`
	Version   int    `json:"version"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type WhiteboardRequest struct {
	Kind    string `json:"kind"`
	Topic   string `json:"topic"`
	Prompt  string `json:"prompt"`
	Content string `json:"content"`
}

type WhiteboardSuggestionRequest struct {
	Message string `json:"message"`
}

type WhiteboardSuggestion struct {
	UseWhiteboard  bool   `json:"use_whiteboard"`
	Kind           string `json:"kind"`
	Topic          string `json:"topic"`
	Prompt         string `json:"prompt"`
	StarterContent string `json:"starter_content"`
	Reason         string `json:"reason"`
}

type WhiteboardSuggestionResponse struct {
	Suggestion *WhiteboardSuggestion `json:"suggestion"`
	Whiteboard *WhiteboardArtifact   `json:"whiteboard,omitempty"`
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

type WeaknessSignal struct {
	ID        string `json:"id,omitempty"`
	AttemptID string `json:"attempt_id,omitempty"`
	ProblemID string `json:"problem_id,omitempty"`
	Category  string `json:"category"`
	Signal    string `json:"signal"`
	Severity  int    `json:"severity"`
	Evidence  string `json:"evidence"`
	Drill     string `json:"drill"`
	CreatedAt string `json:"created_at,omitempty"`
}

type WeaknessTrend struct {
	Category        string `json:"category"`
	Count           int    `json:"count"`
	AverageSeverity int    `json:"average_severity"`
	LastSeenAt      string `json:"last_seen_at"`
	SuggestedDrill  string `json:"suggested_drill"`
}

type AttemptMemory struct {
	ID                     string `json:"id"`
	Status                 string `json:"status"`
	Outcome                string `json:"outcome"`
	AIProvider             string `json:"ai_provider"`
	CompanyPreset          string `json:"company_preset"`
	InterviewMode          string `json:"interview_mode"`
	FollowUpCount          int    `json:"follow_up_count"`
	SolvedWithoutFollowUps bool   `json:"solved_without_followups"`
	MistakeSummary         string `json:"mistake_summary,omitempty"`
	GoogleReadiness        string `json:"google_readiness,omitempty"`
	HireRecommendation     string `json:"hire_recommendation,omitempty"`
	Summary                string `json:"summary,omitempty"`
	CreatedAt              string `json:"created_at"`
}

type ProblemMemory struct {
	Attempts    []AttemptMemory      `json:"attempts"`
	FollowUps   []FollowUpQuestion   `json:"follow_ups"`
	Mistakes    []LearningNote       `json:"mistakes"`
	Weaknesses  []WeaknessSignal     `json:"weaknesses"`
	Whiteboards []WhiteboardArtifact `json:"whiteboards"`
}
