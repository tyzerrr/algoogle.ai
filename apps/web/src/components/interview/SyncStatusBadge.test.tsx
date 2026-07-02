import { describe, expect, it } from "vitest";
import { render } from "@testing-library/react";
import SyncStatusBadge, { syncStateLabel } from "./SyncStatusBadge";
import type { SyncState } from "@/hooks/useCodeSync";

const cases: { state: SyncState; label: string; tone: string }[] = [
  { state: { kind: "preparing" }, label: "同期準備中", tone: "warn" },
  { state: { kind: "pending" }, label: "保存待ち", tone: "warn" },
  { state: { kind: "saving" }, label: "保存中", tone: "warn" },
  { state: { kind: "saved" }, label: "保存済み", tone: "ok" },
  { state: { kind: "external" }, label: "NeoVimから反映", tone: "ok" },
  { state: { kind: "error", detail: "boom" }, label: "同期に失敗しました", tone: "bad" },
];

describe("syncStateLabel", () => {
  it.each(cases)("maps $state.kind to its JA label and tone", ({ state, label, tone }) => {
    const view = syncStateLabel(state);
    expect(view.label).toBe(label);
    expect(view.tone).toBe(tone);
  });
});

describe("SyncStatusBadge", () => {
  it.each(cases)("renders $state.kind with the tone class", ({ state, label, tone }) => {
    const { container } = render(<SyncStatusBadge state={state} />);
    const badge = container.querySelector(".syncBadge");
    expect(badge).not.toBeNull();
    expect(badge).toHaveTextContent(label);
    expect(badge).toHaveClass(`syncBadge-${tone}`);
  });
});
