"use client";

import { useEffect, type MouseEvent } from "react";
import { BrainCircuit, Building2, ShieldCheck, X } from "lucide-react";
import type { AIProvider, CompanyPreset, InterviewMode } from "@/lib/types";

export const AI_PROVIDER_OPTIONS: { id: AIProvider; label: string; note: string }[] = [
  { id: "codex", label: "Codex", note: "Codex CLI subprocess" },
  { id: "claude", label: "Claude Code", note: "Claude Code CLI subprocess" },
];

export const COMPANY_OPTIONS: { id: CompanyPreset; label: string; note: string }[] = [
  { id: "google", label: "Google", note: "曖昧さ、証明、深掘り" },
  { id: "meta", label: "Meta", note: "速度、実装精度、追加問題" },
  { id: "amazon", label: "Amazon", note: "trade-offと行動面" },
  { id: "generic", label: "Generic", note: "総合面接" },
];

export const MODE_OPTIONS: { id: InterviewMode; label: string; note: string }[] = [
  { id: "real", label: "本番", note: "実行なし・補完なし" },
  { id: "practice", label: "練習", note: "練習用に実行可" },
];

export type InterviewSettings = {
  aiProvider: AIProvider;
  companyPreset: CompanyPreset;
  interviewMode: InterviewMode;
};

export default function InterviewSettingsDialog({
  open,
  aiProvider,
  companyPreset,
  interviewMode,
  disabled = false,
  onRequestChange,
  onClose,
}: {
  open: boolean;
  aiProvider: AIProvider;
  companyPreset: CompanyPreset;
  interviewMode: InterviewMode;
  disabled?: boolean;
  onRequestChange: (next: InterviewSettings) => void;
  onClose: () => void;
}) {
  useEffect(() => {
    if (!open) return undefined;
    function handleKey(event: KeyboardEvent) {
      if (event.key === "Escape") onClose();
    }
    document.addEventListener("keydown", handleKey);
    return () => document.removeEventListener("keydown", handleKey);
  }, [open, onClose]);

  if (!open) return null;

  function handleOverlayClick(event: MouseEvent<HTMLDivElement>) {
    if (event.target === event.currentTarget) onClose();
  }

  return (
    <div className="dialogOverlay" onMouseDown={handleOverlayClick}>
      <div
        className="dialogCard settingsDialogCard"
        role="dialog"
        aria-modal="true"
        aria-label="面接設定"
      >
        <div className="settingsDialogHead">
          <h2 className="dialogTitle">面接設定</h2>
          <button className="iconButton" type="button" onClick={onClose} aria-label="設定を閉じる">
            <X size={18} />
          </button>
        </div>

        <div className="settingsGroup">
          <div className="controlHeading">
            <BrainCircuit size={18} />
            <div>
              <strong>AI provider</strong>
              <span>面接官として使うAI CLIを選びます</span>
            </div>
          </div>
          <div className="segmented segmentedTwo">
            {AI_PROVIDER_OPTIONS.map((option) => (
              <button
                className={`segmentButton ${aiProvider === option.id ? "segmentButtonActive" : ""}`}
                type="button"
                key={option.id}
                disabled={disabled}
                title={option.note}
                onClick={() =>
                  option.id !== aiProvider &&
                  onRequestChange({ aiProvider: option.id, companyPreset, interviewMode })
                }
              >
                {option.label}
              </button>
            ))}
          </div>
        </div>

        <div className="settingsGroup">
          <div className="controlHeading">
            <Building2 size={18} />
            <div>
              <strong>Company preset</strong>
              <span>面接官の詰め方を切り替えます</span>
            </div>
          </div>
          <div className="segmented">
            {COMPANY_OPTIONS.map((option) => (
              <button
                className={`segmentButton ${companyPreset === option.id ? "segmentButtonActive" : ""}`}
                type="button"
                key={option.id}
                disabled={disabled}
                title={option.note}
                onClick={() =>
                  option.id !== companyPreset &&
                  onRequestChange({ aiProvider, companyPreset: option.id, interviewMode })
                }
              >
                {option.label}
              </button>
            ))}
          </div>
        </div>

        <div className="settingsGroup">
          <div className="controlHeading">
            <ShieldCheck size={18} />
            <div>
              <strong>Interview mode</strong>
              <span>本番は実行なし・補完なし、練習は実行できます</span>
            </div>
          </div>
          <div className="segmented segmentedTwo">
            {MODE_OPTIONS.map((option) => (
              <button
                className={`segmentButton ${interviewMode === option.id ? "segmentButtonActive" : ""}`}
                type="button"
                key={option.id}
                disabled={disabled}
                title={option.note}
                onClick={() =>
                  option.id !== interviewMode &&
                  onRequestChange({ aiProvider, companyPreset, interviewMode: option.id })
                }
              >
                {option.label}
              </button>
            ))}
          </div>
        </div>

        <p className="settingsNote muted">設定を変更すると新しいattemptが始まります。</p>
      </div>
    </div>
  );
}
