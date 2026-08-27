import { clockTime } from "@/lib/format";
import { cn } from "@/lib/utils";

export function Row({
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
      <time className="hidden w-14 shrink-0 pt-1 text-right font-mono text-2xs text-muted-foreground/50 tabular-nums sm:block">
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
