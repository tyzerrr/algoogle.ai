"use client";

import { FormEvent, useEffect, useRef, useState } from "react";
import { Loader2, Send, Languages } from "lucide-react";
import type { ChatMessage } from "@/lib/types";

export default function AIChat({
  messages,
  loading,
  loadingLabel,
  disabled,
  onSend,
}: {
  messages: ChatMessage[];
  loading: boolean;
  loadingLabel?: string;
  disabled: boolean;
  onSend: (message: string, englishMode: boolean) => Promise<void>;
}) {
  const [message, setMessage] = useState("");
  const [englishMode, setEnglishMode] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ block: "end" });
  }, [messages.length, loading]);

  async function submit(event: FormEvent) {
    event.preventDefault();
    const trimmed = message.trim();
    if (!trimmed) return;
    setMessage("");
    await onSend(trimmed, englishMode);
  }

  return (
    <form className="chat" onSubmit={submit}>
      <div className="messages">
        {messages.length === 0 ? (
          <div className="empty">面接官がまず解法方針を確認します。</div>
        ) : (
          messages.map((item) => (
            <div
              className={`message ${
                item.role === "user" ? "messageUser" : "messageAssistant"
              }`}
              key={item.id}
            >
              {item.content}
            </div>
          ))
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
        <div ref={messagesEndRef} />
      </div>

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
          onChange={(event) => setMessage(event.target.value)}
          placeholder="考え方、詰まっている点、計算量の説明を書いてください"
          disabled={disabled || loading}
        />
        <button
          className="iconButton"
          title="送信"
          aria-label="送信"
          type="submit"
          disabled={disabled || loading || message.trim().length === 0}
        >
          <Send size={18} />
        </button>
      </div>
    </form>
  );
}
