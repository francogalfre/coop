"use client";

import { useCallback, useRef, useState } from "react";
import Link from "next/link";
import type { Route } from "next";
import { useParams } from "next/navigation";
import { motion } from "motion/react";
import { relayApi, RelayError, type Agent } from "@/lib/relay/api";
import { useSession } from "@/lib/auth/auth-client";
import { useVisibilityPolling } from "@/lib/hooks/useVisibilityPolling";
import { AppHeader } from "@/components/app-header";
import { IconFolder } from "@/components/icons";
import { InviteButton } from "./components/invite-button";
import { NoAccess } from "./components/no-access";
import { ProjectSessionsSection } from "./components/project-sessions-section";
import { AgentRoster } from "./components/agent-roster";
import { ProjectContextPanel } from "./components/project-context-panel";
import { ProjectNotesFeed } from "./components/project-notes-feed";

export default function ProjectPage() {
  const params = useParams<{ slug: string }>();
  const slug = params.slug;
  const { data: sessionData } = useSession();

  const [agents, setAgents] = useState<Agent[] | null>(null);
  const [denied, setDenied] = useState(false);
  const latestAgentsLoadRef = useRef<symbol | undefined>(undefined);

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
    }
  }, [slug]);

  useVisibilityPolling(loadAgents, 8000);

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
                  <Link
                    href={"/" as Route}
                    className="transition-colors hover:text-foreground"
                  >
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

                <ProjectSessionsSection
                  slug={slug}
                  userId={sessionData?.user?.id}
                />
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
