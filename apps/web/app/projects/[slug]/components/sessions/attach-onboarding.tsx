"use client";

import Link from "next/link";
import type { Route } from "next";
import { HarnessLogo } from "@/components/brand/harness-logo";
import { CopyButton } from "@/components/ui/copy-button";
import { IconTerminal } from "@/components/icons";
import { buildAttachCommand } from "./attach-command";
import { type ProjectRole } from "./resolve-role";

const HARNESS_IDS = ["claude-code", "opencode", "pi"] as const;
const HARNESS_LABELS: Record<(typeof HARNESS_IDS)[number], string> = {
  "claude-code": "Claude Code",
  opencode: "opencode",
  pi: "pi",
};

function AttachCommand({
  slug,
  className,
}: {
  slug: string;
  className?: string;
}) {
  return (
    <div
      className={`flex items-center gap-2.5 rounded-lg border border-border bg-background px-3 py-2 ${className ?? ""}`}
    >
      <IconTerminal size={14} className="shrink-0 text-muted-foreground/60" />
      <code className="min-w-0 flex-1 truncate font-mono text-[13px] text-foreground/85">
        coop attach --project=<span className="text-agent">{slug}</span>
      </code>
      <CopyButton
        value={buildAttachCommand(slug)}
        label="Copy the attach command"
        className="shrink-0"
      />
    </div>
  );
}

function HarnessChips() {
  return (
    <span className="inline-flex flex-wrap items-center gap-1.5">
      {HARNESS_IDS.map((harness) => (
        <span
          key={harness}
          className="inline-flex items-center gap-1 rounded-md border border-border bg-background px-2 py-1 text-2xs text-muted-foreground"
        >
          <HarnessLogo harness={harness} size={12} />
          {HARNESS_LABELS[harness]}
        </span>
      ))}
    </span>
  );
}

export function AttachOnboarding({
  slug,
  role,
  compact = false,
}: {
  slug: string;
  role: ProjectRole;
  compact?: boolean;
}) {
  if (compact) {
    return (
      <div className="flex flex-col gap-3 rounded-xl border border-border border-dashed bg-card/30 p-4 sm:flex-row sm:items-center sm:gap-4">
        <div className="min-w-0 flex-1">
          <p className="font-display font-medium text-[14px] text-foreground">
            No live sessions right now
          </p>
          <p className="mt-0.5 text-[13px] text-muted-foreground">
            Start one from your terminal and it appears here for the whole team.
          </p>
        </div>
        <AttachCommand slug={slug} className="w-full sm:w-auto sm:max-w-md" />
      </div>
    );
  }

  const owner = role === "owner";

  return (
    <div className="rounded-xl border border-border bg-card/50 p-6 sm:p-8">
      <div className="flex items-start gap-4">
        <span className="grid size-11 shrink-0 place-items-center rounded-xl border border-border bg-background text-muted-foreground">
          <IconTerminal size={19} />
        </span>
        <div className="min-w-0">
          <h2 className="font-display font-medium text-[17px] text-foreground">
            {owner
              ? "Start the first session"
              : "No one is running an agent right now"}
          </h2>
          <p className="mt-1 max-w-lg text-[13.5px] leading-relaxed text-muted-foreground">
            {owner
              ? "Run this in your project directory, then start an agent — the session appears here live, and everyone on this project can watch it, steer it, or take over."
              : "Sessions stream here when anyone on this project starts one — including you. Run this in your project directory, then start an agent, and the whole team can watch it live, steer it, or take over."}
          </p>
        </div>
      </div>

      <div className="mt-5">
        <AttachCommand slug={slug} className="max-w-xl" />
      </div>

      <div className="mt-4 flex flex-wrap items-center gap-x-5 gap-y-2.5 text-[13px] text-muted-foreground">
        <span className="inline-flex items-center gap-2">
          Works with <HarnessChips />
        </span>
        <span className="text-muted-foreground/60">
          First time? Run{" "}
          <code className="font-mono text-[12px] text-foreground/80">
            coop login
          </code>{" "}
          before attaching.
        </span>
        <Link
          href={"/docs/cli" as Route}
          className="text-muted-foreground underline-offset-4 transition-colors hover:text-foreground hover:underline"
        >
          CLI reference
        </Link>
      </div>
    </div>
  );
}
