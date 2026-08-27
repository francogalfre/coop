import type { ReactNode } from "react";
import { cn } from "@/lib/utils";

export function LiveDot({ className }: { className?: string }) {
  return (
    <span className={cn("relative grid size-2 place-items-center", className)}>
      <span className="absolute inline-flex size-2 animate-pulse-live rounded-full bg-live/70 motion-reduce:animate-none" />
      <span className="relative inline-flex size-1.5 rounded-full bg-live" />
    </span>
  );
}

export function StatusPill({
  tone = "neutral",
  children,
  className,
}: {
  tone?: "live" | "ended" | "neutral";
  children: ReactNode;
  className?: string;
}) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 font-medium text-[11px] uppercase tracking-wider",
        tone === "live" && "bg-live/12 text-live",
        tone === "ended" && "bg-secondary text-muted-foreground",
        tone === "neutral" && "bg-secondary/60 text-muted-foreground",
        className,
      )}
    >
      {tone === "live" && <LiveDot />}
      {children}
    </span>
  );
}

export function MetaChip({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-md bg-secondary/50 px-2 py-1 font-mono text-[11px] text-muted-foreground",
        className,
      )}
    >
      {children}
    </span>
  );
}
