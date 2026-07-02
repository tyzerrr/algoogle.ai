"use client";

import { useState } from "react";
import { ClipboardCheck, Copy, FileCode2, LockKeyhole, Loader2, Play } from "lucide-react";
import CodeEditor from "@/components/CodeEditor";
import type { CodeFileResponse } from "@/lib/types";

export default function CodePane({
  code,
  onCodeChange,
  codeFile,
  noAutocomplete,
  noRun,
  running,
  reviewing,
  canAct,
  onRun,
  onReview,
}: {
  code: string;
  onCodeChange: (next: string) => void;
  codeFile: CodeFileResponse | null;
  noAutocomplete: boolean;
  noRun: boolean;
  running: boolean;
  reviewing: boolean;
  canAct: boolean;
  onRun: () => void;
  onReview: () => void;
}) {
  const [copied, setCopied] = useState(false);
  const nvimCommand = codeFile?.path ? `nvim ${codeFile.path}` : "";

  async function copyPath() {
    if (!nvimCommand) return;
    try {
      await navigator.clipboard.writeText(nvimCommand);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1500);
    } catch {
      // Clipboard may be unavailable; the path is still visible to copy manually.
    }
  }

  return (
    <div className="codePane">
      <div className="codePaneToolbar">
        <button
          className="secondaryButton"
          type="button"
          onClick={onRun}
          disabled={!canAct || noRun}
          title={noRun ? "本番モードではローカル実行を使いません" : "ローカルテストを実行"}
        >
          {noRun ? (
            <LockKeyhole size={17} />
          ) : running ? (
            <Loader2 className="spin" size={17} />
          ) : (
            <Play size={17} />
          )}
          {noRun ? "実行不可" : "テスト実行"}
        </button>
        <button className="button" type="button" onClick={onReview} disabled={!canAct}>
          {reviewing ? <Loader2 className="spin" size={17} /> : <ClipboardCheck size={17} />}
          提出してレビュー
        </button>
      </div>

      <div className="syncStrip">
        <div>
          <span className="syncLabel">
            <FileCode2 size={16} /> NeoVim sync
          </span>
          <code className="syncPath">{nvimCommand || "workspace を準備中"}</code>
        </div>
        <div className="syncMeta">
          {noAutocomplete ? <span className="tag">補完なし</span> : null}
          {nvimCommand ? (
            <button
              className="iconButton"
              type="button"
              onClick={copyPath}
              aria-label="NeoVimコマンドをコピー"
              title={copied ? "コピーしました" : "コマンドをコピー"}
            >
              <Copy size={16} />
            </button>
          ) : null}
        </div>
      </div>

      <CodeEditor value={code} onChange={onCodeChange} noAutocomplete={noAutocomplete} />
    </div>
  );
}
