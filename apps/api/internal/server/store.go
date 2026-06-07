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
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS attempts (
			id TEXT PRIMARY KEY,
			problem_id TEXT NOT NULL,
			language TEXT NOT NULL,
			code TEXT NOT NULL,
			status TEXT NOT NULL,
			test_result TEXT,
			ai_review TEXT,
			hints_used INTEGER NOT NULL DEFAULT 0,
			time_spent_seconds INTEGER,
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
		`CREATE INDEX IF NOT EXISTS idx_attempts_problem_created ON attempts(problem_id, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_attempt_created ON chat_messages(attempt_id, created_at ASC)`,
	}
	for _, statement := range statements {
		if _, err := s.db.Exec(statement); err != nil {
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
			`INSERT OR IGNORE INTO problems
			(id, title, difficulty, pattern, tags, statement, examples, constraints_text, starter_code, test_cases, solution_explanation, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
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
			problem.CreatedAt,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) ListProblems() ([]ProblemListItem, error) {
	rows, err := s.db.Query(`SELECT id, title, difficulty, pattern, tags FROM problems ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}

	items := []ProblemListItem{}
	for rows.Next() {
		var item ProblemListItem
		var tagsText string
		if err := rows.Scan(&item.ID, &item.Title, &item.Difficulty, &item.Pattern, &tagsText); err != nil {
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
		items[i].Status, items[i].LastAttemptedAt = s.problemStatus(items[i].ID)
	}
	return items, nil
}

func (s *Store) GetProblem(problemID string) (*Problem, error) {
	row := s.db.QueryRow(
		`SELECT id, title, difficulty, pattern, tags, statement, examples, constraints_text, starter_code, test_cases, COALESCE(solution_explanation, ''), created_at
		FROM problems WHERE id = ?`,
		problemID,
	)
	var p Problem
	var tagsText, examplesText, constraintsText, testCasesText string
	if err := row.Scan(&p.ID, &p.Title, &p.Difficulty, &p.Pattern, &tagsText, &p.Statement, &examplesText, &constraintsText, &p.StarterCode, &testCasesText, &p.SolutionExplanation, &p.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	_ = json.Unmarshal([]byte(tagsText), &p.Tags)
	_ = json.Unmarshal([]byte(examplesText), &p.Examples)
	_ = json.Unmarshal([]byte(constraintsText), &p.Constraints)
	_ = json.Unmarshal([]byte(testCasesText), &p.TestCases)
	p.Status, p.LastAttemptedAt = s.problemStatus(p.ID)
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
	attempt := Attempt{
		ID:        newID(),
		ProblemID: req.ProblemID,
		Language:  language,
		Code:      code,
		Status:    "in_progress",
		CreatedAt: now(),
	}
	_, err = s.db.Exec(
		`INSERT INTO attempts (id, problem_id, language, code, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		attempt.ID, attempt.ProblemID, attempt.Language, attempt.Code, attempt.Status, attempt.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &attempt, nil
}

func (s *Store) GetAttempt(attemptID string) (*Attempt, error) {
	row := s.db.QueryRow(
		`SELECT a.id, a.problem_id, p.title, a.language, a.code, a.status, a.test_result, a.ai_review,
			a.hints_used, a.time_spent_seconds, a.created_at
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
		`SELECT a.id, a.problem_id, p.title, a.language, a.code, a.status, a.test_result, a.ai_review,
			a.hints_used, a.time_spent_seconds, a.created_at
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
		`SELECT a.id, a.problem_id, p.title, a.language, a.code, a.status, a.test_result, a.ai_review,
			a.hints_used, a.time_spent_seconds, a.created_at
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

func (s *Store) UpdateAttemptRun(attemptID, code, status string, result RunResult) (*Attempt, error) {
	payload, _ := json.Marshal(result)
	_, err := s.db.Exec(
		`UPDATE attempts SET code = ?, status = ?, test_result = ? WHERE id = ?`,
		code, status, string(payload), attemptID,
	)
	if err != nil {
		return nil, err
	}
	return s.GetAttempt(attemptID)
}

func (s *Store) UpdateAttemptReview(attemptID, code string, review ReviewResponse) (*Attempt, error) {
	payload, _ := json.Marshal(review)
	_, err := s.db.Exec(
		`UPDATE attempts SET code = ?, ai_review = ? WHERE id = ?`,
		code, string(payload), attemptID,
	)
	if err != nil {
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
	weakPatterns, err := s.weakPatterns()
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
		MistakesToRemember: notes,
		WeakPatterns:       weakPatterns,
		ProblemsForToday:   today,
	}, nil
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

func (s *Store) problemStatus(problemID string) (string, *string) {
	var status string
	var createdAt string
	err := s.db.QueryRow(
		`SELECT status, created_at FROM attempts WHERE problem_id = ? ORDER BY created_at DESC LIMIT 1`,
		problemID,
	).Scan(&status, &createdAt)
	if err != nil {
		return "not_started", nil
	}
	switch status {
	case "passed":
		return "solved", &createdAt
	case "failed", "runtime_error":
		return "needs_review", &createdAt
	default:
		return "in_progress", &createdAt
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
	var testResultText, reviewText sql.NullString
	var timeSpent sql.NullInt64
	if err := row.Scan(
		&attempt.ID,
		&attempt.ProblemID,
		&attempt.ProblemTitle,
		&attempt.Language,
		&attempt.Code,
		&attempt.Status,
		&testResultText,
		&reviewText,
		&attempt.HintsUsed,
		&timeSpent,
		&attempt.CreatedAt,
	); err != nil {
		return nil, err
	}
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

var ErrNotFound = errors.New("not found")

func newID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func now() string {
	return time.Now().UTC().Format(time.RFC3339)
}
