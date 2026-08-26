"use client";

import { signOut, useSession } from "@/lib/auth/auth-client";

export function AuthStatus() {
  const { data, isPending } = useSession();

  if (isPending) return null;

  if (!data?.user) {
    return (
      <a href="/login" style={{ fontSize: "0.85rem" }}>
        Sign in
      </a>
    );
  }

  return (
    <div style={{ display: "flex", alignItems: "center", gap: "0.5rem", fontSize: "0.85rem" }}>
      {data.user.image && (
        <img
          src={data.user.image}
          alt=""
          width={20}
          height={20}
          style={{ borderRadius: "50%" }}
        />
      )}
      <span>{data.user.name}</span>
      <button
        onClick={() => signOut()}
        style={{ border: "none", background: "none", color: "#4b5563", cursor: "pointer", padding: 0 }}
      >
        Sign out
      </button>
    </div>
  );
}
