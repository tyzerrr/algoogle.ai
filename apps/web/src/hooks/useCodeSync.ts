"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { api } from "@/lib/api";
import type { Attempt, CodeFileResponse } from "@/lib/types";

export type SyncState =
  | { kind: "preparing" | "pending" | "saving" | "saved" | "external" }
  | { kind: "error"; detail: string };

export type CodeSync = {
  code: string;
  codeFile: CodeFileResponse | null;
  syncState: SyncState;
  onCodeChange: (next: string) => void;
  seedFromFile: (file: CodeFileResponse) => void;
  setLocalCode: (content: string) => void;
  saveNow: () => Promise<string>;
};

export type PolledFileAction = "external" | "adopt" | "noop";

// Decides how a polled code file should affect local state so unchanged polls
// don't churn React state (which re-rendered the whole interview page every 1.5s).
export function classifyPolledFile(
  file: CodeFileResponse,
  lastKnownUpdatedAt: string,
  currentCode: string,
): PolledFileAction {
  if (file.updated_at === lastKnownUpdatedAt) return "noop";
  if (lastKnownUpdatedAt !== "" && file.content !== currentCode) return "external";
  return "adopt";
}

// Owns the editor buffer and keeps it in sync with the per-attempt code file:
// debounced 700ms save on local edits, 1.5s poll to pick up external (NeoVim) edits.
export function useCodeSync(
  attempt: Attempt | null,
  onActivity?: () => void,
): CodeSync {
  const [code, setCode] = useState("");
  const [codeFile, setCodeFile] = useState<CodeFileResponse | null>(null);
  const [syncState, setSyncState] = useState<SyncState>({ kind: "preparing" });

  const codeRef = useRef("");
  const suppressSaveRef = useRef(false);
  const saveTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const lastKnownUpdatedAtRef = useRef("");
  const attemptId = attempt?.id;

  useEffect(() => {
    codeRef.current = code;
  }, [code]);

  const seedFromFile = useCallback((file: CodeFileResponse) => {
    setCodeFile(file);
    lastKnownUpdatedAtRef.current = file.updated_at;
    if (file.content !== codeRef.current) {
      suppressSaveRef.current = true;
      setCode(file.content);
    }
    setSyncState({ kind: "saved" });
  }, []);

  const setLocalCode = useCallback((content: string) => {
    suppressSaveRef.current = true;
    setCode(content);
    setSyncState({ kind: "preparing" });
  }, []);

  const onCodeChange = useCallback(
    (next: string) => {
      setCode(next);
      onActivity?.();
    },
    [onActivity],
  );

  const saveNow = useCallback(async () => {
    if (!attemptId) return codeRef.current;
    if (saveTimerRef.current) {
      clearTimeout(saveTimerRef.current);
      saveTimerRef.current = null;
    }
    setSyncState({ kind: "saving" });
    try {
      const file = await api.saveCodeFile(attemptId, codeRef.current);
      setCodeFile(file);
      lastKnownUpdatedAtRef.current = file.updated_at;
      setSyncState({ kind: "saved" });
    } catch (err) {
      setSyncState({ kind: "error", detail: (err as Error).message });
    }
    return codeRef.current;
  }, [attemptId]);

  // Debounced save on local edits.
  useEffect(() => {
    if (!attemptId || !codeFile) return undefined;
    if (suppressSaveRef.current) {
      suppressSaveRef.current = false;
      return undefined;
    }
    if (saveTimerRef.current) clearTimeout(saveTimerRef.current);
    setSyncState({ kind: "pending" });
    saveTimerRef.current = setTimeout(() => {
      saveTimerRef.current = null;
      setSyncState({ kind: "saving" });
      api
        .saveCodeFile(attemptId, codeRef.current)
        .then((file) => {
          setCodeFile(file);
          lastKnownUpdatedAtRef.current = file.updated_at;
          setSyncState({ kind: "saved" });
        })
        .catch((err: Error) => setSyncState({ kind: "error", detail: err.message }));
    }, 700);
    return () => {
      if (saveTimerRef.current) {
        clearTimeout(saveTimerRef.current);
        saveTimerRef.current = null;
      }
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [attemptId, code]);

  // Poll for external (NeoVim) edits.
  useEffect(() => {
    if (!attemptId) return undefined;
    let cancelled = false;
    const timer = setInterval(() => {
      if (saveTimerRef.current) return;
      api
        .codeFile(attemptId)
        .then((file) => {
          if (cancelled) return;
          const action = classifyPolledFile(
            file,
            lastKnownUpdatedAtRef.current,
            codeRef.current,
          );
          if (action === "noop") return;
          lastKnownUpdatedAtRef.current = file.updated_at;
          setCodeFile(file);
          if (action === "external") {
            suppressSaveRef.current = true;
            setCode(file.content);
            setSyncState({ kind: "external" });
          }
        })
        .catch((err: Error) => {
          if (cancelled) return;
          const detail = err.message;
          setSyncState((prev) =>
            prev.kind === "error" && prev.detail === detail ? prev : { kind: "error", detail },
          );
        });
    }, 1500);
    return () => {
      cancelled = true;
      clearInterval(timer);
    };
  }, [attemptId]);

  return { code, codeFile, syncState, onCodeChange, seedFromFile, setLocalCode, saveNow };
}
