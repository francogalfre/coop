import { IconReply } from "@/components/icons";
import { clockTime } from "@/lib/format";
import { cn } from "@/lib/utils";

export function Row({
  ts,
  seq,
  rail,
  children,
  className,
  onReply,
  highlighted,
}: {
  ts: string;
  seq?: number;
  rail: React.ReactNode;
  children: React.ReactNode;
  className?: string;
  onReply?: (seq: number) => void;
  highlighted?: boolean;
}) {
  return (
    <div
      id={seq !== undefined ? `seq-${seq}` : undefined}
      className={cn(
        "group relative mx-auto flex max-w-3xl gap-3 rounded-lg px-4 py-2 transition-colors duration-500 sm:px-6",
        highlighted && "bg-human/10",
        className,
      )}
    >
      <time className="hidden w-14 shrink-0 pt-1 text-right font-mono text-2xs text-muted-foreground/50 tabular-nums sm:block">
        {clockTime(ts)}
      </time>
      <div className="relative flex w-6 shrink-0 justify-center">
        <span className="absolute top-7 bottom-[-16px] w-px bg-border/60 group-last:hidden" />
        {rail}
      </div>
      <div className="min-w-0 flex-1 pb-1">{children}</div>
      {onReply && seq !== undefined && (
        <button
          type="button"
          onClick={() => onReply(seq)}
          title={`Reply to step ${seq}`}
          className="absolute top-1.5 right-2 grid size-6 shrink-0 place-items-center rounded-md border border-border bg-card text-muted-foreground opacity-0 transition-opacity hover:text-foreground group-hover:opacity-100"
        >
          <IconReply size={13} />
        </button>
      )}
    </div>
  );
}
