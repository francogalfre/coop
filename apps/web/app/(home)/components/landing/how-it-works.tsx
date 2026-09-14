import { Section } from "@/components/layout/section";
import { SectionHeading } from "@/components/layout/section-heading";
import { CodeBlock } from "@/components/ui/code-block";
import { Reveal } from "@/components/ui/reveal";

const STEPS = [
  {
    title: "Create a project",
    body: "Sign in with GitHub and invite whoever should be able to watch and steer.",
    code: null,
  },
  {
    title: "Attach next to your agent",
    body: "One command in your repo. coop installs the harness hooks and starts listening. Your terminal stays exactly as it was.",
    code: "coop attach --project=storefront",
  },
  {
    title: "Share the link",
    body: "Start the agent as usual. The session appears for everyone on the project.",
    code: null,
  },
];

export function HowItWorks() {
  return (
    <Section id="how-it-works">
      <div className="grid gap-12 lg:grid-cols-[minmax(0,0.85fr)_minmax(0,1.15fr)] lg:gap-20">
        <SectionHeading
          eyebrow="Getting started"
          title="Three steps to a shared session"
          lede="No proxy, no wrapper around your editor, no new way of working."
          className="lg:sticky lg:top-28 lg:self-start"
        />

        <ol>
          {STEPS.map((step, index) => (
            <Reveal key={step.title} delay={index * 0.08}>
              <li className="flex gap-6 border-border/70 border-t py-7 first:border-t-0 first:pt-0">
                <span className="pt-0.5 font-mono text-2xs text-muted-foreground/50 tabular-nums">
                  {String(index + 1).padStart(2, "0")}
                </span>
                <div className="min-w-0 flex-1">
                  <h3 className="font-display font-semibold text-[18px] text-foreground tracking-tight">
                    {step.title}
                  </h3>
                  <p className="mt-2 text-[14px] text-muted-foreground leading-relaxed">
                    {step.body}
                  </p>
                  {step.code && (
                    <CodeBlock
                      body={step.code}
                      copyable
                      className="mt-4 max-w-sm"
                    />
                  )}
                </div>
              </li>
            </Reveal>
          ))}
        </ol>
      </div>
    </Section>
  );
}
