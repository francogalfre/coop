import Link from "next/link";
import type { Route } from "next";
import { relayConfig } from "@/lib/relay/config";

type SessionSummary = {
  session_id: string;
  owner: string;
  started_at: string;
  active: boolean;
  harness?: string;
};

type SessionsResponse = {
  repo: string;
  sessions: SessionSummary[];
};

async function fetchSessions(repo: string): Promise<SessionsResponse | null> {
  try {
    const url = `${relayConfig.httpUrl}/v1/sessions?repo=${encodeURIComponent(repo)}`;
    const res = await fetch(url, { cache: "no-store" });
    if (!res.ok) return null;
    return (await res.json()) as SessionsResponse;
  } catch {
    return null;
  }
}

export default async function Home() {
  const data = relayConfig.repo ? await fetchSessions(relayConfig.repo) : null;

  return (
    <main style={{ maxWidth: 800, margin: "0 auto", padding: "1rem" }}>
      <h1>coop</h1>
      {!relayConfig.repo ? (
        <p>
          Set <code>NEXT_PUBLIC_COOP_REPO</code> to the absolute path of the repo you ran{" "}
          <code>coop attach</code>/<code>coop run</code> in, then restart the dev server.
        </p>
      ) : !data ? (
        <p>Could not reach the relay at {relayConfig.httpUrl}.</p>
      ) : data.sessions.length === 0 ? (
        <p>No sessions yet.</p>
      ) : (
        <ul>
          {data.sessions.map((session) => (
            <li key={session.session_id}>
              <Link href={`/sessions/${session.session_id}` as Route}>{session.session_id}</Link>
              {" — "}
              {session.owner}
              {" — started "}
              {new Date(session.started_at).toLocaleString()}
              {session.active ? " — active" : " — ended"}{" "}
              <span
                style={{
                  fontSize: "0.7rem",
                  fontWeight: 600,
                  color: "white",
                  background: "#6b7280",
                  borderRadius: "3px",
                  padding: "0.05rem 0.35rem",
                }}
              >
                {session.harness ?? "unknown"}
              </span>
            </li>
          ))}
        </ul>
      )}
    </main>
  );
}
