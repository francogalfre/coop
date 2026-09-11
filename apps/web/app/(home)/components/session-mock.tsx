"use client";

import type { ReactNode } from "react";
import { motion, useReducedMotion } from "motion/react";
import { HarnessLogo } from "@/components/harness-logo";
import { PersonAvatar } from "@/components/person-avatar";
import { StatusPill } from "@/components/status-pill";
import { IconAgent, IconLock, IconSend, IconSpinner } from "@/components/icons";

const rail =
  "grid size-6 shrink-0 place-items-center rounded-md border border-border bg-card text-[13px]";

function MockRow({
  delay,
  time,
  leading,
  children,
  bare = false,
  className,
}: {
  delay: number;
  time?: string;
  leading?: ReactNode;
  children: ReactNode;
  bare?: boolean;
  className?: string;
}) {
  const reduce = useReducedMotion();
  return (
    <motion.div
      initial={reduce ? false : { opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={
        reduce
          ? { duration: 0 }
          : { duration: 0.45, delay, ease: [0.16, 1, 0.3, 1] }
      }
      className={bare ? className : "flex items-start gap-3"}
    >
      {time && (
        <time className="hidden w-14 shrink-0 pt-0.5 text-right font-mono text-2xs text-muted-foreground/50 tabular-nums sm:block">
          {time}
        </time>
      )}
      {leading}
      <div className="min-w-0 flex-1">{children}</div>
    </motion.div>
  );
}

/**
 * The signature of the page: a staged play-through of a live session, built
 * from the same row vocabulary the real session viewer uses — agent turns in
 * blue, human steering in amber, a real takeover pausing the next tool call.
 */
export function SessionMock() {
  return (
    <div
      aria-hidden
      className="overflow-hidden rounded-xl border border-border bg-card text-left shadow-2xl shadow-black/50"
    >
      <div className="flex items-center gap-3 border-border border-b px-4 py-3">
        <span className="grid size-7 shrink-0 place-items-center rounded-md border border-border bg-background text-foreground/85">
          <HarnessLogo harness="claude-code" size={14} />
        </span>
        <div className="min-w-0 flex-1">
          <p className="truncate font-display font-medium text-[13px] text-foreground">
            coop/api
          </p>
          <p className="truncate font-mono text-2xs text-muted-foreground">
            claude-code · 3 watching
          </p>
        </div>
        <div className="flex -space-x-1.5">
          <PersonAvatar
            name="Mara"
            className="size-6 text-3xs ring-2 ring-card"
          />
          <PersonAvatar
            name="Jonas"
            className="size-6 text-3xs ring-2 ring-card"
          />
        </div>
        <StatusPill tone="live">Live</StatusPill>
      </div>

      <div className="space-y-3 px-4 py-4 sm:px-5">
        <MockRow
          delay={0.2}
          time="14:32:08"
          leading={
            <span className={`${rail} text-agent`}>
              <IconAgent size={13} />
            </span>
          }
        >
          <p className="pt-0.5 text-[13px] text-foreground/90 leading-relaxed">
            Tracing the N+1 in the sessions list — starting with the query.
          </p>
        </MockRow>

        <MockRow
          delay={0.55}
          time="14:32:21"
          leading={
            <span className={`${rail} text-agent`}>
              <IconSpinner size={13} className="animate-spin" />
            </span>
          }
        >
          <div className="flex items-center gap-2 pt-0.5">
            <span className="shrink-0 font-medium text-[13px] text-foreground">
              bash
            </span>
            <span className="min-w-0 truncate font-mono text-xs text-muted-foreground">
              psql -c &quot;EXPLAIN ANALYZE SELECT …&quot;
            </span>
            <span className="ml-auto shrink-0 rounded bg-agent/15 px-1.5 py-0.5 text-3xs text-agent">
              running
            </span>
          </div>
        </MockRow>

        <MockRow
          delay={0.9}
          time="14:33:02"
          leading={
            <PersonAvatar name="Mara" className="size-6 shrink-0 text-3xs" />
          }
        >
          <div className="rounded-lg rounded-tl-sm border border-human/25 bg-human/[0.07] px-3 py-2">
            <div className="mb-0.5 flex items-center gap-2">
              <span className="font-medium text-xs text-human">Mara</span>
              <span className="inline-flex items-center gap-1 rounded bg-human/15 px-1.5 py-0.5 text-3xs text-human/90">
                <IconSend size={9} />
                to agent
              </span>
            </div>
            <p className="text-[13px] text-foreground/90 leading-relaxed">
              Use the shared db.Pool — don&apos;t open a client per request.
            </p>
          </div>
        </MockRow>

        <MockRow
          delay={1.25}
          time="14:33:10"
          leading={
            <span className={`${rail} text-agent`}>
              <IconAgent size={13} />
            </span>
          }
        >
          <p className="pt-0.5 text-[13px] text-foreground/90 leading-relaxed">
            Good call — switching the handler to the shared pool.
          </p>
        </MockRow>

        <MockRow
          delay={1.6}
          time="14:33:44"
          leading={
            <span className={`${rail} text-muted-foreground`}>
              <IconLock size={13} />
            </span>
          }
        >
          <div className="flex flex-wrap items-center gap-x-2 gap-y-1 pt-0.5">
            <span className="shrink-0 rounded bg-tool/15 px-1.5 py-0.5 font-mono text-3xs text-tool">
              tool.blocked
            </span>
            <p className="text-[13px] text-foreground/80">
              Jonas took over — the next tool call waits for them.
            </p>
          </div>
        </MockRow>
      </div>

      <MockRow
        delay={1.95}
        bare
        className="flex items-center gap-2 border-border border-t px-4 py-3 sm:px-5"
      >
        <span className="min-w-0 flex-1 truncate text-[13px] text-muted-foreground/60">
          Steer the agent, reply to the room, or take over…
        </span>
        <span className="grid size-6 shrink-0 place-items-center rounded-md border border-border bg-card text-muted-foreground">
          <IconSend size={12} />
        </span>
      </MockRow>
    </div>
  );
}
