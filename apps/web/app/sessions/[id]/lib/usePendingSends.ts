"use client";

import { useCallback, useMemo, useState } from "react";
import type { DeliveryState, MessageItem } from "../types";

export type PendingSend = {
  clientId: string;
  author: string;
  authorAvatarUrl?: string;
  text: string;
  toAgent: boolean;
  anchorSeq?: number;
};

export type PendingPatch = Partial<Pick<MessageItem, "delivery" | "queuePosition">>;

type PendingEntry = PendingSend & { delivery: DeliveryState; queuePosition?: number };

// Optimistic sends have no real seq yet — anchored past every real event so
// they always render as the newest item until reconciled or dropped.
function pendingSeq(size: number, index: number): number {
  return Number.MAX_SAFE_INTEGER - size + index;
}

export function usePendingSends() {
  const [byId, setById] = useState<Map<string, PendingEntry>>(new Map());

  const addPending = useCallback((send: PendingSend) => {
    setById((prev) => new Map(prev).set(send.clientId, { ...send, delivery: "sending" }));
  }, []);

  const updatePending = useCallback((clientId: string, patch: PendingPatch) => {
    setById((prev) => {
      const existing = prev.get(clientId);
      if (!existing) return prev;
      const next = new Map(prev);
      next.set(clientId, { ...existing, ...patch });
      return next;
    });
  }, []);

  const resolvePending = useCallback((clientId: string) => {
    setById((prev) => {
      if (!prev.has(clientId)) return prev;
      const next = new Map(prev);
      next.delete(clientId);
      return next;
    });
  }, []);

  const pendingItems = useMemo<MessageItem[]>(() => {
    const entries = [...byId.values()];
    return entries.map((entry, index) => ({
      kind: "message",
      key: `pending-${entry.clientId}`,
      seq: pendingSeq(entries.length, index),
      ts: new Date().toISOString(),
      author: entry.author,
      authorAvatarUrl: entry.authorAvatarUrl,
      text: entry.text,
      toAgent: entry.toAgent,
      anchorSeq: entry.anchorSeq,
      clientId: entry.clientId,
      delivery: entry.delivery,
      queuePosition: entry.queuePosition,
    }));
  }, [byId]);

  return { pendingItems, addPending, updatePending, resolvePending };
}
