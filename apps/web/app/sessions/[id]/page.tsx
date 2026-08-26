"use client";

import { Suspense, useCallback, useMemo, useState } from "react";
import { useParams, useSearchParams } from "next/navigation";
import Link from "next/link";
import type { Route } from "next";
import { toast } from "sonner";
import { useSession } from "@/lib/auth/auth-client";
import { useSessionStream, RETRY_WARNING_THRESHOLD } from "@/lib/relay/useSessionStream";
import { relayApi } from "@/lib/relay/api";
import { buildTimeline, type MessageItem } from "@/lib/session/timeline";
import { SessionHeader } from "@/components/session/session-header";
import { Timeline } from "@/components/session/timeline";
import { Composer } from "@/components/session/composer";
import { IconAlert } from "@/components/icons";
import { Button } from "@/components/ui/button";

const TYPING_WINDOW_MS = 3000;

function Shell({ children }: { children: React.ReactNode }) {
  return <div className="flex h-dvh flex-col bg-background">{children}</div>;
}

function Blocked({ title, body }: { title: string; body: string }) {
  return (
    <Shell>
      <div className="flex flex-1 items-center justify-center px-6">
        <div className="max-w-sm space-y-3 text-center">
          <span className="mx-auto grid size-11 place-items-center rounded-xl border border-border bg-card text-muted-foreground">
            <IconAlert size={19} />
          </span>
          <h1 className="font-display font-semibold text-[17px] text-foreground">{title}</h1>
          <p className="text-[13.5px] text-muted-foreground leading-relaxed">{body}</p>
          <Button asChild variant="secondary" size="sm" className="mt-1">
            <Link href={"/" as Route}>Back to projects</Link>
          </Button>
        </div>
      </div>
    </Shell>
  );
}

function SessionView() {
  const params = useParams<{ id: string }>();
  const searchParams = useSearchParams();
  const { data: authData } = useSession();

  const sessionId = params.id;
  const token = searchParams.get("token") ?? "";
  const from = searchParams.get("from");

  const displayName = authData?.user?.name ?? "Guest";
  const [localMessages, setLocalMessages] = useState<MessageItem[]>([]);

  const { events, presence, connectionState, retryCount, sendPresence } = useSessionStream(
    sessionId,
    token,
    displayName,
  );

  const { items, meta, agentBusy } = useMemo(() => buildTimeline(events), [events]);

  const merged = useMemo(
    () =>
      [...items, ...localMessages].toSorted(
        (a, b) => new Date(a.ts).getTime() - new Date(b.ts).getTime(),
      ),
    [items, localMessages],
  );

  const typingNames = useMemo(() => {
    const now = Date.now();
    return Object.entries(presence)
      .filter(
        ([name, entry]) => entry.active && now - entry.at < TYPING_WINDOW_MS && name !== displayName,
      )
      .map(([name]) => name);
  }, [presence, displayName]);

  const viewers = useMemo(() => {
    const names = new Set<string>(Object.keys(presence));
    names.add(displayName);
    return [...names];
  }, [presence, displayName]);

  const live = !meta.endedAt;

  const handleSend = useCallback(
    async (text: string, toAgent: boolean) => {
      if (!toAgent) {
        setLocalMessages((prev) => [
          ...prev,
          {
            kind: "message",
            key: `local-${Date.now()}`,
            ts: new Date().toISOString(),
            author: displayName,
            text,
            toAgent: false,
          },
        ]);
        return;
      }

      const result = await relayApi.sendMessage(sessionId, token, displayName, text);
      if (result.queued > 1) toast(`Queued — ${result.queued} messages ahead of yours.`);
    },
    [sessionId, token, displayName],
  );

  if (!token) {
    return (
      <Blocked
        title="This link is missing its token"
        body="Ask whoever shared this session to send the full link, or open it from your project dashboard."
      />
    );
  }

  return (
    <Shell>
      <SessionHeader
        sessionId={sessionId}
        meta={meta}
        live={live}
        agentBusy={live && agentBusy}
        viewers={viewers}
        backTo={from ? `/projects/${from}` : "/"}
        backLabel={from ?? "projects"}
      />

      {connectionState === "closed" && retryCount >= RETRY_WARNING_THRESHOLD && (
        <div className="border-destructive/25 border-b bg-destructive/10 px-4 py-2 text-[12.5px] text-destructive sm:px-6">
          Lost connection to the relay — retried {retryCount} times.
        </div>
      )}

      <Timeline items={merged} />

      <Composer
        displayName={displayName}
        disabled={!live}
        typingNames={typingNames}
        onSend={handleSend}
        onTypingChange={sendPresence}
      />
    </Shell>
  );
}

export default function SessionPage() {
  return (
    <Suspense fallback={<Shell>{null}</Shell>}>
      <SessionView />
    </Suspense>
  );
}
