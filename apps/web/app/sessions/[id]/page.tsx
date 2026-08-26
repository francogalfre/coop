"use client";

import { Suspense, useState } from "react";
import Link from "next/link";
import type { Route } from "next";
import { useParams, useSearchParams } from "next/navigation";
import { useSessionStream, RETRY_WARNING_THRESHOLD } from "@/lib/relay/useSessionStream";
import { EventFeed } from "@/components/EventFeed";
import { SteerInput } from "@/components/SteerInput";
import { ViewerList } from "@/components/ViewerList";
import { PresenceBar } from "@/components/PresenceBar";

const DISPLAY_NAME_KEY = "coop:displayName";

function readStoredName(): string {
  if (typeof window === "undefined") return "";
  try {
    return window.localStorage.getItem(DISPLAY_NAME_KEY) ?? "";
  } catch {
    return "";
  }
}

function backLink(from: string | null): string {
  return from ? `/p/${from}` : "/";
}

function MissingToken({ sessionId, from }: { sessionId: string; from: string | null }) {
  return (
    <main style={{ maxWidth: 800, margin: "0 auto", padding: "1rem" }}>
      <p>
        <Link href={backLink(from) as Route}>← {from ? from : "sessions"}</Link>
      </p>
      <h1>Session {sessionId}</h1>
      <p style={{ color: "#dc2626" }}>
        This link is missing its access token — ask whoever shared it to resend the full link.
      </p>
    </main>
  );
}

function SessionPageContent() {
  const params = useParams<{ id: string }>();
  const searchParams = useSearchParams();
  const sessionId = params.id;
  const token = searchParams.get("token") ?? "";
  const from = searchParams.get("from");
  const [name] = useState(() => readStoredName() || searchParams.get("name") || undefined);

  const { events, presence, connectionState, retryCount, sendPresence } = useSessionStream(sessionId, token, name);

  if (!token) {
    return <MissingToken sessionId={sessionId} from={from} />;
  }

  return (
    <main style={{ maxWidth: 800, margin: "0 auto", padding: "1rem" }}>
      <p>
        <Link href={backLink(from) as Route}>← {from ? from : "sessions"}</Link>
      </p>
      <h1>Session {sessionId}</h1>
      <p style={{ fontSize: "0.85rem", color: "#4b5563" }}>connection: {connectionState}</p>
      {connectionState === "closed" && retryCount >= RETRY_WARNING_THRESHOLD && (
        <p style={{ fontSize: "0.85rem", color: "#dc2626" }}>
          Still can&apos;t connect after {retryCount} attempts — check your link.
        </p>
      )}
      <ViewerList sessionId={sessionId} token={token} />
      <PresenceBar events={events} presence={presence} />
      <SteerInput sessionId={sessionId} token={token} sendPresence={sendPresence} />
      <EventFeed events={events} />
    </main>
  );
}

export default function SessionPage() {
  return (
    <Suspense fallback={null}>
      <SessionPageContent />
    </Suspense>
  );
}
