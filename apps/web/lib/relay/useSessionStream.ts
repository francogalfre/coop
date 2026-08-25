"use client";

import { useEffect, useState } from "react";
import { parseEvent } from "@coop/protocol";
import type { Event } from "@coop/protocol";
import { relayConfig } from "./config";

export type ConnectionState = "connecting" | "open" | "closed" | "error";

const INITIAL_RETRY_MS = 1000;
const MAX_RETRY_MS = 15000;

export function useSessionStream(sessionId: string): {
  events: Event[];
  connectionState: ConnectionState;
} {
  const [events, setEvents] = useState<Event[]>([]);
  const [connectionState, setConnectionState] = useState<ConnectionState>("connecting");

  useEffect(() => {
    let cancelled = false;
    let socket: WebSocket | null = null;
    let retryTimer: ReturnType<typeof setTimeout> | null = null;
    let retryDelay = INITIAL_RETRY_MS;

    function connect() {
      if (cancelled) return;

      setConnectionState("connecting");
      setEvents([]);
      socket = new WebSocket(`${relayConfig.wsUrl}/v1/sessions/${sessionId}/stream`);

      socket.addEventListener("open", () => {
        if (cancelled) return;
        retryDelay = INITIAL_RETRY_MS;
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
      if (retryTimer) clearTimeout(retryTimer);
      socket?.close();
    };
  }, [sessionId]);

  return { events, connectionState };
}
