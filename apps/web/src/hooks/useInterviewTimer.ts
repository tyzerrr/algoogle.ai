"use client";

import { useEffect, useState } from "react";
import type { Attempt } from "@/lib/types";

// Counts down the remaining time for the current attempt, recomputing whenever a new
// attempt (new created_at) replaces the old one.
export function useInterviewTimer(attempt: Attempt | null): number | null {
  const [remainingSeconds, setRemainingSeconds] = useState<number | null>(null);
  const createdAt = attempt?.created_at;
  const timeLimit = attempt?.time_limit_seconds;

  useEffect(() => {
    if (!createdAt || !timeLimit) {
      setRemainingSeconds(null);
      return undefined;
    }
    const tick = () => {
      const deadline = new Date(createdAt).getTime() + timeLimit * 1000;
      setRemainingSeconds(Math.max(0, Math.ceil((deadline - Date.now()) / 1000)));
    };
    tick();
    const timer = setInterval(tick, 1000);
    return () => clearInterval(timer);
  }, [createdAt, timeLimit]);

  return remainingSeconds;
}
