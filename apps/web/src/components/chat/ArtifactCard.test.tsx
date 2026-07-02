import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import ArtifactCard from "./ArtifactCard";
import type { ArtifactSubmission } from "@/lib/types";

vi.mock("mermaid", () => ({
  default: {
    initialize: vi.fn(),
    render: vi.fn(async (id: string) => ({ svg: `<svg data-id="${id}"></svg>` })),
  },
}));

function artifact(overrides: Partial<ArtifactSubmission> = {}): ArtifactSubmission {
  return {
    whiteboard_id: "w1",
    kind: "pseudocode",
    topic: "解法方針",
    content: "function solve():\n    return 1",
    ...overrides,
  };
}

describe("ArtifactCard", () => {
  it("renders a pre block for pseudocode", () => {
    const { container } = render(<ArtifactCard artifact={artifact()} />);
    expect(screen.getByText("解法方針")).toBeInTheDocument();
    expect(container.querySelector("pre")).not.toBeNull();
    expect(container.querySelector(".mermaidCanvas")).toBeNull();
  });

  it("renders an svg container for diagrams", async () => {
    const { container } = render(
      <ArtifactCard artifact={artifact({ kind: "diagram", content: "flowchart TD\n A-->B" })} />,
    );
    await waitFor(() => {
      expect(container.querySelector(".mermaidCanvas")).not.toBeNull();
    });
  });

  it("collapses long content behind a toggle", async () => {
    const long = Array.from({ length: 20 }, (_, i) => `line ${i}`).join("\n");
    const { container } = render(<ArtifactCard artifact={artifact({ content: long })} />);
    expect(container.querySelector(".chatArtifactCardCollapsed")).not.toBeNull();
    await userEvent.click(screen.getByRole("button", { name: "すべて表示" }));
    expect(container.querySelector(".chatArtifactCardCollapsed")).toBeNull();
    expect(screen.getByRole("button", { name: "折りたたむ" })).toBeInTheDocument();
  });
});
