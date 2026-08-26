import { IconTerminal } from "@/components/icons";

export function EmptySessions() {
  return (
    <div className="rounded-xl border border-border border-dashed px-6 py-14 text-center">
      <span className="mx-auto mb-3 grid size-11 place-items-center rounded-xl border border-border bg-card text-muted-foreground">
        <IconTerminal size={19} />
      </span>
      <p className="font-display font-medium text-[15px] text-foreground">No sessions yet</p>
      <p className="mx-auto mt-1 max-w-sm text-[13px] text-muted-foreground leading-relaxed">
        Start one from your terminal and it will appear here for the whole team.
      </p>
      <code className="mt-4 inline-block rounded-lg border border-border bg-background px-3 py-2 font-mono text-[12.5px] text-foreground/80">
        coop attach --project=<span className="text-agent">slug</span>
      </code>
    </div>
  );
}
