"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { relayApi, RelayError, type AgentSession } from "@/lib/relay/api";
import { useVisibilityPolling } from "@/lib/hooks/useVisibilityPolling";
import { LiveDot } from "@/components/ui/status-pill";
import { IconFolder } from "@/components/icons";
import { Skeleton } from "@/components/ui/skeleton";
import { SessionCard } from "./session-card";
import { EmptySessions } from "./empty-sessions";
import { AttachOnboarding } from "./attach-onboarding";
import { resolveProjectRole, type ProjectRole } from "./resolve-role";

type ProjectSessionsSectionProps = {
  slug: string;
  userId?: string;
};

export function ProjectSessionsSection({
  slug,
  userId,
}: ProjectSessionsSectionProps) {
  const [sessions, setSessions] = useState<AgentSession[] | null>(null);
  const [failed, setFailed] = useState(false);
  const [role, setRole] = useState<ProjectRole | null>(null);
  const latestLoadRef = useRef<symbol | undefined>(undefined);

  const load = useCallback(async () => {
    const token = Symbol();
    latestLoadRef.current = token;
    try {
      const data = await relayApi.listSessions(slug);
      if (latestLoadRef.current !== token) return;
      setSessions(data.sessions);
      setFailed(false);
    } catch (error) {
      if (latestLoadRef.current !== token) return;
      if (error instanceof RelayError && error.isMissing) return;
      setFailed(true);
    }
  }, [slug]);

  useVisibilityPolling(load, 8000);

  const live = sessions?.filter((s) => s.status === "live") ?? [];
  const ended = sessions?.filter((s) => s.status === "ended") ?? [];
  const needsOnboarding = sessions !== null && live.length === 0;

  useEffect(() => {
    setRole(null);
  }, [slug]);

  useEffect(() => {
    if (!needsOnboarding || role !== null || userId === undefined) return;

    let cancelled = false;

    async function resolveRole() {
      try {
        const { projects } = await relayApi.listProjects();
        if (cancelled) return;
        setRole(resolveProjectRole(projects, slug, userId));
      } catch {
        if (!cancelled) setRole("member");
      }
    }

    void resolveRole();

    return () => {
      cancelled = true;
    };
  }, [needsOnboarding, role, slug, userId]);

  const resolvedRole: ProjectRole = role ?? "member";

  if (failed) {
    return (
      <p className="rounded-xl border border-destructive/25 bg-destructive/10 px-4 py-3 text-sm text-destructive">
        Could not reach the relay. Is it running?
      </p>
    );
  }

  if (sessions === null) {
    return (
      <div className="space-y-2.5">
        {[0, 1].map((i) => (
          <Skeleton key={i} className="h-[112px] rounded-xl" />
        ))}
      </div>
    );
  }

  if (sessions.length === 0) {
    return <EmptySessions slug={slug} role={resolvedRole} />;
  }

  return (
    <div className="space-y-8">
      {live.length === 0 && (
        <AttachOnboarding slug={slug} role={resolvedRole} compact />
      )}

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
          <h2 className="mb-3 flex items-center gap-2 font-medium text-xs text-muted-foreground uppercase tracking-wider">
            <IconFolder size={13} />
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
  );
}
