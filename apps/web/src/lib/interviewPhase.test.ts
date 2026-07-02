import { describe, expect, it } from "vitest";
import { activePhase, phaseFromString } from "./interviewPhase";

describe("activePhase", () => {
  it("uses the server phase as the primary source", () => {
    expect(activePhase({ currentPhase: "clarify", hasRun: false, hasReview: false })).toBe("clarify");
    expect(activePhase({ currentPhase: "code", hasRun: false, hasReview: false })).toBe("code");
  });

  it("lifts an early server phase to dryrun once a run exists", () => {
    expect(activePhase({ currentPhase: "plan", hasRun: true, hasReview: false })).toBe("dryrun");
  });

  it("does not demote a later server phase because of a run", () => {
    expect(activePhase({ currentPhase: "followup", hasRun: true, hasReview: false })).toBe("followup");
  });

  it("lifts to followup once a review exists", () => {
    expect(activePhase({ currentPhase: "code", hasRun: false, hasReview: true })).toBe("followup");
  });

  it("maps the legacy planning phase onto plan", () => {
    expect(phaseFromString("planning")).toBe("plan");
    expect(activePhase({ currentPhase: "planning", hasRun: false, hasReview: false })).toBe("plan");
  });

  it("falls back to clarify-or-later when the server phase is unknown", () => {
    expect(activePhase({ currentPhase: undefined, hasRun: false, hasReview: false })).toBe("plan");
  });
});
