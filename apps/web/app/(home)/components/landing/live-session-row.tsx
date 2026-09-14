import {
  IconFile,
  IconLock,
  IconSparkles,
  IconTerminal,
} from "@/components/icons";
import { PersonAvatar } from "@/components/ui/person-avatar";
import type { LiveRow } from "./live-session-script";

function Rail({ tone, children }: { tone: string; children: React.ReactNode }) {
  return (
    <span
      className={`grid size-5 shrink-0 place-items-center rounded-md ${tone}`}
      aria-hidden
    >
      {children}
    </span>
  );
}

export function LiveSessionRow({ row }: { row: LiveRow }) {
  if (row.kind === "agent") {
    return (
      <div className="flex gap-2.5">
        <Rail tone="bg-agent/12 text-agent">
          <IconSparkles size={12} />
        </Rail>
        <p className="pt-0.5 text-[13px] text-foreground/85 leading-relaxed">
          {row.text}
        </p>
      </div>
    );
  }

  if (row.kind === "tool") {
    return (
      <div className="flex gap-2.5">
        <Rail tone="bg-tool/12 text-tool">
          <IconTerminal size={12} />
        </Rail>
        <p className="truncate pt-0.5 font-mono text-xs text-muted-foreground">
          <span className="text-tool">{row.name}</span> {row.detail}
        </p>
      </div>
    );
  }

  if (row.kind === "file") {
    return (
      <div className="flex gap-2.5">
        <Rail tone="bg-file/12 text-file">
          <IconFile size={12} />
        </Rail>
        <p className="flex min-w-0 items-center gap-2 pt-0.5 font-mono text-xs">
          <span className="truncate text-muted-foreground">{row.path}</span>
          <span className="shrink-0 text-live">+{row.added}</span>
          <span className="shrink-0 text-destructive">−{row.removed}</span>
        </p>
      </div>
    );
  }

  if (row.kind === "human") {
    return (
      <div className="flex gap-2.5">
        <PersonAvatar name={row.who} className="size-5 shrink-0 text-[9px]" />
        <div className="min-w-0 rounded-lg rounded-tl-sm border border-human/25 bg-human/8 px-2.5 py-1.5">
          <p className="font-medium font-mono text-2xs text-human">
            {row.who} via coop
          </p>
          <p className="mt-1 text-[13px] text-foreground/90 leading-relaxed">
            {row.text}
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex gap-2.5">
      <Rail tone="bg-destructive/12 text-destructive">
        <IconLock size={12} />
      </Rail>
      <div className="min-w-0">
        <p className="truncate font-mono text-xs text-muted-foreground">
          <span className="text-destructive">{row.name}</span> {row.detail}
        </p>
        <p className="mt-1 text-2xs text-muted-foreground/80">
          Waiting for the session owner to allow or deny.
        </p>
      </div>
    </div>
  );
}
