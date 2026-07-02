"use client";

import type { Attempt } from "@/lib/types";
import { PHASES, type PhaseId } from "@/lib/interviewPhase";
import type { SyncState } from "@/hooks/useCodeSync";
import SyncStatusBadge from "./SyncStatusBadge";

export default function SessionStrip({
  attempt,
  activePhase,
  syncState,
}: {
  attempt: Attempt | null;
  activePhase: PhaseId;
  syncState: SyncState;
}) {
  return (
    <section className="sessionStrip" aria-label="面接の進行状況">
      <div className="phaseRail" aria-label="interview phase">
        {PHASES.map((phase) => (
          <span
            className={`phaseStep ${phase.id === activePhase ? "phaseStepActive" : ""}`}
            key={phase.id}
          >
            {phase.label}
          </span>
        ))}
      </div>
      <div className="sessionStripMeta">
        <div className="modeFacts">
          {attempt?.no_run ? (
            <span className="status statusTodo">実行不可</span>
          ) : (
            <span className="status">実行OK</span>
          )}
          {attempt?.no_autocomplete ? (
            <span className="status statusTodo">補完OFF</span>
          ) : (
            <span className="status">補完ON</span>
          )}
          {attempt?.requires_plan ? <span className="tag">方針説明必須</span> : null}
        </div>
        <SyncStatusBadge state={syncState} />
      </div>
    </section>
  );
}
