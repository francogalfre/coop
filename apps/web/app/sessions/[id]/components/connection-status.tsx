"use client";

import { IconAlert } from "@/components/icons";
import {
  RETRY_WARNING_THRESHOLD,
  type ConnectionState,
} from "@/lib/relay/useSessionStream";
import { cn } from "@/lib/cn";

const BADGE_LABEL: Record<Exclude<ConnectionState, "open">, string> = {
  connecting: "Connecting",
  closed: "Reconnecting",
  error: "Connection issue",
};

const BADGE_TITLE: Record<ConnectionState, string> = {
  connecting: "Connecting to the relay",
  open: "Connected to the relay",
  closed: "Connection lost — reconnecting automatically",
  error: "The relay connection failed",
};

/**
 * Quiet, always-mounted relay connection marker for the header.
 * Steady green dot when healthy; labeled chip only while the socket is down,
 * so a healthy connection never competes with the Live pill for attention.
 */
export function ConnectionBadge({ state }: { state: ConnectionState }) {
  if (state === "open") {
    return (
      <span title={BADGE_TITLE.open} className="inline-flex items-center">
        <span className="size-1.5 rounded-full bg-live/70" />
        <span className="sr-only">{BADGE_TITLE.open}</span>
      </span>
    );
  }

  return (
    <span
      title={BADGE_TITLE[state]}
      className="inline-flex items-center gap-1.5 text-2xs text-muted-foreground"
    >
      <span
        className={cn(
          "size-1.5 animate-pulse rounded-full motion-reduce:animate-none",
          state === "error" ? "bg-destructive/80" : "bg-muted-foreground/70",
        )}
      />
      {BADGE_LABEL[state]}
      <span className="sr-only">{BADGE_TITLE[state]}</span>
    </span>
  );
}

/**
 * Surfaces once the stream has stayed disconnected past the retry threshold.
 * Stays mounted through the reconnect backoff and clears the moment the
 * socket reopens, so it never flashes on and off per attempt.
 */
export function ConnectionBanner({
  state,
  retryCount,
}: {
  state: ConnectionState;
  retryCount: number;
}) {
  if (state === "open" || retryCount < RETRY_WARNING_THRESHOLD) return null;

  return (
    <div
      role="status"
      aria-live="polite"
      className="border-b border-destructive/25 bg-destructive/10 px-4 py-2 sm:px-6"
    >
      <div className="mx-auto flex max-w-3xl items-start gap-2 text-xs text-destructive">
        <IconAlert size={13} className="mt-0.5 shrink-0" />
        <span>
          Lost connection to the relay — reconnecting automatically (attempt{" "}
          {retryCount}). The timeline catches up as soon as it returns.
        </span>
      </div>
    </div>
  );
}
