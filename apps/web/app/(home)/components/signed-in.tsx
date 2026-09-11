"use client";

import { useEffect, useState } from "react";
import type { Route } from "next";
import { useRouter } from "next/navigation";
import { motion } from "motion/react";
import { relayApi, type Project } from "@/lib/relay/api";
import { IconPeople } from "@/components/icons";
import { AppHeader } from "@/components/app-header";
import { Skeleton } from "@/components/ui/skeleton";
import { ProjectRow } from "./project-row";
import { CreateProject } from "./create-project";

function EmptyProjects() {
  const steps = [
    "Create a project below — it only needs a name.",
    "In your project directory, run coop attach --project=<slug>.",
    "Start Claude Code, opencode, or pi — the session shows up here, live for the team.",
  ];

  return (
    <div className="rounded-xl border border-border border-dashed bg-card/30 px-6 py-12 text-center sm:px-10">
      <span className="mx-auto mb-3 grid size-11 place-items-center rounded-xl border border-border bg-card text-muted-foreground">
        <IconPeople size={19} />
      </span>
      <h2 className="font-display font-medium text-[15px] text-foreground">
        No projects yet
      </h2>
      <p className="mx-auto mt-1 max-w-sm text-[13px] text-muted-foreground leading-relaxed">
        A project is the room your team&apos;s agent sessions live in. Three
        steps to your first live one:
      </p>
      <ol className="mx-auto mt-6 max-w-sm space-y-3 text-left">
        {steps.map((step, i) => (
          <li key={i} className="flex items-start gap-3">
            <span className="grid size-5 shrink-0 place-items-center rounded-full border border-border bg-card font-mono text-2xs text-muted-foreground">
              {i + 1}
            </span>
            <span className="text-[13px] text-muted-foreground leading-relaxed">
              {step}
            </span>
          </li>
        ))}
      </ol>
    </div>
  );
}

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
            Pick a project to drop into a live session — or create one and bring
            your agents into the room.
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
    </>
  );
}
