"use client";

import { useEffect, useState } from "react";
import type { Route } from "next";
import { useRouter } from "next/navigation";
import { relayApi, type Project } from "@/lib/relay/api";
import { AppHeader } from "@/components/layout/app-header";
import { Container } from "@/components/layout/container";
import { Reveal } from "@/components/ui/reveal";
import { Skeleton } from "@/components/ui/skeleton";
import { ProjectRow } from "./project-row";
import { CreateProject } from "./create-project";
import { EmptyProjects } from "./empty-projects";

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
    <Container>
      <AppHeader />
      <main className="px-6 py-12 sm:px-10 sm:py-16">
        <Reveal className="mb-8">
          <h1 className="font-display font-semibold text-[26px] text-foreground tracking-tight">
            Projects
          </h1>
          <p className="mt-1.5 max-w-md text-sm text-muted-foreground leading-relaxed">
            Pick a project to drop into a live session — or create one and bring
            your agents into the room.
          </p>
        </Reveal>

        {failed ? (
          <p className="rounded-xl border border-destructive/25 bg-destructive/10 px-4 py-3 text-sm text-destructive">
            Could not reach the relay. Is it running?
          </p>
        ) : projects === null ? (
          <div className="space-y-2.5">
            {[0, 1, 2].map((i) => (
              <Skeleton key={i} className="h-[76px] rounded-xl" />
            ))}
          </div>
        ) : projects.length === 0 ? (
          <EmptyProjects />
        ) : (
          <ul className="space-y-2.5">
            {projects.map((project, i) => (
              <ProjectRow key={project.id} project={project} index={i} />
            ))}
          </ul>
        )}

        <div className="mt-4">
          <CreateProject
            onCreated={(p) => router.push(`/projects/${p.slug}` as Route)}
          />
        </div>
      </main>
    </Container>
  );
}
