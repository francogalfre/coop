"use client";

import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { HarnessLogo } from "@/components/harness-logo";
import type { TimelineItem } from "../../types";
import { Row } from "../timeline-row-shell";
import { RedactionChip } from "./redaction-chip";

export function AgentTextRow({
  item,
  harness,
  onReply,
  highlighted,
}: {
  item: Extract<TimelineItem, { kind: "agent-text" }>;
  harness?: string;
  onReply?: (seq: number) => void;
  highlighted?: boolean;
}) {
  return (
    <Row
      ts={item.ts}
      seq={item.seq}
      onReply={onReply}
      highlighted={highlighted}
      rail={
        <span className="relative z-10 grid size-6 place-items-center rounded-md border border-agent/30 bg-agent/10 text-agent">
          <HarnessLogo harness={harness} size={13} />
        </span>
      }
    >
      <div className="prose-timeline max-w-none text-sm text-foreground/90 leading-relaxed">
        <ReactMarkdown remarkPlugins={[remarkGfm]}>{item.text}</ReactMarkdown>
      </div>
      <RedactionChip redactions={item.redactions} truncated={item.truncated} />
    </Row>
  );
}
