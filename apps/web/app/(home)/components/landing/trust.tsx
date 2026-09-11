import { Section } from "@/components/layout/section";
import { SectionHeading } from "@/components/layout/section-heading";
import { Reveal } from "@/components/ui/reveal";

const RULES = [
  {
    title: "Secrets never leave the machine",
    body: "The CLI redacts tokens, keys, and environment values before an event is sent. The relay stores what is left, and the count of what was removed.",
  },
  {
    title: "The owner keeps the controls",
    body: "Mode changes, permission prompts, and approving a teammate's steer in restricted mode are owner-only. Being able to watch is not the same as being able to act.",
  },
  {
    title: "Only the project can see it",
    body: "Sessions belong to a project. Watching, steering, and taking over all require membership — there is no public session link.",
  },
];

export function Trust() {
  return (
    <Section id="trust">
      <SectionHeading
        eyebrow="Trust boundary"
        title="Opening a session is not opening your machine"
        lede="Everything a teammate can do is something you decided they could do."
      />

      <dl className="mt-12 grid gap-px overflow-hidden rounded-xl border border-border/70 bg-border/70 sm:grid-cols-3">
        {RULES.map((rule, index) => (
          <Reveal key={rule.title} delay={index * 0.07} className="bg-background">
            <div className="h-full p-6">
              <dt className="font-display font-semibold text-[15px] text-foreground tracking-tight">
                {rule.title}
              </dt>
              <dd className="mt-2.5 text-[13px] text-muted-foreground leading-relaxed">
                {rule.body}
              </dd>
            </div>
          </Reveal>
        ))}
      </dl>
    </Section>
  );
}
