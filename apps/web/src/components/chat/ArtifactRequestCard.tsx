"use client";

import { Code2, GitBranch, ListChecks } from "lucide-react";
import { ARTIFACT_KIND_META } from "@/lib/chatArtifacts";
import type { ArtifactKind, ArtifactRequest } from "@/lib/types";

const ICONS: Record<ArtifactKind, typeof Code2> = {
  pseudocode: Code2,
  diagram: GitBranch,
  notes: ListChecks,
};

export default function ArtifactRequestCard({
  request,
  answered,
  onAnswer,
  disabled,
}: {
  request: ArtifactRequest;
  answered: boolean;
  onAnswer: () => void;
  disabled: boolean;
}) {
  const meta = ARTIFACT_KIND_META[request.kind];
  const Icon = ICONS[request.kind];
  return (
    <div className="chatArtifactRequestCard">
      <div className="chatArtifactRequestHead">
        <Icon size={16} aria-hidden="true" />
        <span className="chatArtifactChip">{meta.title}</span>
      </div>
      <h4 className="chatArtifactTopic">{request.topic}</h4>
      {request.instructions ? (
        <p className="chatArtifactInstructions">{request.instructions}</p>
      ) : null}
      {answered ? (
        <span className="chatArtifactAnsweredBadge">回答済み</span>
      ) : (
        <button className="button" type="button" onClick={onAnswer} disabled={disabled}>
          回答する
        </button>
      )}
    </div>
  );
}
