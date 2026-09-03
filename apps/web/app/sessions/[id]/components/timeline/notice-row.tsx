import { LiveDot } from "@/components/status-pill";
import type { TimelineItem } from "../../types";
import { Row } from "../timeline-row-shell";

export function NoticeRow({
  item,
  onReply,
  highlighted,
}: {
  item: Extract<TimelineItem, { kind: "notice" }>;
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
        item.tone === "start" ? (
          <span className="relative z-10 mt-2 grid place-items-center">
            <LiveDot />
          </span>
        ) : (
          <span className="relative z-10 mt-2.5 size-1.5 rounded-full bg-border" />
        )
      }
    >
      <p className="pt-0.5 text-xs text-muted-foreground">{item.text}</p>
    </Row>
  );
}
