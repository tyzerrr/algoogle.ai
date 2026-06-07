"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { useParams } from "next/navigation";
import {
  Bot,
  BrainCircuit,
  Building2,
  ClipboardCheck,
  FileCode2,
  Loader2,
  LockKeyhole,
  Play,
  RotateCcw,
  ShieldCheck,
  Timer,
} from "lucide-react";
import AIChat from "@/components/AIChat";
import CodeEditor from "@/components/CodeEditor";
import ProblemStatement from "@/components/ProblemStatement";
import ReviewPanel from "@/components/ReviewPanel";
import TestResultPanel from "@/components/TestResultPanel";
import { api } from "@/lib/api";
import type {
  AIProvider,
  Attempt,
  ChatMessage,
  CodeFileResponse,
  CompanyPreset,
  InterviewMode,
  Problem,
  ReviewResponse,
  RunResult,
} from "@/lib/types";

const AI_PROVIDER_OPTIONS: { id: AIProvider; label: string; note: string }[] = [
  { id: "codex", label: "Codex", note: "Codex CLI subprocess" },
  { id: "claude", label: "Claude Code", note: "Claude Code CLI subprocess" },
];

const COMPANY_OPTIONS: { id: CompanyPreset; label: string; note: string }[] = [
  { id: "google", label: "Google", note: "曖昧さ、証明、深掘り" },
  { id: "meta", label: "Meta", note: "速度、実装精度、追加問題" },
  { id: "amazon", label: "Amazon", note: "trade-offと行動面" },
  { id: "generic", label: "Generic", note: "総合面接" },
];

const MODE_OPTIONS: { id: InterviewMode; label: string; note: string }[] = [
  { id: "real", label: "Real", note: "実行なし・補完なし" },
  { id: "practice", label: "Practice", note: "練習用に実行可" },
];

const PHASES = ["Clarify", "Plan", "Code", "Dry run", "Follow-up"];

function formatRemaining(seconds: number | null) {
  if (seconds === null) return "--:--";
  const minutes = Math.floor(seconds / 60);
  const rest = seconds % 60;
  return `${minutes}:${rest.toString().padStart(2, "0")}`;
}

function phaseLabel(phase?: string) {
  if (!phase) return "Plan";
  if (phase.includes("clar")) return "Clarify";
  if (phase.includes("plan")) return "Plan";
  if (phase.includes("code") || phase.includes("implement")) return "Code";
  if (phase.includes("dry") || phase.includes("test")) return "Dry run";
  if (phase.includes("follow") || phase.includes("review")) return "Follow-up";
  return "Plan";
}

export default function ProblemDetailPage() {
  const params = useParams<{ id: string }>();
  const problemId = params.id;
  const [problem, setProblem] = useState<Problem | null>(null);
  const [attempt, setAttempt] = useState<Attempt | null>(null);
  const [code, setCode] = useState("");
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [runResult, setRunResult] = useState<RunResult | undefined>();
  const [review, setReview] = useState<ReviewResponse | undefined>();
  const [codeFile, setCodeFile] = useState<CodeFileResponse | null>(null);
  const [syncStatus, setSyncStatus] = useState("同期準備中");
  const [aiProvider, setAIProvider] = useState<AIProvider>("codex");
  const [companyPreset, setCompanyPreset] = useState<CompanyPreset>("google");
  const [interviewMode, setInterviewMode] = useState<InterviewMode>("real");
  const [remainingSeconds, setRemainingSeconds] = useState<number | null>(null);
  const [lastActivityAt, setLastActivityAt] = useState(() => Date.now());
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState<"chat" | "run" | "review" | "reset" | null>(null);
  const [error, setError] = useState("");
  const codeRef = useRef("");
  const suppressSaveRef = useRef(false);
  const saveTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const lastKnownUpdatedAtRef = useRef("");
  const nudgeInFlightRef = useRef(false);
  const nudgeSentAtRef = useRef(0);

  useEffect(() => {
    codeRef.current = code;
  }, [code]);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      setLoading(true);
      setError("");
      try {
        const nextProblem = await api.problem(problemId);
        if (cancelled) return;
        setProblem(nextProblem);
        setCode(nextProblem.starter_code);
        const nextAttempt = await api.createAttempt(
          problemId,
          nextProblem.starter_code,
          aiProvider,
          companyPreset,
          interviewMode,
        );
        if (cancelled) return;
        setAIProvider(nextAttempt.ai_provider);
        setCompanyPreset(nextAttempt.company_preset);
        setInterviewMode(nextAttempt.interview_mode);
        setAttempt(nextAttempt);
        setRunResult(nextAttempt.test_result);
        setReview(nextAttempt.ai_review);
        const file = await api.codeFile(nextAttempt.id);
        if (cancelled) return;
        applyCodeFile(file, "NeoVim同期が有効です");
        setMessages(await api.messages(nextAttempt.id));
      } catch (err) {
        if (!cancelled) setError((err as Error).message);
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    if (problemId) void load();
    return () => {
      cancelled = true;
    };
  }, [problemId]);

  const canAct = useMemo(() => Boolean(problem && attempt && !busy), [problem, attempt, busy]);
  const activePhase = review ? "Follow-up" : runResult ? "Dry run" : phaseLabel(attempt?.current_phase);

  function markActivity() {
    setLastActivityAt(Date.now());
    nudgeSentAtRef.current = 0;
  }

  function handleCodeChange(nextCode: string) {
    setCode(nextCode);
    markActivity();
  }

  function applyCodeFile(file: CodeFileResponse, status: string) {
    setCodeFile(file);
    lastKnownUpdatedAtRef.current = file.updated_at;
    if (file.content !== codeRef.current) {
      suppressSaveRef.current = true;
      setCode(file.content);
    }
    setSyncStatus(status);
  }

  useEffect(() => {
    if (!attempt?.created_at || !attempt.time_limit_seconds) return undefined;
    const tick = () => {
      const deadline = new Date(attempt.created_at).getTime() + attempt.time_limit_seconds * 1000;
      setRemainingSeconds(Math.max(0, Math.ceil((deadline - Date.now()) / 1000)));
    };
    tick();
    const timer = setInterval(tick, 1000);
    return () => clearInterval(timer);
  }, [attempt?.created_at, attempt?.time_limit_seconds]);

  useEffect(() => {
    if (!attempt || attempt.interview_mode !== "real") return undefined;
    let cancelled = false;
    const timer = setInterval(() => {
      const silentForMS = Date.now() - lastActivityAt;
      if (busy || nudgeInFlightRef.current || nudgeSentAtRef.current || silentForMS < 30000) return;
      nudgeInFlightRef.current = true;
      api
        .nudge(attempt.id, "silence")
        .then(async () => {
          if (cancelled) return;
          nudgeSentAtRef.current = Date.now();
          setMessages(await api.messages(attempt.id));
          setSyncStatus("面接官が発話を促しました");
        })
        .catch((err: Error) => {
          if (!cancelled) setError(err.message);
        })
        .finally(() => {
          nudgeInFlightRef.current = false;
        });
    }, 5000);
    return () => {
      cancelled = true;
      clearInterval(timer);
    };
  }, [attempt, busy, lastActivityAt]);

  async function saveCodeNow() {
    if (!attempt) return codeRef.current;
    if (saveTimerRef.current) {
      clearTimeout(saveTimerRef.current);
      saveTimerRef.current = null;
    }
    setSyncStatus("保存中");
    const file = await api.saveCodeFile(attempt.id, codeRef.current);
    setCodeFile(file);
    lastKnownUpdatedAtRef.current = file.updated_at;
    setSyncStatus("保存済み");
    return codeRef.current;
  }

  useEffect(() => {
    if (!attempt || !codeFile) return undefined;
    if (suppressSaveRef.current) {
      suppressSaveRef.current = false;
      return undefined;
    }
    if (saveTimerRef.current) clearTimeout(saveTimerRef.current);
    setSyncStatus("保存待ち");
    saveTimerRef.current = setTimeout(() => {
      saveTimerRef.current = null;
      api
        .saveCodeFile(attempt.id, codeRef.current)
        .then((file) => {
          setCodeFile(file);
          lastKnownUpdatedAtRef.current = file.updated_at;
          setSyncStatus("保存済み");
        })
        .catch((err: Error) => setSyncStatus(`同期失敗: ${err.message}`));
    }, 700);
    return () => {
      if (saveTimerRef.current) {
        clearTimeout(saveTimerRef.current);
        saveTimerRef.current = null;
      }
    };
  }, [attempt?.id, code]);

  useEffect(() => {
    if (!attempt) return undefined;
    let cancelled = false;
    const timer = setInterval(() => {
      if (saveTimerRef.current) return;
      api
        .codeFile(attempt.id)
        .then((file) => {
          if (cancelled) return;
          if (
            lastKnownUpdatedAtRef.current &&
            file.updated_at !== lastKnownUpdatedAtRef.current &&
            file.content !== codeRef.current
          ) {
            applyCodeFile(file, "NeoVimから反映");
            return;
          }
          setCodeFile(file);
          lastKnownUpdatedAtRef.current = file.updated_at;
        })
        .catch((err: Error) => {
          if (!cancelled) setSyncStatus(`同期失敗: ${err.message}`);
        });
    }, 1500);
    return () => {
      cancelled = true;
      clearInterval(timer);
    };
  }, [attempt?.id]);

  async function resetAttempt(
    nextAIProvider: AIProvider = aiProvider,
    nextCompanyPreset: CompanyPreset = companyPreset,
    nextInterviewMode: InterviewMode = interviewMode,
  ) {
    if (!problem) return;
    setBusy("reset");
    setError("");
    try {
      const nextAttempt = await api.createAttempt(
        problem.id,
        problem.starter_code,
        nextAIProvider,
        nextCompanyPreset,
        nextInterviewMode,
      );
      setAIProvider(nextAttempt.ai_provider);
      setCompanyPreset(nextAttempt.company_preset);
      setInterviewMode(nextAttempt.interview_mode);
      setAttempt(nextAttempt);
      setCode(problem.starter_code);
      setRunResult(undefined);
      setReview(undefined);
      setRemainingSeconds(null);
      markActivity();
      const file = await api.codeFile(nextAttempt.id);
      applyCodeFile(file, "新しいattemptを同期");
      setMessages(await api.messages(nextAttempt.id));
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setBusy(null);
    }
  }

  async function switchInterview(nextCompanyPreset: CompanyPreset, nextInterviewMode: InterviewMode) {
    if (nextCompanyPreset === companyPreset && nextInterviewMode === interviewMode) return;
    await resetAttempt(aiProvider, nextCompanyPreset, nextInterviewMode);
  }

  async function switchAIProvider(nextAIProvider: AIProvider) {
    if (nextAIProvider === aiProvider) return;
    await resetAttempt(nextAIProvider, companyPreset, interviewMode);
  }

  async function runCode() {
    if (!attempt) return;
    if (attempt.no_run) {
      setRunResult({
        passed: false,
        status: "disabled",
        results: [],
        error: "Real Interview Modeではローカル実行を使わず、手でdry runしてください。",
        duration_ms: 0,
      });
      return;
    }
    markActivity();
    setBusy("run");
    setError("");
    try {
      const nextCode = await saveCodeNow();
      const response = await api.run(attempt.id, nextCode);
      setAttempt(response.attempt);
      setRunResult(response.result);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setBusy(null);
    }
  }

  async function requestReview() {
    if (!attempt) return;
    markActivity();
    setBusy("review");
    setError("");
    try {
      const nextCode = await saveCodeNow();
      const response = await api.review(attempt.id, nextCode);
      setAttempt(response.attempt);
      setReview(response.review);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setBusy(null);
    }
  }

  async function sendMessage(message: string, englishMode: boolean) {
    if (!attempt) return;
    markActivity();
    const optimistic: ChatMessage = {
      id: `local-${Date.now()}`,
      attempt_id: attempt.id,
      role: "user",
      content: message,
      created_at: new Date().toISOString(),
    };
    setMessages((current) => current.concat(optimistic));
    setBusy("chat");
    setError("");
    try {
      await saveCodeNow();
      const reply = await api.sendMessage(attempt.id, message, englishMode);
      const storedMessages = await api.messages(attempt.id);
      setMessages(storedMessages.length ? storedMessages : (current) => current.concat(reply));
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setBusy(null);
    }
  }

  if (loading) {
    return (
      <main className="page">
        <div className="empty">面接画面を準備しています...</div>
      </main>
    );
  }

  if (error && !problem) {
    return (
      <main className="page">
        <div className="error">{error}</div>
      </main>
    );
  }

  if (!problem) return null;

  return (
    <main className="pageWide">
      <div className="pageHeader">
        <div>
          <p className="eyebrow">Interview room</p>
          <h1>{problem.title}</h1>
          <p className="muted">
            まず方針を面接官に説明し、納得されたらNeoVimかブラウザで実装します。
          </p>
        </div>
        <div className="buttonRow">
          {problem.source_url ? (
            <a className="secondaryButton" href={problem.source_url} target="_blank" rel="noreferrer">
              公式問題
            </a>
          ) : null}
          <button className="secondaryButton" type="button" onClick={() => resetAttempt()} disabled={!canAct}>
            <RotateCcw size={18} />
            新しいattempt
          </button>
        </div>
      </div>

      {error ? <div className="error" style={{ marginBottom: 12 }}>{error}</div> : null}

      <section className="interviewControl" aria-label="real interview controls">
        <div className="controlCluster">
          <div className="controlHeading">
            <BrainCircuit size={18} />
            <div>
              <strong>AI provider</strong>
              <span>{AI_PROVIDER_OPTIONS.find((option) => option.id === aiProvider)?.note}</span>
            </div>
          </div>
          <div className="segmented segmentedTwo">
            {AI_PROVIDER_OPTIONS.map((option) => (
              <button
                className={`segmentButton ${aiProvider === option.id ? "segmentButtonActive" : ""}`}
                type="button"
                key={option.id}
                disabled={!canAct}
                title={option.note}
                onClick={() => switchAIProvider(option.id)}
              >
                {option.label}
              </button>
            ))}
          </div>
        </div>

        <div className="controlCluster">
          <div className="controlHeading">
            <Building2 size={18} />
            <div>
              <strong>Company preset</strong>
              <span>面接官の詰め方を切り替えます</span>
            </div>
          </div>
          <div className="segmented">
            {COMPANY_OPTIONS.map((option) => (
              <button
                className={`segmentButton ${companyPreset === option.id ? "segmentButtonActive" : ""}`}
                type="button"
                key={option.id}
                disabled={!canAct}
                title={option.note}
                onClick={() => switchInterview(option.id, interviewMode)}
              >
                {option.label}
              </button>
            ))}
          </div>
        </div>

        <div className="controlCluster">
          <div className="controlHeading">
            <ShieldCheck size={18} />
            <div>
              <strong>Interview mode</strong>
              <span>{MODE_OPTIONS.find((option) => option.id === interviewMode)?.note}</span>
            </div>
          </div>
          <div className="segmented segmentedTwo">
            {MODE_OPTIONS.map((option) => (
              <button
                className={`segmentButton ${interviewMode === option.id ? "segmentButtonActive" : ""}`}
                type="button"
                key={option.id}
                disabled={!canAct}
                title={option.note}
                onClick={() => switchInterview(companyPreset, option.id)}
              >
                {option.label}
              </button>
            ))}
          </div>
        </div>

        <div className="controlCluster controlClusterWide">
          <div className="modeFacts">
            <span className="timerBadge">
              <Timer size={16} />
              {formatRemaining(remainingSeconds)}
            </span>
            {attempt?.no_run ? <span className="status statusTodo">No local run</span> : <span className="status">Run allowed</span>}
            {attempt?.no_autocomplete ? (
              <span className="status statusTodo">Autocomplete off</span>
            ) : (
              <span className="status">Autocomplete on</span>
            )}
            {attempt?.requires_plan ? <span className="tag">Plan gate</span> : null}
          </div>
          <div className="phaseRail" aria-label="interview phase">
            {PHASES.map((phase) => (
              <span className={`phaseStep ${phase === activePhase ? "phaseStepActive" : ""}`} key={phase}>
                {phase}
              </span>
            ))}
          </div>
        </div>
      </section>

      <div className="interviewShell">
        <section className="pane">
          <div className="paneHeader">
            <h2>問題</h2>
          </div>
          <div className="paneBody">
            <ProblemStatement problem={problem} />
          </div>

          <div className="paneHeader">
            <h2>
              <Bot size={18} /> AI面接官
            </h2>
            {busy === "chat" ? <Loader2 className="spin" size={17} /> : null}
          </div>
          <div className="paneBody">
            <AIChat
              messages={messages}
              loading={busy === "chat"}
              disabled={!attempt}
              onSend={sendMessage}
            />
          </div>
        </section>

        <section className="pane editorPane">
          <div className="paneHeader">
            <h2>Python</h2>
            <div className="buttonRow">
              <button
                className="secondaryButton"
                type="button"
                onClick={runCode}
                disabled={!canAct || Boolean(attempt?.no_run)}
                title={attempt?.no_run ? "Real Interview Modeではローカル実行を使いません" : "ローカルテストを実行"}
              >
                {attempt?.no_run ? (
                  <LockKeyhole size={17} />
                ) : busy === "run" ? (
                  <Loader2 className="spin" size={17} />
                ) : (
                  <Play size={17} />
                )}
                {attempt?.no_run ? "No run" : "テスト"}
              </button>
              <button className="button" type="button" onClick={requestReview} disabled={!canAct}>
                {busy === "review" ? (
                  <Loader2 className="spin" size={17} />
                ) : (
                  <ClipboardCheck size={17} />
                )}
                Submit
              </button>
            </div>
          </div>
          <div className="syncStrip">
            <div>
              <span className="syncLabel">
                <FileCode2 size={16} /> NeoVim sync
              </span>
              <code className="syncPath">{codeFile?.path ? `nvim ${codeFile.path}` : "workspace を準備中"}</code>
            </div>
            <div className="syncMeta">
              {attempt?.no_autocomplete ? <span className="tag">補完なし</span> : null}
              <span className="syncState">{syncStatus}</span>
            </div>
          </div>
          <CodeEditor value={code} onChange={handleCodeChange} noAutocomplete={Boolean(attempt?.no_autocomplete)} />

          <div className="paneHeader">
            <h2>テスト結果</h2>
          </div>
          <div className="paneBody">
            <TestResultPanel result={runResult} />
          </div>

          <div className="paneHeader">
            <h2>AIレビュー</h2>
          </div>
          <div className="paneBody">
            <ReviewPanel review={review} />
          </div>
        </section>
      </div>
    </main>
  );
}
