"use client";

import { useState } from "react";
import { toast } from "sonner";
import { IconCheck, IconClose, IconSend } from "@/components/icons";
import { Button } from "@/components/ui/button";
import { initials, tintFor } from "@/lib/format";
import { relayApi } from "@/lib/relay/api";
import { Row } from "./timeline-row-shell";
import type { TimelineItem } from "../types";

export function SteerRequestRow({
  item,
  sessionId,
  isOwner,
}: {
  item: Extract<TimelineItem, { kind: "steer-request" }>;
  sessionId: string;
  isOwner: boolean;
}) {
  const [resolving, setResolving] = useState(false);

  async function resolve(decision: "allow" | "deny") {
    setResolving(true);
    try {
      await relayApi.resolveSteer(sessionId, item.requestId, decision);
    } catch {
      toast.error("Failed to resolve the request.");
    } finally {
      setResolving(false);
    }
  }

  if (item.status === "denied") {
    return (
      <Row
        ts={item.ts}
        rail={
          <span
            className="relative z-10 grid size-6 place-items-center rounded-full font-medium text-3xs text-background opacity-50"
            style={{ background: tintFor(item.author) }}
          >
            {initials(item.author)}
          </span>
        }
      >
        <div className="rounded-lg rounded-tl-sm border border-border/60 bg-secondary/30 px-3 py-2 opacity-60">
          <div className="mb-0.5 flex items-center gap-2">
            <span className="font-medium text-xs text-muted-foreground">{item.author}</span>
            <span className="text-3xs text-muted-foreground">
              denied by {item.resolvedBy ?? "the owner"}
            </span>
          </div>
          <p className="whitespace-pre-wrap text-sm text-muted-foreground leading-relaxed">{item.text}</p>
        </div>
      </Row>
    );
  }

  return (
    <Row
      ts={item.ts}
      rail={
        <span
          className="relative z-10 grid size-6 place-items-center rounded-full font-medium text-3xs text-background"
          style={{ background: tintFor(item.author) }}
        >
          {initials(item.author)}
        </span>
      }
    >
      <div className="rounded-lg rounded-tl-sm border border-human/25 bg-human/[0.07] px-3 py-2">
        <div className="mb-0.5 flex items-center gap-2">
          <span className="font-medium text-xs text-human">{item.author}</span>
          {item.status === "pending" ? (
            <span className="inline-flex items-center gap-1 rounded bg-secondary px-1.5 py-0.5 text-3xs text-muted-foreground">
              awaiting approval
            </span>
          ) : (
            <span className="inline-flex items-center gap-1 rounded bg-human/15 px-1.5 py-0.5 text-3xs text-human/90">
              <IconSend size={9} />
              to agent
            </span>
          )}
        </div>
        <p className="whitespace-pre-wrap text-sm text-foreground/90 leading-relaxed">{item.text}</p>

        {item.status === "pending" && isOwner && (
          <div className="mt-2 flex gap-1.5">
            <Button
              size="sm"
              variant="outline"
              disabled={resolving}
              onClick={() => void resolve("allow")}
              className="h-6.5 gap-1 rounded-md px-2 text-2xs"
            >
              <IconCheck size={11} />
              Approve
            </Button>
            <Button
              size="sm"
              variant="outline"
              disabled={resolving}
              onClick={() => void resolve("deny")}
              className="h-6.5 gap-1 rounded-md px-2 text-2xs"
            >
              <IconClose size={11} />
              Deny
            </Button>
          </div>
        )}
      </div>
    </Row>
  );
}
