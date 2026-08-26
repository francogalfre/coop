import { relayConfig } from "./config";

export type Project = {
  id: number;
  name: string;
  slug: string;
  created_by: string;
  created_at: string;
};

export type AgentSession = {
  id: string;
  repo: string;
  cwd: string;
  harness: string;
  status: "live" | "ended";
  started_at: string;
  ended_at?: string | null;
};

export class RelayError extends Error {
  constructor(
    message: string,
    readonly status: number,
  ) {
    super(message);
    this.name = "RelayError";
  }

  get isMissing() {
    return this.status === 404;
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  let res: Response;

  try {
    res = await fetch(`${relayConfig.httpUrl}${path}`, {
      ...init,
      credentials: "include",
      headers: {
        ...(init?.body ? { "Content-Type": "application/json" } : {}),
        ...init?.headers,
      },
    });
  } catch {
    throw new RelayError("Could not reach the relay.", 0);
  }

  if (!res.ok) {
    throw new RelayError(`Request failed (${res.status}).`, res.status);
  }

  if (res.status === 204) return undefined as T;

  return (await res.json()) as T;
}

export const relayApi = {
  listProjects: () => request<{ projects: Project[] }>("/v1/projects"),

  createProject: (name: string, slug: string) =>
    request<Project>("/v1/projects", {
      method: "POST",
      body: JSON.stringify({ name, slug }),
    }),

  listSessions: (slug: string) =>
    request<{ sessions: AgentSession[] }>(`/v1/projects/${encodeURIComponent(slug)}/sessions`),

  createInvite: (slug: string) =>
    request<{ token: string }>(`/v1/projects/${encodeURIComponent(slug)}/invites`, {
      method: "POST",
    }),

  acceptInvite: (token: string) =>
    request<Project>(`/v1/projects/invites/${encodeURIComponent(token)}/accept`, {
      method: "POST",
    }),

  listViewers: (sessionId: string, token: string) =>
    request<{ viewers: string[] }>(`/v1/sessions/${encodeURIComponent(sessionId)}/viewers`, {
      headers: { "X-Coop-Session-Token": token },
    }),

  sendMessage: (sessionId: string, token: string, from: string, text: string) =>
    request<{ status: string; queued: number }>(
      `/v1/sessions/${encodeURIComponent(sessionId)}/steer`,
      {
        method: "POST",
        headers: { "X-Coop-Session-Token": token },
        body: JSON.stringify({ from, text }),
      },
    ),
};
