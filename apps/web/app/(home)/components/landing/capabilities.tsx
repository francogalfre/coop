import { IconMessage, IconPeople, IconSparkles, IconUnlock } from "@/components/icons";
import { Section } from "@/components/layout/section";
import { SectionHeading } from "@/components/layout/section-heading";
import { Reveal } from "@/components/ui/reveal";

const CAPABILITIES = [
  {
    icon: IconSparkles,
    tone: "bg-agent/12 text-agent",
    title: "Watch it work",
    body: "Every tool call, command, result, and file change, in order, as it happens. Plus what every other agent in the project is touching, so two sessions do not collide on the same file.",
  },
  {
    icon: IconMessage,
    tone: "bg-human/12 text-human",
    title: "Steer mid-task",
    body: "Send the agent a message without stopping it. It arrives attributed — the agent sees who asked, and so does everyone else reading the timeline.",
  },
  {
    icon: IconPeople,
    tone: "bg-tool/12 text-tool",
    title: "Answer its questions",
    body: "When the agent asks the team something, the question lands in the session instead of blocking one person's terminal. Whoever knows the answer replies.",
  },
  {
    icon: IconUnlock,
    tone: "bg-live/12 text-live",
    title: "Take over",
    body: "Claim the session and drive the agent yourself while the rest of the team watches. Hand it back when you are done.",
  },
];

export function Capabilities() {
  return (
    <Section id="capabilities">
      <SectionHeading
        eyebrow="In the browser"
        title="Reading the log is not the same as being there"
        lede="A transcript tells you what happened. coop lets you change what happens next."
      />

      <div className="mt-12 grid gap-px overflow-hidden rounded-xl border border-border/70 bg-border/70 sm:grid-cols-2">
        {CAPABILITIES.map((item, index) => (
          <Reveal key={item.title} delay={index * 0.06} className="bg-background">
            <article className="flex h-full gap-4 p-6 sm:p-7">
              <span className={`grid size-8 shrink-0 place-items-center rounded-lg ${item.tone}`}>
                <item.icon size={15} />
              </span>
              <div>
                <h3 className="font-display font-semibold text-[17px] text-foreground tracking-tight">
                  {item.title}
                </h3>
                <p className="mt-2 text-[13px] text-muted-foreground leading-relaxed">
                  {item.body}
                </p>
              </div>
            </article>
          </Reveal>
        ))}
      </div>
    </Section>
  );
}
