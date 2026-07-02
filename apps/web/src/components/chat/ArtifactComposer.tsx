"use client";

import { useState, type KeyboardEvent } from "react";
import { Code2, GitBranch, ListChecks } from "lucide-react";
import { useMermaid } from "@/hooks/useMermaid";
import { ARTIFACT_KIND_META, defaultArtifactContent } from "@/lib/chatArtifacts";
import type { ArtifactKind } from "@/lib/types";

export type ArtifactComposerPayload = {
  kind: ArtifactKind;
  topic: string;
  content: string;
  message: string;
  requestMessageId?: string;
};

export type ArtifactComposerRequest = {
  messageId: string;
  kind: ArtifactKind;
  topic: string;
  instructions: string;
  starterContent: string;
};

const KIND_ORDER: ArtifactKind[] = ["pseudocode", "diagram", "notes"];
const ICONS: Record<ArtifactKind, typeof Code2> = {
  pseudocode: Code2,
  diagram: GitBranch,
  notes: ListChecks,
};

export default function ArtifactComposer({
  request,
  busy,
  onSubmit,
  onCancel,
  onActivity,
}: {
  request: ArtifactComposerRequest | null;
  busy: boolean;
  onSubmit: (payload: ArtifactComposerPayload) => Promise<void>;
  onCancel: () => void;
  onActivity?: () => void;
}) {
  const initialKind = request?.kind ?? "pseudocode";
  const [kind, setKind] = useState<ArtifactKind>(initialKind);
  const [topic, setTopic] = useState(request?.topic ?? "");
  const [content, setContent] = useState(
    request
      ? request.starterContent || defaultArtifactContent(initialKind, request.topic)
      : defaultArtifactContent(initialKind, ""),
  );
  const [message, setMessage] = useState("");
  const [contentTouched, setContentTouched] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const { svg, error } = useMermaid(content, kind === "diagram");

  function selectKind(next: ArtifactKind) {
    if (request) return;
    setKind(next);
    if (!contentTouched) setContent(defaultArtifactContent(next, topic));
  }

  async function submit() {
    if (content.trim().length === 0 || busy || submitting) return;
    setSubmitting(true);
    try {
      await onSubmit({
        kind,
        topic: topic.trim() || ARTIFACT_KIND_META[kind].title,
        content,
        message: message.trim(),
        requestMessageId: request?.messageId,
      });
    } finally {
      setSubmitting(false);
    }
  }

  function handleContentKeyDown(event: KeyboardEvent<HTMLTextAreaElement>) {
    if (event.key !== "Enter" || !event.metaKey || event.nativeEvent.isComposing) return;
    event.preventDefault();
    void submit();
  }

  const disabled = busy || submitting;

  return (
    <div className="artifactComposer" aria-label="ホワイトボード回答">
      {request ? (
        <div className="artifactComposerRequest">
          <span className="chatArtifactChip">{ARTIFACT_KIND_META[kind].title}</span>
          {request.instructions ? (
            <p className="chatArtifactInstructions">{request.instructions}</p>
          ) : null}
        </div>
      ) : (
        <div className="artifactComposerKinds" role="group" aria-label="回答の種類">
          {KIND_ORDER.map((option) => {
            const Icon = ICONS[option];
            return (
              <button
                className={`artifactKind ${option === kind ? "artifactKindActive" : ""}`}
                key={option}
                type="button"
                aria-pressed={option === kind}
                onClick={() => selectKind(option)}
                disabled={disabled}
              >
                <Icon size={15} aria-hidden="true" />
                {ARTIFACT_KIND_META[option].label}
              </button>
            );
          })}
        </div>
      )}

      <label className="artifactComposerTopic">
        <span>トピック</span>
        <input
          value={topic}
          onChange={(event) => {
            setTopic(event.target.value);
            onActivity?.();
          }}
          placeholder="トピック"
          disabled={disabled}
        />
      </label>

      <div className="artifactComposerGrid">
        <label className="artifactComposerContent">
          <span>内容</span>
          <textarea
            value={content}
            onChange={(event) => {
              setContent(event.target.value);
              setContentTouched(true);
              onActivity?.();
            }}
            onKeyDown={handleContentKeyDown}
            disabled={disabled}
          />
        </label>
        {kind === "diagram" ? (
          <div className="artifactComposerPreview" aria-live="polite">
            {error ? (
              <pre className="chatArtifactError">{error}</pre>
            ) : svg ? (
              <div className="mermaidCanvas" data-testid="mermaid-preview" dangerouslySetInnerHTML={{ __html: svg }} />
            ) : (
              <div className="empty">プレビューを生成しています…</div>
            )}
          </div>
        ) : null}
      </div>

      <label className="artifactComposerComment">
        <span>コメント（任意）</span>
        <input
          value={message}
          onChange={(event) => {
            setMessage(event.target.value);
            onActivity?.();
          }}
          placeholder="面接官へ一言"
          disabled={disabled}
        />
      </label>

      <div className="artifactComposerActions">
        <button className="secondaryButton" type="button" onClick={onCancel} disabled={disabled}>
          キャンセル
        </button>
        <button
          className="button"
          type="button"
          onClick={() => void submit()}
          disabled={disabled || content.trim().length === 0}
        >
          提出
        </button>
      </div>
    </div>
  );
}
