"use client";

import { useEffect, useRef, type RefObject } from "react";
import { api } from "@/lib/api";
import type { Attempt, ChatMessage } from "@/lib/types";

// In real-interview mode, if the candidate goes silent for 30s the interviewer nudges them.
// The nudge message lands in chat; we no longer surface a separate status string for it.
// Activity arrives via a ref so per-keystroke updates never re-render the page.
export function useSilenceNudge({
  attempt,
  busy,
  lastActivityRef,
  onMessages,
  onError,
}: {
  attempt: Attempt | null;
  busy: boolean;
  lastActivityRef: RefObject<number>;
  onMessages: (messages: ChatMessage[]) => void;
  onError: (message: string) => void;
}) {
  const nudgeInFlightRef = useRef(false);
  const nudgeSentAtRef = useRef(0);

  useEffect(() => {
    if (!attempt || attempt.interview_mode !== "real") return undefined;
    let cancelled = false;
    const timer = setInterval(() => {
      const lastActivityAt = lastActivityRef.current;
      const silentForMS = Date.now() - lastActivityAt;
      // A nudge newer than the last activity means we already nudged this
      // silence; fresh activity moves the ref past it and re-arms the nudge.
      if (
        busy ||
        nudgeInFlightRef.current ||
        nudgeSentAtRef.current >= lastActivityAt ||
        silentForMS < 30000
      ) {
        return;
      }
      nudgeInFlightRef.current = true;
      api
        .nudge(attempt.id, "silence")
        .then(async () => {
          if (cancelled) return;
          nudgeSentAtRef.current = Date.now();
          onMessages(await api.messages(attempt.id));
        })
        .catch((err: Error) => {
          if (!cancelled) onError(err.message);
        })
        .finally(() => {
          nudgeInFlightRef.current = false;
        });
    }, 5000);
    return () => {
      cancelled = true;
      clearInterval(timer);
    };
  }, [attempt, busy, lastActivityRef, onMessages, onError]);
}
