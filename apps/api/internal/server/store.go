package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func NewStore(databaseURL string) (*Store, error) {
	path := sqlitePath(databaseURL)
	if path != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil && filepath.Dir(path) != "." {
			return nil, err
		}
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	store := &Store{db: db}
	if err := store.migrate(); err != nil {
		return nil, err
	}
	return store, nil
}

func sqlitePath(databaseURL string) string {
	if strings.HasPrefix(databaseURL, "sqlite:///") {
		return strings.TrimPrefix(databaseURL, "sqlite:///")
	}
	if strings.HasPrefix(databaseURL, "sqlite://") {
		return strings.TrimPrefix(databaseURL, "sqlite://")
	}
	return databaseURL
}

func (s *Store) migrate() error {
	statements := []string{
		`PRAGMA foreign_keys = ON`,
		`CREATE TABLE IF NOT EXISTS problems (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			difficulty TEXT NOT NULL,
			pattern TEXT NOT NULL,
			tags TEXT NOT NULL,
			statement TEXT NOT NULL,
			examples TEXT NOT NULL,
			constraints_text TEXT NOT NULL,
			starter_code TEXT NOT NULL,
			test_cases TEXT NOT NULL,
			solution_explanation TEXT,
			source_url TEXT,
			list_name TEXT NOT NULL DEFAULT 'Arai60',
			order_index INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS attempts (
			id TEXT PRIMARY KEY,
			problem_id TEXT NOT NULL,
			language TEXT NOT NULL,
			code TEXT NOT NULL,
			status TEXT NOT NULL,
			outcome TEXT NOT NULL DEFAULT 'in_progress',
			company_preset TEXT NOT NULL DEFAULT 'google',
			interview_mode TEXT NOT NULL DEFAULT 'real',
			current_phase TEXT NOT NULL DEFAULT 'planning',
			time_limit_seconds INTEGER NOT NULL DEFAULT 2700,
			no_run INTEGER NOT NULL DEFAULT 1,
			no_autocomplete INTEGER NOT NULL DEFAULT 1,
			requires_plan INTEGER NOT NULL DEFAULT 1,
			follow_up_count INTEGER NOT NULL DEFAULT 0,
			solved_without_followups INTEGER NOT NULL DEFAULT 0,
			mistake_summary TEXT,
			test_result TEXT,
			ai_review TEXT,
			hints_used INTEGER NOT NULL DEFAULT 0,
			time_spent_seconds INTEGER,
			completed_at TEXT,
			created_at TEXT NOT NULL,
			FOREIGN KEY (problem_id) REFERENCES problems(id)
		)`,
		`CREATE TABLE IF NOT EXISTS chat_messages (
			id TEXT PRIMARY KEY,
			attempt_id TEXT NOT NULL,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at TEXT NOT NULL,
			FOREIGN KEY (attempt_id) REFERENCES attempts(id)
		)`,
		`CREATE TABLE IF NOT EXISTS learning_notes (
			id TEXT PRIMARY KEY,
			problem_id TEXT NOT NULL,
			mistake_type TEXT NOT NULL,
			note TEXT NOT NULL,
			next_review_date TEXT,
			created_at TEXT NOT NULL,
			FOREIGN KEY (problem_id) REFERENCES problems(id)
		)`,
		`CREATE TABLE IF NOT EXISTS follow_up_questions (
			id TEXT PRIMARY KEY,
			attempt_id TEXT NOT NULL,
			problem_id TEXT NOT NULL,
			question TEXT NOT NULL,
			source TEXT NOT NULL,
			answered INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			FOREIGN KEY (attempt_id) REFERENCES attempts(id),
			FOREIGN KEY (problem_id) REFERENCES problems(id)
		)`,
		`CREATE TABLE IF NOT EXISTS weakness_signals (
			id TEXT PRIMARY KEY,
			attempt_id TEXT NOT NULL,
			problem_id TEXT NOT NULL,
			category TEXT NOT NULL,
			signal TEXT NOT NULL,
			severity INTEGER NOT NULL DEFAULT 1,
			evidence TEXT NOT NULL,
			drill TEXT NOT NULL,
			created_at TEXT NOT NULL,
			FOREIGN KEY (attempt_id) REFERENCES attempts(id),
			FOREIGN KEY (problem_id) REFERENCES problems(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_attempts_problem_created ON attempts(problem_id, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_attempt_created ON chat_messages(attempt_id, created_at ASC)`,
		`CREATE INDEX IF NOT EXISTS idx_followups_problem_created ON follow_up_questions(problem_id, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_weakness_problem_created ON weakness_signals(problem_id, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_weakness_category_created ON weakness_signals(category, created_at DESC)`,
	}
	for _, statement := range statements {
		if _, err := s.db.Exec(statement); err != nil {
			return err
		}
	}
	for _, column := range []string{
		`ALTER TABLE problems ADD COLUMN source_url TEXT`,
		`ALTER TABLE problems ADD COLUMN list_name TEXT NOT NULL DEFAULT 'Arai60'`,
		`ALTER TABLE problems ADD COLUMN order_index INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE attempts ADD COLUMN outcome TEXT NOT NULL DEFAULT 'in_progress'`,
		`ALTER TABLE attempts ADD COLUMN company_preset TEXT NOT NULL DEFAULT 'google'`,
		`ALTER TABLE attempts ADD COLUMN interview_mode TEXT NOT NULL DEFAULT 'real'`,
		`ALTER TABLE attempts ADD COLUMN current_phase TEXT NOT NULL DEFAULT 'planning'`,
		`ALTER TABLE attempts ADD COLUMN time_limit_seconds INTEGER NOT NULL DEFAULT 2700`,
		`ALTER TABLE attempts ADD COLUMN no_run INTEGER NOT NULL DEFAULT 1`,
		`ALTER TABLE attempts ADD COLUMN no_autocomplete INTEGER NOT NULL DEFAULT 1`,
		`ALTER TABLE attempts ADD COLUMN requires_plan INTEGER NOT NULL DEFAULT 1`,
		`ALTER TABLE attempts ADD COLUMN follow_up_count INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE attempts ADD COLUMN solved_without_followups INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE attempts ADD COLUMN mistake_summary TEXT`,
		`ALTER TABLE attempts ADD COLUMN completed_at TEXT`,
	} {
		if _, err := s.db.Exec(column); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
			return err
		}
	}
	return nil
}

func (s *Store) Seed(problems []Problem) error {
	for _, problem := range problems {
		tags, _ := json.Marshal(problem.Tags)
		examples, _ := json.Marshal(problem.Examples)
		constraintsText, _ := json.Marshal(problem.Constraints)
		testCases, _ := json.Marshal(problem.TestCases)
		_, err := s.db.Exec(
			`INSERT INTO problems
			(id, title, difficulty, pattern, tags, statement, examples, constraints_text, starter_code, test_cases, solution_explanation, source_url, list_name, order_index, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(id) DO UPDATE SET
				title = excluded.title,
				difficulty = excluded.difficulty,
				pattern = excluded.pattern,
				tags = excluded.tags,
				statement = excluded.statement,
				examples = excluded.examples,
				constraints_text = excluded.constraints_text,
				starter_code = excluded.starter_code,
				test_cases = excluded.test_cases,
				solution_explanation = excluded.solution_explanation,
				source_url = excluded.source_url,
				list_name = excluded.list_name,
				order_index = excluded.order_index`,
			problem.ID,
			problem.Title,
			problem.Difficulty,
			problem.Pattern,
			string(tags),
			problem.Statement,
			string(examples),
			string(constraintsText),
			problem.StarterCode,
			string(testCases),
			nullableString(problem.SolutionExplanation),
			nullableString(problem.SourceURL),
			defaultString(problem.ListName, "Arai60"),
			problem.OrderIndex,
			problem.CreatedAt,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) ListProblems() ([]ProblemListItem, error) {
	rows, err := s.db.Query(
		`SELECT id, title, difficulty, pattern, tags, COALESCE(source_url, ''), COALESCE(list_name, ''), order_index
		FROM problems
		WHERE list_name = 'Arai60' AND order_index > 0
		ORDER BY order_index ASC, created_at ASC`,
	)
	if err != nil {
		return nil, err
	}

	items := []ProblemListItem{}
	for rows.Next() {
		var item ProblemListItem
		var tagsText string
		if err := rows.Scan(&item.ID, &item.Title, &item.Difficulty, &item.Pattern, &tagsText, &item.SourceURL, &item.ListName, &item.OrderIndex); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(tagsText), &item.Tags)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	for i := range items {
		s.applyProgress(&items[i])
	}
	return items, nil
}

func (s *Store) GetProblem(problemID string) (*Problem, error) {
	row := s.db.QueryRow(
		`SELECT id, title, difficulty, pattern, tags, statement, examples, constraints_text, starter_code,
			test_cases, COALESCE(solution_explanation, ''), COALESCE(source_url, ''), COALESCE(list_name, ''),
			order_index, created_at
		FROM problems WHERE id = ?`,
		problemID,
	)
	var p Problem
	var tagsText, examplesText, constraintsText, testCasesText string
	if err := row.Scan(
		&p.ID,
		&p.Title,
		&p.Difficulty,
		&p.Pattern,
		&tagsText,
		&p.Statement,
		&examplesText,
		&constraintsText,
		&p.StarterCode,
		&testCasesText,
		&p.SolutionExplanation,
		&p.SourceURL,
		&p.ListName,
		&p.OrderIndex,
		&p.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	_ = json.Unmarshal([]byte(tagsText), &p.Tags)
	_ = json.Unmarshal([]byte(examplesText), &p.Examples)
	_ = json.Unmarshal([]byte(constraintsText), &p.Constraints)
	_ = json.Unmarshal([]byte(testCasesText), &p.TestCases)
	item := ProblemListItem{ID: p.ID}
	s.applyProgress(&item)
	p.Status = item.Status
	p.MasteryStatus = item.MasteryStatus
	p.AttemptCount = item.AttemptCount
	p.FollowUpCount = item.FollowUpCount
	p.SolvedWithoutFollowUps = item.SolvedWithoutFollowUps
	p.LastAttemptedAt = item.LastAttemptedAt
	return &p, nil
}

func (s *Store) CreateAttempt(req CreateAttemptRequest) (*Attempt, error) {
	problem, err := s.GetProblem(req.ProblemID)
	if err != nil {
		return nil, err
	}
	language := req.Language
	if language == "" {
		language = "python"
	}
	code := req.Code
	if code == "" {
		code = problem.StarterCode
	}
	settings := defaultsForInterview(req.CompanyPreset, req.InterviewMode)
	attempt := Attempt{
		ID:               newID(),
		ProblemID:        req.ProblemID,
		Language:         language,
		Code:             code,
		Status:           "in_progress",
		Outcome:          "in_progress",
		CompanyPreset:    settings.CompanyPreset,
		InterviewMode:    settings.InterviewMode,
		CurrentPhase:     settings.CurrentPhase,
		TimeLimitSeconds: settings.TimeLimitSeconds,
		NoRun:            settings.NoRun,
		NoAutocomplete:   settings.NoAutocomplete,
		RequiresPlan:     settings.RequiresPlan,
		CreatedAt:        now(),
	}
	_, err = s.db.Exec(
		`INSERT INTO attempts
		(id, problem_id, language, code, status, outcome, company_preset, interview_mode, current_phase,
			time_limit_seconds, no_run, no_autocomplete, requires_plan, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		attempt.ID,
		attempt.ProblemID,
		attempt.Language,
		attempt.Code,
		attempt.Status,
		attempt.Outcome,
		attempt.CompanyPreset,
		attempt.InterviewMode,
		attempt.CurrentPhase,
		attempt.TimeLimitSeconds,
		boolInt(attempt.NoRun),
		boolInt(attempt.NoAutocomplete),
		boolInt(attempt.RequiresPlan),
		attempt.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &attempt, nil
}

func (s *Store) GetAttempt(attemptID string) (*Attempt, error) {
	row := s.db.QueryRow(
		`SELECT a.id, a.problem_id, p.title, a.language, a.code, a.status, a.outcome,
			a.company_preset, a.interview_mode, a.current_phase, a.time_limit_seconds,
			a.no_run, a.no_autocomplete, a.requires_plan, a.follow_up_count,
			a.solved_without_followups, COALESCE(a.mistake_summary, ''), a.test_result, a.ai_review,
			a.hints_used, a.time_spent_seconds, a.completed_at, a.created_at
		FROM attempts a
		JOIN problems p ON p.id = a.problem_id
		WHERE a.id = ?`,
		attemptID,
	)
	attempt, err := scanAttempt(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return attempt, nil
}

func (s *Store) ListAttemptsForProblem(problemID string) ([]Attempt, error) {
	rows, err := s.db.Query(
		`SELECT a.id, a.problem_id, p.title, a.language, a.code, a.status, a.outcome,
			a.company_preset, a.interview_mode, a.current_phase, a.time_limit_seconds,
			a.no_run, a.no_autocomplete, a.requires_plan, a.follow_up_count,
			a.solved_without_followups, COALESCE(a.mistake_summary, ''), a.test_result, a.ai_review,
			a.hints_used, a.time_spent_seconds, a.completed_at, a.created_at
		FROM attempts a
		JOIN problems p ON p.id = a.problem_id
		WHERE a.problem_id = ?
		ORDER BY a.created_at DESC`,
		problemID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAttempts(rows)
}

func (s *Store) RecentAttempts(limit int) ([]Attempt, error) {
	rows, err := s.db.Query(
		`SELECT a.id, a.problem_id, p.title, a.language, a.code, a.status, a.outcome,
			a.company_preset, a.interview_mode, a.current_phase, a.time_limit_seconds,
			a.no_run, a.no_autocomplete, a.requires_plan, a.follow_up_count,
			a.solved_without_followups, COALESCE(a.mistake_summary, ''), a.test_result, a.ai_review,
			a.hints_used, a.time_spent_seconds, a.completed_at, a.created_at
		FROM attempts a
		JOIN problems p ON p.id = a.problem_id
		ORDER BY a.created_at DESC
		LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAttempts(rows)
}

func (s *Store) AddMessage(attemptID, role, content string) (*ChatMessage, error) {
	message := ChatMessage{
		ID:        newID(),
		AttemptID: attemptID,
		Role:      role,
		Content:   content,
		CreatedAt: now(),
	}
	_, err := s.db.Exec(
		`INSERT INTO chat_messages (id, attempt_id, role, content, created_at) VALUES (?, ?, ?, ?, ?)`,
		message.ID, message.AttemptID, message.Role, message.Content, message.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (s *Store) EnsureInitialMessage(attemptID, content string) (*ChatMessage, error) {
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM chat_messages WHERE attempt_id = ?`, attemptID).Scan(&count); err != nil {
		return nil, err
	}
	if count > 0 {
		messages, err := s.ListMessages(attemptID)
		if err != nil {
			return nil, err
		}
		if len(messages) == 0 {
			return nil, ErrNotFound
		}
		return &messages[0], nil
	}
	return s.AddMessage(attemptID, "assistant", content)
}

func (s *Store) ListMessages(attemptID string) ([]ChatMessage, error) {
	rows, err := s.db.Query(
		`SELECT id, attempt_id, role, content, created_at
		FROM chat_messages WHERE attempt_id = ? ORDER BY created_at ASC`,
		attemptID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := []ChatMessage{}
	for rows.Next() {
		var message ChatMessage
		if err := rows.Scan(&message.ID, &message.AttemptID, &message.Role, &message.Content, &message.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	return messages, rows.Err()
}

func (s *Store) RecordFollowUps(attemptID, source string, questions []string) error {
	attempt, err := s.GetAttempt(attemptID)
	if err != nil {
		return err
	}
	inserted := 0
	for _, question := range compactStrings(questions) {
		_, err := s.db.Exec(
			`INSERT INTO follow_up_questions (id, attempt_id, problem_id, question, source, answered, created_at)
			VALUES (?, ?, ?, ?, ?, 0, ?)`,
			newID(), attempt.ID, attempt.ProblemID, question, source, now(),
		)
		if err != nil {
			return err
		}
		inserted++
	}
	if inserted == 0 {
		return nil
	}
	_, err = s.db.Exec(
		`UPDATE attempts
		SET follow_up_count = (SELECT COUNT(*) FROM follow_up_questions WHERE attempt_id = ?)
		WHERE id = ?`,
		attemptID, attemptID,
	)
	return err
}

func (s *Store) recentFollowUps(limit int) ([]FollowUpQuestion, error) {
	rows, err := s.db.Query(
		`SELECT id, attempt_id, problem_id, question, source, answered, created_at
		FROM follow_up_questions ORDER BY created_at DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFollowUps(rows)
}

func (s *Store) followUpsForProblem(problemID string, limit int) ([]FollowUpQuestion, error) {
	rows, err := s.db.Query(
		`SELECT id, attempt_id, problem_id, question, source, answered, created_at
		FROM follow_up_questions WHERE problem_id = ? ORDER BY created_at DESC LIMIT ?`,
		problemID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFollowUps(rows)
}

func scanFollowUps(rows *sql.Rows) ([]FollowUpQuestion, error) {
	items := []FollowUpQuestion{}
	for rows.Next() {
		var item FollowUpQuestion
		var answered int
		if err := rows.Scan(&item.ID, &item.AttemptID, &item.ProblemID, &item.Question, &item.Source, &answered, &item.CreatedAt); err != nil {
			return nil, err
		}
		item.Answered = answered == 1
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) RecordWeaknesses(attemptID string, weaknesses []WeaknessSignal) error {
	attempt, err := s.GetAttempt(attemptID)
	if err != nil {
		return err
	}
	for _, weakness := range weaknesses {
		if strings.TrimSpace(weakness.Category) == "" || strings.TrimSpace(weakness.Signal) == "" {
			continue
		}
		severity := weakness.Severity
		if severity < 1 {
			severity = 1
		}
		if severity > 5 {
			severity = 5
		}
		_, err := s.db.Exec(
			`INSERT INTO weakness_signals
			(id, attempt_id, problem_id, category, signal, severity, evidence, drill, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			newID(),
			attempt.ID,
			attempt.ProblemID,
			strings.TrimSpace(weakness.Category),
			strings.TrimSpace(weakness.Signal),
			severity,
			strings.TrimSpace(weakness.Evidence),
			strings.TrimSpace(weakness.Drill),
			now(),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) weaknessGraph(limit int) ([]WeaknessTrend, error) {
	rows, err := s.db.Query(
		`SELECT category, COUNT(*) as count, CAST(ROUND(AVG(severity)) AS INTEGER) as average_severity,
			MAX(created_at) as last_seen_at,
			COALESCE((SELECT drill FROM weakness_signals w2 WHERE w2.category = weakness_signals.category ORDER BY created_at DESC LIMIT 1), '')
		FROM weakness_signals
		GROUP BY category
		ORDER BY average_severity DESC, count DESC, last_seen_at DESC
		LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	trends := []WeaknessTrend{}
	for rows.Next() {
		var trend WeaknessTrend
		if err := rows.Scan(&trend.Category, &trend.Count, &trend.AverageSeverity, &trend.LastSeenAt, &trend.SuggestedDrill); err != nil {
			return nil, err
		}
		trends = append(trends, trend)
	}
	return trends, rows.Err()
}

func (s *Store) weaknessesForProblem(problemID string, limit int) ([]WeaknessSignal, error) {
	rows, err := s.db.Query(
		`SELECT id, attempt_id, problem_id, category, signal, severity, evidence, drill, created_at
		FROM weakness_signals WHERE problem_id = ? ORDER BY created_at DESC LIMIT ?`,
		problemID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanWeaknesses(rows)
}

func scanWeaknesses(rows *sql.Rows) ([]WeaknessSignal, error) {
	items := []WeaknessSignal{}
	for rows.Next() {
		var item WeaknessSignal
		if err := rows.Scan(
			&item.ID,
			&item.AttemptID,
			&item.ProblemID,
			&item.Category,
			&item.Signal,
			&item.Severity,
			&item.Evidence,
			&item.Drill,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) UpdateAttemptCode(attemptID, code string) (*Attempt, error) {
	_, err := s.db.Exec(`UPDATE attempts SET code = ? WHERE id = ?`, code, attemptID)
	if err != nil {
		return nil, err
	}
	return s.GetAttempt(attemptID)
}

func (s *Store) UpdateAttemptRun(attemptID, code, status string, result RunResult) (*Attempt, error) {
	payload, _ := json.Marshal(result)
	outcome := "in_progress"
	completedAt := interface{}(nil)
	if status == "passed" {
		outcome = "local_tests_passed"
		nowValue := now()
		completedAt = nowValue
	} else if status == "failed" || status == "runtime_error" {
		outcome = "needs_review"
	}
	_, err := s.db.Exec(
		`UPDATE attempts SET code = ?, status = ?, outcome = ?, test_result = ?, completed_at = COALESCE(completed_at, ?) WHERE id = ?`,
		code, status, outcome, string(payload), completedAt, attemptID,
	)
	if err != nil {
		return nil, err
	}
	return s.GetAttempt(attemptID)
}

func (s *Store) UpdateAttemptReview(attemptID, code string, review ReviewResponse) (*Attempt, error) {
	currentAttempt, err := s.GetAttempt(attemptID)
	if err != nil {
		return nil, err
	}
	payload, _ := json.Marshal(review)
	status := "needs_review"
	outcome := "needs_review"
	solvedWithoutFollowUps := 0
	completedAt := interface{}(nil)
	mistakeSummary := strings.Join(compactStrings(review.MistakesToRemember), " / ")
	followUpCount := currentAttempt.FollowUpCount + len(compactStrings(review.ComplexityQuestions)) + len(compactStrings(review.FollowUpQuestions))
	if review.IsCorrect {
		status = "passed"
		if followUpCount == 0 {
			outcome = "solved_clean"
			solvedWithoutFollowUps = 1
		} else {
			outcome = "solved_with_followups"
		}
		nowValue := now()
		completedAt = nowValue
	}
	_, err = s.db.Exec(
		`UPDATE attempts
		SET code = ?, status = ?, outcome = ?, solved_without_followups = ?,
			mistake_summary = ?, ai_review = ?, completed_at = COALESCE(completed_at, ?)
		WHERE id = ?`,
		code, status, outcome, solvedWithoutFollowUps, nullableString(mistakeSummary), string(payload), completedAt, attemptID,
	)
	if err != nil {
		return nil, err
	}
	if err := s.RecordFollowUps(attemptID, "review_complexity", review.ComplexityQuestions); err != nil {
		return nil, err
	}
	if err := s.RecordFollowUps(attemptID, "review_followup", review.FollowUpQuestions); err != nil {
		return nil, err
	}
	if err := s.RecordWeaknesses(attemptID, review.DetectedWeaknesses); err != nil {
		return nil, err
	}
	for _, note := range review.MistakesToRemember {
		if strings.TrimSpace(note) == "" {
			continue
		}
		attempt, err := s.GetAttempt(attemptID)
		if err != nil {
			return nil, err
		}
		_, err = s.db.Exec(
			`INSERT INTO learning_notes (id, problem_id, mistake_type, note, next_review_date, created_at)
			VALUES (?, ?, ?, ?, ?, ?)`,
			newID(), attempt.ProblemID, "ai_review", note, time.Now().AddDate(0, 0, 3).Format("2006-01-02"), now(),
		)
		if err != nil {
			return nil, err
		}
	}
	return s.GetAttempt(attemptID)
}

func (s *Store) ReviewDashboard() (*ReviewDashboard, error) {
	attempts, err := s.RecentAttempts(8)
	if err != nil {
		return nil, err
	}
	notes, err := s.learningNotes(10)
	if err != nil {
		return nil, err
	}
	followUps, err := s.recentFollowUps(12)
	if err != nil {
		return nil, err
	}
	weakPatterns, err := s.weakPatterns()
	if err != nil {
		return nil, err
	}
	weaknessGraph, err := s.weaknessGraph(8)
	if err != nil {
		return nil, err
	}
	problems, err := s.ListProblems()
	if err != nil {
		return nil, err
	}
	today := []ProblemListItem{}
	for _, problem := range problems {
		if problem.Status == "needs_review" || problem.Status == "not_started" {
			today = append(today, problem)
		}
		if len(today) == 4 {
			break
		}
	}
	return &ReviewDashboard{
		RecentAttempts:     attempts,
		RecentFollowUps:    followUps,
		MistakesToRemember: notes,
		WeakPatterns:       weakPatterns,
		WeaknessGraph:      weaknessGraph,
		ProblemsForToday:   today,
	}, nil
}

func (s *Store) ProblemMemory(problemID string) (ProblemMemory, error) {
	attemptRows, err := s.db.Query(
		`SELECT id, status, outcome, company_preset, interview_mode, follow_up_count, solved_without_followups,
			COALESCE(mistake_summary, ''), COALESCE(ai_review, ''), created_at
		FROM attempts WHERE problem_id = ? ORDER BY created_at DESC LIMIT 6`,
		problemID,
	)
	if err != nil {
		return ProblemMemory{}, err
	}
	defer attemptRows.Close()

	memory := ProblemMemory{Attempts: []AttemptMemory{}, FollowUps: []FollowUpQuestion{}, Mistakes: []LearningNote{}, Weaknesses: []WeaknessSignal{}}
	for attemptRows.Next() {
		var item AttemptMemory
		var solvedWithout int
		var reviewText string
		if err := attemptRows.Scan(
			&item.ID,
			&item.Status,
			&item.Outcome,
			&item.CompanyPreset,
			&item.InterviewMode,
			&item.FollowUpCount,
			&solvedWithout,
			&item.MistakeSummary,
			&reviewText,
			&item.CreatedAt,
		); err != nil {
			return ProblemMemory{}, err
		}
		item.SolvedWithoutFollowUps = solvedWithout == 1
		if reviewText != "" {
			var review ReviewResponse
			if err := json.Unmarshal([]byte(reviewText), &review); err == nil {
				item.GoogleReadiness = review.GoogleReadiness
				item.HireRecommendation = review.HireRecommendation
				item.Summary = review.Summary
			}
		}
		memory.Attempts = append(memory.Attempts, item)
	}
	if err := attemptRows.Err(); err != nil {
		return ProblemMemory{}, err
	}

	followUps, err := s.followUpsForProblem(problemID, 12)
	if err != nil {
		return ProblemMemory{}, err
	}
	memory.FollowUps = followUps

	weaknesses, err := s.weaknessesForProblem(problemID, 10)
	if err != nil {
		return ProblemMemory{}, err
	}
	memory.Weaknesses = weaknesses

	noteRows, err := s.db.Query(
		`SELECT n.id, n.problem_id, p.title, n.mistake_type, n.note, n.next_review_date, n.created_at
		FROM learning_notes n
		JOIN problems p ON p.id = n.problem_id
		WHERE n.problem_id = ?
		ORDER BY n.created_at DESC LIMIT 8`,
		problemID,
	)
	if err != nil {
		return ProblemMemory{}, err
	}
	defer noteRows.Close()
	for noteRows.Next() {
		var note LearningNote
		var next sql.NullString
		if err := noteRows.Scan(&note.ID, &note.ProblemID, &note.ProblemTitle, &note.MistakeType, &note.Note, &next, &note.CreatedAt); err != nil {
			return ProblemMemory{}, err
		}
		if next.Valid {
			note.NextReviewDate = &next.String
		}
		memory.Mistakes = append(memory.Mistakes, note)
	}
	return memory, noteRows.Err()
}

func (s *Store) Daily() (*DailyResponse, error) {
	problems, err := s.ListProblems()
	if err != nil {
		return nil, err
	}
	if len(problems) == 0 {
		return nil, ErrNotFound
	}
	chosen := problems[0]
	for _, problem := range problems {
		if problem.Status == "needs_review" {
			chosen = problem
			break
		}
		if problem.Status == "not_started" && chosen.Status != "needs_review" {
			chosen = problem
			break
		}
	}
	reason := "未着手の典型パターンなので、今日のウォームアップに向いています。"
	if chosen.Status == "needs_review" {
		reason = "直近の提出で詰まりが残っているため、忘れる前に復習すると効果が高いです。"
	}
	return &DailyResponse{
		Problem: chosen,
		Reason:  reason,
		Focus: []string{
			"まず全探索を説明してから、どこを削れるか言語化する",
			"境界条件をコード前に3つ挙げる",
			"最後に時間計算量と空間計算量を面接官へ説明する",
		},
		RecommendedURL: "/problems/" + chosen.ID,
	}, nil
}

func (s *Store) applyProgress(item *ProblemListItem) {
	var attempts int
	var passedAttempts int
	var firstStatus string
	var firstFollowUps int
	var firstSolvedWithout int
	var latestStatus string
	var latestCreated string
	var followUps int
	err := s.db.QueryRow(
		`SELECT
			COUNT(*),
			COALESCE(SUM(CASE WHEN status = 'passed' THEN 1 ELSE 0 END), 0),
			COALESCE((SELECT status FROM attempts WHERE problem_id = ? ORDER BY created_at ASC LIMIT 1), ''),
			COALESCE((SELECT follow_up_count FROM attempts WHERE problem_id = ? ORDER BY created_at ASC LIMIT 1), 0),
			COALESCE((SELECT solved_without_followups FROM attempts WHERE problem_id = ? ORDER BY created_at ASC LIMIT 1), 0),
			COALESCE((SELECT status FROM attempts WHERE problem_id = ? ORDER BY created_at DESC LIMIT 1), ''),
			COALESCE((SELECT created_at FROM attempts WHERE problem_id = ? ORDER BY created_at DESC LIMIT 1), ''),
			COALESCE((SELECT COUNT(*) FROM follow_up_questions WHERE problem_id = ?), 0)
		FROM attempts WHERE problem_id = ?`,
		item.ID, item.ID, item.ID, item.ID, item.ID, item.ID, item.ID,
	).Scan(&attempts, &passedAttempts, &firstStatus, &firstFollowUps, &firstSolvedWithout, &latestStatus, &latestCreated, &followUps)
	if err != nil || attempts == 0 {
		item.Status = "not_started"
		item.MasteryStatus = "not_started"
		return
	}
	item.AttemptCount = attempts
	item.FollowUpCount = followUps
	if latestCreated != "" {
		item.LastAttemptedAt = &latestCreated
	}
	item.SolvedWithoutFollowUps = firstSolvedWithout == 1
	if passedAttempts > 0 {
		item.Status = "solved"
		if firstStatus == "passed" && firstSolvedWithout == 1 && firstFollowUps == 0 {
			item.MasteryStatus = "first_try_clean"
		} else if firstStatus == "passed" {
			item.MasteryStatus = "first_try_with_followups"
		} else {
			item.MasteryStatus = "solved_after_retry"
		}
		return
	}
	switch latestStatus {
	case "failed", "runtime_error", "needs_review":
		item.Status = "needs_review"
		item.MasteryStatus = "needs_review"
	default:
		item.Status = "in_progress"
		item.MasteryStatus = "in_progress"
	}
}

func (s *Store) learningNotes(limit int) ([]LearningNote, error) {
	rows, err := s.db.Query(
		`SELECT n.id, n.problem_id, p.title, n.mistake_type, n.note, n.next_review_date, n.created_at
		FROM learning_notes n
		JOIN problems p ON p.id = n.problem_id
		ORDER BY n.created_at DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notes := []LearningNote{}
	for rows.Next() {
		var note LearningNote
		var next sql.NullString
		if err := rows.Scan(&note.ID, &note.ProblemID, &note.ProblemTitle, &note.MistakeType, &note.Note, &next, &note.CreatedAt); err != nil {
			return nil, err
		}
		if next.Valid {
			note.NextReviewDate = &next.String
		}
		notes = append(notes, note)
	}
	return notes, rows.Err()
}

func (s *Store) weakPatterns() ([]WeakPattern, error) {
	rows, err := s.db.Query(
		`SELECT p.pattern, COUNT(*) as count
		FROM attempts a
		JOIN problems p ON p.id = a.problem_id
		WHERE a.status != 'passed'
		GROUP BY p.pattern
		ORDER BY count DESC LIMIT 5`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	patterns := []WeakPattern{}
	for rows.Next() {
		var pattern WeakPattern
		if err := rows.Scan(&pattern.Pattern, &pattern.Count); err != nil {
			return nil, err
		}
		patterns = append(patterns, pattern)
	}
	return patterns, rows.Err()
}

type attemptScanner interface {
	Scan(dest ...interface{}) error
}

func scanAttempt(row attemptScanner) (*Attempt, error) {
	var attempt Attempt
	var testResultText, reviewText, completedAt sql.NullString
	var timeSpent sql.NullInt64
	var noRun, noAutocomplete, requiresPlan, solvedWithout int
	if err := row.Scan(
		&attempt.ID,
		&attempt.ProblemID,
		&attempt.ProblemTitle,
		&attempt.Language,
		&attempt.Code,
		&attempt.Status,
		&attempt.Outcome,
		&attempt.CompanyPreset,
		&attempt.InterviewMode,
		&attempt.CurrentPhase,
		&attempt.TimeLimitSeconds,
		&noRun,
		&noAutocomplete,
		&requiresPlan,
		&attempt.FollowUpCount,
		&solvedWithout,
		&attempt.MistakeSummary,
		&testResultText,
		&reviewText,
		&attempt.HintsUsed,
		&timeSpent,
		&completedAt,
		&attempt.CreatedAt,
	); err != nil {
		return nil, err
	}
	attempt.NoRun = noRun == 1
	attempt.NoAutocomplete = noAutocomplete == 1
	attempt.RequiresPlan = requiresPlan == 1
	attempt.SolvedWithoutFollowUps = solvedWithout == 1
	if testResultText.Valid {
		var result RunResult
		if err := json.Unmarshal([]byte(testResultText.String), &result); err == nil {
			attempt.TestResult = &result
		}
	}
	if reviewText.Valid {
		var review ReviewResponse
		if err := json.Unmarshal([]byte(reviewText.String), &review); err == nil {
			attempt.AIReview = &review
		}
	}
	if timeSpent.Valid {
		value := int(timeSpent.Int64)
		attempt.TimeSpentSeconds = &value
	}
	if completedAt.Valid {
		attempt.CompletedAt = &completedAt.String
	}
	return &attempt, nil
}

func scanAttempts(rows *sql.Rows) ([]Attempt, error) {
	attempts := []Attempt{}
	for rows.Next() {
		attempt, err := scanAttempt(rows)
		if err != nil {
			return nil, err
		}
		attempts = append(attempts, *attempt)
	}
	return attempts, rows.Err()
}

func nullableString(value string) interface{} {
	if value == "" {
		return nil
	}
	return value
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func compactStrings(values []string) []string {
	result := []string{}
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}

var ErrNotFound = errors.New("not found")

func newID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func now() string {
	return time.Now().UTC().Format(time.RFC3339)
}
