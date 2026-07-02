import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import InterviewSettingsDialog from "./InterviewSettingsDialog";

function setup() {
  const onRequestChange = vi.fn();
  const onClose = vi.fn();
  render(
    <InterviewSettingsDialog
      open
      aiProvider="codex"
      companyPreset="google"
      interviewMode="real"
      onRequestChange={onRequestChange}
      onClose={onClose}
    />,
  );
  return { onRequestChange, onClose };
}

describe("InterviewSettingsDialog", () => {
  it("shows the persistent note about starting a new attempt", () => {
    setup();
    expect(screen.getByText("設定を変更すると新しいattemptが始まります。")).toBeInTheDocument();
  });

  it("fires onRequestChange with the new combo when a different option is picked", async () => {
    const { onRequestChange } = setup();
    await userEvent.click(screen.getByRole("button", { name: "Meta" }));
    expect(onRequestChange).toHaveBeenCalledWith({
      aiProvider: "codex",
      companyPreset: "meta",
      interviewMode: "real",
    });
  });

  it("switches interview mode using JA labels", async () => {
    const { onRequestChange } = setup();
    await userEvent.click(screen.getByRole("button", { name: "練習" }));
    expect(onRequestChange).toHaveBeenCalledWith({
      aiProvider: "codex",
      companyPreset: "google",
      interviewMode: "practice",
    });
  });

  it("does nothing when the already-active option is clicked", async () => {
    const { onRequestChange } = setup();
    await userEvent.click(screen.getByRole("button", { name: "Google" }));
    await userEvent.click(screen.getByRole("button", { name: "本番" }));
    await userEvent.click(screen.getByRole("button", { name: "Codex" }));
    expect(onRequestChange).not.toHaveBeenCalled();
  });
});
