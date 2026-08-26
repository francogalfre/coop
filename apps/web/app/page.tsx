"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import type { Route } from "next";
import { useRouter } from "next/navigation";
import { motion } from "motion/react";
import { toast } from "sonner";
import { signIn, useSession } from "@/lib/auth/auth-client";
import { relayApi, RelayError, type Project } from "@/lib/relay/api";
import { AppHeader } from "@/components/app-header";
import { IconAgent, IconChevronRight, IconGithub, IconPlus, IconSpinner } from "@/components/icons";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { relativeTime } from "@/lib/format";

function slugify(value: string) {
  return value
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9\s-]/g, "")
    .replace(/\s+/g, "-")
    .slice(0, 40);
}

function SignedOut() {
  return (
    <main className="mx-auto flex min-h-dvh max-w-2xl flex-col items-center justify-center px-6 text-center">
      <motion.div
        initial={{ opacity: 0, y: 12 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, ease: [0.16, 1, 0.3, 1] }}
        className="space-y-6"
      >
        <span className="inline-flex items-center gap-2 rounded-full border border-border bg-card px-3 py-1.5 text-[12px] text-muted-foreground">
          <IconAgent size={13} />
          Multiplayer coding agents
        </span>

        <h1 className="text-balance font-display font-semibold text-[42px] text-foreground leading-[1.05] tracking-tight sm:text-[56px]">
          Your agent, <span className="text-muted-foreground">everyone&apos;s screen.</span>
        </h1>

        <p className="mx-auto max-w-md text-balance text-[15px] text-muted-foreground leading-relaxed">
          One person runs a coding agent. Teammates open a link and watch it work live — reading
          every tool call, messaging it, and steering it together.
        </p>

        <Button
          size="lg"
          onClick={() => void signIn.social({ provider: "github" })}
          className="h-11 gap-2 rounded-xl px-5 text-[14px]"
        >
          <IconGithub size={17} />
          Continue with GitHub
        </Button>
      </motion.div>
    </main>
  );
}

function ProjectRow({ project, index }: { project: Project; index: number }) {
  return (
    <motion.li
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3, delay: index * 0.04, ease: [0.16, 1, 0.3, 1] }}
    >
      <Link
        href={`/projects/${project.slug}` as Route}
        className="group flex items-center gap-3 rounded-xl border border-border bg-card/50 px-4 py-3.5 transition-all hover:border-border/80 hover:bg-card"
      >
        <div className="min-w-0 flex-1">
          <p className="truncate font-display font-medium text-[15px] text-foreground">
            {project.name}
          </p>
          <p className="truncate font-mono text-[12px] text-muted-foreground">
            {project.slug} · created {relativeTime(project.created_at)}
          </p>
        </div>
        <IconChevronRight
          size={16}
          className="shrink-0 text-muted-foreground/50 transition-transform group-hover:translate-x-0.5 group-hover:text-foreground"
        />
      </Link>
    </motion.li>
  );
}

function CreateProject({ onCreated }: { onCreated: (p: Project) => void }) {
  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");
  const [busy, setBusy] = useState(false);

  async function create() {
    const slug = slugify(name);
    if (!name.trim() || !slug) return;

    setBusy(true);
    try {
      onCreated(await relayApi.createProject(name.trim(), slug));
    } catch (error) {
      toast.error(
        error instanceof RelayError && error.status === 500
          ? "That slug may already be taken."
          : "Could not create the project.",
      );
    } finally {
      setBusy(false);
    }
  }

  if (!open) {
    return (
      <Button
        variant="secondary"
        onClick={() => setOpen(true)}
        className="h-10 w-full gap-2 rounded-xl border border-border border-dashed bg-transparent text-[13.5px] text-muted-foreground hover:bg-card hover:text-foreground"
      >
        <IconPlus size={15} />
        New project
      </Button>
    );
  }

  return (
    <motion.div
      initial={{ opacity: 0, height: 0 }}
      animate={{ opacity: 1, height: "auto" }}
      className="overflow-hidden rounded-xl border border-border bg-card/50 p-3"
    >
      <Input
        autoFocus
        value={name}
        placeholder="Project name"
        onChange={(e) => setName(e.target.value)}
        onKeyDown={(e) => e.key === "Enter" && void create()}
        className="h-9 border-0 bg-transparent px-1 text-[14px] shadow-none focus-visible:ring-0"
      />
      <div className="mt-2 flex items-center gap-2 border-border/60 border-t pt-2">
        <span className="flex-1 truncate font-mono text-[11.5px] text-muted-foreground">
          {name.trim() ? `/projects/${slugify(name)}` : "…"}
        </span>
        <Button variant="ghost" size="sm" onClick={() => setOpen(false)} className="h-8 text-[12.5px]">
          Cancel
        </Button>
        <Button
          size="sm"
          onClick={() => void create()}
          disabled={busy || !name.trim()}
          className="h-8 gap-1.5 text-[12.5px]"
        >
          {busy && <IconSpinner size={13} className="animate-spin" />}
          Create
        </Button>
      </div>
    </motion.div>
  );
}

function SignedIn() {
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

export default function Home() {
  const { data, isPending } = useSession();

  if (isPending) return <div className="min-h-dvh bg-background" />;
  return data?.user ? <SignedIn /> : <SignedOut />;
}
