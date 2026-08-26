"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import type { Route } from "next";
import { useParams } from "next/navigation";
import { relayApi, RelayError, type AgentSession } from "@/lib/relay/api";
import { AppHeader } from "@/components/app-header";
import { LiveDot } from "@/components/status-pill";
import { IconFolder } from "@/components/icons";
import { Skeleton } from "@/components/ui/skeleton";
import { SessionCard } from "./components/session-card";
import { InviteButton } from "./components/invite-button";
import { NoAccess } from "./components/no-access";
import { EmptySessions } from "./components/empty-sessions";

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
