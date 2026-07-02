import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import ArtifactComposer, { type ArtifactComposerRequest } from "./ArtifactComposer";

vi.mock("mermaid", () => ({
  default: {
    initialize: vi.fn(),
    render: vi.fn(async (id: string) => ({ svg: `<svg data-id="${id}"></svg>` })),
  },
}));

const request: ArtifactComposerRequest = {
  messageId: "req-1",
  kind: "notes",
  topic: "計算量",
  instructions: "時間と空間を比較して",
  starterContent: "## 計算量\n- seed",
};

describe("ArtifactComposer", () => {
  it("seeds the starter content and fixes the kind in request mode", () => {
    render(
      <ArtifactComposer request={request} busy={false} onSubmit={vi.fn()} onCancel={vi.fn()} />,
    );
    expect(screen.getByText("時間と空間を比較して")).toBeInTheDocument();
    expect(screen.getByRole("textbox", { name: "内容" })).toHaveValue("## 計算量\n- seed");
    // No kind selector in request mode.
    expect(screen.queryByRole("group", { name: "回答の種類" })).toBeNull();
  });

  it("submits the payload including the request message id", async () => {
    const onSubmit = vi.fn().mockResolvedValue(undefined);
    render(
      <ArtifactComposer request={request} busy={false} onSubmit={onSubmit} onCancel={vi.fn()} />,
    );
    await userEvent.click(screen.getByRole("button", { name: "提出" }));
    expect(onSubmit).toHaveBeenCalledWith(
      expect.objectContaining({
        kind: "notes",
        topic: "計算量",
        requestMessageId: "req-1",
        content: "## 計算量\n- seed",
      }),
    );
  });

  it("reseeds content when switching kind in voluntary mode if untouched", async () => {
    render(<ArtifactComposer request={null} busy={false} onSubmit={vi.fn()} onCancel={vi.fn()} />);
    const content = screen.getByRole("textbox", { name: "内容" }) as HTMLTextAreaElement;
    expect(content.value).toContain("function solve");
    await userEvent.click(screen.getByRole("button", { name: /ノート/ }));
    expect(content.value).toContain("## ");
  });

  it("does not reseed content once the user has typed", async () => {
    render(<ArtifactComposer request={null} busy={false} onSubmit={vi.fn()} onCancel={vi.fn()} />);
    const content = screen.getByRole("textbox", { name: "内容" });
    await userEvent.clear(content);
    await userEvent.type(content, "my own plan");
    await userEvent.click(screen.getByRole("button", { name: /ノート/ }));
    expect(content).toHaveValue("my own plan");
  });

  it("fires onActivity while typing in the content textarea", async () => {
    const onActivity = vi.fn();
    render(
      <ArtifactComposer
        request={null}
        busy={false}
        onSubmit={vi.fn()}
        onCancel={vi.fn()}
        onActivity={onActivity}
      />,
    );
    const content = screen.getByRole("textbox", { name: "内容" });
    await userEvent.type(content, "x");
    expect(onActivity).toHaveBeenCalled();
  });

  it("fires onCancel", async () => {
    const onCancel = vi.fn();
    render(<ArtifactComposer request={null} busy={false} onSubmit={vi.fn()} onCancel={onCancel} />);
    await userEvent.click(screen.getByRole("button", { name: "キャンセル" }));
    expect(onCancel).toHaveBeenCalledTimes(1);
  });
});
