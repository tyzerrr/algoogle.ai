"use client";

import type { ReactNode } from "react";

export type WorkspaceTabId = "code" | "test" | "review";

export type TestBadgeTone = "green" | "red" | "gray";

type TabConfig = {
  id: WorkspaceTabId;
  label: string;
};

const TABS: TabConfig[] = [
  { id: "code", label: "コード" },
  { id: "test", label: "テスト" },
  { id: "review", label: "レビュー" },
];

export default function WorkspaceTabs({
  active,
  onChange,
  codePanel,
  testPanel,
  reviewPanel,
  testBadge,
  reviewBadge = false,
}: {
  active: WorkspaceTabId;
  onChange: (tab: WorkspaceTabId) => void;
  codePanel: ReactNode;
  testPanel: ReactNode;
  reviewPanel: ReactNode;
  testBadge?: TestBadgeTone | null;
  reviewBadge?: boolean;
}) {
  const panels: Record<WorkspaceTabId, ReactNode> = {
    code: codePanel,
    test: testPanel,
    review: reviewPanel,
  };

  function badgeFor(id: WorkspaceTabId): ReactNode {
    if (id === "test" && testBadge) {
      return <span className={`workspaceTabDot workspaceTabDot-${testBadge}`} aria-hidden="true" />;
    }
    if (id === "review" && reviewBadge) {
      return <span className="workspaceTabDot workspaceTabDot-green" aria-hidden="true" />;
    }
    return null;
  }

  return (
    <div className="workspaceTabsRoot">
      <div className="workspaceTabs" role="tablist" aria-label="ワークスペース">
        {TABS.map((tab) => (
          <button
            className={`workspaceTab ${active === tab.id ? "workspaceTabActive" : ""}`}
            key={tab.id}
            id={`workspace-tab-${tab.id}`}
            type="button"
            role="tab"
            aria-selected={active === tab.id}
            aria-controls={`workspace-panel-${tab.id}`}
            onClick={() => onChange(tab.id)}
          >
            {tab.label}
            {badgeFor(tab.id)}
          </button>
        ))}
      </div>
      {TABS.map((tab) => (
        <div
          className="workspacePanel"
          key={tab.id}
          id={`workspace-panel-${tab.id}`}
          role="tabpanel"
          aria-labelledby={`workspace-tab-${tab.id}`}
          hidden={active !== tab.id}
        >
          {panels[tab.id]}
        </div>
      ))}
    </div>
  );
}
