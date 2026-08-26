"use client";

import { Suspense, useCallback, useEffect, useMemo, useState } from "react";
import { useParams, useSearchParams } from "next/navigation";
import { toast } from "sonner";
import { parseEvent } from "@coop/protocol";
import { useSession } from "@/lib/auth/auth-client";
import { useSessionStream, RETRY_WARNING_THRESHOLD } from "@/lib/relay/useSessionStream";
import { relayApi } from "@/lib/relay/api";
import { buildTimeline } from "./lib/build-timeline";
import type { MessageItem } from "./types";
import { SessionHeader } from "./components/session-header";
import { Timeline } from "./components/timeline";
import { Composer } from "./components/composer";
import { ViewTabs, type SessionViewTab } from "./components/view-tabs";
import { PtyTerminal } from "./components/pty-terminal";

const TYPING_WINDOW_MS = 3000;

function Shell({ children }: { children: React.ReactNode }) {
  return <div className="flex h-dvh flex-col bg-background">{children}</div>;
}

function SessionView() {
  const params = useParams<{ id: string }>();
  const searchParams = useSearchParams();
  const { data: authData } = useSession();

  const sessionId = params.id;
  const from = searchParams.get("from");

  const displayName = authData?.user?.name ?? "Guest";
  const [localMessages, setLocalMessages] = useState<MessageItem[]>([]);
  const [hasEarlier, setHasEarlier] = useState(true);
  const [loadingEarlier, setLoadingEarlier] = useState(false);
  const [viewTab, setViewTab] = useState<SessionViewTab>("timeline");

  const { events, presence, connectionState, retryCount, sendPresence, mergeEvents } =
    useSessionStream(sessionId);

  const { items, meta, agentBusy } = useMemo(() => buildTimeline(events), [events]);

  const loadEarlier = useCallback(async () => {
    const oldestSeq = events[0]?.seq;
    if (oldestSeq === undefined || loadingEarlier) return;

    setLoadingEarlier(true);
    try {
      const page = await relayApi.listEvents(sessionId, oldestSeq);
      const earlier = page.events.flatMap((raw) => {
        const result = parseEvent(raw);
        return result.ok ? [result.value] : [];
      });

      mergeEvents(earlier);
      setHasEarlier(page.has_more);
    } catch {
      // leave hasEarlier as-is so the control stays available to retry
    } finally {
      setLoadingEarlier(false);
    }
  }, [sessionId, events, loadingEarlier, mergeEvents]);

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
  const confirmedHeldByMe = Boolean(meta.takeover?.active && meta.takeover.by === displayName);
  const [pendingTakeover, setPendingTakeover] = useState<boolean | null>(null);

  useEffect(() => {
    if (pendingTakeover !== null && confirmedHeldByMe === pendingTakeover) setPendingTakeover(null);
  }, [confirmedHeldByMe, pendingTakeover]);

  const heldByMe = pendingTakeover ?? confirmedHeldByMe;
  const takeoverHeldBy = meta.takeover?.active && !heldByMe ? meta.takeover.by : undefined;

  const handleToggleTakeover = useCallback(async () => {
    const next = !heldByMe;
    setPendingTakeover(next);
    try {
      await relayApi.setTakeover(sessionId, next);
    } catch {
      setPendingTakeover(null);
      toast.error("Failed to update takeover.");
    }
  }, [sessionId, heldByMe]);

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

      const result = await relayApi.sendMessage(sessionId, text);
      if (result.queued > 1) toast(`Queued — ${result.queued} messages ahead of yours.`);
    },
    [sessionId, displayName],
  );

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
        displayName={displayName}
        onToggleTakeover={handleToggleTakeover}
      />

      {connectionState === "closed" && retryCount >= RETRY_WARNING_THRESHOLD && (
        <div className="border-destructive/25 border-b bg-destructive/10 px-4 py-2 text-[12.5px] text-destructive sm:px-6">
          Lost connection to the relay — retried {retryCount} times.
        </div>
      )}

      <ViewTabs active={viewTab} onChange={setViewTab} />

      {viewTab === "timeline" ? (
        <Timeline
          items={merged}
          harness={meta.harness}
          onLoadEarlier={loadEarlier}
          hasEarlier={hasEarlier}
          loadingEarlier={loadingEarlier}
        />
      ) : (
        <PtyTerminal sessionId={sessionId} heldByMe={heldByMe} />
      )}

      <Composer
        displayName={displayName}
        disabled={!live}
        typingNames={typingNames}
        takeoverHeldBy={takeoverHeldBy}
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
