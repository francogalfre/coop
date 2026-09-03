"use client";

import { memo } from "react";
import { IconSpinner } from "@/components/icons";
import type { TimelineItem } from "../types";
import { AgentTextRow } from "./timeline/agent-text-row";
import { BlockedRow } from "./timeline/blocked-row";
import { CommandRow } from "./timeline/command-row";
import { MessageRow } from "./timeline/message-row";
import { NoticeRow } from "./timeline/notice-row";
import { PermissionRow } from "./timeline/permission-row";
import { QuestionRow } from "./timeline/question-row";
import { ToolRow } from "./timeline/tool-row";
import { Row } from "./timeline-row-shell";
import { SteerRequestRow } from "./steer-request-row";

function TurnBoundaryRow({ item }: { item: Extract<TimelineItem, { kind: "turn-start" | "turn-end" }> }) {
  return (
    <Row
      ts={item.ts}
      seq={item.seq}
      rail={
        item.kind === "turn-start" ? (
          <IconSpinner size={11} className="relative z-10 mt-1.5 animate-spin text-muted-foreground/50" />
        ) : (
          <span className="relative z-10 mt-2.5 size-1.5 rounded-full bg-border" />
        )
      }
    >
      <p className="pt-0.5 text-2xs text-muted-foreground/70 uppercase tracking-wider">
        {item.kind === "turn-start" ? "turn started" : "turn ended"}
      </p>
    </Row>
  );
}

export const TimelineRow = memo(function TimelineRow({
  item,
  harness,
  sessionId,
  isOwner,
  onReply,
  onJumpToAnchor,
  highlighted,
}: {
  item: TimelineItem;
  harness?: string;
  sessionId: string;
  isOwner: boolean;
  onReply?: (seq: number) => void;
  onJumpToAnchor?: (seq: number) => void;
  highlighted?: boolean;
}) {
  switch (item.kind) {
    case "tool":
      return <ToolRow item={item} onReply={onReply} highlighted={highlighted} />;
    case "agent-text":
      return <AgentTextRow item={item} harness={harness} onReply={onReply} highlighted={highlighted} />;
    case "message":
      return <MessageRow item={item} onReply={onReply} onJumpToAnchor={onJumpToAnchor} highlighted={highlighted} />;
    case "steer-request":
      return <SteerRequestRow item={item} sessionId={sessionId} isOwner={isOwner} />;
    case "permission":
      return <PermissionRow item={item} sessionId={sessionId} isOwner={isOwner} />;
    case "question":
      return <QuestionRow item={item} sessionId={sessionId} isOwner={isOwner} />;
    case "blocked":
      return <BlockedRow item={item} onReply={onReply} highlighted={highlighted} />;
    case "command":
      return <CommandRow item={item} onReply={onReply} highlighted={highlighted} />;
    case "notice":
      return <NoticeRow item={item} onReply={onReply} highlighted={highlighted} />;
    case "turn-start":
    case "turn-end":
      return <TurnBoundaryRow item={item} />;
  }
});
