import type { ChatMessage } from "./types";

export type AttemptProgressInput = {
  code: string;
  starterCode: string;
  messages: ChatMessage[];
};

// Progress = the candidate produced something that a fresh attempt would discard.
// A submitted artifact already counts as a user message.
export function hasAttemptProgress({
  code,
  starterCode,
  messages,
}: AttemptProgressInput): boolean {
  if (code !== starterCode) return true;
  if (messages.some((message) => message.role === "user")) return true;
  return false;
}
