import { motion } from "motion/react";
import { signIn } from "@/lib/auth/auth-client";
import { IconAgent, IconGithub, IconMessage, IconTerminal, IconUnlock } from "@/components/icons";
import { Button } from "@/components/ui/button";

const STEPS = [
  {
    icon: IconTerminal,
    title: "Watch",
    description: "Drop into a live session and see every tool call and message as it happens.",
  },
  {
    icon: IconMessage,
    title: "Redirect",
    description: "Send a message the agent sees mid-task, attributed to you.",
  },
  {
    icon: IconUnlock,
    title: "Hand off",
    description: "Take over completely — the agent pauses, you drive.",
  },
];

export function SignedOut() {
  return (
    <main className="mx-auto flex min-h-dvh max-w-3xl flex-col items-center justify-center px-6 py-16 text-center">
      <motion.div
        initial={{ opacity: 0, y: 12 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, ease: [0.16, 1, 0.3, 1] }}
        className="space-y-6"
      >
        <span className="inline-flex items-center gap-2 rounded-full border border-border bg-card px-3 py-1.5 text-[12px] text-muted-foreground">
          <IconAgent size={13} />
          Multiplayer coding agents
        </span>

        <h1 className="text-balance font-display font-semibold text-[42px] text-foreground leading-[1.05] tracking-tight sm:text-[56px]">
          Your agent, <span className="text-muted-foreground">everyone&apos;s screen.</span>
        </h1>

        <p className="mx-auto max-w-md text-balance text-[15px] text-muted-foreground leading-relaxed">
          One person runs a coding agent. Teammates open a link and watch it work live — reading
          every tool call, messaging it, and steering it together.
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
          <div key={title} className="rounded-xl border border-border bg-card/50 p-5 text-left">
            <Icon size={18} className="text-muted-foreground" />
            <p className="mt-3 font-display font-medium text-md text-foreground">{title}</p>
            <p className="mt-1 text-[13px] text-muted-foreground leading-relaxed">{description}</p>
          </div>
        ))}
      </motion.div>
    </main>
  );
}
