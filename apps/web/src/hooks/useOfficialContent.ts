"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { api } from "@/lib/api";
import type { OfficialProblemContent } from "@/lib/types";

export type OfficialContentState = {
  content: OfficialProblemContent | null;
  loading: boolean;
  error: string;
  retry: () => void;
};

// Fetches the persisted official (LeetCode) problem content for a problem, keyed by
// problemId and cancellation-safe so stale problem responses never overwrite the current one.
export function useOfficialContent(problemId: string | undefined): OfficialContentState {
  const [content, setContent] = useState<OfficialProblemContent | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [reloadTick, setReloadTick] = useState(0);
  const activeProblemRef = useRef<string | undefined>(undefined);

  useEffect(() => {
    if (!problemId) return undefined;
    let cancelled = false;
    activeProblemRef.current = problemId;
    setContent(null);
    setError("");
    setLoading(true);
    api
      .officialProblem(problemId)
      .then((next) => {
        if (!cancelled) setContent(next);
      })
      .catch((err: Error) => {
        if (!cancelled) setError(err.message);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [problemId, reloadTick]);

  const retry = useCallback(() => setReloadTick((tick) => tick + 1), []);

  return { content, loading, error, retry };
}
