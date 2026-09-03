import { useState } from "react";
import { IconChevronRight, IconFile, IconReply, IconSend } from "@/components/icons";
import { PersonAvatar } from "@/components/person-avatar";
import { cn } from "@/lib/utils";
import type { DeliveryState, TimelineItem } from "../../types";
import { Row } from "../timeline-row-shell";

const DELIVERY_LABEL: Record<DeliveryState, string> = {
  sending: "sending…",
  queued: "queued",
  delivered: "delivered",
  seen: "seen by agent",
  dropped: "dropped — not delivered",
};

const COLLAPSE_THRESHOLD = 320;

function ProjectContextRow({
  item,
  onReply,
  highlighted,
}: {
  item: Extract<TimelineItem, { kind: "message" }>;
  onReply?: (seq: number) => void;
  highlighted?: boolean;
}) {
  const collapsible = item.text.length > COLLAPSE_THRESHOLD;
  const [open, setOpen] = useState(!collapsible);

  return (
    <Row
      ts={item.ts}
      seq={item.seq}
      onReply={onReply}
      highlighted={highlighted}
      rail={
        <span className="relative z-10 grid size-6 place-items-center rounded-md border border-border bg-card text-file">
          <IconFile size={13} />
        </span>
      }
    >
      <div className="rounded-lg border border-border bg-card/50 px-3 py-2">
        <button
          type="button"
          onClick={() => collapsible && setOpen((v) => !v)}
          disabled={!collapsible}
          className={cn("flex w-full items-center gap-1.5 text-left", collapsible && "cursor-pointer")}
        >
          {collapsible && (
            <IconChevronRight
              size={11}
              className={cn("shrink-0 text-muted-foreground/50 transition-transform duration-200", open && "rotate-90")}
            />
          )}
          <span className="font-medium text-2xs text-muted-foreground uppercase tracking-wider">
            Project context shared (v{item.projectContextVersion})
          </span>
        </button>
        {open && (
          <p className="mt-1.5 whitespace-pre-wrap text-sm text-foreground/80 leading-relaxed">{item.text}</p>
        )}
      </div>
    </Row>
  );
}

export function MessageRow({
  item,
  onReply,
  highlighted,
  onJumpToAnchor,
}: {
  item: Extract<TimelineItem, { kind: "message" }>;
  onReply?: (seq: number) => void;
  highlighted?: boolean;
  onJumpToAnchor?: (seq: number) => void;
}) {
  if (item.projectContextVersion !== undefined) {
    return <ProjectContextRow item={item} onReply={onReply} highlighted={highlighted} />;
  }

  return (
    <Row
      ts={item.ts}
      seq={item.seq}
      onReply={onReply}
      highlighted={highlighted}
      rail={<PersonAvatar name={item.author} avatarUrl={item.authorAvatarUrl} className="relative z-10 size-6 text-3xs" />}
    >
      <div className="rounded-lg rounded-tl-sm border border-human/25 bg-human/[0.07] px-3 py-2">
        <div className="mb-0.5 flex items-center gap-2">
          <span className="font-medium text-xs text-human">{item.author}</span>
          {item.toAgent && (
            <span className="inline-flex items-center gap-1 rounded bg-human/15 px-1.5 py-0.5 text-3xs text-human/90">
              <IconSend size={9} />
              to agent
            </span>
          )}
        </div>
        {item.anchorSeq !== undefined && (
          <button
            type="button"
            onClick={() => onJumpToAnchor?.(item.anchorSeq!)}
            className="mb-1 flex items-center gap-1 text-2xs text-human/80 hover:underline"
          >
            <IconReply size={10} />
            replying to step {item.anchorSeq}
          </button>
        )}
        <p className="whitespace-pre-wrap text-sm text-foreground/90 leading-relaxed">{item.text}</p>
        {item.delivery && <p className="mt-1 text-2xs text-muted-foreground/70">{DELIVERY_LABEL[item.delivery]}</p>}
      </div>
    </Row>
  );
}
