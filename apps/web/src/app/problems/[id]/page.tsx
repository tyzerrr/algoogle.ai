"use client";

import { useEffect, useMemo, useState } from "react";
import { useParams } from "next/navigation";
import { Bot, ClipboardCheck, Loader2, Play, RotateCcw } from "lucide-react";
import AIChat from "@/components/AIChat";
import CodeEditor from "@/components/CodeEditor";
import ProblemStatement from "@/components/ProblemStatement";
import ReviewPanel from "@/components/ReviewPanel";
import TestResultPanel from "@/components/TestResultPanel";
import { api } from "@/lib/api";
import type { Attempt, ChatMessage, Problem, ReviewResponse, RunResult } from "@/lib/types";

export default function ProblemDetailPage() {
  const params = useParams<{ id: string }>();
  const problemId = params.id;
  const [problem, setProblem] = useState<Problem | null>(null);
  const [attempt, setAttempt] = useState<Attempt | null>(null);
  const [code, setCode] = useState("");
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [runResult, setRunResult] = useState<RunResult | undefined>();
  const [review, setReview] = useState<ReviewResponse | undefined>();
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState<"chat" | "run" | "review" | "reset" | null>(null);
  const [error, setError] = useState("");

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
      setMessages([]);
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
      const response = await api.run(attempt.id, code);
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
      const response = await api.review(attempt.id, code);
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
            問題を読み、方針を面接官に説明し、コードを書いてからテストとレビューに進みます。
          </p>
        </div>
        <div className="buttonRow">
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
                レビュー
              </button>
            </div>
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
