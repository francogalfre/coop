import { HarnessLogo } from "@/components/brand/harness-logo";
import { Section } from "@/components/layout/section";
import { SectionHeading } from "@/components/layout/section-heading";
import { Reveal } from "@/components/ui/reveal";

const COLUMNS = ["Harness", "Mechanism", "Installed at", "Concurrent sessions"];

const HARNESSES = [
  {
    id: "claude-code",
    name: "Claude Code",
    mechanism: "Native hooks",
    hook: ".claude/settings.local.json",
    concurrent: "Several per directory",
  },
  {
    id: "opencode",
    name: "opencode",
    mechanism: "Plugin",
    hook: ".opencode/plugin/coop.js",
    concurrent: "One per directory",
  },
  {
    id: "pi",
    name: "pi",
    mechanism: "Extension",
    hook: ".pi/extensions/coop.ts",
    concurrent: "One per directory",
  },
];

const GRID =
  "sm:grid sm:grid-cols-[1.1fr_0.9fr_1.5fr_1.1fr] sm:items-center sm:gap-6";

export function Harnesses() {
  return (
    <Section id="harnesses">
      <SectionHeading
        eyebrow="Harnesses"
        title="It listens through the harness, not the terminal"
        lede="coop installs each harness's own hook or plugin. Nothing is scraped from your screen, so a tool call arrives as a tool call — not as a line of text someone has to parse."
      />

      <Reveal delay={0.1} className="mt-12">
        <div
          className={`hidden border-border/70 border-b pb-3 font-mono text-2xs text-muted-foreground/60 uppercase tracking-[0.14em] ${GRID}`}
        >
          {COLUMNS.map((column) => (
            <span key={column}>{column}</span>
          ))}
        </div>

        <ul>
          {HARNESSES.map((harness) => (
            <li
              key={harness.id}
              className={`flex flex-col gap-1.5 border-border/50 border-b py-5 ${GRID}`}
            >
              <span className="flex items-center gap-2.5 font-medium text-[14px] text-foreground">
                <HarnessLogo
                  harness={harness.id}
                  size={16}
                  className="shrink-0"
                />
                {harness.name}
              </span>
              <span className="text-[13px] text-muted-foreground">
                {harness.mechanism}
              </span>
              <code className="truncate font-mono text-xs text-muted-foreground/80">
                {harness.hook}
              </code>
              <span className="text-[13px] text-muted-foreground">
                {harness.concurrent}
              </span>
            </li>
          ))}
        </ul>
      </Reveal>

      <Reveal delay={0.18}>
        <p className="mt-6 text-[13px] text-muted-foreground/70">
          Running something else?{" "}
          <code className="font-mono text-foreground/75">coop run</code> wraps
          any command in a pty and still streams the session, with less detail.
        </p>
      </Reveal>
    </Section>
  );
}
