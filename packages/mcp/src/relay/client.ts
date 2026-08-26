import type { Config } from "../config/config.js";
import type { PresenceResponse, SessionSummary } from "./types.js";
import { RelayUnreachableError } from "./types.js";

type RawSignal = {
  session_id: string;
  owner?: string;
  mode: "read" | "write";
  at: string;
  active: boolean;
};

type RawPresenceResponse = {
  repo: string;
  window_seconds: number;
  paths: { path: string; signals: RawSignal[] }[];
};

type RawSessionsResponse = {
  repo: string;
  sessions: { session_id: string; owner: string; started_at: string; active: boolean }[];
};

async function getJson(url: URL, config: Config): Promise<unknown> {
  let response: Response;

  try {
    response = await fetch(url, {
      headers: config.cliCredential ? { Authorization: `Bearer ${config.cliCredential}` } : undefined,
    });
  } catch (cause) {
    throw new RelayUnreachableError(
      `failed to reach relay at ${url}: ${cause instanceof Error ? cause.message : String(cause)}`,
    );
  }

  if (!response.ok) {
    const body = await response.text().catch(() => "");
    throw new RelayUnreachableError(`relay responded ${response.status} for ${url}: ${body}`);
  }

  return response.json();
}

export async function fetchPresence(
  config: Config,
  paths: string[],
  windowSeconds?: number,
): Promise<PresenceResponse> {
  const params = new URLSearchParams({ repo: config.repo, paths: paths.join(",") });

  if (windowSeconds !== undefined) params.set("window_seconds", String(windowSeconds));

  const url = new URL(`/v1/presence?${params.toString()}`, config.relayUrl);
  const raw = (await getJson(url, config)) as RawPresenceResponse;

  return {
    repo: raw.repo,
    windowSeconds: raw.window_seconds,
    paths: raw.paths.map((p) => ({
      path: p.path,
      signals: p.signals.map((s) => ({
        path: p.path,
        sessionId: s.session_id,
        owner: s.owner,
        mode: s.mode,
        at: s.at,
        active: s.active,
      })),
    })),
  };
}

export async function fetchActiveSessions(config: Config): Promise<SessionSummary[]> {
  const params = new URLSearchParams({ repo: config.repo });
  const url = new URL(`/v1/sessions?${params.toString()}`, config.relayUrl);
  const raw = (await getJson(url, config)) as RawSessionsResponse;

  return raw.sessions.map((s) => ({
    sessionId: s.session_id,
    owner: s.owner,
    startedAt: s.started_at,
    active: s.active,
  }));
}
