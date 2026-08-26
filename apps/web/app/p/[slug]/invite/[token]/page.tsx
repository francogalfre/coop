"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import type { Route } from "next";
import { signIn, useSession } from "@/lib/auth/auth-client";
import { relayConfig } from "@/lib/relay/config";

export default function InviteAcceptPage() {
  const params = useParams<{ slug: string; token: string }>();
  const router = useRouter();
  const { data, isPending } = useSession();
  const [status, setStatus] = useState<"idle" | "accepting" | "invalid" | "error">("idle");

  useEffect(() => {
    if (isPending || !data?.user || status !== "idle") return;

    async function accept() {
      setStatus("accepting");
      try {
        const res = await fetch(
          `${relayConfig.httpUrl}/v1/projects/invites/${params.token}/accept`,
          { method: "POST", credentials: "include" },
        );
        if (res.status === 404) {
          setStatus("invalid");
          return;
        }
        if (!res.ok) {
          setStatus("error");
          return;
        }
        const project = (await res.json()) as { slug: string };
        router.push(`/p/${project.slug}` as Route);
      } catch {
        setStatus("error");
      }
    }

    accept();
  }, [isPending, data, status, params.token, router]);

  if (isPending) return null;

  if (!data?.user) {
    return (
      <main style={{ maxWidth: 400, margin: "4rem auto", padding: "1rem", textAlign: "center" }}>
        <h1>Join {params.slug}</h1>
        <p style={{ fontSize: "0.85rem", color: "#4b5563" }}>Sign in to accept this invite.</p>
        <button
          onClick={() =>
            signIn.social({ provider: "github", callbackURL: window.location.pathname })
          }
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

  return (
    <main style={{ maxWidth: 400, margin: "4rem auto", padding: "1rem", textAlign: "center" }}>
      <h1>Join {params.slug}</h1>
      {status === "accepting" || status === "idle" ? (
        <p style={{ fontSize: "0.85rem", color: "#4b5563" }}>Joining…</p>
      ) : status === "invalid" ? (
        <p style={{ color: "#dc2626" }}>This invite is invalid, expired, or already used.</p>
      ) : (
        <p style={{ color: "#dc2626" }}>Could not join the project.</p>
      )}
    </main>
  );
}
