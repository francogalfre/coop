"use client";

import { useState } from "react";
import type { ReactNode } from "react";
import Link from "next/link";
import type { Route } from "next";
import { motion, useReducedMotion } from "motion/react";
import { toast } from "sonner";
import { signIn } from "@/lib/auth/auth-client";
import { Mark } from "@/components/mark";
import { HarnessLogo } from "@/components/harness-logo";
import { Button } from "@/components/ui/button";
import { IconCheck, IconCopy, IconGithub } from "@/components/icons";
import { SessionMock } from "./session-mock";

const HARNESS_IDS = ["claude-code", "opencode", "pi"] as const;
const HARNESS_LABELS: Record<(typeof HARNESS_IDS)[number], string> = {
  "claude-code": "Claude Code",
  opencode: "opencode",
  pi: "pi",
};

function Reveal({
  delay = 0,
  children,
}: {
  delay?: number;
  children: ReactNode;
}) {
  const reduce = useReducedMotion();
  return (
    <motion.div
      initial={reduce ? false : { opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={
        reduce
          ? { duration: 0 }
          : { duration: 0.5, delay, ease: [0.16, 1, 0.3, 1] }
      }
    >
      {children}
    </motion.div>
  );
}

function signInWithGithub() {
  return void signIn.social({ provider: "github" });
}

const BEATS = [
  {
    event: "session.start",
    chip: "border-live/25 bg-live/10 text-live",
    title: "Watch every step, live",
    text: "Tool calls, diffs, and terminal output stream to the whole team the moment they happen — a living session, not a transcript after the fact.",
  },
  {
    event: "human.steer",
    chip: "border-human/25 bg-human/10 text-human",
    title: "Steer it mid-task",
    text: "Send a message into the running session and it reaches the agent with your name on it — so you can redirect before an hour-long task finishes wrong.",
  },
  {
    event: "tool.blocked",
    chip: "border-tool/25 bg-tool/10 text-tool",
    title: "Take over, for real",
    text: "Taking over pauses the agent's next tool call through the harness's own permission hook. Hand it back and it keeps going.",
  },
] as const;

function AttachChip() {
  const [copied, setCopied] = useState(false);

  async function copy() {
    try {
      await navigator.clipboard.writeText("coop attach --project=your-app");
      setCopied(true);
      setTimeout(() => setCopied(false), 1800);
    } catch {
      toast.error("Could not copy — type the command manually.");
    }
  }

  return (
    <div className="inline-flex max-w-full items-center gap-2.5 rounded-lg border border-border bg-card px-3 py-2 font-mono text-[13px]">
      <span className="shrink-0 text-muted-foreground/60">$</span>
      <span className="min-w-0 truncate text-foreground/85">
        coop attach --project=<span className="text-agent">your-app</span>
      </span>
      <Button
        variant="ghost"
        size="icon-sm"
        onClick={() => void copy()}
        aria-label="Copy command to clipboard"
        className="shrink-0 text-muted-foreground hover:text-foreground"
      >
        {copied ? (
          <IconCheck size={13} className="text-live" />
        ) : (
          <IconCopy size={13} />
        )}
      </Button>
    </div>
  );
}

export function SignedOut() {
  return (
    <div className="relative min-h-dvh overflow-x-clip bg-background">
      <div
        aria-hidden
        className="pointer-events-none absolute -top-40 left-1/2 -z-10 h-[480px] w-[min(760px,92vw)] -translate-x-1/2 rounded-full bg-agent/10 blur-[130px]"
      />

      <header className="mx-auto flex h-16 w-full max-w-3xl items-center justify-between px-6">
        <Link
          href={"/" as Route}
          className="flex items-center gap-2 font-display font-semibold text-[16px] text-foreground tracking-tight"
        >
          <Mark size={22} />
          coop
        </Link>
        <nav className="flex items-center gap-1.5">
          <Button
            asChild
            variant="ghost"
            size="sm"
            className="text-[13px] text-muted-foreground"
          >
            <Link href={"/docs" as Route}>Docs</Link>
          </Button>
          <Button
            size="sm"
            onClick={signInWithGithub}
            className="gap-1.5 text-[13px]"
          >
            <IconGithub size={14} />
            Sign in
          </Button>
        </nav>
      </header>

      <main className="mx-auto max-w-3xl px-6 pb-24">
        <section className="pt-10 text-center sm:pt-16">
          <Reveal>
            <h1 className="text-balance font-display text-[38px] font-semibold leading-[1.06] tracking-tight text-foreground sm:text-[56px]">
              Your agents shouldn&apos;t work alone.
            </h1>
          </Reveal>

          <Reveal delay={0.08}>
            <p className="mx-auto mt-5 max-w-xl text-balance text-[15px] leading-relaxed text-muted-foreground">
              Google Docs made writing multiplayer. Figma made design
              multiplayer. coop does it for coding agents: one teammate runs the
              agent, everyone else watches it live, steers it mid-task, and
              takes over when it counts.
            </p>
          </Reveal>

          <Reveal delay={0.16}>
            <div className="mt-7 flex flex-col items-center justify-center gap-3 sm:flex-row">
              <Button
                size="lg"
                onClick={signInWithGithub}
                className="h-11 gap-2 rounded-xl px-5 text-[14px]"
              >
                <IconGithub size={17} />
                Continue with GitHub
              </Button>
              <Button
                asChild
                variant="outline"
                size="lg"
                className="h-11 rounded-xl px-5 text-[14px]"
              >
                <Link href={"/docs" as Route}>How it works</Link>
              </Button>
            </div>
          </Reveal>

          <Reveal delay={0.24}>
            <div className="mt-7 flex flex-wrap items-center justify-center gap-2">
              {HARNESS_IDS.map((harness) => (
                <span
                  key={harness}
                  className="inline-flex items-center gap-1.5 rounded-full border border-border bg-card/60 px-3 py-1.5 text-xs text-muted-foreground"
                >
                  <HarnessLogo harness={harness} size={13} />
                  {HARNESS_LABELS[harness]}
                </span>
              ))}
            </div>
          </Reveal>
        </section>

        <section className="mt-14 sm:mt-16">
          <Reveal>
            <SessionMock />
            <p className="mt-3 text-center text-xs text-muted-foreground/70">
              A live session, as the rest of the team sees it.
            </p>
          </Reveal>
        </section>

        <section className="mt-20 sm:mt-28">
          <Reveal>
            <h2 className="font-display text-[22px] font-semibold tracking-tight text-foreground">
              From a transcript to a room
            </h2>
          </Reveal>
          <div className="mt-7 space-y-7">
            {BEATS.map((beat, i) => (
              <Reveal key={beat.event} delay={0.05 * i}>
                <div className="flex flex-col gap-2.5 sm:flex-row sm:gap-5">
                  <span
                    className={`self-start shrink-0 rounded-md border px-2 py-1 font-mono text-xs sm:w-[6.5rem] sm:text-center ${beat.chip}`}
                  >
                    {beat.event}
                  </span>
                  <div>
                    <h3 className="font-display text-[16px] font-medium text-foreground">
                      {beat.title}
                    </h3>
                    <p className="mt-1 max-w-lg text-[14px] leading-relaxed text-muted-foreground">
                      {beat.text}
                    </p>
                  </div>
                </div>
              </Reveal>
            ))}
          </div>
        </section>

        <section className="mt-20 sm:mt-28">
          <Reveal>
            <h2 className="font-display text-[22px] font-semibold tracking-tight text-foreground">
              Safe to share by default
            </h2>
            <p className="mt-3 max-w-xl text-[14px] leading-relaxed text-muted-foreground">
              coop rides the harness&apos;s own hooks — it never scrapes your
              terminal. Secrets are redacted on your machine before an event
              leaves it, and steering is attributed to the person who sent it,
              never framed as a system instruction.
            </p>
          </Reveal>
        </section>

        <section className="mt-20 text-center sm:mt-28">
          <Reveal>
            <h2 className="text-balance font-display text-[26px] font-semibold tracking-tight text-foreground sm:text-[30px]">
              Bring your team into the room
            </h2>
            <p className="mx-auto mt-3 max-w-md text-[14px] leading-relaxed text-muted-foreground">
              Sign in with GitHub, create a project, and run one command in your
              project directory — your next agent session is a room.
            </p>
            <div className="mt-6 flex flex-col items-center gap-4">
              <Button
                size="lg"
                onClick={signInWithGithub}
                className="h-11 gap-2 rounded-xl px-5 text-[14px]"
              >
                <IconGithub size={17} />
                Continue with GitHub
              </Button>
              <AttachChip />
            </div>
          </Reveal>
        </section>
      </main>

      <footer className="border-border/60 border-t py-8">
        <p className="text-center text-xs text-muted-foreground/70">
          coop — multiplayer sessions for coding agents. Runs on Claude Code,
          opencode, and pi.
        </p>
      </footer>
    </div>
  );
}
