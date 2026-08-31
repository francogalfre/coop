"use client";

import { Suspense, useCallback, useEffect, useMemo, useState } from "react";
import dynamic from "next/dynamic";
import { useParams, useSearchParams } from "next/navigation";
import { toast } from "sonner";
import { parseEvent } from "@coop/protocol";
import { useSession } from "@/lib/auth/auth-client";
import { useSessionStream, RETRY_WARNING_THRESHOLD } from "@/lib/relay/useSessionStream";
import { relayApi } from "@/lib/relay/api";
import { buildTimeline } from "./lib/build-timeline";
import { useTypingNames } from "./lib/useTypingNames";
import { SessionHeader } from "./components/session-header";
import { Timeline } from "./components/timeline";
import { Composer } from "./components/composer";
import { ViewTabs, type SessionViewTab } from "./components/view-tabs";

const PtyTerminal = dynamic(() => import("./components/pty-terminal").then((m) => m.PtyTerminal), {
  ssr: false,
});

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
  const [viewTab, setViewTab] = useState<SessionViewTab>("timeline");
  const [terminalMounted, setTerminalMounted] = useState(false);

  useEffect(() => {
    if (viewTab === "terminal") setTerminalMounted(true);
  }, [viewTab]);

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

  const handleSend = useCallback(
    async (text: string, toAgent: boolean) => {
      if (!toAgent) {
        await relayApi.sendTeamMessage(sessionId, text, replyingToSeq ?? undefined);
        setReplyingToSeq(null);
        return;
      }

      const result = await relayApi.sendMessage(sessionId, text);
      if (result.status === "pending") {
        toast("Waiting for the owner to approve.");
      } else if (result.queued > 1) {
        toast(`Queued — ${result.queued} messages ahead of yours.`);
      }
    },
    [sessionId, replyingToSeq],
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
        isOwner={isOwner}
        onToggleTakeover={handleToggleTakeover}
      />

      {connectionState === "closed" && retryCount >= RETRY_WARNING_THRESHOLD && (
        <div className="border-destructive/25 border-b bg-destructive/10 px-4 py-2 text-xs text-destructive sm:px-6">
          Lost connection to the relay — retried {retryCount} times.
        </div>
      )}

      <ViewTabs active={viewTab} onChange={setViewTab} />

      <Timeline
        items={items}
        harness={meta.harness}
        sessionId={sessionId}
        isOwner={isOwner}
        onLoadEarlier={loadEarlier}
        hasEarlier={hasEarlier}
        loadingEarlier={loadingEarlier}
        visible={viewTab === "timeline"}
        onReply={setReplyingToSeq}
      />
      {terminalMounted && (
        <PtyTerminal sessionId={sessionId} heldByMe={heldByMe} visible={viewTab === "terminal"} />
      )}

      <Composer
        displayName={displayName}
        disabled={!live}
        typingNames={typingNames}
        takeoverHeldBy={takeoverHeldBy}
        onSend={handleSend}
        onTypingChange={sendPresence}
        replyingToSeq={replyingToSeq ?? undefined}
        onDismissReply={() => setReplyingToSeq(null)}
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
