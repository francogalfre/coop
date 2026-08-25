"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useSessionStream } from "@/lib/relay/useSessionStream";
import { EventFeed } from "@/components/EventFeed";
import { SteerInput } from "@/components/SteerInput";

export default function SessionPage() {
  const params = useParams<{ id: string }>();
  const sessionId = params.id;
  const { events, connectionState } = useSessionStream(sessionId);

  return (
    <main style={{ maxWidth: 800, margin: "0 auto", padding: "1rem" }}>
      <p>
        <Link href="/">← sessions</Link>
      </p>
      <h1>Session {sessionId}</h1>
      <p style={{ fontSize: "0.85rem", color: "#4b5563" }}>connection: {connectionState}</p>
      <SteerInput sessionId={sessionId} />
      <EventFeed events={events} />
    </main>
  );
}
