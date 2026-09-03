"use client";

import { cn } from "@/lib/utils";

export function QueueStrip({
  agentBusy,
  agentState,
  queueDepth,
}: {
  agentBusy: boolean;
  agentState?: string;
  queueDepth?: number;
}) {
  if (!agentBusy && !queueDepth) return null;

  return (
    <div className="flex min-w-0 items-center gap-1.5 text-2xs text-muted-foreground">
      <span
        className={cn(
          "size-1.5 shrink-0 rounded-full",
          agentBusy ? "animate-pulse-live bg-live" : "bg-muted-foreground/50",
        )}
      />
      <span className="shrink-0">{agentBusy ? "working" : "waiting"}</span>
      {agentBusy && agentState && (
        <>
          <span className="text-muted-foreground/40">·</span>
          <span className="hidden truncate text-foreground/80 sm:inline">{agentState}</span>
        </>
      )}
      {Boolean(queueDepth) && (
        <>
          <span className="text-muted-foreground/40">·</span>
          <span className="shrink-0">{queueDepth} queued</span>
        </>
      )}
    </div>
  );
}
