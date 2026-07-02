import { describe, expect, it } from "vitest";
import {
  ARTIFACT_KIND_META,
  defaultArtifactContent,
  openArtifactRequest,
} from "./chatArtifacts";
import type { ArtifactRequest, ChatMessage, ChatMessageKind } from "./types";

let seq = 0;

function message(
  kind: ChatMessageKind,
  extra?: Partial<ChatMessage>,
): ChatMessage {
  seq += 1;
  return {
    id: `m${seq}`,
    attempt_id: "a1",
    role: kind === "artifact" ? "user" : "assistant",
    content: "…",
    created_at: "2024-01-01T00:00:00Z",
    kind,
    ...extra,
  };
}

function request(topic = "解法方針"): ArtifactRequest {
  return { kind: "pseudocode", topic, instructions: "説明して", starter_content: "" };
}

function requestMessage(topic = "解法方針"): ChatMessage {
  return message("artifact_request", { payload: { artifact_request: request(topic) } });
}

function artifactMessage(): ChatMessage {
  return message("artifact", {
    payload: {
      artifact: { whiteboard_id: "w1", kind: "pseudocode", topic: "解法方針", content: "code" },
    },
  });
}

describe("openArtifactRequest", () => {
  it("returns null when there are no requests", () => {
    expect(openArtifactRequest([message("text"), message("text")])).toBeNull();
  });

  it("returns the latest unanswered request", () => {
    const req = requestMessage("トピックA");
    const result = openArtifactRequest([message("text"), req]);
    expect(result?.message.id).toBe(req.id);
    expect(result?.request.topic).toBe("トピックA");
  });

  it("returns null once an artifact appears at a later index", () => {
    const req = requestMessage();
    const messages = [req, artifactMessage()];
    expect(openArtifactRequest(messages)).toBeNull();
  });

  it("treats a request that occurs after an artifact as open", () => {
    const req = requestMessage("後発");
    const messages = [requestMessage("先発"), artifactMessage(), req];
    const result = openArtifactRequest(messages);
    expect(result?.message.id).toBe(req.id);
    expect(result?.request.topic).toBe("後発");
  });

  it("treats legacy messages without kind as text", () => {
    const legacy = {
      id: "legacy",
      attempt_id: "a1",
      role: "assistant",
      content: "old",
      created_at: "2024-01-01T00:00:00Z",
    } as unknown as ChatMessage;
    expect(openArtifactRequest([legacy])).toBeNull();
  });
});

describe("ARTIFACT_KIND_META", () => {
  it("covers all three kinds", () => {
    expect(ARTIFACT_KIND_META.pseudocode.title).toBe("疑似コード");
    expect(ARTIFACT_KIND_META.diagram.title).toBe("図（Mermaid）");
    expect(ARTIFACT_KIND_META.notes.title).toBe("ノート");
  });
});

describe("defaultArtifactContent", () => {
  it("seeds a mermaid flowchart for diagrams", () => {
    expect(defaultArtifactContent("diagram", "x")).toContain("flowchart TD");
  });

  it("injects the topic heading for notes", () => {
    expect(defaultArtifactContent("notes", "二分探索")).toContain("## 二分探索");
  });

  it("produces a function scaffold for pseudocode", () => {
    expect(defaultArtifactContent("pseudocode", "topic")).toContain("function solve");
  });
});
