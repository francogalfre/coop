"use client";

import { useState } from "react";
import { toast } from "sonner";
import { IconCheck, IconClose, IconLock } from "@/components/icons";
import { Button } from "@/components/ui/button";
import { relayConfig } from "@/lib/relay/config";
import { prettyJson } from "@/lib/format";
import type { TimelineItem } from "../../types";
import { Row } from "../timeline-row-shell";
import { CodeBlock } from "./code-block";
import { RedactionChip } from "./redaction-chip";

// relayApi (lib/relay/api.ts) has no permission-resolution endpoint yet — this
// posts directly against the shape the relay is expected to expose, so the
// row still typechecks and works once that endpoint lands.
async function resolvePermission(sessionId: string, requestId: string, decision: "allow" | "deny") {
  const res = await fetch(
    `${relayConfig.httpUrl}/v1/sessions/${encodeURIComponent(sessionId)}/permissions/${encodeURIComponent(requestId)}/resolve`,
    {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ decision }),
    },
  );
  if (!res.ok) throw new Error(`Request failed (${res.status}).`);
}

export function PermissionRow({
  item,
  sessionId,
  isOwner,
}: {
  item: Extract<TimelineItem, { kind: "permission" }>;
  sessionId: string;
  isOwner: boolean;
}) {
  const [resolving, setResolving] = useState(false);

  async function resolve(decision: "allow" | "deny") {
    setResolving(true);
    try {
      await resolvePermission(sessionId, item.requestId, decision);
    } catch {
      toast.error("Failed to resolve the permission request.");
    } finally {
      setResolving(false);
    }
  }

  return (
    <Row
      ts={item.ts}
      seq={item.seq}
      rail={
        <span className="relative z-10 grid size-6 place-items-center rounded-md border border-tool/30 bg-tool/10 text-tool">
          <IconLock size={12} />
        </span>
      }
    >
      <div className="rounded-lg rounded-tl-sm border border-tool/25 bg-tool/[0.06] px-3 py-2">
        <div className="mb-1 flex items-center gap-2">
          <span className="font-medium text-xs text-tool">Permission requested</span>
          <span className="font-mono text-2xs text-muted-foreground">{item.toolName}</span>
          {item.status !== "pending" && (
            <span className="ml-auto text-3xs text-muted-foreground">
              {item.status} {item.resolvedBy && `by ${item.resolvedBy}`}
            </span>
          )}
        </div>
        {item.input && <CodeBlock label="input" body={prettyJson(item.input)} />}
        <RedactionChip redactions={item.inputRedactions} truncated={item.inputTruncated} />
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
              Allow
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
