export function RedactionChip({ redactions, truncated }: { redactions: number; truncated: boolean }) {
  if (redactions <= 0 && !truncated) return null;

  return (
    <div className="mt-1.5 flex flex-wrap items-center gap-1.5">
      {redactions > 0 && (
        <span className="rounded bg-destructive/15 px-1.5 py-0.5 text-3xs text-destructive">
          {redactions} secret{redactions === 1 ? "" : "s"} redacted
        </span>
      )}
      {truncated && (
        <span className="rounded bg-secondary px-1.5 py-0.5 text-3xs text-muted-foreground">output truncated</span>
      )}
    </div>
  );
}
