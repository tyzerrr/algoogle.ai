import type { ArtifactKind, ArtifactRequest, ChatMessage } from "./types";

// The interviewer's latest open artifact request, or null when none awaits an answer.
// A request is OPEN iff no artifact message exists at a later index than it, so a
// single backward scan suffices: the first artifact closes everything earlier.
export function openArtifactRequest(
  messages: ChatMessage[],
): { message: ChatMessage; request: ArtifactRequest } | null {
  for (let i = messages.length - 1; i >= 0; i -= 1) {
    const message = messages[i];
    if (message.kind === "artifact") return null;
    if (message.kind !== "artifact_request") continue;
    const request = message.payload?.artifact_request;
    if (request) return { message, request };
  }
  return null;
}

export const ARTIFACT_KIND_META: Record<ArtifactKind, { label: string; title: string }> = {
  pseudocode: { label: "疑似コード", title: "疑似コード" },
  diagram: { label: "図", title: "図（Mermaid）" },
  notes: { label: "ノート", title: "ノート" },
};

export function defaultArtifactContent(kind: ArtifactKind, topic: string): string {
  switch (kind) {
    case "diagram":
      return ["flowchart TD", "  A[入力] --> B{判定}", "  B -->|yes| C[更新]", "  B -->|no| D[次へ]"].join(
        "\n",
      );
    case "notes":
      return [
        `## ${topic || "トピック"}`,
        "",
        "- 不変条件: ",
        "- エッジケース: ",
        "",
        "| Approach | Time | Space |",
        "| --- | --- | --- |",
        "| Brute force |  |  |",
        "| Optimized |  |  |",
      ].join("\n");
    default:
      return [
        "function solve(input):",
        `    # ${topic || "approach"}`,
        "    initialize state",
        "    for each element in input:",
        "        update state",
        "    return answer",
      ].join("\n");
  }
}
