package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

type App struct {
	cfg       Config
	store     *Store
	runner    *CodeRunner
	codex     *CodexClient
	codeFiles *CodeFileManager
	router    http.Handler
}

func New(cfg Config) (*App, error) {
	store, err := NewStore(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	seed, err := LoadSeedProblems()
	if err != nil {
		return nil, err
	}
	if err := store.Seed(seed); err != nil {
		return nil, err
	}

	app := &App{
		cfg:       cfg,
		store:     store,
		runner:    NewCodeRunner(cfg.PythonBin),
		codex:     NewCodexClient(cfg),
		codeFiles: NewCodeFileManager(cfg.CodeWorkspaceDir, cfg.CodeWorkspacePublic),
	}
	app.router = app.routes()
	return app, nil
}

func (a *App) Listen() error {
	return http.ListenAndServe(":"+a.cfg.Port, a.router)
}

func (a *App) routes() http.Handler {
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/health", a.health)
	r.Get("/problems", a.listProblems)
	r.Get("/problems/{problemID}", a.getProblem)
	r.Post("/attempts", a.createAttempt)
	r.Get("/attempts/{attemptID}", a.getAttempt)
	r.Get("/problems/{problemID}/attempts", a.listAttemptsForProblem)
	r.Post("/attempts/{attemptID}/chat", a.chat)
	r.Get("/attempts/{attemptID}/chat", a.listChat)
	r.Get("/attempts/{attemptID}/code-file", a.getCodeFile)
	r.Put("/attempts/{attemptID}/code-file", a.updateCodeFile)
	r.Post("/attempts/{attemptID}/run", a.runCode)
	r.Post("/attempts/{attemptID}/review", a.review)
	r.Get("/daily", a.daily)
	r.Get("/review", a.reviewDashboard)
	return r
}

func (a *App) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *App) listProblems(w http.ResponseWriter, _ *http.Request) {
	items, err := a.store.ListProblems()
	respond(w, items, err)
}

func (a *App) getProblem(w http.ResponseWriter, r *http.Request) {
	problem, err := a.store.GetProblem(chi.URLParam(r, "problemID"))
	respond(w, problem, err)
}

func (a *App) createAttempt(w http.ResponseWriter, r *http.Request) {
	var req CreateAttemptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.ProblemID == "" {
		writeError(w, http.StatusBadRequest, "problem_id is required")
		return
	}
	attempt, err := a.store.CreateAttempt(req)
	if err == nil {
		a.codeFiles.Decorate(attempt)
		memory, _ := a.store.ProblemMemory(attempt.ProblemID)
		_, _ = a.store.EnsureInitialMessage(attempt.ID, initialInterviewMessage(memory))
	}
	respond(w, attempt, err)
}

func (a *App) getAttempt(w http.ResponseWriter, r *http.Request) {
	attempt, err := a.store.GetAttempt(chi.URLParam(r, "attemptID"))
	if err == nil {
		a.codeFiles.Decorate(attempt)
	}
	respond(w, attempt, err)
}

func (a *App) listAttemptsForProblem(w http.ResponseWriter, r *http.Request) {
	attempts, err := a.store.ListAttemptsForProblem(chi.URLParam(r, "problemID"))
	respond(w, attempts, err)
}

func (a *App) chat(w http.ResponseWriter, r *http.Request) {
	attemptID := chi.URLParam(r, "attemptID")
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.Message == "" {
		writeError(w, http.StatusBadRequest, "message is required")
		return
	}
	attempt, err := a.store.GetAttempt(attemptID)
	if err != nil {
		respond(w, nil, err)
		return
	}
	if synced, syncErr := a.readSyncedCode(*attempt); syncErr == nil {
		attempt.Code = synced
	}
	problem, err := a.store.GetProblem(attempt.ProblemID)
	if err != nil {
		respond(w, nil, err)
		return
	}
	if _, err := a.store.AddMessage(attemptID, "user", req.Message); err != nil {
		respond(w, nil, err)
		return
	}
	messages, _ := a.store.ListMessages(attemptID)
	memory, _ := a.store.ProblemMemory(attempt.ProblemID)
	reply, err := a.codex.Chat(r.Context(), *problem, *attempt, messages, req.Message, req.EnglishMode, memory)
	if err != nil {
		reply = "Codex CLIを使ったAI面接官を起動できませんでした。CODEX_CLI_PATH、ログイン状態、Docker利用時の ~/.codex マウントを確認してください。\n\n詳細: " + err.Error()
	}
	message, err := a.store.AddMessage(attemptID, "assistant", reply)
	if err == nil {
		_ = a.store.RecordFollowUps(attemptID, "chat", extractQuestions(reply))
	}
	respond(w, message, err)
}

func (a *App) listChat(w http.ResponseWriter, r *http.Request) {
	messages, err := a.store.ListMessages(chi.URLParam(r, "attemptID"))
	respond(w, messages, err)
}

func (a *App) getCodeFile(w http.ResponseWriter, r *http.Request) {
	attempt, err := a.store.GetAttempt(chi.URLParam(r, "attemptID"))
	if err != nil {
		respond(w, nil, err)
		return
	}
	info, err := a.codeFiles.Ensure(*attempt)
	respond(w, info, err)
}

func (a *App) updateCodeFile(w http.ResponseWriter, r *http.Request) {
	attemptID := chi.URLParam(r, "attemptID")
	var req CodeFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	attempt, err := a.store.GetAttempt(attemptID)
	if err != nil {
		respond(w, nil, err)
		return
	}
	info, err := a.codeFiles.Write(*attempt, req.Code)
	if err != nil {
		respond(w, nil, err)
		return
	}
	if _, err := a.store.UpdateAttemptCode(attemptID, req.Code); err != nil {
		respond(w, nil, err)
		return
	}
	writeJSON(w, http.StatusOK, info)
}

func (a *App) runCode(w http.ResponseWriter, r *http.Request) {
	attemptID := chi.URLParam(r, "attemptID")
	var req RunRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	attempt, err := a.store.GetAttempt(attemptID)
	if err != nil {
		respond(w, nil, err)
		return
	}
	problem, err := a.store.GetProblem(attempt.ProblemID)
	if err != nil {
		respond(w, nil, err)
		return
	}
	code := req.Code
	if code == "" {
		synced, syncErr := a.readSyncedCode(*attempt)
		if syncErr == nil {
			code = synced
		} else {
			code = attempt.Code
		}
	}
	attempt.Code = code
	_, _ = a.codeFiles.Write(*attempt, code)
	result := a.runner.Run(r.Context(), problem.ID, code, problem.TestCases)
	status := result.Status
	if status == "" {
		status = "failed"
	}
	updated, err := a.store.UpdateAttemptRun(attemptID, code, status, result)
	if err != nil {
		respond(w, nil, err)
		return
	}
	a.codeFiles.Decorate(updated)
	writeJSON(w, http.StatusOK, map[string]interface{}{"attempt": updated, "result": result})
}

func (a *App) review(w http.ResponseWriter, r *http.Request) {
	attemptID := chi.URLParam(r, "attemptID")
	var req ReviewRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	attempt, err := a.store.GetAttempt(attemptID)
	if err != nil {
		respond(w, nil, err)
		return
	}
	problem, err := a.store.GetProblem(attempt.ProblemID)
	if err != nil {
		respond(w, nil, err)
		return
	}
	code := req.Code
	if code == "" {
		synced, syncErr := a.readSyncedCode(*attempt)
		if syncErr == nil {
			code = synced
		} else {
			code = attempt.Code
		}
	}
	attempt.Code = code
	_, _ = a.codeFiles.Write(*attempt, code)
	memory, _ := a.store.ProblemMemory(attempt.ProblemID)
	review, _ := a.codex.Review(r.Context(), *problem, *attempt, code, memory)
	updated, err := a.store.UpdateAttemptReview(attemptID, code, review)
	if err != nil {
		respond(w, nil, err)
		return
	}
	a.codeFiles.Decorate(updated)
	writeJSON(w, http.StatusOK, map[string]interface{}{"attempt": updated, "review": review})
}

func (a *App) daily(w http.ResponseWriter, _ *http.Request) {
	daily, err := a.store.Daily()
	respond(w, daily, err)
}

func (a *App) reviewDashboard(w http.ResponseWriter, _ *http.Request) {
	dashboard, err := a.store.ReviewDashboard()
	respond(w, dashboard, err)
}

func respond(w http.ResponseWriter, value interface{}, err error) {
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, value)
}

func writeJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func (a *App) readSyncedCode(attempt Attempt) (string, error) {
	info, err := a.codeFiles.Ensure(attempt)
	if err != nil {
		return "", err
	}
	if _, err := a.store.UpdateAttemptCode(attempt.ID, info.Content); err != nil {
		return "", err
	}
	return info.Content, nil
}

func initialInterviewMessage(memory ProblemMemory) string {
	prefix := ""
	if len(memory.Attempts) > 1 || len(memory.Mistakes) > 0 || len(memory.FollowUps) > 0 {
		prefix = "この問題は過去の履歴も見ながら少し厳しめに確認します。前回のミスやフォローアップも踏まえます。\n\n"
	}
	return prefix + "まず実装に入る前に、どのように解くつもりかを説明してください。全探索の方針、より良い解法の見込み、使うデータ構造、気になるエッジケースを1つずつ短く述べてください。コードはまだ書かなくて大丈夫です。"
}

func extractQuestions(content string) []string {
	questions := []string{}
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(line, "-"))
		line = strings.TrimSpace(strings.TrimPrefix(line, "・"))
		if line == "" {
			continue
		}
		if strings.Contains(line, "?") || strings.Contains(line, "？") || strings.Contains(line, "ですか") || strings.Contains(line, "ますか") {
			questions = append(questions, line)
		}
	}
	return questions
}
