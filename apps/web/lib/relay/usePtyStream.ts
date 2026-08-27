"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { relayConfig } from "./config";
import type { ConnectionState } from "./useSessionStream";

const INITIAL_RETRY_MS = 1000;
const MAX_RETRY_MS = 15000;

type PtyFrame =
  | { type: "pty.output"; data: string }
  | { type: "pty.input"; data: string }
  | { type: "pty.resize"; cols: number; rows: number };

// native toBase64/fromBase64 aren't in our TS lib target yet, so feature-detect via a widened type
type Base64Encodable = Uint8Array & { toBase64?: () => string };
type Base64Decodable = typeof Uint8Array & { fromBase64?: (data: string) => Uint8Array };

const CHUNK_SIZE = 0x8000;

function encodeBase64(bytes: Uint8Array): string {
  const fast = (bytes as Base64Encodable).toBase64;
  if (fast) return fast.call(bytes);

  let binary = "";
  for (let i = 0; i < bytes.length; i += CHUNK_SIZE) {
    binary += String.fromCharCode(...bytes.subarray(i, i + CHUNK_SIZE));
  }
  return btoa(binary);
}

function decodeBase64(data: string): Uint8Array {
  const fast = (Uint8Array as Base64Decodable).fromBase64;
  if (fast) return fast(data);

  const binary = atob(data);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i);
  return bytes;
}

export function usePtyStream(sessionId: string): {
  connectionState: ConnectionState;
  onOutput: (cb: (bytes: Uint8Array) => void) => () => void;
  onResize: (cb: (cols: number, rows: number) => void) => () => void;
  sendInput: (bytes: Uint8Array) => void;
  sendResize: (cols: number, rows: number) => void;
} {
  const [connectionState, setConnectionState] = useState<ConnectionState>("connecting");
  const socketRef = useRef<WebSocket | null>(null);
  const outputListeners = useRef(new Set<(bytes: Uint8Array) => void>());
  const resizeListeners = useRef(new Set<(cols: number, rows: number) => void>());

  const onOutput = useCallback((cb: (bytes: Uint8Array) => void) => {
    outputListeners.current.add(cb);
    return () => outputListeners.current.delete(cb);
  }, []);

  const onResize = useCallback((cb: (cols: number, rows: number) => void) => {
    resizeListeners.current.add(cb);
    return () => resizeListeners.current.delete(cb);
  }, []);

  useEffect(() => {
    let cancelled = false;
    let socket: WebSocket | null = null;
    let retryTimer: ReturnType<typeof setTimeout> | null = null;
    let retryDelay = INITIAL_RETRY_MS;

    function connect() {
      if (cancelled) return;

      setConnectionState("connecting");
      socket = new WebSocket(`${relayConfig.wsUrl}/v1/sessions/${sessionId}/pty`);
      socketRef.current = socket;

      socket.addEventListener("open", () => {
        if (cancelled) return;
        retryDelay = INITIAL_RETRY_MS;
        setConnectionState("open");
      });

      socket.addEventListener("message", (message) => {
        if (cancelled) return;

        let parsed: PtyFrame;
        try {
          parsed = JSON.parse(message.data as string);
        } catch {
          return;
        }

        if (parsed.type === "pty.output") {
          const bytes = decodeBase64(parsed.data);
          for (const cb of outputListeners.current) cb(bytes);
        } else if (parsed.type === "pty.resize") {
          for (const cb of resizeListeners.current) cb(parsed.cols, parsed.rows);
        }
      });

      socket.addEventListener("error", () => {
        if (cancelled) return;
        setConnectionState("error");
      });

      socket.addEventListener("close", () => {
        if (cancelled) return;
        socketRef.current = null;
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
  }, [sessionId]);

  const sendInput = useCallback((bytes: Uint8Array) => {
    const socket = socketRef.current;
    if (!socket || socket.readyState !== WebSocket.OPEN) return;
    socket.send(JSON.stringify({ type: "pty.input", data: encodeBase64(bytes) }));
  }, []);

  const sendResize = useCallback((cols: number, rows: number) => {
    const socket = socketRef.current;
    if (!socket || socket.readyState !== WebSocket.OPEN) return;
    socket.send(JSON.stringify({ type: "pty.resize", cols, rows }));
  }, []);

  return { connectionState, onOutput, onResize, sendInput, sendResize };
}
