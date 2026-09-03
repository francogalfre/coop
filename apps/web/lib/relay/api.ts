import type { HarnessCommand } from "@coop/protocol";
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

export type ProjectContext = {
  text: string;
  version: number;
  updated_by?: string;
  updated_at?: string;
};

export type ProjectNote = {
  id: string;
  author_id: string;
  author_display_name: string;
  author_avatar_url?: string;
  source: "human" | "agent";
  session_id?: string;
  text: string;
  created_at: string;
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

  steerAgent: (sessionId: string, text: string, clientId: string, anchorSeq?: number) =>
    request<{ status: "accepted"; queued: number } | { status: "pending"; request_id: string }>(
      `/v1/sessions/${encodeURIComponent(sessionId)}/steer`,
      {
        method: "POST",
        body: JSON.stringify({ text, client_id: clientId, anchor_seq: anchorSeq }),
      },
    ),

  sendTeamMessage: (sessionId: string, text: string, clientId: string, anchorSeq?: number) =>
    request<{ status: "sent"; seq: number }>(
      `/v1/sessions/${encodeURIComponent(sessionId)}/message`,
      {
        method: "POST",
        body: JSON.stringify({ text, client_id: clientId, anchor_seq: anchorSeq }),
      },
    ),

  runCommand: (sessionId: string, command: HarnessCommand, args?: string) =>
    request<{ status: "ok" }>(`/v1/sessions/${encodeURIComponent(sessionId)}/command`, {
      method: "POST",
      body: JSON.stringify({ command, args }),
    }),

  answerQuestion: (sessionId: string, questionId: string, text: string) =>
    request<{ status: "answered" }>(
      `/v1/sessions/${encodeURIComponent(sessionId)}/questions/${encodeURIComponent(questionId)}/answer`,
      {
        method: "POST",
        body: JSON.stringify({ text }),
      },
    ),

  resolvePermission: (sessionId: string, requestId: string, decision: "allow" | "deny") =>
    request<{ decision: "allow" | "deny" }>(
      `/v1/sessions/${encodeURIComponent(sessionId)}/permissions/${encodeURIComponent(requestId)}/resolve`,
      {
        method: "POST",
        body: JSON.stringify({ decision }),
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

  getProjectContext: (slug: string) =>
    request<ProjectContext>(`/v1/projects/${encodeURIComponent(slug)}/context`),

  setProjectContext: (slug: string, text: string) =>
    request<ProjectContext>(`/v1/projects/${encodeURIComponent(slug)}/context`, {
      method: "PUT",
      body: JSON.stringify({ text }),
    }),

  listProjectNotes: (slug: string, limit?: number) => {
    const params = new URLSearchParams();
    if (limit !== undefined) params.set("limit", String(limit));
    const query = params.toString();

    return request<{ notes: ProjectNote[] }>(
      `/v1/projects/${encodeURIComponent(slug)}/notes${query ? `?${query}` : ""}`,
    );
  },

  postProjectNote: (slug: string, text: string) =>
    request<{ note: ProjectNote }>(`/v1/projects/${encodeURIComponent(slug)}/notes`, {
      method: "POST",
      body: JSON.stringify({ text }),
    }),

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
