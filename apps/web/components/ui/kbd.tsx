import type { ReactNode } from "react";
import { cn } from "@/lib/cn";

export function Kbd({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <kbd
      className={cn(
        "inline-flex h-5 min-w-5 items-center justify-center rounded border border-border bg-secondary/60 px-1.5 font-mono text-2xs text-muted-foreground",
        className,
      )}
    >
      {children}
    </kbd>
  );
}
