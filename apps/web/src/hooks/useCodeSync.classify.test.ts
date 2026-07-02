import { describe, expect, it } from "vitest";
import { classifyPolledFile } from "./useCodeSync";
import type { CodeFileResponse } from "@/lib/types";

function file(overrides: Partial<CodeFileResponse> = {}): CodeFileResponse {
  return { path: "main.py", content: "code", updated_at: "t1", size: 4, ...overrides };
}

describe("classifyPolledFile", () => {
  it("returns 'noop' when updated_at is unchanged", () => {
    expect(classifyPolledFile(file({ updated_at: "t1" }), "t1", "anything")).toBe("noop");
  });

  it("returns 'external' when updated_at moved and content differs from local", () => {
    expect(
      classifyPolledFile(file({ updated_at: "t2", content: "remote" }), "t1", "local"),
    ).toBe("external");
  });

  it("returns 'adopt' when updated_at moved but content matches local", () => {
    expect(
      classifyPolledFile(file({ updated_at: "t2", content: "same" }), "t1", "same"),
    ).toBe("adopt");
  });

  it("returns 'adopt' on first observation (no last-known updated_at)", () => {
    expect(
      classifyPolledFile(file({ updated_at: "t2", content: "remote" }), "", "local"),
    ).toBe("adopt");
  });
});
