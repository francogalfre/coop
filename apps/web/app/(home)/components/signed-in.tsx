"use client";

import { useEffect, useState } from "react";
import type { Route } from "next";
import { useRouter } from "next/navigation";
import { motion } from "motion/react";
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
      <main className="mx-auto max-w-5xl px-6 py-14">
        <motion.div
          initial={{ opacity: 0, y: 6 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.35, ease: [0.16, 1, 0.3, 1] }}
          className="mb-8"
        >
          <h1 className="font-display font-semibold text-[26px] text-foreground tracking-tight">
            Projects
          </h1>
          <p className="mt-1.5 text-sm text-muted-foreground">
            A project is where your team&apos;s agent sessions live.
          </p>
        </motion.div>

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
        ) : (
          <ul className="space-y-2.5">
            {projects.map((project, i) => (
              <ProjectRow key={project.id} project={project} index={i} />
            ))}
          </ul>
        )}

        <div className="mt-4">
          <CreateProject onCreated={(p) => router.push(`/projects/${p.slug}` as Route)} />
        </div>
      </main>
    </>
  );
}
