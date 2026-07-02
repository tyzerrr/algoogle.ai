"use client";

import { useState } from "react";
import { useMermaid } from "@/hooks/useMermaid";
import { ARTIFACT_KIND_META } from "@/lib/chatArtifacts";
import type { ArtifactSubmission } from "@/lib/types";

const COLLAPSE_LINE_THRESHOLD = 12;

export default function ArtifactCard({ artifact }: { artifact: ArtifactSubmission }) {
  const isDiagram = artifact.kind === "diagram";
  const { svg, error } = useMermaid(artifact.content, isDiagram);
  const [expanded, setExpanded] = useState(false);
  const meta = ARTIFACT_KIND_META[artifact.kind];
  const collapsible = artifact.content.split("\n").length > COLLAPSE_LINE_THRESHOLD;
  const showCollapsed = collapsible && !expanded;

  return (
    <div className="chatArtifactCard">
      <div className="chatArtifactCardHeader">
        <span className="chatArtifactChip">{meta.title}</span>
        <span className="chatArtifactTopic">{artifact.topic}</span>
      </div>
      <div
        className={`chatArtifactCardBody ${showCollapsed ? "chatArtifactCardCollapsed" : ""}`}
      >
        {isDiagram ? (
          error ? (
            <pre className="chatArtifactError">{error}</pre>
          ) : svg ? (
            <div className="mermaidCanvas" data-testid="mermaid-canvas" dangerouslySetInnerHTML={{ __html: svg }} />
          ) : (
            <pre>{artifact.content}</pre>
          )
        ) : (
          <pre>{artifact.content}</pre>
        )}
      </div>
      {collapsible ? (
        <button
          className="secondaryButton"
          type="button"
          onClick={() => setExpanded((current) => !current)}
        >
          {expanded ? "折りたたむ" : "すべて表示"}
        </button>
      ) : null}
    </div>
  );
}
