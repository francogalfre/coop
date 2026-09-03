import { motion } from "motion/react";
import { signIn } from "@/lib/auth/auth-client";
import { IconGithub, IconLock, IconMessage, IconUnlock } from "@/components/icons";
import { Mark } from "@/components/mark";
import { Button } from "@/components/ui/button";

const STEPS = [
  {
    icon: IconLock,
    title: "Redacted at the source",
    description:
      "Secrets are filtered on your machine before an event is sent — never on the relay, never after the fact.",
  },
  {
    icon: IconMessage,
    title: "Steering, attributed",
    description:
      "A teammate's message reaches the agent mid-task with their name on it — never disguised as a system instruction.",
  },
  {
    icon: IconUnlock,
    title: "A takeover that's real",
    description:
      "Taking over pauses the agent's next tool call through the harness's own permission hook, not a UI-only lock.",
  },
];

export function SignedOut() {
  return (
    <main className="relative mx-auto flex min-h-dvh max-w-3xl flex-col items-center justify-center px-6 py-16 text-center">
      <div
        aria-hidden
        className="-z-10 -translate-x-1/2 pointer-events-none absolute top-[-8%] left-1/2 h-[420px] w-[min(680px,90vw)] rounded-full bg-agent/10 blur-[120px]"
      />

      <motion.div
        initial={{ opacity: 0, y: 12 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, ease: [0.16, 1, 0.3, 1] }}
        className="space-y-6"
      >
        <Mark size={44} />

        <span className="inline-flex items-center gap-2 rounded-full border border-border bg-card px-3 py-1.5 text-[12px] text-muted-foreground">
          <span className="size-1.5 rounded-full bg-live" />
          Multiplayer coding agents
        </span>

        <h1 className="text-balance font-display font-semibold text-[42px] text-foreground leading-[1.05] tracking-tight sm:text-[56px]">
          The agent runs alone. <span className="text-muted-foreground">Your team doesn&apos;t have to.</span>
        </h1>

        <p className="mx-auto max-w-md text-balance text-[15px] text-muted-foreground leading-relaxed">
          One person runs the agent; everyone else gets a read-only transcript. Coop makes it a room
          your team can enter — watch every tool call, steer mid-task, take over for real. Starting
          with coding agents.
        </p>

        <Button
          size="lg"
          onClick={() => void signIn.social({ provider: "github" })}
          className="h-11 gap-2 rounded-xl px-5 text-[14px]"
        >
          <IconGithub size={17} />
          Continue with GitHub
        </Button>
      </motion.div>

      <motion.div
        initial={{ opacity: 0, y: 12 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, delay: 0.15, ease: [0.16, 1, 0.3, 1] }}
        className="mt-16 grid w-full gap-3 sm:grid-cols-3"
      >
        {STEPS.map(({ icon: Icon, title, description }) => (
          <div
            key={title}
            className="rounded-xl border border-border bg-card/50 p-5 text-left transition-colors hover:border-border/80 hover:bg-card"
          >
            <Icon size={18} className="text-muted-foreground" />
            <p className="mt-3 font-display font-medium text-md text-foreground">{title}</p>
            <p className="mt-1 text-[13px] text-muted-foreground leading-relaxed">{description}</p>
          </div>
        ))}
      </motion.div>
    </main>
  );
}
