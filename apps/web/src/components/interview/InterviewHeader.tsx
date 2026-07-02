"use client";

import { RotateCcw, Settings2, Timer } from "lucide-react";
import type { Problem } from "@/lib/types";

function formatRemaining(seconds: number | null) {
  if (seconds === null) return "--:--";
  const minutes = Math.floor(seconds / 60);
  const rest = seconds % 60;
  return `${minutes}:${rest.toString().padStart(2, "0")}`;
}

export default function InterviewHeader({
  problem,
  remainingSeconds,
  settingsSummary,
  onOpenSettings,
  onNewInterview,
  disabled = false,
}: {
  problem: Problem;
  remainingSeconds: number | null;
  settingsSummary: string;
  onOpenSettings: () => void;
  onNewInterview: () => void;
  disabled?: boolean;
}) {
  return (
    <div className="pageHeader interviewHeader">
      <div>
        <p className="eyebrow">Interview room</p>
        <h1>{problem.title}</h1>
        <div className="resultMeta interviewHeaderMeta">
          {problem.order_index ? <span className="difficulty">#{problem.order_index}</span> : null}
          <span className="difficulty">{problem.difficulty}</span>
          <span className="tag">{problem.pattern}</span>
          {problem.list_name ? <span className="tag">{problem.list_name}</span> : null}
        </div>
      </div>
      <div className="interviewHeaderActions">
        <span className="timerBadge">
          <Timer size={16} />
          {formatRemaining(remainingSeconds)}
        </span>
        <button className="secondaryButton" type="button" onClick={onOpenSettings} disabled={disabled}>
          <Settings2 size={17} />
          {settingsSummary}
        </button>
        <button className="secondaryButton" type="button" onClick={onNewInterview} disabled={disabled}>
          <RotateCcw size={17} />
          新しい面接
        </button>
      </div>
    </div>
  );
}
