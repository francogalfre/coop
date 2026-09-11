import { IconPeople } from "@/components/icons";

const STEPS = [
  "Create a project below — it only needs a name.",
  "In your project directory, run coop attach --project=<slug>.",
  "Start Claude Code, opencode, or pi — the session shows up here, live for the team.",
];

export function EmptyProjects() {
  return (
    <div className="rounded-xl border border-border border-dashed bg-card/30 px-6 py-12 text-center sm:px-10">
      <span className="mx-auto mb-3 grid size-11 place-items-center rounded-xl border border-border bg-card text-muted-foreground">
        <IconPeople size={19} />
      </span>
      <h2 className="font-display font-medium text-[15px] text-foreground">No projects yet</h2>
      <p className="mx-auto mt-1 max-w-sm text-[13px] text-muted-foreground leading-relaxed">
        A project is the room your team&apos;s agent sessions live in. Three steps to your first
        live one:
      </p>
      <ol className="mx-auto mt-6 max-w-sm space-y-3 text-left">
        {STEPS.map((step, index) => (
          <li key={step} className="flex items-start gap-3">
            <span className="grid size-5 shrink-0 place-items-center rounded-full border border-border bg-card font-mono text-2xs text-muted-foreground">
              {index + 1}
            </span>
            <span className="text-[13px] text-muted-foreground leading-relaxed">{step}</span>
          </li>
        ))}
      </ol>
    </div>
  );
}
