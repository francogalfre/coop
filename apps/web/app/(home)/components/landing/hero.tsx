"use client";

import { IconGithub } from "@/components/icons";
import { Button } from "@/components/ui/button";
import { CopyButton } from "@/components/ui/copy-button";
import { Reveal } from "@/components/ui/reveal";
import { LiveSession } from "./live-session";

const ATTACH_COMMAND = "coop attach --project=storefront";

export function Hero({ onSignIn }: { onSignIn: () => void }) {
  return (
    <section className="relative px-6 pt-16 pb-20 sm:px-10 sm:pt-24 sm:pb-28">
      <div className="grid items-center gap-12 lg:grid-cols-[minmax(0,1fr)_minmax(0,1.05fr)] lg:gap-16">
        <div>
          <Reveal>
            <p className="font-mono text-2xs text-muted-foreground/70 uppercase tracking-[0.18em]">
              Multiplayer sessions for coding agents
            </p>
          </Reveal>

          <Reveal delay={0.06}>
            <h1 className="mt-5 text-balance font-display font-semibold text-[40px] text-foreground leading-[1.04] tracking-tight sm:text-[54px]">
              One agent.
              <br />
              The whole team
              <br />
              in the room.
            </h1>
          </Reveal>

          <Reveal delay={0.12}>
            <p className="mt-6 max-w-md text-[15px] text-muted-foreground leading-relaxed">
              A coding agent session is usually one person staring at a terminal.
              coop opens that session to your team: they watch every tool call as
              it lands, redirect the agent mid-task, and take the wheel when it
              matters — from a browser, without touching your machine.
            </p>
          </Reveal>

          <Reveal delay={0.18}>
            <div className="mt-8 flex flex-wrap items-center gap-3">
              <Button size="lg" onClick={onSignIn} className="gap-2">
                <IconGithub size={15} />
                Start with GitHub
              </Button>

              <div className="flex items-center gap-1 rounded-lg border border-border/70 bg-card/50 py-1 pr-1 pl-3">
                <code className="font-mono text-xs text-muted-foreground">{ATTACH_COMMAND}</code>
                <CopyButton value={ATTACH_COMMAND} label="Copy the attach command" />
              </div>
            </div>
          </Reveal>

          <Reveal delay={0.24}>
            <p className="mt-5 text-xs text-muted-foreground/70">
              Free while in beta. Works with Claude Code, opencode, and pi.
            </p>
          </Reveal>
        </div>

        <Reveal delay={0.1}>
          <LiveSession />
        </Reveal>
      </div>
    </section>
  );
}
