export type Signal = {
  path: string;
  sessionId: string;
  owner?: string;
  mode: "read" | "write";
  at: string;
  active: boolean;
};

export type PresenceResponse = {
  repo: string;
  windowSeconds: number;
  paths: { path: string; signals: Signal[] }[];
};

export type SessionSummary = {
  sessionId: string;
  owner: string;
  startedAt: string;
  active: boolean;
};

export class RelayUnreachableError extends Error {}
