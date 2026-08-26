"use client";

import { useEffect, useState } from "react";
import type { Route } from "next";
import { useRouter } from "next/navigation";
import { relayApi, type Project } from "@/lib/relay/api";
import { AppHeader } from "@/components/app-header";
import { Skeleton } from "@/components/ui/skeleton";
import { ProjectRow } from "./project-row";
import { CreateProject } from "./create-project";

export function SignedIn() {
  const router = useRouter();
  const [projects, setProjects] = useState<Project[] | null>(null);
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    let cancelled = false;

    relayApi
      .listProjects()
      .then((data) => !cancelled && setProjects(data.projects))
      .catch(() => !cancelled && setFailed(true));

    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <>
      <AppHeader />
      <main className="mx-auto max-w-5xl px-5 py-10">
        <div className="mb-6">
          <h1 className="font-display font-semibold text-[26px] text-foreground tracking-tight">
            Projects
          </h1>
          <p className="mt-1 text-[13.5px] text-muted-foreground">
            A project is where your team&apos;s agent sessions live.
          </p>
        </div>

        {failed ? (
          <p className="rounded-xl border border-destructive/25 bg-destructive/10 px-4 py-3 text-[13.5px] text-destructive">
            Could not reach the relay. Is it running?
          </p>
        ) : projects === null ? (
          <div className="space-y-2">
            {[0, 1, 2].map((i) => (
              <Skeleton key={i} className="h-[68px] rounded-xl" />
            ))}
          </div>
        ) : (
          <ul className="space-y-2">
            {projects.map((project, i) => (
              <ProjectRow key={project.id} project={project} index={i} />
            ))}
          </ul>
        )}

        <div className="mt-3">
          <CreateProject onCreated={(p) => router.push(`/projects/${p.slug}` as Route)} />
        </div>
      </main>
    </>
  );
}
