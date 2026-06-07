"use client";

import { useEffect, useMemo, useState } from "react";
import {
  Bot,
  Braces,
  Code2,
  GitBranch,
  ListChecks,
  MessageSquare,
  Plus,
  Save,
  SplitSquareHorizontal,
  Table2,
} from "lucide-react";
import type {
  WhiteboardArtifact,
  WhiteboardKind,
  WhiteboardRequest,
  WhiteboardSuggestion,
} from "@/lib/types";

type WhiteboardMode = {
  id: WhiteboardKind;
  label: string;
  title: string;
  icon: typeof Code2;
};

const MODES: WhiteboardMode[] = [
  { id: "pseudocode", label: "Pseudo", title: "疑似コード", icon: Code2 },
  { id: "mermaid_sequence", label: "Mermaid", title: "シーケンス図", icon: GitBranch },
  { id: "data_structure", label: "Data", title: "データ構造", icon: SplitSquareHorizontal },
  { id: "invariants", label: "Invariant", title: "不変条件", icon: ListChecks },
  { id: "state_transition", label: "State", title: "状態遷移", icon: Braces },
  { id: "complexity_table", label: "Cost", title: "計算量比較", icon: Table2 },
];

function modeFor(kind: WhiteboardKind) {
  return MODES.find((mode) => mode.id === kind) ?? MODES[0];
}

function defaultPrompt(kind: WhiteboardKind) {
  switch (kind) {
    case "mermaid_sequence":
      return "操作の流れをMermaidのsequenceDiagramで表し、各ステップで何を検証するか説明してください。";
    case "data_structure":
      return "使うデータ構造、保持する情報、各操作の計算量を説明してください。";
    case "invariants":
      return "初期化、維持、終了時に成立する不変条件を説明してください。";
    case "state_transition":
      return "状態、遷移条件、不正な遷移を整理してください。";
    case "complexity_table":
      return "全探索と最適解を比較し、時間・空間・trade-offを説明してください。";
    default:
      return "実装前に、全探索から最適化までの方針を疑似コードで説明してください。";
  }
}

function defaultContent(kind: WhiteboardKind, topic: string) {
  switch (kind) {
    case "mermaid_sequence":
      return [
        "sequenceDiagram",
        "  participant Candidate",
        "  participant Interviewer",
        "  Candidate->>Interviewer: Explain the approach",
        "  Interviewer-->>Candidate: Probe invariants and edge cases",
      ].join("\n");
    case "data_structure":
      return ["Data structure:", "- ", "", "Operations:", "- lookup:", "- update:", "", "Invariant:", "- "].join("\n");
    case "invariants":
      return ["Invariant:", "- ", "", "Why it holds:", "- Initialization:", "- Maintenance:", "- Termination:"].join("\n");
    case "state_transition":
      return ["States:", "- ", "", "Transitions:", "- ", "", "Invalid transitions:", "- "].join("\n");
    case "complexity_table":
      return [
        "| Approach | Time | Space | Trade-off |",
        "| --- | --- | --- | --- |",
        "| Brute force |  |  |  |",
        "| Optimized |  |  |  |",
      ].join("\n");
    default:
      return [
        "function solve(input):",
        `    # ${topic || "approach"}`,
        "    clarify constraints",
        "    outline brute force",
        "    derive optimized invariant",
        "    handle edge cases",
        "    return answer",
      ].join("\n");
  }
}

function buildRequest(kind: WhiteboardKind, topic: string, prompt: string, content: string): WhiteboardRequest {
  const nextTopic = topic.trim() || modeFor(kind).title;
  return {
    kind,
    topic: nextTopic,
    prompt: prompt.trim() || defaultPrompt(kind),
    content: content.trim() || defaultContent(kind, nextTopic),
  };
}

export default function WhiteboardPanel({
  whiteboards,
  activeId,
  loading,
  disabled,
  suggestion,
  onSelect,
  onCreate,
  onUpdate,
  onSuggest,
  onDiscuss,
}: {
  whiteboards: WhiteboardArtifact[];
  activeId: string | null;
  loading: boolean;
  disabled: boolean;
  suggestion?: WhiteboardSuggestion | null;
  onSelect: (id: string) => void;
  onCreate: (payload: WhiteboardRequest) => Promise<void>;
  onUpdate: (id: string, payload: WhiteboardRequest) => Promise<void>;
  onSuggest: (message: string) => Promise<void>;
  onDiscuss: (
    whiteboard: WhiteboardArtifact | null,
    payload: WhiteboardRequest,
    message: string,
  ) => Promise<void>;
}) {
  const activeWhiteboard = useMemo(
    () => whiteboards.find((item) => item.id === activeId) ?? null,
    [activeId, whiteboards],
  );
  const [kind, setKind] = useState<WhiteboardKind>("pseudocode");
  const [topic, setTopic] = useState("解法方針");
  const [prompt, setPrompt] = useState(defaultPrompt("pseudocode"));
  const [content, setContent] = useState(defaultContent("pseudocode", "解法方針"));
  const [suggestContext, setSuggestContext] = useState("");
  const [discussionMessage, setDiscussionMessage] = useState("");
  const [previewSvg, setPreviewSvg] = useState("");
  const [previewError, setPreviewError] = useState("");
  const [themeTick, setThemeTick] = useState(0);

  useEffect(() => {
    if (!activeWhiteboard) return;
    setKind(activeWhiteboard.kind);
    setTopic(activeWhiteboard.topic);
    setPrompt(activeWhiteboard.prompt);
    setContent(activeWhiteboard.content);
  }, [activeWhiteboard?.id, activeWhiteboard?.version, activeWhiteboard]);

  useEffect(() => {
    function handleThemeChange() {
      setThemeTick((current) => current + 1);
    }
    window.addEventListener("algosensei-theme-change", handleThemeChange);
    return () => window.removeEventListener("algosensei-theme-change", handleThemeChange);
  }, []);

  useEffect(() => {
    if (kind !== "mermaid_sequence") {
      setPreviewSvg("");
      setPreviewError("");
      return undefined;
    }
    let cancelled = false;
    async function renderMermaid() {
      try {
        const mermaidModule = await import("mermaid");
        const mermaid = mermaidModule.default;
        const theme = document.documentElement.dataset.theme === "light" ? "default" : "dark";
        mermaid.initialize({ startOnLoad: false, securityLevel: "strict", theme });
        const id = `whiteboard-${Date.now()}-${Math.random().toString(36).slice(2)}`;
        const { svg } = await mermaid.render(id, content || defaultContent("mermaid_sequence", topic));
        if (!cancelled) {
          setPreviewSvg(svg);
          setPreviewError("");
        }
      } catch (err) {
        if (!cancelled) {
          setPreviewSvg("");
          setPreviewError((err as Error).message);
        }
      }
    }
    void renderMermaid();
    return () => {
      cancelled = true;
    };
  }, [kind, content, topic, themeTick]);

  const isDirty =
    !activeWhiteboard ||
    activeWhiteboard.kind !== kind ||
    activeWhiteboard.topic !== topic ||
    activeWhiteboard.prompt !== prompt ||
    activeWhiteboard.content !== content;
  const canSave = !disabled && !loading && isDirty;
  const selectedMode = modeFor(kind);

  function resetForKind(nextKind: WhiteboardKind) {
    setKind(nextKind);
    if (!activeWhiteboard) {
      const nextTopic = modeFor(nextKind).title;
      setTopic(nextTopic);
      setPrompt(defaultPrompt(nextKind));
      setContent(defaultContent(nextKind, nextTopic));
    }
  }

  async function createNew() {
    const nextTopic = modeFor(kind).title;
    await onCreate({
      kind,
      topic: nextTopic,
      prompt: defaultPrompt(kind),
      content: defaultContent(kind, nextTopic),
    });
  }

  async function saveCurrent() {
    const payload = buildRequest(kind, topic, prompt, content);
    if (activeWhiteboard) {
      await onUpdate(activeWhiteboard.id, payload);
      return;
    }
    await onCreate(payload);
  }

  return (
    <section className="whiteboardPanel" aria-label="WhiteBoard">
      <div className="whiteboardHeader">
        <div>
          <p className="eyebrow">WhiteBoard</p>
          <h2>{selectedMode.title}</h2>
        </div>
        <div className="buttonRow">
          <button
            className="secondaryButton"
            type="button"
            onClick={() => onSuggest(suggestContext)}
            disabled={disabled || loading}
            title="AI面接官にWhiteBoardが必要か判断させる"
          >
            <Bot size={17} />
            AI判定
          </button>
          <button className="iconButton" type="button" onClick={createNew} disabled={disabled || loading} title="新規">
            <Plus size={18} />
          </button>
          <button className="iconButton" type="button" onClick={saveCurrent} disabled={!canSave} title="保存">
            <Save size={18} />
          </button>
        </div>
      </div>

      <div className="whiteboardTabs" aria-label="saved whiteboards">
        {whiteboards.length === 0 ? (
          <span className="whiteboardEmpty">No board</span>
        ) : (
          whiteboards.map((item) => (
            <button
              className={`whiteboardTab ${item.id === activeWhiteboard?.id ? "whiteboardTabActive" : ""}`}
              key={item.id}
              type="button"
              onClick={() => onSelect(item.id)}
              title={item.prompt}
            >
              {modeFor(item.kind).label}
              <span>{item.topic}</span>
            </button>
          ))
        )}
      </div>

      <div className="whiteboardModes" aria-label="whiteboard mode">
        {MODES.map((mode) => {
          const Icon = mode.icon;
          return (
            <button
              className={`whiteboardMode ${mode.id === kind ? "whiteboardModeActive" : ""}`}
              key={mode.id}
              type="button"
              onClick={() => resetForKind(mode.id)}
              title={mode.title}
            >
              <Icon size={16} />
              {mode.label}
            </button>
          );
        })}
      </div>

      {suggestion?.reason ? <div className="whiteboardReason">{suggestion.reason}</div> : null}

      <div className="whiteboardMetaGrid">
        <label>
          <span>Topic</span>
          <input value={topic} onChange={(event) => setTopic(event.target.value)} disabled={disabled || loading} />
        </label>
        <label>
          <span>AI context</span>
          <input
            value={suggestContext}
            onChange={(event) => setSuggestContext(event.target.value)}
            placeholder="今の迷い、議論したい点"
            disabled={disabled || loading}
          />
        </label>
      </div>

      <label className="whiteboardPrompt">
        <span>Interviewer prompt</span>
        <textarea value={prompt} onChange={(event) => setPrompt(event.target.value)} disabled={disabled || loading} />
      </label>

      <div className="whiteboardEditorGrid">
        <label className="whiteboardEditor">
          <span>Board</span>
          <textarea value={content} onChange={(event) => setContent(event.target.value)} disabled={disabled || loading} />
        </label>
        <div className="whiteboardPreview" aria-live="polite">
          <div className="whiteboardPreviewTitle">Preview</div>
          {kind === "mermaid_sequence" ? (
            previewError ? (
              <pre>{previewError}</pre>
            ) : previewSvg ? (
              <div className="mermaidCanvas" dangerouslySetInnerHTML={{ __html: previewSvg }} />
            ) : (
              <div className="empty">Rendering...</div>
            )
          ) : (
            <pre>{content}</pre>
          )}
        </div>
      </div>

      <div className="whiteboardDiscuss">
        <textarea
          value={discussionMessage}
          onChange={(event) => setDiscussionMessage(event.target.value)}
          placeholder="この白板で面接官に聞きたいこと"
          disabled={disabled || loading}
        />
        <button
          className="button"
          type="button"
          onClick={() => onDiscuss(activeWhiteboard, buildRequest(kind, topic, prompt, content), discussionMessage)}
          disabled={disabled || loading}
        >
          <MessageSquare size={17} />
          議論
        </button>
      </div>
    </section>
  );
}
