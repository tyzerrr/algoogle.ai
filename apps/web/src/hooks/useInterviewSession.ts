"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { api } from "@/lib/api";
import { hasAttemptProgress } from "@/lib/attemptProgress";
import type { ArtifactComposerPayload } from "@/components/chat/ArtifactComposer";
import type {
  AIProvider,
  Attempt,
  ChatMessage,
  CompanyPreset,
  InterviewMode,
  Problem,
  ReviewResponse,
  RunResult,
} from "@/lib/types";
import { useCodeSync } from "./useCodeSync";
import { useInterviewTimer } from "./useInterviewTimer";
import { useSilenceNudge } from "./useSilenceNudge";

export type InterviewSettings = {
  aiProvider: AIProvider;
  companyPreset: CompanyPreset;
  interviewMode: InterviewMode;
};

const AI_PROVIDER_LABELS: Record<AIProvider, string> = {
  codex: "Codex",
  claude: "Claude Code",
};

type BusyKind = "chat" | "run" | "review" | "reset" | "artifact" | null;

export type ArtifactSubmitPayload = ArtifactComposerPayload & { englishMode: boolean };

export function useInterviewSession(problemId: string | undefined) {
  const [problem, setProblem] = useState<Problem | null>(null);
  const [attempt, setAttempt] = useState<Attempt | null>(null);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [runResult, setRunResult] = useState<RunResult | undefined>();
  const [review, setReview] = useState<ReviewResponse | undefined>();
  const [aiProvider, setAIProvider] = useState<AIProvider>("codex");
  const [companyPreset, setCompanyPreset] = useState<CompanyPreset>("google");
  const [interviewMode, setInterviewMode] = useState<InterviewMode>("real");
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState<BusyKind>(null);
  const [chatLoadingLabel, setChatLoadingLabel] = useState("");
  const [error, setError] = useState("");
  // Ref, not state: activity is recorded on every keystroke and must not
  // re-render the interview page (that re-render was the jank being fixed).
  const lastActivityRef = useRef(Date.now());

  const markActivity = useCallback(() => {
    lastActivityRef.current = Date.now();
  }, []);
  const codeSync = useCodeSync(attempt, markActivity);
  const remainingSeconds = useInterviewTimer(attempt);

  // Keep the live settings available to the mount-time load effect without re-running it.
  const settingsRef = useRef<InterviewSettings>({ aiProvider, companyPreset, interviewMode });
  useEffect(() => {
    settingsRef.current = { aiProvider, companyPreset, interviewMode };
  }, [aiProvider, companyPreset, interviewMode]);

  const applyAttemptSettings = useCallback((next: Attempt) => {
    setAIProvider(next.ai_provider);
    setCompanyPreset(next.company_preset);
    setInterviewMode(next.interview_mode);
  }, []);

  useSilenceNudge({
    attempt,
    busy: busy !== null,
    lastActivityRef,
    onMessages: setMessages,
    onError: setError,
  });

  useEffect(() => {
    if (!problemId) return undefined;
    let cancelled = false;
    async function load() {
      setLoading(true);
      setError("");
      try {
        const nextProblem = await api.problem(problemId!);
        if (cancelled) return;
        setProblem(nextProblem);
        codeSync.setLocalCode(nextProblem.starter_code);
        const settings = settingsRef.current;
        const nextAttempt = await api.createAttempt(
          problemId!,
          nextProblem.starter_code,
          settings.aiProvider,
          settings.companyPreset,
          settings.interviewMode,
        );
        if (cancelled) return;
        applyAttemptSettings(nextAttempt);
        setAttempt(nextAttempt);
        setRunResult(nextAttempt.test_result);
        setReview(nextAttempt.ai_review);
        const [file, nextMessages] = await Promise.all([
          api.codeFile(nextAttempt.id),
          api.messages(nextAttempt.id),
        ]);
        if (cancelled) return;
        codeSync.seedFromFile(file);
        setMessages(nextMessages);
      } catch (err) {
        if (!cancelled) setError((err as Error).message);
      } finally {
        if (!cancelled) setLoading(false);
      }
    }
    void load();
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [problemId]);

  const canAct = useMemo(() => Boolean(problem && attempt && !busy), [problem, attempt, busy]);
  const currentAIProviderLabel = AI_PROVIDER_LABELS[aiProvider] ?? "AI面接官";

  const hasProgress = useMemo(
    () =>
      Boolean(problem) &&
      hasAttemptProgress({
        code: codeSync.code,
        starterCode: problem?.starter_code ?? "",
        messages,
      }),
    [problem, codeSync.code, messages],
  );

  const resetAttempt = useCallback(
    async (next?: Partial<InterviewSettings>) => {
      if (!problem) return;
      const settings: InterviewSettings = {
        aiProvider: next?.aiProvider ?? aiProvider,
        companyPreset: next?.companyPreset ?? companyPreset,
        interviewMode: next?.interviewMode ?? interviewMode,
      };
      setBusy("reset");
      setError("");
      try {
        const nextAttempt = await api.createAttempt(
          problem.id,
          problem.starter_code,
          settings.aiProvider,
          settings.companyPreset,
          settings.interviewMode,
          true,
        );
        applyAttemptSettings(nextAttempt);
        setAttempt(nextAttempt);
        codeSync.setLocalCode(problem.starter_code);
        setRunResult(undefined);
        setReview(undefined);
        markActivity();
        const [file, nextMessages] = await Promise.all([
          api.codeFile(nextAttempt.id),
          api.messages(nextAttempt.id),
        ]);
        codeSync.seedFromFile(file);
        setMessages(nextMessages);
      } catch (err) {
        setError((err as Error).message);
      } finally {
        setBusy(null);
      }
    },
    [problem, aiProvider, companyPreset, interviewMode, applyAttemptSettings, codeSync, markActivity],
  );

  const runCode = useCallback(async () => {
    if (!attempt) return;
    if (attempt.no_run) {
      setRunResult({
        passed: false,
        status: "disabled",
        results: [],
        error: "本番モードではローカル実行を使わず、手でdry runしてください。",
        duration_ms: 0,
      });
      return;
    }
    markActivity();
    setBusy("run");
    setError("");
    try {
      const nextCode = await codeSync.saveNow();
      const response = await api.run(attempt.id, nextCode);
      setAttempt(response.attempt);
      setRunResult(response.result);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setBusy(null);
    }
  }, [attempt, codeSync, markActivity]);

  const requestReview = useCallback(async () => {
    if (!attempt) return;
    markActivity();
    setBusy("review");
    setError("");
    try {
      const nextCode = await codeSync.saveNow();
      const response = await api.review(attempt.id, nextCode);
      setAttempt(response.attempt);
      setReview(response.review);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setBusy(null);
    }
  }, [attempt, codeSync, markActivity]);

  const sendMessage = useCallback(
    async (message: string, englishMode: boolean) => {
      if (!attempt) return;
      markActivity();
      const optimistic: ChatMessage = {
        id: `local-${Date.now()}`,
        attempt_id: attempt.id,
        role: "user",
        content: message,
        created_at: new Date().toISOString(),
        kind: "text",
      };
      setMessages((current) => current.concat(optimistic));
      setBusy("chat");
      setChatLoadingLabel(`${currentAIProviderLabel}へ送信しています`);
      setError("");
      try {
        await codeSync.saveNow();
        setChatLoadingLabel(`${currentAIProviderLabel}が考えています`);
        const response = await api.sendMessage(attempt.id, message, englishMode);
        const storedMessages = await api.messages(attempt.id);
        setMessages(
          storedMessages.length ? storedMessages : (current) => current.concat(response.message),
        );
        setAttempt((current) =>
          current ? { ...current, current_phase: response.current_phase } : current,
        );
      } catch (err) {
        setError((err as Error).message);
      } finally {
        setBusy(null);
        setChatLoadingLabel("");
      }
    },
    [attempt, codeSync, currentAIProviderLabel, markActivity],
  );

  const submitArtifact = useCallback(
    async (payload: ArtifactSubmitPayload) => {
      if (!attempt) return;
      markActivity();
      setBusy("artifact");
      setError("");
      try {
        await codeSync.saveNow();
        const res = await api.artifactReply(attempt.id, {
          kind: payload.kind,
          topic: payload.topic,
          content: payload.content,
          message: payload.message,
          request_message_id: payload.requestMessageId,
          english_mode: payload.englishMode,
        });
        const storedMessages = await api.messages(attempt.id);
        setMessages(
          storedMessages.length
            ? storedMessages
            : (current) => current.concat(res.user_message, res.message),
        );
        setAttempt((current) =>
          current ? { ...current, current_phase: res.current_phase } : current,
        );
      } catch (err) {
        setError((err as Error).message);
        // Rethrow so the composer knows the submission failed and keeps the typed content.
        throw err;
      } finally {
        setBusy(null);
      }
    },
    [attempt, codeSync, markActivity],
  );

  return {
    problem,
    attempt,
    messages,
    runResult,
    review,
    aiProvider,
    companyPreset,
    interviewMode,
    loading,
    busy,
    chatLoadingLabel,
    error,
    remainingSeconds,
    canAct,
    hasProgress,
    markActivity,
    code: codeSync.code,
    codeFile: codeSync.codeFile,
    syncState: codeSync.syncState,
    onCodeChange: codeSync.onCodeChange,
    resetAttempt,
    runCode,
    requestReview,
    sendMessage,
    submitArtifact,
  };
}
