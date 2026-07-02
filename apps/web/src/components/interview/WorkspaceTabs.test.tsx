import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import WorkspaceTabs, { type WorkspaceTabId } from "./WorkspaceTabs";
import { useState } from "react";

function Harness({ testBadge }: { testBadge?: "green" | "red" | "gray" | null }) {
  const [active, setActive] = useState<WorkspaceTabId>("code");
  return (
    <WorkspaceTabs
      active={active}
      onChange={setActive}
      codePanel={<div data-testid="code-panel">CODE</div>}
      testPanel={<div data-testid="test-panel">TEST</div>}
      reviewPanel={<div data-testid="review-panel">REVIEW</div>}
      testBadge={testBadge}
      reviewBadge
    />
  );
}

describe("WorkspaceTabs", () => {
  it("renders the three tabs", () => {
    render(<Harness />);
    expect(screen.getByRole("tab", { name: /コード/ })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: /テスト/ })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: /レビュー/ })).toBeInTheDocument();
    expect(screen.queryByRole("tab", { name: /ホワイトボード/ })).toBeNull();
  });

  it("switching tabs toggles hidden without unmounting the panels", async () => {
    render(<Harness />);
    const codePanelNode = screen.getByTestId("code-panel");
    const testPanelNode = screen.getByTestId("test-panel");

    // Initially: code visible, test hidden but still mounted.
    expect(codePanelNode.closest("[role=tabpanel]")).not.toHaveAttribute("hidden");
    expect(testPanelNode.closest("[role=tabpanel]")).toHaveAttribute("hidden");

    await userEvent.click(screen.getByRole("tab", { name: /テスト/ }));

    // Same DOM nodes persist (Monaco/undo must not unmount).
    expect(screen.getByTestId("code-panel")).toBe(codePanelNode);
    expect(screen.getByTestId("test-panel")).toBe(testPanelNode);
    expect(codePanelNode.closest("[role=tabpanel]")).toHaveAttribute("hidden");
    expect(testPanelNode.closest("[role=tabpanel]")).not.toHaveAttribute("hidden");
  });

  it("marks the active tab with aria-selected", async () => {
    render(<Harness />);
    expect(screen.getByRole("tab", { name: /コード/ })).toHaveAttribute("aria-selected", "true");
    await userEvent.click(screen.getByRole("tab", { name: /レビュー/ }));
    expect(screen.getByRole("tab", { name: /レビュー/ })).toHaveAttribute("aria-selected", "true");
    expect(screen.getByRole("tab", { name: /コード/ })).toHaveAttribute("aria-selected", "false");
  });

  it("renders a test dot with the tone class and a review badge", () => {
    const { container } = render(<Harness testBadge="red" />);
    expect(container.querySelector(".workspaceTabDot-red")).not.toBeNull();
    expect(container.querySelector(".workspaceTabDot-green")).not.toBeNull();
  });

  it("omits the test dot when no run result", () => {
    const { container } = render(<Harness testBadge={null} />);
    expect(container.querySelector(".workspaceTabDot-red")).toBeNull();
    expect(container.querySelector(".workspaceTabDot-green")).not.toBeNull(); // review badge only
  });
});
