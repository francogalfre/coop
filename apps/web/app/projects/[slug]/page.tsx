"use client";

import { useCallback, useRef, useState } from "react";
import Link from "next/link";
import type { Route } from "next";
import { useParams } from "next/navigation";
import { motion } from "motion/react";
import { relayApi, RelayError, type Agent, type AgentSession } from "@/lib/relay/api";
import { useVisibilityPolling } from "@/lib/hooks/useVisibilityPolling";
import { AppHeader } from "@/components/app-header";
import { LiveDot } from "@/components/status-pill";
import { IconFolder } from "@/components/icons";
import { Skeleton } from "@/components/ui/skeleton";
import { SessionCard } from "./components/session-card";
import { InviteButton } from "./components/invite-button";
import { NoAccess } from "./components/no-access";
import { EmptySessions } from "./components/empty-sessions";
import { AgentRoster } from "./components/agent-roster";
import { ProjectContextPanel } from "./components/project-context-panel";
import { ProjectNotesFeed } from "./components/project-notes-feed";

export default function ProjectPage() {
  const params = useParams<{ slug: string }>();
  const slug = params.slug;

  const [sessions, setSessions] = useState<AgentSession[] | null>(null);
  const [agents, setAgents] = useState<Agent[] | null>(null);
  const [denied, setDenied] = useState(false);
  const [failed, setFailed] = useState(false);
  const latestLoadRef = useRef<symbol | undefined>(undefined);
  const latestAgentsLoadRef = useRef<symbol | undefined>(undefined);

  const load = useCallback(async () => {
    const token = Symbol();
    latestLoadRef.current = token;
    try {
      const data = await relayApi.listSessions(slug);
      if (latestLoadRef.current !== token) return;
      setSessions(data.sessions);
    } catch (error) {
      if (latestLoadRef.current !== token) return;
      if (error instanceof RelayError && error.isMissing) setDenied(true);
      else setFailed(true);
    }
  }, [slug]);

  const loadAgents = useCallback(async () => {
    const token = Symbol();
    latestAgentsLoadRef.current = token;
    try {
      const data = await relayApi.listAgents(slug);
      if (latestAgentsLoadRef.current !== token) return;
      setAgents(data.agents);
    } catch (error) {
      if (latestAgentsLoadRef.current !== token) return;
      if (error instanceof RelayError && error.isMissing) setDenied(true);
      else setFailed(true);
    }
  }, [slug]);

  useVisibilityPolling(load, 8000);
  useVisibilityPolling(loadAgents, 8000);

  const live = sessions?.filter((s) => s.status === "live") ?? [];
  const ended = sessions?.filter((s) => s.status === "ended") ?? [];

  return (
    <>
      <AppHeader />
      <main className="mx-auto max-w-6xl px-6 py-14">
        {denied ? (
          <NoAccess />
        ) : (
          <>
            <motion.div
              initial={{ opacity: 0, y: 6 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.35, ease: [0.16, 1, 0.3, 1] }}
              className="mb-8 flex flex-wrap items-end justify-between gap-3"
            >
              <div>
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                  <IconFolder size={13} />
                  <Link href={"/" as Route} className="transition-colors hover:text-foreground">
                    projects
                  </Link>
                  <span className="text-muted-foreground/40">/</span>
                </div>
                <h1 className="mt-1.5 font-display font-semibold text-[26px] text-foreground tracking-tight">
                  {slug}
                </h1>
              </div>
              <InviteButton slug={slug} />
            </motion.div>

            <div className="grid grid-cols-1 gap-8 lg:grid-cols-[minmax(0,1fr)_340px] lg:items-start">
              <div>
                {agents !== null && agents.length > 0 && (
                  <section className="mb-8">
                    <h2 className="mb-3 font-medium text-xs text-muted-foreground uppercase tracking-wider">
                      Agents
                    </h2>
                    <AgentRoster slug={slug} agents={agents} />
                  </section>
                )}

                {failed ? (
                  <p className="rounded-xl border border-destructive/25 bg-destructive/10 px-4 py-3 text-sm text-destructive">
                    Could not reach the relay. Is it running?
                  </p>
                ) : sessions === null ? (
                  <div className="space-y-2.5">
                    {[0, 1].map((i) => (
                      <Skeleton key={i} className="h-[112px] rounded-xl" />
                    ))}
                  </div>
                ) : sessions.length === 0 ? (
                  <EmptySessions />
                ) : (
                  <div className="space-y-8">
                    {live.length > 0 && (
                      <section>
                        <h2 className="mb-3 flex items-center gap-2 font-medium text-xs text-muted-foreground uppercase tracking-wider">
                          <LiveDot />
                          Live now
                        </h2>
                        <ul className="space-y-2.5">
                          {live.map((s, i) => (
                            <SessionCard key={s.id} session={s} slug={slug} index={i} />
                          ))}
                        </ul>
                      </section>
                    )}

                    {ended.length > 0 && (
                      <section>
                        <h2 className="mb-3 font-medium text-xs text-muted-foreground uppercase tracking-wider">
                          Earlier
                        </h2>
                        <ul className="space-y-2.5">
                          {ended.map((s, i) => (
                            <SessionCard key={s.id} session={s} slug={slug} index={i} />
                          ))}
                        </ul>
                      </section>
                    )}
                  </div>
                )}
              </div>

              <aside className="space-y-6 lg:sticky lg:top-14 lg:mt-7">
                <ProjectContextPanel slug={slug} />
                <ProjectNotesFeed slug={slug} />
              </aside>
            </div>
          </>
        )}
      </main>
    </>
  );
}
