"use client";

import { IconGithub } from "@/components/icons";
import { Button } from "@/components/ui/button";
import { CopyButton } from "@/components/ui/copy-button";
import { DottedMountain } from "@/components/ui/dotted-mountain";
import { Reveal } from "@/components/ui/reveal";
import { YCPill } from "./yc-pill";

const ATTACH_COMMAND = "coop attach --project=storefront";

export function Hero({ onSignIn }: { onSignIn: () => void }) {
  return (
    <section className="relative overflow-hidden pt-16 sm:pt-20">
      <div className="relative z-10 max-w-2xl px-6 pb-8 sm:px-10">
        <Reveal>
          <YCPill />
        </Reveal>

        <Reveal delay={0.06}>
          <h1 className="mt-7 text-balance text-editorial text-[46px] text-foreground leading-[0.96] sm:text-[72px]">
            Your coding agent
            <br />
            shouldn&apos;t work alone.
          </h1>
        </Reveal>

        <Reveal delay={0.12}>
          <p className="mt-7 max-w-lg text-[15px] text-muted-foreground leading-relaxed sm:text-md">
            It runs on your machine. Your team watches every tool call in the
            browser and redirects it mid-task — before the diff gets too big to
            review.
          </p>
        </Reveal>

        <Reveal delay={0.18}>
          <div className="mt-9 flex flex-wrap items-center gap-3">
            <Button size="lg" onClick={onSignIn} className="gap-2">
              <IconGithub size={15} />
              Start with GitHub
            </Button>

            <div className="flex items-center gap-1 rounded-lg border border-border/70 bg-card/60 py-1 pr-1 pl-3 backdrop-blur-sm">
              <code className="font-mono text-xs text-muted-foreground">
                {ATTACH_COMMAND}
              </code>
              <CopyButton
                value={ATTACH_COMMAND}
                label="Copy the attach command"
              />
            </div>
          </div>
        </Reveal>

        <Reveal delay={0.24}>
          <p className="mt-7 font-mono text-2xs text-muted-foreground/60 uppercase tracking-[0.16em]">
            Free in beta · Claude Code · opencode · pi
          </p>
        </Reveal>
      </div>

      <div className="relative h-[180px] sm:h-[250px]">
        <div
          aria-hidden
          className="-translate-x-1/2 pointer-events-none absolute bottom-0 left-1/2 h-[260px] w-[900px] rounded-full bg-agent/20 blur-[120px]"
        />
        <DottedMountain className="absolute inset-0 size-full text-agent drop-shadow-[0_0_18px_var(--agent)]" />
      </div>
    </section>
  );
}
