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
    <div className="flex items-center gap-1.5 px-1 pb-1.5 text-2xs text-muted-foreground">
      <span
        className={cn(
          "size-1.5 shrink-0 rounded-full",
          agentBusy ? "animate-pulse-live bg-live" : "bg-muted-foreground/50",
        )}
      />
      <span>{agentBusy ? "working" : "waiting for the agent's next step"}</span>
      {agentBusy && agentState && (
        <>
          <span className="text-muted-foreground/40">·</span>
          <span className="truncate text-foreground/80">{agentState}</span>
        </>
      )}
      {Boolean(queueDepth) && (
        <>
          <span className="text-muted-foreground/40">·</span>
          <span>
            {queueDepth} queued
          </span>
        </>
      )}
    </div>
  );
}
