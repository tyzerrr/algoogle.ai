"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { useParams } from "next/navigation";
import { Bot, ClipboardCheck, FileCode2, Loader2, Play, RotateCcw } from "lucide-react";
import AIChat from "@/components/AIChat";
import CodeEditor from "@/components/CodeEditor";
import ProblemStatement from "@/components/ProblemStatement";
import ReviewPanel from "@/components/ReviewPanel";
import TestResultPanel from "@/components/TestResultPanel";
import { api } from "@/lib/api";
import type {
  Attempt,
  ChatMessage,
  CodeFileResponse,
  Problem,
  ReviewResponse,
  RunResult,
} from "@/lib/types";

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
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState<"chat" | "run" | "review" | "reset" | null>(null);
  const [error, setError] = useState("");
  const codeRef = useRef("");
  const suppressSaveRef = useRef(false);
  const saveTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const lastKnownUpdatedAtRef = useRef("");

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
        const nextAttempt = await api.createAttempt(problemId, nextProblem.starter_code);
        if (cancelled) return;
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

  function applyCodeFile(file: CodeFileResponse, status: string) {
    setCodeFile(file);
    lastKnownUpdatedAtRef.current = file.updated_at;
    if (file.content !== codeRef.current) {
      suppressSaveRef.current = true;
      setCode(file.content);
    }
    setSyncStatus(status);
  }

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

  async function resetAttempt() {
    if (!problem) return;
    setBusy("reset");
    setError("");
    try {
      const nextAttempt = await api.createAttempt(problem.id, problem.starter_code);
      setAttempt(nextAttempt);
      setCode(problem.starter_code);
      setRunResult(undefined);
      setReview(undefined);
      const file = await api.codeFile(nextAttempt.id);
      applyCodeFile(file, "新しいattemptを同期");
      setMessages(await api.messages(nextAttempt.id));
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setBusy(null);
    }
  }

  async function runCode() {
    if (!attempt) return;
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
          <button className="secondaryButton" type="button" onClick={resetAttempt} disabled={!canAct}>
            <RotateCcw size={18} />
            新しいattempt
          </button>
        </div>
      </div>

      {error ? <div className="error" style={{ marginBottom: 12 }}>{error}</div> : null}

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
              <button className="secondaryButton" type="button" onClick={runCode} disabled={!canAct}>
                {busy === "run" ? <Loader2 className="spin" size={17} /> : <Play size={17} />}
                テスト
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
            <span className="syncState">{syncStatus}</span>
          </div>
          <CodeEditor value={code} onChange={setCode} />

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
