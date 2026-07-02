package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

type App struct {
	cfg           Config
	store         *Store
	runner        *CodeRunner
	ai            InterviewAI
	codeFiles     *CodeFileManager
	sourceFetcher ProblemSourceProvider
	router        http.Handler
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
		cfg:           cfg,
		store:         store,
		runner:        NewCodeRunner(cfg.PythonBin),
		ai:            NewAIService(cfg),
		codeFiles:     NewCodeFileManager(cfg.CodeWorkspaceDir, cfg.CodeWorkspacePublic),
		sourceFetcher: NewProblemSourceFetcher(),
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
	r.Get("/problems/{problemID}/official", a.getOfficialProblem)
	r.Post("/attempts", a.createAttempt)
	r.Get("/attempts/{attemptID}", a.getAttempt)
	r.Get("/problems/{problemID}/attempts", a.listAttemptsForProblem)
	r.Post("/attempts/{attemptID}/chat", a.chat)
	r.Get("/attempts/{attemptID}/chat", a.listChat)
	r.Post("/attempts/{attemptID}/nudge", a.nudge)
	r.Get("/attempts/{attemptID}/whiteboards", a.listWhiteboards)
	r.Post("/attempts/{attemptID}/artifact-reply", a.artifactReply)
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

func (a *App) getOfficialProblem(w http.ResponseWriter, r *http.Request) {
	problem, err := a.store.GetProblem(chi.URLParam(r, "problemID"))
	if err != nil {
		respond(w, nil, err)
		return
	}
	if cached, cacheErr := a.store.GetOfficialContent(problem.ID); cacheErr == nil {
		respond(w, cached, nil)
		return
	}
	content, err := a.sourceFetcher.Fetch(r.Context(), *problem)
	if err == nil && content != nil {
		_ = a.store.SaveOfficialContent(problem.ID, *content)
	}
	respond(w, content, err)
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
		if req.Reset {
			// Reset semantics: overwrite the shared per-problem file with the
			// starter code so a fresh attempt discards prior on-disk edits.
			_, _ = a.codeFiles.Write(*attempt, attempt.Code)
		}
		a.codeFiles.Decorate(attempt)
		memory, _ := a.store.ProblemMemory(attempt.ProblemID)
		_, _ = a.store.EnsureInitialMessage(attempt.ID, initialInterviewMessage(*attempt, memory))
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
	message, phase, err := a.completeInterviewerTurn(r.Context(), attempt, problem, req.Message, req.EnglishMode)
	if err != nil {
		respond(w, nil, err)
		return
	}
	writeJSON(w, http.StatusOK, ChatResponse{Message: *message, CurrentPhase: phase})
}

// completeInterviewerTurn runs one interviewer turn: it asks the AI, advances the
// phase, persists the assistant message (as an artifact request when one is
// present), and records follow-ups. A CLI failure degrades to a plain text turn
// so the interview never stalls.
func (a *App) completeInterviewerTurn(ctx context.Context, attempt *Attempt, problem *Problem, userMessage string, englishMode bool) (*ChatMessage, string, error) {
	messages, _ := a.store.ListMessages(attempt.ID)
	memory, _ := a.store.ProblemMemory(attempt.ProblemID)
	turn, err := a.ai.Chat(ctx, *problem, *attempt, messages, userMessage, englishMode, memory)
	if err != nil {
		reply := fmt.Sprintf("%s CLIを使ったAI面接官を起動できませんでした。CLI path、ログイン状態、Docker利用時のcredentialマウントを確認してください。\n\n詳細: %s", aiProviderLabel(attempt.AIProvider), err.Error())
		turn = InterviewerTurn{Reply: reply, Phase: attempt.CurrentPhase, ArtifactRequest: nil}
	}

	phase := advancePhase(attempt.CurrentPhase, turn.Phase)
	if phase != attempt.CurrentPhase {
		if updated, phaseErr := a.store.UpdateAttemptPhase(attempt.ID, phase); phaseErr == nil {
			attempt = updated
		}
	}

	kind := "text"
	var payload *ChatMessagePayload
	if turn.ArtifactRequest != nil {
		kind = "artifact_request"
		payload = &ChatMessagePayload{ArtifactRequest: turn.ArtifactRequest}
	}
	message, err := a.store.AddStructuredMessage(attempt.ID, "assistant", turn.Reply, kind, payload)
	if err != nil {
		return nil, "", err
	}
	_ = a.store.RecordFollowUps(attempt.ID, "chat", extractQuestions(turn.Reply))
	if turn.ArtifactRequest != nil {
		_ = a.store.RecordFollowUps(attempt.ID, "artifact_request", []string{turn.ArtifactRequest.Instructions})
	}
	return message, phase, nil
}

func (a *App) listChat(w http.ResponseWriter, r *http.Request) {
	messages, err := a.store.ListMessages(chi.URLParam(r, "attemptID"))
	respond(w, messages, err)
}

func (a *App) nudge(w http.ResponseWriter, r *http.Request) {
	attemptID := chi.URLParam(r, "attemptID")
	var req NudgeRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	attempt, err := a.store.GetAttempt(attemptID)
	if err != nil {
		respond(w, nil, err)
		return
	}
	message := silenceNudgeMessage(*attempt, req.Reason)
	created, err := a.store.AddMessage(attemptID, "assistant", message)
	if err == nil {
		_ = a.store.RecordFollowUps(attemptID, "silence_nudge", []string{message})
	}
	respond(w, created, err)
}

func (a *App) listWhiteboards(w http.ResponseWriter, r *http.Request) {
	items, err := a.store.ListWhiteboardsForAttempt(chi.URLParam(r, "attemptID"))
	respond(w, items, err)
}

func (a *App) artifactReply(w http.ResponseWriter, r *http.Request) {
	attemptID := chi.URLParam(r, "attemptID")
	var req ArtifactReplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if strings.TrimSpace(req.Content) == "" {
		writeError(w, http.StatusBadRequest, "content is required")
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

	// Resolve the originating request (tolerant): a missing or invalid id simply
	// means no linkage, and we fall back to the body's kind/topic.
	kind := req.Kind
	topic := req.Topic
	prompt := ""
	if strings.TrimSpace(req.RequestMessageID) != "" {
		if requestMessage, msgErr := a.store.GetMessage(attemptID, req.RequestMessageID); msgErr == nil && requestMessage.Payload != nil && requestMessage.Payload.ArtifactRequest != nil {
			ar := requestMessage.Payload.ArtifactRequest
			prompt = ar.Instructions
			if strings.TrimSpace(kind) == "" {
				kind = ar.Kind
			}
			if strings.TrimSpace(topic) == "" {
				topic = ar.Topic
			}
		}
	}
	kind = normalizeWhiteboardKind(kind)
	topic = trimOrDefault(topic, whiteboardKindLabel(kind))

	whiteboard, err := a.store.CreateWhiteboard(attemptID, WhiteboardRequest{
		Kind:    kind,
		Topic:   topic,
		Prompt:  prompt,
		Content: req.Content,
	})
	if err != nil {
		respond(w, nil, err)
		return
	}

	submission := ArtifactSubmission{
		WhiteboardID:     whiteboard.ID,
		Kind:             kind,
		Topic:            topic,
		Content:          req.Content,
		RequestMessageID: strings.TrimSpace(req.RequestMessageID),
	}
	userContent := trimOrDefault(req.Message, "ホワイトボードで回答します。")
	userMessage, err := a.store.AddStructuredMessage(attemptID, "user", userContent, "artifact", &ChatMessagePayload{Artifact: &submission})
	if err != nil {
		respond(w, nil, err)
		return
	}

	aiUserMessage := formatArtifactSubmissionForPrompt(userContent, submission)
	message, phase, err := a.completeInterviewerTurn(r.Context(), attempt, problem, aiUserMessage, req.EnglishMode)
	if err != nil {
		respond(w, nil, err)
		return
	}
	writeJSON(w, http.StatusOK, ArtifactReplyResponse{
		Artifact:     whiteboard,
		UserMessage:  *userMessage,
		Message:      *message,
		CurrentPhase: phase,
	})
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
	if attempt.NoRun {
		updatedAttempt, _ := a.store.UpdateAttemptCode(attemptID, code)
		if updatedAttempt != nil {
			attempt = updatedAttempt
		}
		result := RunResult{
			Passed:     false,
			Status:     "disabled",
			Results:    []TestCaseResult{},
			Error:      "Real Interview Modeではローカルテスト実行を禁止しています。dry runで検証してからSubmitしてください。",
			DurationMS: 0,
		}
		a.codeFiles.Decorate(attempt)
		writeJSON(w, http.StatusOK, map[string]interface{}{"attempt": attempt, "result": result})
		return
	}
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
	review, _ := a.ai.Review(r.Context(), *problem, *attempt, code, memory)
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

func initialInterviewMessage(attempt Attempt, memory ProblemMemory) string {
	prefix := ""
	if len(memory.Attempts) > 1 || len(memory.Mistakes) > 0 || len(memory.FollowUps) > 0 {
		prefix = "この問題は過去の履歴も見ながら少し厳しめに確認します。前回のミスやフォローアップも踏まえます。\n\n"
	}
	mode := "本番モードです。テスト実行や補完に頼らず、声に出して検証する前提で進めます。\n\n"
	if attempt.InterviewMode == "practice" {
		mode = "練習モードです。ただし本番同様、実装前の説明は省略しません。\n\n"
	}
	return prefix + mode + "まず問題の確認から始めます。問題文を自分の言葉で言い直し、入力の制約、返すべき値、気になるエッジケースを確認してください。曖昧な点があれば私に質問してください。コードはまだ書かなくて大丈夫です。"
}

func silenceNudgeMessage(attempt Attempt, reason string) string {
	if strings.TrimSpace(reason) == "" {
		reason = "silence"
	}
	switch attempt.CompanyPreset {
	case "meta":
		return "少し止まっています。Metaの面接ではペースも見ます。今の仮説、次に試す分岐、詰まっている一点を30秒で説明してください。"
	case "amazon":
		return "少し止まっています。Amazonの面接では判断過程も評価対象です。今の制約理解と、顧客影響のある失敗ケースを1つ説明してください。"
	case "google":
		return "少し止まっていますね。考えていることをそのまま声に出してもらえますか？今の方針と、次に確認しようとしていることを教えてください。"
	default:
		return "少し止まっています。今何を考えているか、次に検証することを短く説明してください。"
	}
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
