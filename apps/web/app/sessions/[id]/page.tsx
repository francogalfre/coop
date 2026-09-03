"use client";

import { Suspense, useCallback, useEffect, useMemo, useState } from "react";
import { useParams, useSearchParams } from "next/navigation";
import { toast } from "sonner";
import { parseEvent } from "@coop/protocol";
import { useSession } from "@/lib/auth/auth-client";
import { useSessionStream, RETRY_WARNING_THRESHOLD } from "@/lib/relay/useSessionStream";
import { relayApi } from "@/lib/relay/api";
import { basename } from "@/lib/format";
import { buildTimeline } from "./lib/build-timeline";
import { useTypingNames } from "./lib/useTypingNames";
import { usePendingSends } from "./lib/usePendingSends";
import { SessionHeader } from "./components/session-header";
import { Timeline } from "./components/timeline";
import { Composer } from "./components/composer";

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
  const [replyingToSeq, setReplyingToSeq] = useState<number | null>(null);
  const [hasEarlier, setHasEarlier] = useState(true);
  const [loadingEarlier, setLoadingEarlier] = useState(false);

  const { events, presence, connectionState, retryCount, sendPresence, mergeEvents } =
    useSessionStream(sessionId);

  const { items, meta, agentBusy } = useMemo(() => buildTimeline(events), [events]);
  const { pendingItems, addPending, updatePending, resolvePending } = usePendingSends();

  useEffect(() => {
    for (const item of items) {
      if (item.kind === "message" && item.clientId) resolvePending(item.clientId);
    }
  }, [items, resolvePending]);

  const timelineItems = useMemo(() => [...items, ...pendingItems], [items, pendingItems]);

  const agentState = useMemo(() => {
    for (let i = items.length - 1; i >= 0; i -= 1) {
      const item = items[i];
      if (item?.kind === "tool" && item.status === "running") {
        const file = item.files[0];
        return file ? `${item.toolName}(${basename(file.path)})` : item.toolName;
      }
    }
    return undefined;
  }, [items]);

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

  const typingNames = useTypingNames(presence, displayName, TYPING_WINDOW_MS);

  const viewers = useMemo(() => {
    const names = new Set<string>(Object.keys(presence));
    names.add(displayName);
    return [...names];
  }, [presence, displayName]);

  const live = !meta.endedAt;
  const isOwner = Boolean(authData?.user?.id && authData.user.id === meta.owner?.id);
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
        isOwner={isOwner}
        onToggleTakeover={handleToggleTakeover}
      />

      {connectionState === "closed" && retryCount >= RETRY_WARNING_THRESHOLD && (
        <div className="border-destructive/25 border-b bg-destructive/10 px-4 py-2 text-xs text-destructive sm:px-6">
          Lost connection to the relay — retried {retryCount} times.
        </div>
      )}

      <Timeline
        items={timelineItems}
        harness={meta.harness}
        sessionId={sessionId}
        isOwner={isOwner}
        onLoadEarlier={loadEarlier}
        hasEarlier={hasEarlier}
        loadingEarlier={loadingEarlier}
        onReply={setReplyingToSeq}
      />

      <Composer
        sessionId={sessionId}
        displayName={displayName}
        isOwner={isOwner}
        disabled={!live}
        typingNames={typingNames}
        takeoverHeldBy={takeoverHeldBy}
        heldByMe={heldByMe}
        agentBusy={live && agentBusy}
        agentState={agentState}
        capabilities={meta.capabilities}
        onTypingChange={sendPresence}
        replyingToSeq={replyingToSeq ?? undefined}
        onDismissReply={() => setReplyingToSeq(null)}
        onPendingSend={addPending}
        onPendingUpdate={updatePending}
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
