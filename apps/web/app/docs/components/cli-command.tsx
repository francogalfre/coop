import { CodeBlock } from "./code-block";

type Flag = { flag: string; description: string };

export function CliCommand({
  name,
  description,
  usage,
  flags,
  example,
  note,
}: {
  name: string;
  description: string;
  usage: string;
  flags?: Flag[];
  example: string;
  note?: string;
}) {
  return (
    <section className="rounded-xl border border-border bg-card/50 p-6">
      <h2 className="font-display font-semibold text-foreground text-lg">{name}</h2>
      <p className="mt-2 text-[14px] text-muted-foreground leading-relaxed">{description}</p>

      <CodeBlock label="Usage">{usage}</CodeBlock>

      {flags && flags.length > 0 && (
        <dl className="mb-4 space-y-2.5 border-border border-t pt-4">
          {flags.map((f) => (
            <div key={f.flag} className="flex flex-col gap-0.5 sm:flex-row sm:gap-3">
              <dt className="shrink-0 font-mono text-[13px] text-foreground/90 sm:w-40">{f.flag}</dt>
              <dd className="text-[13px] text-muted-foreground leading-relaxed">{f.description}</dd>
            </div>
          ))}
        </dl>
      )}

      <CodeBlock label="Example">{example}</CodeBlock>

      {note && <p className="text-[13px] text-muted-foreground leading-relaxed">{note}</p>}
    </section>
  );
}
