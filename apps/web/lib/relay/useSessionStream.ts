"use client";

import { useEffect, useRef, useState } from "react";
import { parseEvent } from "@coop/protocol";
import type { Event } from "@coop/protocol";
import { relayConfig } from "./config";
import type { PresenceState } from "./presenceState";
import { emptyPresenceState, reducePresence } from "./presenceState";
import type { PresenceFrame } from "./presenceFrame";
import { parsePresenceFrame } from "./presenceFrame";

export type ConnectionState = "connecting" | "open" | "closed" | "error";

export const RETRY_WARNING_THRESHOLD = 3;

const INITIAL_RETRY_MS = 1000;
const MAX_RETRY_MS = 15000;

export function useSessionStream(
  sessionId: string,
  token: string,
  name?: string,
): {
  events: Event[];
  presence: PresenceState;
  connectionState: ConnectionState;
  retryCount: number;
  sendPresence: (active: boolean) => void;
} {
  const [events, setEvents] = useState<Event[]>([]);
  const [presence, setPresence] = useState<PresenceState>(emptyPresenceState);
  const [connectionState, setConnectionState] = useState<ConnectionState>("connecting");
  const [retryCount, setRetryCount] = useState(0);
  const socketRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    let cancelled = false;
    let socket: WebSocket | null = null;
    let retryTimer: ReturnType<typeof setTimeout> | null = null;
    let retryDelay = INITIAL_RETRY_MS;
    let attempts = 0;

    if (!token) {
      setConnectionState("closed");
      return () => {
        cancelled = true;
      };
    }

    function connect() {
      if (cancelled) return;

      setConnectionState("connecting");
      setEvents([]);
      setPresence(emptyPresenceState());
      const query = `token=${encodeURIComponent(token)}${name ? `&name=${encodeURIComponent(name)}` : ""}`;
      socket = new WebSocket(`${relayConfig.wsUrl}/v1/sessions/${sessionId}/stream?${query}`);
      socketRef.current = socket;

      socket.addEventListener("open", () => {
        if (cancelled) return;
        retryDelay = INITIAL_RETRY_MS;
        attempts = 0;
        setRetryCount(0);
        setConnectionState("open");
      });

      socket.addEventListener("message", (message) => {
        if (cancelled) return;

        let parsed: unknown;
        try {
          parsed = JSON.parse(message.data as string);
        } catch {
          return;
        }

        const presenceFrame = parsePresenceFrame(parsed);
        if (presenceFrame) {
          setPresence((prev) => reducePresence(prev, presenceFrame as PresenceFrame));
          return;
        }

        const result = parseEvent(parsed);
        if (result.ok) {
          setEvents((prev) => [...prev, result.value]);
        }
      });

      socket.addEventListener("error", () => {
        if (cancelled) return;
        setConnectionState("error");
      });

      socket.addEventListener("close", () => {
        if (cancelled) return;
        socketRef.current = null;
        attempts += 1;
        setRetryCount(attempts);
        setConnectionState("closed");
        retryTimer = setTimeout(() => {
          retryDelay = Math.min(retryDelay * 2, MAX_RETRY_MS);
          connect();
        }, retryDelay);
      });
    }

    connect();

    return () => {
      cancelled = true;
      socketRef.current = null;
      if (retryTimer) clearTimeout(retryTimer);
      socket?.close();
    };
  }, [sessionId, token, name]);

  function sendPresence(active: boolean) {
    const socket = socketRef.current;
    if (!socket || socket.readyState !== WebSocket.OPEN) return;
    socket.send(JSON.stringify({ type: "presence.typing", active }));
  }

  return { events, presence, connectionState, retryCount, sendPresence };
}
