"use client";

import { useState } from "react";
import {
  IconCheck,
  IconChevronRight,
  IconCode,
  IconEdit,
  IconFile,
  IconSpinner,
  IconTerminal,
} from "@/components/icons";
import { prettyJson, summarizeToolInput } from "@/lib/format";
import { cn } from "@/lib/utils";
import type { ToolStatus, TimelineItem } from "../../types";
import { Row } from "../timeline-row-shell";
import { CodeBlock } from "./code-block";
import { DiffView } from "./diff-view";
import { RedactionChip } from "./redaction-chip";

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

function matchCountLabel(toolName: string, output?: string): string | null {
  if (!output) return null;
  const name = toolName.toLowerCase();
  if (name === "grep") return `${output.split("\n").filter(Boolean).length} matches`;
  if (name === "glob") return `${output.split("\n").filter(Boolean).length} paths`;
  if (name === "read") return `${output.split("\n").length} lines`;
  return null;
}

function TerminalBlock({ body, status }: { body: string; status: ToolStatus }) {
  return (
    <div className="overflow-hidden rounded-lg border border-border/70 bg-background/80">
      <div className="flex items-center gap-1.5 border-border/70 border-b px-3 py-1.5 font-mono text-3xs text-muted-foreground/70">
        <span
          className={cn(
            "size-1.5 rounded-full",
            status === "failed" ? "bg-destructive" : status === "running" ? "bg-agent" : "bg-live",
          )}
        />
        <span className="uppercase tracking-wider">
          {status === "failed" ? "exit ≠ 0" : status === "running" ? "running" : "exit 0"}
        </span>
      </div>
      <pre className="max-h-72 overflow-auto p-3 font-mono text-xs text-foreground/85 leading-relaxed">{body}</pre>
    </div>
  );
}

type TodoEntry = { text: string; status: string };

function parseTodos(raw: string): TodoEntry[] | null {
  let parsed: unknown;
  try {
    parsed = JSON.parse(raw);
  } catch {
    return null;
  }
  if (!parsed || typeof parsed !== "object") return null;
  const todos = (parsed as Record<string, unknown>).todos;
  if (!Array.isArray(todos)) return null;

  const entries: TodoEntry[] = [];
  for (const todo of todos) {
    if (!todo || typeof todo !== "object") return null;
    const record = todo as Record<string, unknown>;
    const text = typeof record.content === "string" ? record.content : typeof record.text === "string" ? record.text : null;
    if (text === null) return null;
    entries.push({ text, status: typeof record.status === "string" ? record.status : "pending" });
  }
  return entries;
}

function TodoChecklist({ entries }: { entries: TodoEntry[] }) {
  return (
    <ul className="space-y-1">
      {entries.map((entry, index) => (
        <li key={index} className="flex items-center gap-2 text-sm">
          <span
            className={cn(
              "size-2.5 shrink-0 rounded-full border",
              entry.status === "completed" && "border-live bg-live",
              entry.status === "in_progress" && "border-agent bg-agent/60 animate-pulse",
              entry.status !== "completed" && entry.status !== "in_progress" && "border-border bg-transparent",
            )}
          />
          <span className={cn(entry.status === "completed" ? "text-muted-foreground line-through" : "text-foreground/90")}>
            {entry.text}
          </span>
        </li>
      ))}
    </ul>
  );
}

export function ToolRow({
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
  const isBash = item.toolName.toLowerCase() === "bash";
  const todos = item.toolName.toLowerCase() === "todowrite" ? parseTodos(item.input) : null;
  const summary = todos ? `${todos.length} todo${todos.length === 1 ? "" : "s"}` : summarizeToolInput(item.toolName, item.input);
  const matchCount = matchCountLabel(item.toolName, item.output);
  const expandable = Boolean(!todos && (item.input || item.output));

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
            item.status === "failed" ? "border-destructive/40 text-destructive" : "border-border text-tool",
          )}
        >
          {item.status === "running" ? <IconSpinner size={13} className="animate-spin" /> : <Icon size={13} />}
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
            className={cn("shrink-0 text-muted-foreground/50 transition-transform duration-200", open && "rotate-90")}
          />
        )}
        <span className="shrink-0 font-medium text-sm text-foreground">{item.toolName}</span>
        <span className="truncate font-mono text-xs text-muted-foreground">{summary}</span>
        {matchCount && <span className="shrink-0 font-mono text-3xs text-muted-foreground/70">{matchCount}</span>}
        {item.durationMs !== undefined && (
          <span className="shrink-0 font-mono text-3xs text-muted-foreground/70">{item.durationMs}ms</span>
        )}
        {item.status === "failed" ? (
          <span className="ml-auto shrink-0 rounded bg-destructive/15 px-1.5 py-0.5 text-3xs text-destructive">failed</span>
        ) : item.status === "ok" ? (
          <IconCheck size={12} className="ml-auto shrink-0 text-live/70" />
        ) : null}
      </button>

      {todos && (
        <div className="mt-1.5">
          <TodoChecklist entries={todos} />
        </div>
      )}

      <RedactionChip redactions={item.inputRedactions + item.outputRedactions} truncated={item.inputTruncated || item.outputTruncated} />

      {item.files.length > 0 && (
        <div className="mt-1.5 flex flex-wrap gap-1.5">
          {item.files.map((file) => (
            <DiffView key={`${file.mode}-${file.path}`} file={file} />
          ))}
        </div>
      )}

      {expandable && (
        <div
          className={cn(
            "grid transition-all duration-250 ease-[cubic-bezier(0.16,1,0.3,1)]",
            open ? "grid-rows-[1fr] opacity-100" : "grid-rows-[0fr] opacity-0",
          )}
        >
          <div className="min-h-0 overflow-hidden">
            <div className="mt-2 space-y-2">
              {item.input && <CodeBlock label="input" body={prettyJson(item.input)} />}
              {item.output &&
                (isBash ? (
                  <TerminalBlock body={item.output} status={item.status} />
                ) : (
                  <CodeBlock label="output" body={prettyJson(item.output)} />
                ))}
            </div>
          </div>
        </div>
      )}
    </Row>
  );
}
