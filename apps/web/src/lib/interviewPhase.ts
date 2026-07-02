export type PhaseId = "clarify" | "plan" | "code" | "dryrun" | "followup";

export const PHASES: { id: PhaseId; label: string }[] = [
  { id: "clarify", label: "確認" },
  { id: "plan", label: "方針" },
  { id: "code", label: "実装" },
  { id: "dryrun", label: "検証" },
  { id: "followup", label: "深掘り" },
];

// Maps the backend's free-form phase string onto one of the 5 rail steps.
export function phaseFromString(phase?: string): PhaseId {
  if (!phase) return "plan";
  if (phase.includes("clar")) return "clarify";
  if (phase.includes("plan")) return "plan";
  if (phase.includes("code") || phase.includes("implement")) return "code";
  if (phase.includes("dry") || phase.includes("test")) return "dryrun";
  if (phase.includes("follow") || phase.includes("review")) return "followup";
  return "plan";
}

export function activePhase({
  currentPhase,
  hasRun,
  hasReview,
}: {
  currentPhase?: string;
  hasRun: boolean;
  hasReview: boolean;
}): PhaseId {
  const order = PHASES.map((phase) => phase.id);
  const index = Math.max(
    order.indexOf(phaseFromString(currentPhase)),
    hasRun ? order.indexOf("dryrun") : -1,
    hasReview ? order.indexOf("followup") : -1,
  );
  return order[Math.max(index, 0)];
}
