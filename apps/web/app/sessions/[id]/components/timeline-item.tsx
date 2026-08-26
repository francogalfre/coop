"use client";

import { useState } from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import {
  IconCheck,
  IconChevronRight,
  IconCode,
  IconEdit,
  IconFile,
  IconSend,
  IconSpinner,
  IconTerminal,
} from "@/components/icons";
import { HarnessLogo } from "@/components/harness-logo";
import { LiveDot } from "@/components/status-pill";
import { basename, clockTime, initials, prettyJson, summarizeToolInput, tintFor } from "@/lib/format";
import type { TimelineItem } from "../types";
import { cn } from "@/lib/utils";

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

function Row({
  ts,
  rail,
  children,
  className,
}: {
  ts: string;
  rail: React.ReactNode;
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <div className={cn("group relative mx-auto flex max-w-3xl gap-3 px-4 py-2 sm:px-6", className)}>
      <time className="hidden w-14 shrink-0 pt-1 text-right font-mono text-[11px] text-muted-foreground/50 tabular-nums sm:block">
        {clockTime(ts)}
      </time>
      <div className="relative flex w-6 shrink-0 justify-center">
        <span className="absolute top-7 bottom-[-16px] w-px bg-border/60 group-last:hidden" />
        {rail}
      </div>
      <div className="min-w-0 flex-1 pb-1">{children}</div>
    </div>
  );
}

function ToolRow({ item }: { item: Extract<TimelineItem, { kind: "tool" }> }) {
  const [open, setOpen] = useState(false);
  const Icon = toolIconFor(item.toolName);
  const summary = summarizeToolInput(item.toolName, item.input);
  const expandable = Boolean(item.input || item.output);

  return (
    <Row
      ts={item.ts}
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
        <span className="shrink-0 font-medium text-[13px] text-foreground">{item.toolName}</span>
        <span className="truncate font-mono text-[12px] text-muted-foreground">{summary}</span>
        {item.status === "failed" ? (
          <span className="ml-auto shrink-0 rounded bg-destructive/15 px-1.5 py-0.5 text-[10px] text-destructive">
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
              className="inline-flex items-center gap-1 rounded bg-secondary/60 px-1.5 py-0.5 font-mono text-[11px] text-muted-foreground"
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
      <div className="border-border/70 border-b px-3 py-1.5 font-mono text-[10px] text-muted-foreground/70 uppercase tracking-wider">
        {label}
      </div>
      <pre className="max-h-72 overflow-auto p-3 font-mono text-[12px] text-foreground/85 leading-relaxed">
        {body}
      </pre>
    </div>
  );
}

function AgentTextRow({
  item,
  harness,
}: {
  item: Extract<TimelineItem, { kind: "agent-text" }>;
  harness?: string;
}) {
  return (
    <Row
      ts={item.ts}
      rail={
        <span className="relative z-10 grid size-6 place-items-center rounded-md border border-agent/30 bg-agent/10 text-agent">
          <HarnessLogo harness={harness} size={13} />
        </span>
      }
    >
      <div className="prose-timeline max-w-none text-[13.5px] text-foreground/90 leading-relaxed">
        <ReactMarkdown remarkPlugins={[remarkGfm]}>{item.text}</ReactMarkdown>
      </div>
    </Row>
  );
}

function MessageRow({ item }: { item: Extract<TimelineItem, { kind: "message" }> }) {
  return (
    <Row
      ts={item.ts}
      rail={
        <span
          className="relative z-10 grid size-6 place-items-center rounded-full font-medium text-[10px] text-background"
          style={{ background: tintFor(item.author) }}
        >
          {initials(item.author)}
        </span>
      }
    >
      <div className="rounded-lg rounded-tl-sm border border-human/25 bg-human/[0.07] px-3 py-2">
        <div className="mb-0.5 flex items-center gap-2">
          <span className="font-medium text-[12.5px] text-human">{item.author}</span>
          {item.toAgent && (
            <span className="inline-flex items-center gap-1 rounded bg-human/15 px-1.5 py-0.5 text-[10px] text-human/90">
              <IconSend size={9} />
              to agent
            </span>
          )}
        </div>
        <p className="whitespace-pre-wrap text-[13.5px] text-foreground/90 leading-relaxed">{item.text}</p>
      </div>
    </Row>
  );
}

function NoticeRow({ item }: { item: Extract<TimelineItem, { kind: "notice" }> }) {
  return (
    <Row
      ts={item.ts}
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
      <p className="pt-0.5 text-[12.5px] text-muted-foreground">{item.text}</p>
    </Row>
  );
}

export function TimelineRow({ item, harness }: { item: TimelineItem; harness?: string }) {
  switch (item.kind) {
    case "tool":
      return <ToolRow item={item} />;
    case "agent-text":
      return <AgentTextRow item={item} harness={harness} />;
    case "message":
      return <MessageRow item={item} />;
    case "notice":
      return <NoticeRow item={item} />;
  }
}
