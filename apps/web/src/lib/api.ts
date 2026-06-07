import type {
  Attempt,
  ChatMessage,
  DailyResponse,
  Problem,
  ProblemListItem,
  ReviewDashboard,
  ReviewResponse,
  RunResult,
} from "./types";

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL?.replace(/\/$/, "") ?? "http://localhost:8000";

async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers ?? {}),
    },
    cache: "no-store",
  });

  if (!response.ok) {
    let message = `API request failed: ${response.status}`;
    try {
      const body = (await response.json()) as { error?: string };
      if (body.error) message = body.error;
    } catch {
      // Keep the status-based message.
    }
    throw new Error(message);
  }
  return response.json() as Promise<T>;
}

export const api = {
  daily: () => apiFetch<DailyResponse>("/daily"),
  problems: () => apiFetch<ProblemListItem[]>("/problems"),
  problem: (id: string) => apiFetch<Problem>(`/problems/${id}`),
  createAttempt: (problemId: string, code?: string) =>
    apiFetch<Attempt>("/attempts", {
      method: "POST",
      body: JSON.stringify({ problem_id: problemId, language: "python", code: code ?? "" }),
    }),
  messages: (attemptId: string) => apiFetch<ChatMessage[]>(`/attempts/${attemptId}/chat`),
  sendMessage: (attemptId: string, message: string, englishMode: boolean) =>
    apiFetch<ChatMessage>(`/attempts/${attemptId}/chat`, {
      method: "POST",
      body: JSON.stringify({ message, english_mode: englishMode }),
    }),
  run: (attemptId: string, code: string) =>
    apiFetch<{ attempt: Attempt; result: RunResult }>(`/attempts/${attemptId}/run`, {
      method: "POST",
      body: JSON.stringify({ code, language: "python" }),
    }),
  review: (attemptId: string, code: string) =>
    apiFetch<{ attempt: Attempt; review: ReviewResponse }>(`/attempts/${attemptId}/review`, {
      method: "POST",
      body: JSON.stringify({ code, language: "python" }),
    }),
  reviewDashboard: () => apiFetch<ReviewDashboard>("/review"),
};
