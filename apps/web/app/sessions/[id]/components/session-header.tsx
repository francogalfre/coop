"use client";

import Link from "next/link";
import type { Route } from "next";
import { motion, useReducedMotion } from "motion/react";
import { IconChevronLeft, IconFolder, IconLock, IconUnlock } from "@/components/icons";
import { HarnessLogo } from "@/components/harness-logo";
import { PresenceStack } from "./presence-stack";
import { SessionModeToggle } from "./session-mode-toggle";
import { StatusPill } from "@/components/status-pill";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import type { SessionMeta } from "../types";

export function SessionHeader({
  sessionId,
  meta,
  live,
  agentBusy,
  viewers,
  backTo,
  backLabel,
  displayName,
  isOwner,
  onToggleTakeover,
}: {
  sessionId: string;
  meta: SessionMeta;
  live: boolean;
  agentBusy: boolean;
  viewers: string[];
  backTo: string;
  backLabel: string;
  displayName: string;
  isOwner: boolean;
  onToggleTakeover: () => void;
}) {
  const agentName = meta.harness ?? "agent";
  const takeover = meta.takeover;
  const heldByMe = Boolean(takeover?.active && takeover.by === displayName);
  const heldByOther = Boolean(takeover?.active && !heldByMe);
  const prefersReducedMotion = useReducedMotion();

  return (
    <header className="border-border/70 border-b bg-card/30 backdrop-blur-sm">
      <div className="mx-auto flex max-w-3xl items-center gap-3 px-4 pt-5 sm:px-6">
        <Link
          href={backTo as Route}
          className="inline-flex items-center gap-1 rounded-md py-1 text-xs text-muted-foreground transition-colors hover:text-foreground"
        >
          <IconChevronLeft size={13} />
          {backLabel}
        </Link>
        <div className="ml-auto flex items-center gap-3">
          {viewers.length > 0 && <PresenceStack names={viewers} />}
          {live && <SessionModeToggle sessionId={sessionId} mode={meta.mode ?? "auto"} isOwner={isOwner} />}
          {live && (
            <Button
              size="sm"
              variant={heldByMe ? "default" : "outline"}
              disabled={heldByOther}
              onClick={onToggleTakeover}
              className="h-7 gap-1.5 rounded-full px-2.5 text-xs"
            >
              {heldByMe ? <IconUnlock size={12} /> : <IconLock size={12} />}
              {heldByMe ? "Release" : heldByOther ? `Held by ${takeover?.by}` : "Take over"}
            </Button>
          )}
          <StatusPill tone={live ? "live" : "ended"}>{live ? "Live" : "Ended"}</StatusPill>
        </div>
      </div>

      {takeover?.active && (
        <div
          className={cn(
            "mx-auto flex max-w-3xl items-center gap-1.5 px-4 pb-2 text-xs sm:px-6",
            heldByMe ? "text-human" : "text-muted-foreground",
          )}
        >
          <IconLock size={11} />
          {heldByMe
            ? "You have taken over — the agent's tool calls are paused."
            : `${takeover.by} has taken over — the agent is paused.`}
        </div>
      )}

      <div className="mx-auto max-w-3xl px-4 pt-2 pb-4 sm:px-6">
        <div className="flex flex-wrap items-center gap-x-3 gap-y-2">
          <h1 className="font-display font-semibold text-lg text-foreground tracking-tight">
            {meta.repo ? meta.repo.split("/").slice(-1)[0] : "Session"}
          </h1>

          {live && agentBusy ? (
            <span className="inline-flex items-center gap-1.5 rounded-full bg-agent/12 px-2.5 py-1 text-xs text-agent">
              <HarnessLogo harness={meta.harness} size={12} className="shrink-0" />
              <span className="font-medium">{agentName}</span>
              <span className="text-agent/80">is working</span>
              <span className="flex gap-0.5">
                {[0, 1, 2].map((i) => (
                  <motion.span
                    key={i}
                    className="size-1 rounded-full bg-agent"
                    animate={prefersReducedMotion ? undefined : { opacity: [0.3, 1, 0.3] }}
                    transition={{ duration: 1.1, repeat: Infinity, delay: i * 0.18 }}
                  />
                ))}
              </span>
            </span>
          ) : (
            <span className="inline-flex items-center gap-1.5 rounded-full bg-secondary/60 px-2.5 py-1 text-xs text-muted-foreground">
              <HarnessLogo harness={meta.harness} size={12} className="shrink-0" />
              {agentName} {live ? "idle" : "stopped"}
            </span>
          )}
        </div>

        {(meta.repo || meta.owner) && (
          <p className="mt-1 flex items-center gap-1 truncate text-muted-foreground text-xs">
            {meta.repo && (
              <span className="inline-flex items-center gap-1">
                <IconFolder size={11} className="shrink-0" />
                {meta.repo}
              </span>
            )}
            {meta.repo && meta.owner && <span className="text-muted-foreground/40">·</span>}
            {meta.owner && <span>run by {meta.owner.name}</span>}
          </p>
        )}
      </div>
    </header>
  );
}
