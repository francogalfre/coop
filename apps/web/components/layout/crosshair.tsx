import { cn } from "@/lib/cn";

export function Crosshair({ className }: { className?: string }) {
  return (
    <span
      aria-hidden
      className={cn(
        "pointer-events-none absolute hidden size-[7px] lg:block",
        className,
      )}
    >
      <span className="absolute top-1/2 left-0 h-px w-full -translate-y-1/2 bg-border" />
      <span className="absolute top-0 left-1/2 h-full w-px -translate-x-1/2 bg-border" />
    </span>
  );
}
