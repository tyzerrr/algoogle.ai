import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import AIChat from "./AIChat";
import type { ChatMessage } from "@/lib/types";

vi.mock("mermaid", () => ({
  default: {
    initialize: vi.fn(),
    render: vi.fn(async (id: string) => ({ svg: `<svg data-id="${id}"></svg>` })),
  },
}));

let seq = 0;
function base(role: ChatMessage["role"], content: string): ChatMessage {
  seq += 1;
  return {
    id: `m${seq}`,
    attempt_id: "a1",
    role,
    content,
    created_at: "2024-01-01T00:00:00Z",
    kind: "text",
  };
}

function openRequestMessage(): ChatMessage {
  return {
    ...base("assistant", "疑似コードで方針を示してください"),
    kind: "artifact_request",
    payload: {
      artifact_request: {
        kind: "pseudocode",
        topic: "解法方針",
        instructions: "全探索から最適化まで",
        starter_content: "SEED-CONTENT",
      },
    },
  };
}

function artifactMessage(): ChatMessage {
  return {
    ...base("user", "提出しました"),
    kind: "artifact",
    payload: {
      artifact: { whiteboard_id: "w1", kind: "pseudocode", topic: "解法方針", content: "function solve():" },
    },
  };
}

function renderChat(messages: ChatMessage[], overrides: Partial<Parameters<typeof AIChat>[0]> = {}) {
  const onSend = vi.fn().mockResolvedValue(undefined);
  const onSubmitArtifact = vi.fn().mockResolvedValue(undefined);
  const view = render(
    <AIChat
      messages={messages}
      loading={false}
      disabled={false}
      onSend={onSend}
      onSubmitArtifact={onSubmitArtifact}
      artifactBusy={false}
      {...overrides}
    />,
  );
  return { onSend, onSubmitArtifact, container: view.container };
}

describe("AIChat", () => {
  it("renders text, artifact_request and artifact messages", () => {
    const messages = [base("assistant", "こんにちは"), openRequestMessage(), artifactMessage()];
    const { container } = renderChat(messages);
    expect(screen.getByText("こんにちは")).toBeInTheDocument();
    expect(screen.getByText("疑似コードで方針を示してください")).toBeInTheDocument();
    // request is answered (an artifact follows) -> badge, ArtifactCard rendered.
    expect(screen.getByText("回答済み")).toBeInTheDocument();
    expect(container.querySelector(".chatArtifactCard")).not.toBeNull();
  });

  it("shows the pending affordance for an open request and opens a seeded composer", async () => {
    renderChat([openRequestMessage()]);
    expect(screen.getByText("面接官がホワイトボード回答を求めています")).toBeInTheDocument();
    await userEvent.click(screen.getAllByRole("button", { name: "回答する" })[0]);
    const content = screen.getByRole("textbox", { name: "内容" }) as HTMLTextAreaElement;
    expect(content.value).toBe("SEED-CONTENT");
    // request mode: no kind selector, affordance replaced by composer.
    expect(screen.queryByText("面接官がホワイトボード回答を求めています")).toBeNull();
  });

  it("opens a voluntary composer with the kind selector", async () => {
    renderChat([base("assistant", "どうぞ")]);
    await userEvent.click(screen.getByRole("button", { name: /ホワイトボードで答える/ }));
    expect(screen.getByRole("group", { name: "回答の種類" })).toBeInTheDocument();
  });

  it("submits an artifact and closes the composer", async () => {
    const { onSubmitArtifact } = renderChat([base("assistant", "どうぞ")]);
    await userEvent.click(screen.getByRole("button", { name: /ホワイトボードで答える/ }));
    await userEvent.click(screen.getByRole("button", { name: "提出" }));
    expect(onSubmitArtifact).toHaveBeenCalledTimes(1);
    expect(onSubmitArtifact).toHaveBeenCalledWith(
      expect.objectContaining({ kind: "pseudocode", englishMode: false }),
    );
    // Composer closed -> normal input row returns.
    expect(screen.getByRole("button", { name: /送信/ })).toBeInTheDocument();
  });

  it("fires onActivity while typing in the chat textarea", async () => {
    const onActivity = vi.fn();
    renderChat([base("assistant", "どうぞ")], { onActivity });
    const textarea = screen.getByPlaceholderText(
      "考え方、詰まっている点、計算量の説明を書いてください",
    );
    await userEvent.type(textarea, "hi");
    expect(onActivity).toHaveBeenCalled();
  });

  it("keeps the composer (and typed content) open when submission fails", async () => {
    const { onSubmitArtifact } = renderChat([base("assistant", "どうぞ")]);
    onSubmitArtifact.mockRejectedValueOnce(new Error("network down"));
    await userEvent.click(screen.getByRole("button", { name: /ホワイトボードで答える/ }));
    const content = screen.getByRole("textbox", { name: "内容" }) as HTMLTextAreaElement;
    await userEvent.clear(content);
    await userEvent.type(content, "my pseudocode");
    await userEvent.click(screen.getByRole("button", { name: "提出" }));
    expect(onSubmitArtifact).toHaveBeenCalledTimes(1);
    // Composer stays open with the typed content preserved.
    expect((screen.getByRole("textbox", { name: "内容" }) as HTMLTextAreaElement).value).toBe(
      "my pseudocode",
    );
  });
});
