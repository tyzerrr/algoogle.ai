"use client";

import { useEffect, useMemo, useRef, useState, type FormEvent, type KeyboardEvent } from "react";
import { Loader2, Send, Languages, PenLine } from "lucide-react";
import ArtifactRequestCard from "@/components/chat/ArtifactRequestCard";
import ArtifactCard from "@/components/chat/ArtifactCard";
import ArtifactComposer, {
  type ArtifactComposerPayload,
  type ArtifactComposerRequest,
} from "@/components/chat/ArtifactComposer";
import { openArtifactRequest } from "@/lib/chatArtifacts";
import type { ChatMessage } from "@/lib/types";

export type ArtifactSubmitPayload = ArtifactComposerPayload & { englishMode: boolean };

type ComposerState = { mode: "request" } | { mode: "voluntary" } | null;

export default function AIChat({
  messages,
  loading,
  loadingLabel,
  disabled,
  onSend,
  onSubmitArtifact,
  artifactBusy,
  onActivity,
}: {
  messages: ChatMessage[];
  loading: boolean;
  loadingLabel?: string;
  disabled: boolean;
  onSend: (message: string, englishMode: boolean) => Promise<void>;
  onSubmitArtifact: (payload: ArtifactSubmitPayload) => Promise<void>;
  artifactBusy: boolean;
  onActivity?: () => void;
}) {
  const [message, setMessage] = useState("");
  const [englishMode, setEnglishMode] = useState(false);
  const [composer, setComposer] = useState<ComposerState>(null);
  const messagesRef = useRef<HTMLDivElement | null>(null);
  const pinnedToBottomRef = useRef(true);
  const sendingRef = useRef(false);

  const openRequest = useMemo(() => openArtifactRequest(messages), [messages]);

  // Only auto-scroll when the user is already parked at the bottom, so a silence
  // nudge arriving while they read history doesn't yank their scroll position.
  useEffect(() => {
    const container = messagesRef.current;
    if (!container || !pinnedToBottomRef.current) return;
    container.scrollTop = container.scrollHeight;
  }, [messages.length, loading, composer]);

  function handleMessagesScroll() {
    const container = messagesRef.current;
    if (!container) return;
    pinnedToBottomRef.current =
      container.scrollHeight - container.scrollTop - container.clientHeight < 80;
  }

  async function sendCurrentMessage() {
    const trimmed = message.trim();
    if (!trimmed || disabled || loading || sendingRef.current) return;
    sendingRef.current = true;
    setMessage("");
    try {
      await onSend(trimmed, englishMode);
    } finally {
      sendingRef.current = false;
    }
  }

  async function submit(event: FormEvent) {
    event.preventDefault();
    await sendCurrentMessage();
  }

  function handlePromptKeyDown(event: KeyboardEvent<HTMLTextAreaElement>) {
    if (event.key !== "Enter" || !event.metaKey || event.nativeEvent.isComposing) return;
    event.preventDefault();
    void sendCurrentMessage();
  }

  async function handleArtifactSubmit(payload: ArtifactComposerPayload) {
    try {
      await onSubmitArtifact({ ...payload, englishMode });
      setComposer(null);
    } catch {
      // Keep the composer open so the typed artifact survives; the session
      // hook has already surfaced the error banner.
    }
  }

  const composerRequest: ArtifactComposerRequest | null =
    composer?.mode === "request" && openRequest
      ? {
          messageId: openRequest.message.id,
          kind: openRequest.request.kind,
          topic: openRequest.request.topic,
          instructions: openRequest.request.instructions,
          starterContent: openRequest.request.starter_content,
        }
      : null;

  return (
    <form className="chat" onSubmit={submit}>
      <div className="messages" ref={messagesRef} onScroll={handleMessagesScroll}>
        {messages.length === 0 ? (
          <div className="empty">面接官がまず解法方針を確認します。</div>
        ) : (
          messages.map((item) => {
            const kind = item.kind ?? "text";
            const bubbleClass = `message ${item.role === "user" ? "messageUser" : "messageAssistant"}`;
            if (kind === "artifact_request" && item.payload?.artifact_request) {
              return (
                <div className="chatMessageGroup" key={item.id}>
                  {item.content ? <div className={bubbleClass}>{item.content}</div> : null}
                  <ArtifactRequestCard
                    request={item.payload.artifact_request}
                    answered={item.id !== openRequest?.message.id}
                    onAnswer={() => setComposer({ mode: "request" })}
                    disabled={disabled || loading || artifactBusy}
                  />
                </div>
              );
            }
            if (kind === "artifact" && item.payload?.artifact) {
              return (
                <div className="chatMessageGroup" key={item.id}>
                  {item.content ? <div className={bubbleClass}>{item.content}</div> : null}
                  <ArtifactCard artifact={item.payload.artifact} />
                </div>
              );
            }
            return (
              <div className={bubbleClass} key={item.id}>
                {item.content}
              </div>
            );
          })
        )}
        {loading ? (
          <div className="message messageAssistant messageThinking" role="status" aria-live="polite">
            <Loader2 className="spin" size={16} />
            <span>{loadingLabel || "AI面接官が考えています"}</span>
            <span className="thinkingPulse" aria-hidden="true">
              <span />
              <span />
              <span />
            </span>
          </div>
        ) : null}
      </div>

      {composer ? (
        <ArtifactComposer
          request={composerRequest}
          busy={artifactBusy}
          onSubmit={handleArtifactSubmit}
          onCancel={() => setComposer(null)}
          onActivity={onActivity}
        />
      ) : (
        <>
          {openRequest ? (
            <div className="chatArtifactAffordance">
              <span>面接官がホワイトボード回答を求めています</span>
              <button
                className="button"
                type="button"
                onClick={() => setComposer({ mode: "request" })}
                disabled={disabled || loading || artifactBusy}
              >
                回答する
              </button>
            </div>
          ) : null}

          <label className="toggle">
            <input
              type="checkbox"
              checked={englishMode}
              onChange={(event) => setEnglishMode(event.target.checked)}
            />
            <Languages size={17} />
            English interview mode
          </label>

          <div className="chatInput">
            <textarea
              value={message}
              onChange={(event) => {
                setMessage(event.target.value);
                onActivity?.();
              }}
              onKeyDown={handlePromptKeyDown}
              placeholder="考え方、詰まっている点、計算量の説明を書いてください"
              disabled={disabled || loading}
            />
            <div className="chatInputActions">
              <button
                className="secondaryButton"
                type="button"
                onClick={() => setComposer({ mode: "voluntary" })}
                disabled={disabled || loading || artifactBusy}
              >
                <PenLine size={17} />
                ホワイトボードで答える
              </button>
              <button
                className="button chatSendButton"
                title="送信"
                type="submit"
                disabled={disabled || loading || message.trim().length === 0}
              >
                <Send size={18} />
                送信
              </button>
            </div>
          </div>
        </>
      )}
    </form>
  );
}
