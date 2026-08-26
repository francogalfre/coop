"use client";

import { useEffect, useState } from "react";
import type { Event } from "@coop/protocol";
import type { PresenceState } from "@/lib/relay/presenceState";

const TYPING_TIMEOUT_MS = 3000;
const TICK_MS = 500;

function typingNames(presence: PresenceState, now: number): string[] {
  return Object.entries(presence)
    .filter(([, entry]) => entry.active && now - entry.at < TYPING_TIMEOUT_MS)
    .map(([name]) => name);
}

function agentDisplayName(events: Event[]): string {
  for (const event of events) {
    if (event.type === "session.start" && "harness" in event && typeof event.harness === "string") {
      return event.harness;
    }
  }
  return "Agent";
}

function isAgentWorking(events: Event[]): boolean {
  const openToolUseIds = new Set<string>();
  let lastRelevantType: string | null = null;

  for (const event of events) {
    if (event.type === "tool.call" && "tool_use_id" in event && typeof event.tool_use_id === "string") {
      openToolUseIds.add(event.tool_use_id);
      lastRelevantType = event.type;
    } else if (event.type === "tool.result" && "tool_use_id" in event && typeof event.tool_use_id === "string") {
      openToolUseIds.delete(event.tool_use_id);
      lastRelevantType = event.type;
    } else if (event.type === "agent.turn_start" || event.type === "agent.turn_end") {
      lastRelevantType = event.type;
    }
  }

  if (openToolUseIds.size > 0) return true;
  return lastRelevantType !== null && lastRelevantType !== "agent.turn_end";
}

export function PresenceBar({ events, presence }: { events: Event[]; presence: PresenceState }) {
  const [now, setNow] = useState(() => Date.now());

  useEffect(() => {
    const interval = setInterval(() => setNow(Date.now()), TICK_MS);
    return () => clearInterval(interval);
  }, []);

  const typing = typingNames(presence, now);
  const working = isAgentWorking(events);
  const agentName = agentDisplayName(events);

  if (typing.length === 0 && !working) return null;

  return (
    <p style={{ fontSize: "0.85rem", color: "#4b5563" }}>
      {working && <span>{agentName} is working…</span>}
      {working && typing.length > 0 && " "}
      {typing.length > 0 && (
        <span>
          {typing.map((name, i) => (
            <span key={name}>
              <strong>{name}</strong>
              {i < typing.length - 1 ? ", " : ""}
            </span>
          ))}{" "}
          {typing.length === 1 ? "is" : "are"} typing…
        </span>
      )}
    </p>
  );
}
