"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
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

export function useSessionStream(sessionId: string): {
  events: Event[];
  presence: PresenceState;
  connectionState: ConnectionState;
  retryCount: number;
  sendPresence: (active: boolean) => void;
  mergeEvents: (events: Event[]) => void;
} {
  const [eventsBySeq, setEventsBySeq] = useState<Map<number, Event>>(new Map());
  const [presence, setPresence] = useState<PresenceState>(emptyPresenceState);
  const [connectionState, setConnectionState] = useState<ConnectionState>("connecting");
  const [retryCount, setRetryCount] = useState(0);
  const socketRef = useRef<WebSocket | null>(null);

  const mergeEvents = useCallback((incoming: Event[]) => {
    if (incoming.length === 0) return;
    setEventsBySeq((prev) => {
      const next = new Map(prev);
      for (const event of incoming) next.set(event.seq, event);
      return next;
    });
  }, []);

  useEffect(() => {
    let cancelled = false;
    let socket: WebSocket | null = null;
    let retryTimer: ReturnType<typeof setTimeout> | null = null;
    let retryDelay = INITIAL_RETRY_MS;
    let attempts = 0;

    setEventsBySeq(new Map());
    setPresence(emptyPresenceState());

    function connect() {
      if (cancelled) return;

      setConnectionState("connecting");
      socket = new WebSocket(`${relayConfig.wsUrl}/v1/sessions/${sessionId}/stream`);
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
          mergeEvents([result.value]);
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
  }, [sessionId, mergeEvents]);

  const events = useMemo(
    () => [...eventsBySeq.values()].toSorted((a, b) => a.seq - b.seq),
    [eventsBySeq],
  );

  function sendPresence(active: boolean) {
    const socket = socketRef.current;
    if (!socket || socket.readyState !== WebSocket.OPEN) return;
    socket.send(JSON.stringify({ type: "presence.typing", active }));
  }

  return { events, presence, connectionState, retryCount, sendPresence, mergeEvents };
}
