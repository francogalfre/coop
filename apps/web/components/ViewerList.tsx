"use client";

import { useEffect, useState } from "react";
import { relayConfig } from "@/lib/relay/config";

const POLL_INTERVAL_MS = 5000;

export function ViewerList({ sessionId, token }: { sessionId: string; token: string }) {
  const [viewers, setViewers] = useState<string[]>([]);

  useEffect(() => {
    let cancelled = false;

    async function poll() {
      try {
        const res = await fetch(`${relayConfig.httpUrl}/v1/sessions/${sessionId}/viewers`, {
          headers: { "X-Coop-Session-Token": token },
        });
        if (!res.ok) {
          if (!cancelled) setViewers([]);
          return;
        }
        const data = (await res.json()) as { viewers: string[] };
        if (!cancelled) setViewers(data.viewers);
      } catch {
        if (!cancelled) setViewers([]);
      }
    }

    poll();
    const interval = setInterval(poll, POLL_INTERVAL_MS);

    return () => {
      cancelled = true;
      clearInterval(interval);
    };
  }, [sessionId, token]);

  if (viewers.length === 0) return null;

  return (
    <p style={{ fontSize: "0.85rem", color: "#4b5563" }}>
      Watching now:{" "}
      {viewers.map((name) => (
        <span
          key={name}
          style={{
            fontSize: "0.7rem",
            fontWeight: 600,
            color: "white",
            background: "#6b7280",
            borderRadius: "3px",
            padding: "0.05rem 0.35rem",
            marginRight: "0.25rem",
          }}
        >
          {name}
        </span>
      ))}
    </p>
  );
}
