"use client";

import { memo, useState } from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import {
  IconCheck,
  IconChevronRight,
  IconCode,
  IconEdit,
  IconFile,
  IconReply,
  IconSend,
  IconSpinner,
  IconTerminal,
} from "@/components/icons";
import { HarnessLogo } from "@/components/harness-logo";
import { LiveDot } from "@/components/status-pill";
import { basename, initials, prettyJson, summarizeToolInput, tintFor } from "@/lib/format";
import type { TimelineItem } from "../types";
import { cn } from "@/lib/utils";
import { Row } from "./timeline-row-shell";
import { SteerRequestRow } from "./steer-request-row";

const TOOL_ICONS: Record<string, typeof IconTerminal> = {
  bash: IconTerminal,
  read: IconFile,
  edit: IconEdit,
  write: IconEdit,
  grep: IconCode,
  glob: IconCode,
};

function toolIconFor(name: string) {
  return TOOL_ICONS[name.toLowerCase()] ?? IconCode;
}

function ToolRow({
  item,
  onReply,
  highlighted,
}: {
  item: Extract<TimelineItem, { kind: "tool" }>;
  onReply?: (seq: number) => void;
  highlighted?: boolean;
}) {
  const [open, setOpen] = useState(false);
  const Icon = toolIconFor(item.toolName);
  const summary = summarizeToolInput(item.toolName, item.input);
  const expandable = Boolean(item.input || item.output);

  return (
    <Row
      ts={item.ts}
      seq={item.seq}
      onReply={onReply}
      highlighted={highlighted}
      rail={
        <span
          className={cn(
            "relative z-10 grid size-6 place-items-center rounded-md border bg-card transition-colors",
            item.status === "failed"
              ? "border-destructive/40 text-destructive"
              : "border-border text-tool",
          )}
        >
          {item.status === "running" ? (
            <IconSpinner size={13} className="animate-spin" />
          ) : (
            <Icon size={13} />
          )}
        </span>
      }
    >
      <button
        type="button"
        onClick={() => expandable && setOpen((v) => !v)}
        disabled={!expandable}
        className={cn(
          "flex w-full items-center gap-2 rounded-md text-left transition-colors",
          expandable && "hover:bg-secondary/40 -mx-2 px-2 py-0.5 cursor-pointer",
        )}
      >
        {expandable && (
          <IconChevronRight
            size={12}
            className={cn(
              "shrink-0 text-muted-foreground/50 transition-transform duration-200",
              open && "rotate-90",
            )}
          />
        )}
        <span className="shrink-0 font-medium text-sm text-foreground">{item.toolName}</span>
        <span className="truncate font-mono text-xs text-muted-foreground">{summary}</span>
        {item.status === "failed" ? (
          <span className="ml-auto shrink-0 rounded bg-destructive/15 px-1.5 py-0.5 text-3xs text-destructive">
            failed
          </span>
        ) : item.status === "ok" ? (
          <IconCheck size={12} className="ml-auto shrink-0 text-live/70" />
        ) : null}
      </button>

      {item.files.length > 0 && (
        <div className="mt-1.5 flex flex-wrap gap-1.5">
          {item.files.map((file) => (
            <span
              key={`${file.mode}-${file.path}`}
              className="inline-flex items-center gap-1 rounded bg-secondary/60 px-1.5 py-0.5 font-mono text-2xs text-muted-foreground"
            >
              <span className={cn("size-1 rounded-full", file.mode === "write" ? "bg-human" : "bg-file")} />
              {basename(file.path)}
            </span>
          ))}
        </div>
      )}

      <div
        className={cn(
          "grid transition-all duration-250 ease-[cubic-bezier(0.16,1,0.3,1)]",
          open ? "grid-rows-[1fr] opacity-100" : "grid-rows-[0fr] opacity-0",
        )}
      >
        <div className="min-h-0 overflow-hidden">
          <div className="mt-2 space-y-2">
            {item.input && <CodeBlock label="input" body={prettyJson(item.input)} />}
            {item.output && <CodeBlock label="output" body={prettyJson(item.output)} />}
          </div>
        </div>
      </div>
    </Row>
  );
}

function CodeBlock({ label, body }: { label: string; body: string }) {
  return (
    <div className="overflow-hidden rounded-lg border border-border/70 bg-background/60">
      <div className="border-border/70 border-b px-3 py-1.5 font-mono text-3xs text-muted-foreground/70 uppercase tracking-wider">
        {label}
      </div>
      <pre className="max-h-72 overflow-auto p-3 font-mono text-xs text-foreground/85 leading-relaxed">
        {body}
      </pre>
    </div>
  );
}

function AgentTextRow({
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
    </Row>
  );
}

function MessageRow({
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
  return (
    <Row
      ts={item.ts}
      seq={item.seq}
      onReply={onReply}
      highlighted={highlighted}
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
      </div>
    </Row>
  );
}

function NoticeRow({
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
      return (
        <MessageRow
          item={item}
          onReply={onReply}
          onJumpToAnchor={onJumpToAnchor}
          highlighted={highlighted}
        />
      );
    case "steer-request":
      return <SteerRequestRow item={item} sessionId={sessionId} isOwner={isOwner} />;
    case "notice":
      return <NoticeRow item={item} onReply={onReply} highlighted={highlighted} />;
  }
});
