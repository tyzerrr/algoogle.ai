import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import ArtifactRequestCard from "./ArtifactRequestCard";
import type { ArtifactRequest } from "@/lib/types";

const request: ArtifactRequest = {
  kind: "pseudocode",
  topic: "二分探索の方針",
  instructions: "全探索から最適化までを疑似コードで示してください",
  starter_content: "",
};

describe("ArtifactRequestCard", () => {
  it("renders the topic and instructions", () => {
    render(
      <ArtifactRequestCard request={request} answered={false} onAnswer={() => {}} disabled={false} />,
    );
    expect(screen.getByText("二分探索の方針")).toBeInTheDocument();
    expect(screen.getByText(/最適化までを疑似コード/)).toBeInTheDocument();
  });

  it("fires onAnswer when the button is clicked", async () => {
    const onAnswer = vi.fn();
    render(
      <ArtifactRequestCard request={request} answered={false} onAnswer={onAnswer} disabled={false} />,
    );
    await userEvent.click(screen.getByRole("button", { name: "回答する" }));
    expect(onAnswer).toHaveBeenCalledTimes(1);
  });

  it("shows a badge and no button once answered", () => {
    render(
      <ArtifactRequestCard request={request} answered onAnswer={() => {}} disabled={false} />,
    );
    expect(screen.getByText("回答済み")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "回答する" })).toBeNull();
  });
});
