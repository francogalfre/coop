"use client";

import { useEffect, useRef } from "react";
import { motion } from "motion/react";
import { IconChevronUp, IconSparkles } from "@/components/icons";
import { TimelineRow } from "./timeline-item";
import type { TimelineItem } from "../types";

function EmptyState() {
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-3 px-6 py-20 text-center">
      <span className="grid size-11 place-items-center rounded-xl border border-border bg-card text-muted-foreground">
        <IconSparkles size={19} />
      </span>
      <div className="space-y-1">
        <p className="font-display font-medium text-[15px] text-foreground">Waiting for the agent</p>
        <p className="max-w-xs text-[13px] text-muted-foreground leading-relaxed">
          Nothing has happened yet. Tool calls and messages will stream in here as they occur.
        </p>
      </div>
    </div>
  );
}

export function Timeline({
  items,
  harness,
  onLoadEarlier,
  hasEarlier,
  loadingEarlier,
}: {
  items: TimelineItem[];
  harness?: string;
  onLoadEarlier?: () => Promise<void>;
  hasEarlier?: boolean;
  loadingEarlier?: boolean;
}) {
  const endRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const pinnedRef = useRef(true);

  useEffect(() => {
    const el = containerRef.current;
    if (!el) return;

    function onScroll() {
      if (!el) return;
      pinnedRef.current = el.scrollHeight - el.scrollTop - el.clientHeight < 120;
    }

    el.addEventListener("scroll", onScroll, { passive: true });
    return () => el.removeEventListener("scroll", onScroll);
  }, []);

  useEffect(() => {
    if (pinnedRef.current) endRef.current?.scrollIntoView({ behavior: "smooth", block: "end" });
  }, [items.length]);

  async function handleLoadEarlier() {
    const el = containerRef.current;
    if (!el || !onLoadEarlier) return;

    const prevHeight = el.scrollHeight;
    await onLoadEarlier();

    requestAnimationFrame(() => {
      const el2 = containerRef.current;
      if (!el2) return;
      el2.scrollTop += el2.scrollHeight - prevHeight;
    });
  }

  if (items.length === 0) return <EmptyState />;

  return (
    <div ref={containerRef} className="flex-1 overflow-y-auto py-3">
      {hasEarlier && (
        <div className="flex justify-center pb-2">
          <button
            type="button"
            onClick={handleLoadEarlier}
            disabled={loadingEarlier}
            className="rounded-full border border-border bg-card px-3 py-1 text-[12px] text-muted-foreground transition-colors hover:text-foreground disabled:opacity-50"
          >
            <span className="inline-flex items-center gap-1">
              <IconChevronUp size={12} />
              {loadingEarlier ? "Loading…" : "Load earlier"}
            </span>
          </button>
        </div>
      )}
      {items.map((item, index) => (
        <motion.div
          key={item.key}
          initial={index > items.length - 4 ? { opacity: 0, y: 6 } : false}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.28, ease: [0.16, 1, 0.3, 1] }}
        >
          <TimelineRow item={item} harness={harness} />
        </motion.div>
      ))}
      <div ref={endRef} />
    </div>
  );
}
