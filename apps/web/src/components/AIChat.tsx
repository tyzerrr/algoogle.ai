"use client";

import { FormEvent, useState } from "react";
import { Send, Languages } from "lucide-react";
import type { ChatMessage } from "@/lib/types";

export default function AIChat({
  messages,
  loading,
  disabled,
  onSend,
}: {
  messages: ChatMessage[];
  loading: boolean;
  disabled: boolean;
  onSend: (message: string, englishMode: boolean) => Promise<void>;
}) {
  const [message, setMessage] = useState("");
  const [englishMode, setEnglishMode] = useState(false);

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
          <div className="empty">面接官に方針、詰まり、計算量の説明を投げられます。</div>
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
