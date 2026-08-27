"use client";

import Link from "next/link";
import type { Route } from "next";
import { motion } from "motion/react";
import type { AgentSession } from "@/lib/relay/api";
import { LiveDot, MetaChip, StatusPill } from "@/components/status-pill";
import { IconTerminal } from "@/components/icons";
import { HarnessLogo } from "@/components/harness-logo";
import { relativeTime } from "@/lib/format";

export function SessionCard({
  session,
  slug,
  index,
}: {
  session: AgentSession;
  slug: string;
  index: number;
}) {
  const live = session.status === "live";

  return (
    <motion.li
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3, delay: index * 0.04, ease: [0.16, 1, 0.3, 1] }}
    >
      <Link
        href={`/sessions/${session.id}?from=${slug}` as Route}
        className="group block rounded-xl border border-border bg-card/50 p-5 transition-all hover:border-border/80 hover:bg-card"
      >
        <div className="flex items-start gap-4">
          <span
            className={`grid size-10 shrink-0 place-items-center rounded-lg border ${
              live ? "border-live/30 bg-live/10 text-live" : "border-border bg-secondary/50 text-muted-foreground"
            }`}
          >
            <HarnessLogo harness={session.harness} size={17} />
          </span>

          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <p className="truncate font-display font-medium text-md text-foreground">
                {session.repo ? session.repo.split("/").slice(-1)[0] : session.id.slice(0, 18)}
              </p>
              {live && <LiveDot />}
            </div>
            <p className="mt-1 truncate font-mono text-xs text-muted-foreground">
              {session.repo || session.cwd}
            </p>
            <div className="mt-2.5 flex flex-wrap items-center gap-1.5">
              <MetaChip>
                <IconTerminal size={11} />
                {session.harness}
              </MetaChip>
              <MetaChip>
                {live ? "started" : "ran"} {relativeTime(session.started_at)}
              </MetaChip>
            </div>
          </div>

          <StatusPill tone={live ? "live" : "ended"}>{live ? "Live" : "Ended"}</StatusPill>
        </div>
      </Link>
    </motion.li>
  );
}
