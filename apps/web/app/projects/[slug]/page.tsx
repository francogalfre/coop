"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import type { Route } from "next";
import { useParams } from "next/navigation";
import { motion } from "motion/react";
import { toast } from "sonner";
import { relayApi, RelayError, type AgentSession } from "@/lib/relay/api";
import { AppHeader } from "@/components/app-header";
import { LiveDot, MetaChip, StatusPill } from "@/components/status-pill";
import {
  IconAgent,
  IconAlert,
  IconCheck,
  IconCopy,
  IconFolder,
  IconLink,
  IconSpinner,
  IconTerminal,
} from "@/components/icons";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { relativeTime } from "@/lib/format";

function SessionCard({
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
        className="group block rounded-xl border border-border bg-card/50 p-4 transition-all hover:border-border/80 hover:bg-card"
      >
        <div className="flex items-start gap-3">
          <span
            className={`grid size-9 shrink-0 place-items-center rounded-lg border ${
              live ? "border-live/30 bg-live/10 text-live" : "border-border bg-secondary/50 text-muted-foreground"
            }`}
          >
            <IconAgent size={17} />
          </span>

          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <p className="truncate font-display font-medium text-[14.5px] text-foreground">
                {session.repo ? session.repo.split("/").slice(-1)[0] : session.id.slice(0, 18)}
              </p>
              {live && <LiveDot />}
            </div>
            <p className="mt-0.5 truncate font-mono text-[11.5px] text-muted-foreground">
              {session.repo || session.cwd}
            </p>
            <div className="mt-2 flex flex-wrap items-center gap-1.5">
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

function InviteButton({ slug }: { slug: string }) {
  const [url, setUrl] = useState("");
  const [busy, setBusy] = useState(false);
  const [copied, setCopied] = useState(false);

  async function createInvite() {
    setBusy(true);
    try {
      const { token } = await relayApi.createInvite(slug);
      setUrl(`${window.location.origin}/projects/${slug}/invite/${token}`);
    } catch (error) {
      toast.error(
        error instanceof RelayError && error.isMissing
          ? "Only the project owner can invite people."
          : "Could not create an invite.",
      );
    } finally {
      setBusy(false);
    }
  }

  async function copy() {
    try {
      await navigator.clipboard.writeText(url);
      setCopied(true);
      setTimeout(() => setCopied(false), 1800);
    } catch {
      toast.error("Could not copy — select the link manually.");
    }
  }

  if (!url) {
    return (
      <Button
        variant="secondary"
        size="sm"
        onClick={() => void createInvite()}
        disabled={busy}
        className="h-9 gap-1.5 text-[13px]"
      >
        {busy ? <IconSpinner size={14} className="animate-spin" /> : <IconLink size={14} />}
        Invite teammate
      </Button>
    );
  }

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.97 }}
      animate={{ opacity: 1, scale: 1 }}
      className="flex items-center gap-1.5 rounded-lg border border-border bg-card p-1 pl-3"
    >
      <span className="max-w-[240px] truncate font-mono text-[11.5px] text-muted-foreground">
        {url.replace(/^https?:\/\//, "")}
      </span>
      <Button size="sm" onClick={() => void copy()} className="h-7 gap-1.5 text-[12px]">
        {copied ? <IconCheck size={12} /> : <IconCopy size={12} />}
        {copied ? "Copied" : "Copy"}
      </Button>
    </motion.div>
  );
}

function NoAccess() {
  return (
    <div className="mx-auto max-w-sm space-y-3 py-24 text-center">
      <span className="mx-auto grid size-11 place-items-center rounded-xl border border-border bg-card text-muted-foreground">
        <IconAlert size={19} />
      </span>
      <h1 className="font-display font-semibold text-[17px] text-foreground">
        You don&apos;t have access to this project
      </h1>
      <p className="text-[13.5px] text-muted-foreground leading-relaxed">
        Ask a member to send you an invite link.
      </p>
      <Button asChild variant="secondary" size="sm" className="mt-1">
        <Link href={"/" as Route}>Back to projects</Link>
      </Button>
    </div>
  );
}

function EmptySessions() {
  return (
    <div className="rounded-xl border border-border border-dashed px-6 py-14 text-center">
      <span className="mx-auto mb-3 grid size-11 place-items-center rounded-xl border border-border bg-card text-muted-foreground">
        <IconTerminal size={19} />
      </span>
      <p className="font-display font-medium text-[15px] text-foreground">No sessions yet</p>
      <p className="mx-auto mt-1 max-w-sm text-[13px] text-muted-foreground leading-relaxed">
        Start one from your terminal and it will appear here for the whole team.
      </p>
      <code className="mt-4 inline-block rounded-lg border border-border bg-background px-3 py-2 font-mono text-[12.5px] text-foreground/80">
        coop attach --project=<span className="text-agent">slug</span>
      </code>
    </div>
  );
}

export default function ProjectPage() {
  const params = useParams<{ slug: string }>();
  const slug = params.slug;

  const [sessions, setSessions] = useState<AgentSession[] | null>(null);
  const [denied, setDenied] = useState(false);
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      try {
        const data = await relayApi.listSessions(slug);
        if (!cancelled) setSessions(data.sessions);
      } catch (error) {
        if (cancelled) return;
        if (error instanceof RelayError && error.isMissing) setDenied(true);
        else setFailed(true);
      }
    }

    void load();
    const timer = setInterval(load, 8000);

    return () => {
      cancelled = true;
      clearInterval(timer);
    };
  }, [slug]);

  const live = sessions?.filter((s) => s.status === "live") ?? [];
  const ended = sessions?.filter((s) => s.status === "ended") ?? [];

  return (
    <>
      <AppHeader />
      <main className="mx-auto max-w-5xl px-5 py-10">
        {denied ? (
          <NoAccess />
        ) : (
          <>
            <div className="mb-6 flex flex-wrap items-end justify-between gap-3">
              <div>
                <div className="flex items-center gap-2 text-[12.5px] text-muted-foreground">
                  <IconFolder size={13} />
                  <Link href={"/" as Route} className="transition-colors hover:text-foreground">
                    projects
                  </Link>
                  <span className="text-muted-foreground/40">/</span>
                </div>
                <h1 className="mt-1 font-display font-semibold text-[26px] text-foreground tracking-tight">
                  {slug}
                </h1>
              </div>
              <InviteButton slug={slug} />
            </div>

            {failed ? (
              <p className="rounded-xl border border-destructive/25 bg-destructive/10 px-4 py-3 text-[13.5px] text-destructive">
                Could not reach the relay. Is it running?
              </p>
            ) : sessions === null ? (
              <div className="space-y-2">
                {[0, 1].map((i) => (
                  <Skeleton key={i} className="h-[104px] rounded-xl" />
                ))}
              </div>
            ) : sessions.length === 0 ? (
              <EmptySessions />
            ) : (
              <div className="space-y-7">
                {live.length > 0 && (
                  <section>
                    <h2 className="mb-2.5 flex items-center gap-2 font-medium text-[12px] text-muted-foreground uppercase tracking-wider">
                      <LiveDot />
                      Live now
                    </h2>
                    <ul className="space-y-2">
                      {live.map((s, i) => (
                        <SessionCard key={s.id} session={s} slug={slug} index={i} />
                      ))}
                    </ul>
                  </section>
                )}

                {ended.length > 0 && (
                  <section>
                    <h2 className="mb-2.5 font-medium text-[12px] text-muted-foreground uppercase tracking-wider">
                      Earlier
                    </h2>
                    <ul className="space-y-2">
                      {ended.map((s, i) => (
                        <SessionCard key={s.id} session={s} slug={slug} index={i} />
                      ))}
                    </ul>
                  </section>
                )}
              </div>
            )}
          </>
        )}
      </main>
    </>
  );
}
