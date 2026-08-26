"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import type { Route } from "next";
import { useRouter } from "next/navigation";
import { signIn, useSession } from "@/lib/auth/auth-client";
import { relayConfig } from "@/lib/relay/config";

type Project = {
  id: number;
  name: string;
  slug: string;
  created_by: string;
  created_at: string;
};

function LoggedOut() {
  return (
    <main style={{ maxWidth: 400, margin: "4rem auto", padding: "1rem", textAlign: "center" }}>
      <h1>coop</h1>
      <p style={{ fontSize: "0.85rem", color: "#4b5563" }}>
        Multiplayer sessions for coding agents. One person runs an agent; teammates watch it live,
        steer it, and take it over.
      </p>
      <button
        onClick={() => signIn.social({ provider: "github" })}
        style={{
          padding: "0.6rem 1.2rem",
          borderRadius: "4px",
          border: "1px solid #d1d5db",
          background: "#111827",
          color: "#fff",
          cursor: "pointer",
        }}
      >
        Continue with GitHub
      </button>
    </main>
  );
}

function CreateProjectForm({ onCreated }: { onCreated: (project: Project) => void }) {
  const [name, setName] = useState("");
  const [slug, setSlug] = useState("");
  const [status, setStatus] = useState<"idle" | "creating" | "error">("idle");
  const [error, setError] = useState("");

  async function create() {
    if (!name.trim() || !slug.trim()) return;
    setStatus("creating");
    setError("");
    try {
      const res = await fetch(`${relayConfig.httpUrl}/v1/projects`, {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, slug }),
      });
      if (!res.ok) {
        setError("Could not create project.");
        setStatus("error");
        return;
      }
      const project = (await res.json()) as Project;
      onCreated(project);
    } catch {
      setError("Could not reach the relay.");
      setStatus("error");
    }
  }

  return (
    <div style={{ border: "1px solid #d1d5db", borderRadius: "4px", padding: "0.75rem", margin: "1rem 0" }}>
      <input
        aria-label="project name"
        placeholder="project name"
        value={name}
        onChange={(e) => setName(e.target.value)}
        style={{ display: "block", width: "100%", marginBottom: "0.5rem", boxSizing: "border-box" }}
      />
      <input
        aria-label="project slug"
        placeholder="project-slug"
        value={slug}
        onChange={(e) => setSlug(e.target.value)}
        style={{ display: "block", width: "100%", marginBottom: "0.5rem", boxSizing: "border-box" }}
      />
      <button onClick={create} disabled={status === "creating" || !name.trim() || !slug.trim()}>
        Create project
      </button>
      {status === "error" && <span style={{ marginLeft: "0.5rem", color: "#dc2626" }}>{error}</span>}
    </div>
  );
}

function LoggedIn() {
  const router = useRouter();
  const [projects, setProjects] = useState<Project[] | null>(null);
  const [error, setError] = useState(false);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      try {
        const res = await fetch(`${relayConfig.httpUrl}/v1/projects`, { credentials: "include" });
        if (!res.ok) {
          if (!cancelled) setError(true);
          return;
        }
        const data = (await res.json()) as { projects: Project[] };
        if (!cancelled) setProjects(data.projects);
      } catch {
        if (!cancelled) setError(true);
      }
    }

    load();
    return () => {
      cancelled = true;
    };
  }, []);

  function handleCreated(project: Project) {
    router.push(`/p/${project.slug}` as Route);
  }

  return (
    <main style={{ maxWidth: 800, margin: "0 auto", padding: "1rem" }}>
      <h1>coop</h1>
      {error ? (
        <p style={{ color: "#dc2626" }}>Could not reach the relay at {relayConfig.httpUrl}.</p>
      ) : projects === null ? (
        <p style={{ fontSize: "0.85rem", color: "#4b5563" }}>Loading projects…</p>
      ) : projects.length === 0 ? (
        <p style={{ fontSize: "0.85rem", color: "#4b5563" }}>You&apos;re not in a project yet.</p>
      ) : (
        <ul>
          {projects.map((project) => (
            <li key={project.id}>
              <Link href={`/p/${project.slug}` as Route}>{project.name}</Link>{" "}
              <span style={{ fontSize: "0.8rem", color: "#6b7280" }}>({project.slug})</span>
            </li>
          ))}
        </ul>
      )}
      <h2 style={{ fontSize: "1rem" }}>New project</h2>
      <CreateProjectForm onCreated={handleCreated} />
    </main>
  );
}

export default function Home() {
  const { data, isPending } = useSession();

  if (isPending) return null;
  if (!data?.user) return <LoggedOut />;
  return <LoggedIn />;
}
