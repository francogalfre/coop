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

export type Agent = {
  id: string;
  name: string;
  display_name: string;
  status: "online" | "idle" | "offline";
  current_session_id: string | null;
};

export type EventsPage = {
  events: unknown[];
  oldest_seq: number;
  has_more: boolean;
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

  listAgents: (slug: string) =>
    request<{ agents: Agent[] }>(`/v1/projects/${encodeURIComponent(slug)}/agents`),

  messageAgent: (slug: string, name: string, text: string) =>
    request<{ status: "accepted"; queued: number } | { status: "pending"; request_id: string }>(
      `/v1/projects/${encodeURIComponent(slug)}/agents/${encodeURIComponent(name)}/message`,
      {
        method: "POST",
        body: JSON.stringify({ text }),
      },
    ),

  sendMessage: (sessionId: string, text: string) =>
    request<{ status: "accepted"; queued: number } | { status: "pending"; request_id: string }>(
      `/v1/sessions/${encodeURIComponent(sessionId)}/steer`,
      {
        method: "POST",
        body: JSON.stringify({ text }),
      },
    ),

  sendTeamMessage: (sessionId: string, text: string, anchorSeq?: number) =>
    request<{ status: "sent"; seq: number }>(
      `/v1/sessions/${encodeURIComponent(sessionId)}/message`,
      {
        method: "POST",
        body: JSON.stringify({ text, anchor_seq: anchorSeq }),
      },
    ),

  setTakeover: (sessionId: string, active: boolean) =>
    request<{ active: boolean; by?: string }>(
      `/v1/sessions/${encodeURIComponent(sessionId)}/takeover`,
      {
        method: "POST",
        body: JSON.stringify({ active }),
      },
    ),

  setSessionMode: (sessionId: string, mode: "auto" | "restricted") =>
    request<{ mode: "auto" | "restricted" }>(`/v1/sessions/${encodeURIComponent(sessionId)}/mode`, {
      method: "POST",
      body: JSON.stringify({ mode }),
    }),

  resolveSteer: (sessionId: string, requestId: string, decision: "allow" | "deny") =>
    request<{ decision: "allow" | "deny" }>(
      `/v1/sessions/${encodeURIComponent(sessionId)}/steer/${encodeURIComponent(requestId)}/resolve`,
      {
        method: "POST",
        body: JSON.stringify({ decision }),
      },
    ),

  listEvents: (sessionId: string, before?: number, limit?: number) => {
    const params = new URLSearchParams();
    if (before !== undefined) params.set("before", String(before));
    if (limit !== undefined) params.set("limit", String(limit));
    const query = params.toString();

    return request<EventsPage>(
      `/v1/sessions/${encodeURIComponent(sessionId)}/events${query ? `?${query}` : ""}`,
    );
  },
};
