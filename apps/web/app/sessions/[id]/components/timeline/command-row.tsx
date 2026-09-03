import type { TimelineItem } from "../../types";
import { Row } from "../timeline-row-shell";

export function CommandRow({
  item,
  onReply,
  highlighted,
}: {
  item: Extract<TimelineItem, { kind: "command" }>;
  onReply?: (seq: number) => void;
  highlighted?: boolean;
}) {
  return (
    <Row
      ts={item.ts}
      seq={item.seq}
      onReply={onReply}
      highlighted={highlighted}
      rail={<span className="relative z-10 mt-2 size-1.5 rounded-full bg-muted-foreground/40" />}
    >
      <p className="pt-0.5 font-mono text-xs text-muted-foreground">
        <span className="text-foreground">{item.author}</span> ran{" "}
        <span className="text-tool">
          /{item.command}
          {item.args ? ` ${item.args}` : ""}
        </span>
      </p>
    </Row>
  );
}
