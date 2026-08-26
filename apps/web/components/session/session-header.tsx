"use client";

import Link from "next/link";
import type { Route } from "next";
import { motion } from "motion/react";
import { IconAgent, IconChevronLeft, IconFolder } from "@/components/icons";
import { PresenceStack } from "@/components/presence-stack";
import { MetaChip, StatusPill } from "@/components/status-pill";
import type { SessionMeta } from "@/lib/session/timeline";

export function SessionHeader({
  sessionId,
  meta,
  live,
  agentBusy,
  viewers,
  backTo,
  backLabel,
}: {
  sessionId: string;
  meta: SessionMeta;
  live: boolean;
  agentBusy: boolean;
  viewers: string[];
  backTo: string;
  backLabel: string;
}) {
  const agentName = meta.harness ?? "agent";

  return (
    <header className="border-border/70 border-b bg-card/30 backdrop-blur-sm">
      <div className="mx-auto flex max-w-3xl items-center gap-3 px-4 pt-3 sm:px-6">
        <Link
          href={backTo as Route}
          className="inline-flex items-center gap-1 rounded-md py-1 text-[12.5px] text-muted-foreground transition-colors hover:text-foreground"
        >
          <IconChevronLeft size={13} />
          {backLabel}
        </Link>
        <div className="ml-auto flex items-center gap-3">
          {viewers.length > 0 && <PresenceStack names={viewers} />}
          <StatusPill tone={live ? "live" : "ended"}>{live ? "Live" : "Ended"}</StatusPill>
        </div>
      </div>

      <div className="mx-auto flex max-w-3xl flex-wrap items-center gap-x-3 gap-y-2 px-4 pt-2 pb-3.5 sm:px-6">
        <h1 className="font-display font-semibold text-[19px] text-foreground tracking-tight">
          {meta.repo ? meta.repo.split("/").slice(-1)[0] : "Session"}
        </h1>

        {live && agentBusy ? (
          <span className="inline-flex items-center gap-1.5 rounded-full bg-agent/12 px-2.5 py-1 text-[12px] text-agent">
            <IconAgent size={12} />
            <span className="font-medium">{agentName}</span>
            <span className="text-agent/80">is working</span>
            <span className="flex gap-0.5">
              {[0, 1, 2].map((i) => (
                <motion.span
                  key={i}
                  className="size-1 rounded-full bg-agent"
                  animate={{ opacity: [0.3, 1, 0.3] }}
                  transition={{ duration: 1.1, repeat: Infinity, delay: i * 0.18 }}
                />
              ))}
            </span>
          </span>
        ) : (
          <span className="inline-flex items-center gap-1.5 rounded-full bg-secondary/60 px-2.5 py-1 text-[12px] text-muted-foreground">
            <IconAgent size={12} />
            {agentName} {live ? "idle" : "stopped"}
          </span>
        )}

        <div className="flex flex-wrap items-center gap-1.5">
          {meta.repo && (
            <MetaChip>
              <IconFolder size={11} />
              {meta.repo}
            </MetaChip>
          )}
          {meta.owner && <MetaChip>run by {meta.owner}</MetaChip>}
          <MetaChip className="hidden sm:inline-flex">{sessionId.slice(0, 16)}</MetaChip>
        </div>
      </div>
    </header>
  );
}
