import { IconAlert } from "@/components/icons";
import { PersonAvatar } from "@/components/person-avatar";
import type { TimelineItem } from "../../types";
import { Row } from "../timeline-row-shell";
import { RedactionChip } from "./redaction-chip";

export function BlockedRow({
  item,
  onReply,
  highlighted,
}: {
  item: Extract<TimelineItem, { kind: "blocked" }>;
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
        <span className="relative z-10 grid size-6 place-items-center rounded-md border border-destructive/30 bg-destructive/10 text-destructive">
          <IconAlert size={12} />
        </span>
      }
    >
      <div className="rounded-lg rounded-tl-sm border border-destructive/25 bg-destructive/[0.06] px-3 py-2">
        <div className="mb-0.5 flex items-center gap-2">
          <PersonAvatar name={item.blockedBy} avatarUrl={item.blockedByAvatarUrl} className="size-4 text-3xs" />
          <span className="font-medium text-xs text-destructive">{item.toolName} blocked</span>
          <span className="text-3xs text-muted-foreground">by {item.blockedBy}</span>
        </div>
        <p className="whitespace-pre-wrap text-sm text-foreground/80 leading-relaxed">{item.reason}</p>
        <RedactionChip redactions={item.reasonRedactions} truncated={item.reasonTruncated} />
      </div>
    </Row>
  );
}
