import { describe, expect, it } from "vitest";
import { hasAttemptProgress } from "./attemptProgress";
import type { ChatMessage } from "./types";

const starter = "def solve():\n    pass";

function userMessage(): ChatMessage {
  return {
    id: "m1",
    attempt_id: "a1",
    role: "user",
    content: "方針を説明します",
    created_at: "2024-01-01T00:00:00Z",
    kind: "text",
  };
}

function assistantMessage(): ChatMessage {
  return {
    id: "m0",
    attempt_id: "a1",
    role: "assistant",
    content: "まず制約を確認しましょう",
    created_at: "2024-01-01T00:00:00Z",
    kind: "text",
  };
}

describe("hasAttemptProgress", () => {
  it("is false for a pristine attempt", () => {
    expect(hasAttemptProgress({ code: starter, starterCode: starter, messages: [] })).toBe(false);
  });

  it("ignores assistant-only messages (interviewer opener)", () => {
    expect(
      hasAttemptProgress({
        code: starter,
        starterCode: starter,
        messages: [assistantMessage()],
      }),
    ).toBe(false);
  });

  it("is true once the code differs from the starter", () => {
    expect(
      hasAttemptProgress({
        code: starter + "\n# my work",
        starterCode: starter,
        messages: [],
      }),
    ).toBe(true);
  });

  it("is true once the candidate has sent a message", () => {
    expect(
      hasAttemptProgress({
        code: starter,
        starterCode: starter,
        messages: [assistantMessage(), userMessage()],
      }),
    ).toBe(true);
  });

  it("is true once the candidate submitted an artifact (a user message)", () => {
    const artifact: ChatMessage = {
      id: "m2",
      attempt_id: "a1",
      role: "user",
      content: "図を提出します",
      created_at: "2024-01-01T00:00:00Z",
      kind: "artifact",
      payload: {
        artifact: { whiteboard_id: "w1", kind: "diagram", topic: "flow", content: "flowchart TD" },
      },
    };
    expect(
      hasAttemptProgress({ code: starter, starterCode: starter, messages: [artifact] }),
    ).toBe(true);
  });
});
