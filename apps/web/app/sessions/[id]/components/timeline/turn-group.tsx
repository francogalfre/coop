"use client";

import type { ReactNode } from "react";
import { Collapsible as CollapsiblePrimitive } from "radix-ui";
import { IconChevronRight, IconSpinner } from "@/components/icons";
import { MetaChip } from "@/components/status-pill";
import { cn } from "@/lib/utils";
import type { TimelineGroup } from "../../lib/timeline/group-turns";
import type { TimelineItem } from "../../types";

function formatDuration(ms: number): string {
  return ms < 1000 ? `${ms}ms` : `${(ms / 1000).toFixed(1)}s`;
}

function formatTokens(inputTokens?: number, outputTokens?: number): string | null {
  const total = (inputTokens ?? 0) + (outputTokens ?? 0);
  if (total <= 0) return null;
  return total >= 1000 ? `${(total / 1000).toFixed(1)}k tok` : `${total} tok`;
}

export function TurnGroup({
  group,
  renderItem,
}: {
  group: Extract<TimelineGroup, { kind: "turn" }>;
  renderItem: (item: TimelineItem) => ReactNode;
}) {
  const { start, end, items } = group;
  const tokens = end ? formatTokens(end.usage?.inputTokens, end.usage?.outputTokens) : null;

  return (
    <CollapsiblePrimitive.Root defaultOpen className="mx-auto max-w-3xl px-4 sm:px-6">
      <div className="relative flex gap-3 py-1">
        <span className="w-14 shrink-0 sm:block" />
        <div className="relative flex w-6 shrink-0 justify-center">
          <span className="absolute top-3 bottom-0 w-px bg-border/60" />
        </div>
        <CollapsiblePrimitive.Trigger className="group flex min-w-0 flex-1 items-center gap-1.5 rounded-md py-1 text-left text-muted-foreground hover:text-foreground">
          <IconChevronRight size={11} className="shrink-0 transition-transform duration-200 group-data-[state=open]:rotate-90" />
          <span className="text-2xs uppercase tracking-wider">turn{start.turnId ? ` ${start.turnId}` : ""}</span>
          {end ? (
            <span className="flex items-center gap-1.5">
              {end.durationMs !== undefined && <MetaChip>{formatDuration(end.durationMs)}</MetaChip>}
              {tokens && <MetaChip>{tokens}</MetaChip>}
              {end.usage?.costUsd !== undefined && <MetaChip>${end.usage.costUsd.toFixed(3)}</MetaChip>}
            </span>
          ) : (
            <span className="inline-flex items-center gap-1 text-3xs">
              <IconSpinner size={10} className="animate-spin" />
              in progress
            </span>
          )}
        </CollapsiblePrimitive.Trigger>
      </div>
      <CollapsiblePrimitive.Content className={cn("space-y-0")}>
        {items.map((item) => renderItem(item))}
      </CollapsiblePrimitive.Content>
    </CollapsiblePrimitive.Root>
  );
}
