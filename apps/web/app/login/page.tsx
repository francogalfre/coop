"use client";

import { signIn } from "@/lib/auth/auth-client";

export default function LoginPage() {
  return (
    <main style={{ maxWidth: 400, margin: "4rem auto", padding: "1rem", textAlign: "center" }}>
      <h1>Sign in to coop</h1>
      <p style={{ fontSize: "0.85rem", color: "#4b5563" }}>
        Sign in to watch and steer live agent sessions with your team.
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
