"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import type { Route } from "next";
import { relayConfig } from "@/lib/relay/config";

type Session = {
  id: string;
  repo: string;
  cwd: string;
  harness: string;
  status: "live" | "ended";
  started_at: string;
  ended_at: string | null;
};

const STATUS_COLORS: Record<string, string> = {
  live: "#059669",
  ended: "#6b7280",
};

function SessionRow({ session, slug }: { session: Session; slug: string }) {
  return (
    <li style={{ padding: "0.4rem 0", borderBottom: "1px solid #e5e7eb", fontSize: "0.9rem" }}>
      <Link href={`/sessions/${session.id}?from=${slug}` as Route}>{session.id}</Link>{" "}
      <span
        style={{
          fontSize: "0.7rem",
          fontWeight: 600,
          color: "white",
          background: STATUS_COLORS[session.status] ?? "#374151",
          borderRadius: "3px",
          padding: "0.05rem 0.35rem",
        }}
      >
        {session.status}
      </span>{" "}
      <span style={{ fontSize: "0.8rem", color: "#6b7280" }}>
        {session.harness} — {session.repo}
      </span>
    </li>
  );
}

function InviteFlow({ slug }: { slug: string }) {
  const [url, setUrl] = useState("");
  const [status, setStatus] = useState<"idle" | "inviting" | "error" | "denied" | "copied">("idle");

  async function invite() {
    setStatus("inviting");
    try {
      const res = await fetch(`${relayConfig.httpUrl}/v1/projects/${slug}/invites`, {
        method: "POST",
        credentials: "include",
      });
      if (res.status === 404) {
        setStatus("denied");
        return;
      }
      if (!res.ok) {
        setStatus("error");
        return;
      }
      const data = (await res.json()) as { token: string };
      setUrl(`${window.location.origin}/p/${slug}/invite/${data.token}`);
      setStatus("idle");
    } catch {
      setStatus("error");
    }
  }

  async function copy() {
    try {
      await navigator.clipboard.writeText(url);
      setStatus("copied");
    } catch {
      setStatus("error");
    }
  }

  return (
    <div style={{ margin: "1rem 0" }}>
      <button onClick={invite} disabled={status === "inviting"}>
        Invite teammate
      </button>
      {status === "denied" && (
        <p style={{ fontSize: "0.8rem", color: "#dc2626" }}>Only the project owner can invite.</p>
      )}
      {status === "error" && (
        <p style={{ fontSize: "0.8rem", color: "#dc2626" }}>Could not create invite.</p>
      )}
      {url && (
        <p style={{ fontSize: "0.85rem" }}>
          <code>{url}</code>{" "}
          <button onClick={copy}>{status === "copied" ? "Copied!" : "Copy"}</button>
        </p>
      )}
    </div>
  );
}

function NoAccess({ slug }: { slug: string }) {
  return (
    <main style={{ maxWidth: 800, margin: "0 auto", padding: "1rem" }}>
      <p>
        <Link href="/">← projects</Link>
      </p>
      <p style={{ color: "#dc2626" }}>You don&apos;t have access to {slug}.</p>
    </main>
  );
}

export default function ProjectPage() {
  const params = useParams<{ slug: string }>();
  const slug = params.slug;
  const [sessions, setSessions] = useState<Session[] | null>(null);
  const [denied, setDenied] = useState(false);
  const [error, setError] = useState(false);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      try {
        const res = await fetch(`${relayConfig.httpUrl}/v1/projects/${slug}/sessions`, {
          credentials: "include",
        });
        if (res.status === 404) {
          if (!cancelled) setDenied(true);
          return;
        }
        if (!res.ok) {
          if (!cancelled) setError(true);
          return;
        }
        const data = (await res.json()) as { sessions: Session[] };
        if (!cancelled) setSessions(data.sessions);
      } catch {
        if (!cancelled) setError(true);
      }
    }

    load();
    return () => {
      cancelled = true;
    };
  }, [slug]);

  if (denied) return <NoAccess slug={slug} />;

  const live = sessions?.filter((s) => s.status === "live") ?? [];
  const ended = sessions?.filter((s) => s.status === "ended") ?? [];

  return (
    <main style={{ maxWidth: 800, margin: "0 auto", padding: "1rem" }}>
      <p>
        <Link href="/">← projects</Link>
      </p>
      <h1>{slug}</h1>
      <InviteFlow slug={slug} />
      {error ? (
        <p style={{ color: "#dc2626" }}>Could not reach the relay at {relayConfig.httpUrl}.</p>
      ) : sessions === null ? (
        <p style={{ fontSize: "0.85rem", color: "#4b5563" }}>Loading sessions…</p>
      ) : sessions.length === 0 ? (
        <p style={{ fontSize: "0.85rem", color: "#4b5563" }}>No sessions yet.</p>
      ) : (
        <>
          <h2 style={{ fontSize: "1rem" }}>Live</h2>
          {live.length === 0 ? (
            <p style={{ fontSize: "0.85rem", color: "#4b5563" }}>No live sessions.</p>
          ) : (
            <ul style={{ listStyle: "none", padding: 0, margin: 0 }}>
              {live.map((s) => (
                <SessionRow key={s.id} session={s} slug={slug} />
              ))}
            </ul>
          )}
          <h2 style={{ fontSize: "1rem" }}>Ended</h2>
          {ended.length === 0 ? (
            <p style={{ fontSize: "0.85rem", color: "#4b5563" }}>No ended sessions.</p>
          ) : (
            <ul style={{ listStyle: "none", padding: 0, margin: 0 }}>
              {ended.map((s) => (
                <SessionRow key={s.id} session={s} slug={slug} />
              ))}
            </ul>
          )}
        </>
      )}
    </main>
  );
}
