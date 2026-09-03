"use client";

import { Collapsible as CollapsiblePrimitive } from "radix-ui";
import { IconChevronRight } from "@/components/icons";
import { basename } from "@/lib/format";
import { cn } from "@/lib/utils";
import type { TouchedFile } from "../../types";

function diffLineTone(text: string): "add" | "remove" | "context" {
  if (text.startsWith("+")) return "add";
  if (text.startsWith("-")) return "remove";
  return "context";
}

export function DiffView({ file }: { file: TouchedFile }) {
  const hunks = file.hunks;
  const stats = (file.added !== undefined || file.removed !== undefined) && (
    <span className="font-mono text-3xs">
      {file.added !== undefined && <span className="text-live">+{file.added}</span>}{" "}
      {file.removed !== undefined && <span className="text-destructive">−{file.removed}</span>}
    </span>
  );

  if (!hunks || hunks.length === 0) {
    return (
      <span className="inline-flex items-center gap-1.5 rounded bg-secondary/60 px-1.5 py-0.5 font-mono text-2xs text-muted-foreground">
        <span className={cn("size-1 rounded-full", file.mode === "write" ? "bg-human" : "bg-file")} />
        {basename(file.path)}
        {stats}
      </span>
    );
  }

  return (
    <CollapsiblePrimitive.Root className="overflow-hidden rounded-lg border border-border/70">
      <CollapsiblePrimitive.Trigger className="group flex w-full items-center gap-1.5 bg-secondary/40 px-2 py-1 text-left font-mono text-2xs text-muted-foreground hover:bg-secondary/60">
        <IconChevronRight size={11} className="shrink-0 transition-transform duration-200 group-data-[state=open]:rotate-90" />
        <span className={cn("size-1 shrink-0 rounded-full", file.mode === "write" ? "bg-human" : "bg-file")} />
        <span className="truncate">{basename(file.path)}</span>
        {stats}
      </CollapsiblePrimitive.Trigger>
      <CollapsiblePrimitive.Content className="border-border/70 border-t bg-background/60">
        <pre className="max-h-96 overflow-auto p-2 font-mono text-2xs leading-relaxed">
          {hunks.map((hunk, hunkIndex) => (
            <div key={`${hunk.old_start}-${hunk.new_start}-${hunkIndex}`}>
              <div className="text-agent/70">
                @@ -{hunk.old_start},{hunk.old_lines} +{hunk.new_start},{hunk.new_lines} @@
              </div>
              {hunk.lines.map((line, lineIndex) => {
                const tone = diffLineTone(line.text);
                return (
                  <div
                    key={lineIndex}
                    className={cn(
                      "whitespace-pre-wrap",
                      tone === "add" && "bg-live/10 text-live",
                      tone === "remove" && "bg-destructive/10 text-destructive",
                      tone === "context" && "text-foreground/70",
                    )}
                  >
                    {line.text}
                  </div>
                );
              })}
            </div>
          ))}
        </pre>
      </CollapsiblePrimitive.Content>
    </CollapsiblePrimitive.Root>
  );
}
