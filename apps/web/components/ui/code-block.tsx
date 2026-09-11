import { cn } from "@/lib/cn";
import { CopyButton } from "@/components/ui/copy-button";

export function CodeBlock({
  body,
  label,
  copyable = false,
  scroll = false,
  className,
}: {
  body: string;
  label?: string;
  copyable?: boolean;
  scroll?: boolean;
  className?: string;
}) {
  return (
    <div
      className={cn(
        "group relative overflow-hidden rounded-lg border border-border/70 bg-[oklch(0.12_0_0)]",
        className,
      )}
    >
      {label && (
        <div className="flex items-center border-border/60 border-b px-3 py-1.5">
          <span className="truncate font-mono text-2xs text-muted-foreground/80 uppercase tracking-wider">
            {label}
          </span>
        </div>
      )}

      {copyable && (
        <CopyButton
          value={body}
          className={cn(
            "absolute right-1.5 size-7 bg-[oklch(0.12_0_0)]/90",
            label ? "top-0.5" : "top-1.5",
          )}
        />
      )}

      <pre
        className={cn(
          "overflow-x-auto px-3 py-2.5 font-mono text-xs text-foreground/85 leading-relaxed",
          scroll && "max-h-72 overflow-y-auto",
          copyable && !label && "pr-11",
        )}
      >
        <code>{body}</code>
      </pre>
    </div>
  );
}
