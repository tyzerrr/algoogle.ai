import {
  AlertTriangle,
  ArrowDownToLine,
  Check,
  Loader2,
  type LucideIcon,
} from "lucide-react";
import type { SyncState } from "@/hooks/useCodeSync";

export type SyncBadgeTone = "ok" | "warn" | "bad";

export type SyncBadgeView = {
  label: string;
  tone: SyncBadgeTone;
  Icon: LucideIcon;
  spin: boolean;
};

export function syncStateLabel(state: SyncState): SyncBadgeView {
  switch (state.kind) {
    case "preparing":
      return { label: "同期準備中", tone: "warn", Icon: Loader2, spin: true };
    case "pending":
      return { label: "保存待ち", tone: "warn", Icon: Loader2, spin: false };
    case "saving":
      return { label: "保存中", tone: "warn", Icon: Loader2, spin: true };
    case "saved":
      return { label: "保存済み", tone: "ok", Icon: Check, spin: false };
    case "external":
      return { label: "NeoVimから反映", tone: "ok", Icon: ArrowDownToLine, spin: false };
    case "error":
      return { label: "同期に失敗しました", tone: "bad", Icon: AlertTriangle, spin: false };
  }
}

export default function SyncStatusBadge({ state }: { state: SyncState }) {
  const { label, tone, Icon, spin } = syncStateLabel(state);
  const title = state.kind === "error" ? state.detail : undefined;
  return (
    <span className={`syncBadge syncBadge-${tone}`} title={title}>
      <Icon size={15} className={spin ? "spin" : undefined} aria-hidden="true" />
      {label}
    </span>
  );
}
