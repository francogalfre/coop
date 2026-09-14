"use client";

import { useCallback, useRef, useState } from "react";
import { useParams } from "next/navigation";
import { relayApi, RelayError, type Agent } from "@/lib/relay/api";
import { useSession } from "@/lib/auth/auth-client";
import { useVisibilityPolling } from "@/lib/hooks/useVisibilityPolling";
import { AppHeader } from "@/components/layout/app-header";
import { Container } from "@/components/layout/container";
import { NoAccess } from "./components/no-access";
import { ProjectHeader } from "./components/project-header";
import { ProjectSessionsSection } from "./components/sessions/project-sessions-section";
import { AgentRoster } from "./components/agent-roster";
import { ProjectContextPanel } from "./components/context/project-context-panel";
import { ProjectNotesFeed } from "./components/context/project-notes-feed";

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
    <Container>
      <AppHeader />
      <main className="px-6 py-12 sm:px-10 sm:py-16">
        {denied ? (
          <NoAccess />
        ) : (
          <>
            <ProjectHeader slug={slug} />

            <div className="grid grid-cols-1 gap-8 lg:grid-cols-[minmax(0,1fr)_320px] lg:items-start">
              <div>
                {agents !== null && agents.length > 0 && (
                  <section className="mb-8">
                    <h2 className="mb-3 font-medium text-muted-foreground text-xs uppercase tracking-wider">
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
    </Container>
  );
}
