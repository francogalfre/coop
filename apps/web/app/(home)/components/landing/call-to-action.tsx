"use client";

import { IconGithub } from "@/components/icons";
import { Section } from "@/components/layout/section";
import { Button } from "@/components/ui/button";
import { CodeBlock } from "@/components/ui/code-block";
import { Reveal } from "@/components/ui/reveal";

export function CallToAction({ onSignIn }: { onSignIn: () => void }) {
  return (
    <Section className="bg-dotgrid">
      <Reveal className="mx-auto max-w-lg text-center">
        <h2 className="text-balance font-display font-semibold text-[30px] text-foreground leading-[1.1] tracking-tight sm:text-[38px]">
          Put your next agent session in front of your team
        </h2>
        <p className="mx-auto mt-4 max-w-sm text-balance text-[15px] text-muted-foreground leading-relaxed">
          The session your team joins is the one you were going to run anyway.
        </p>

        <div className="mt-8 flex flex-col items-center gap-4">
          <Button size="lg" onClick={onSignIn} className="gap-2">
            <IconGithub size={15} />
            Start with GitHub
          </Button>
          <CodeBlock
            body="coop attach --project=storefront"
            copyable
            className="w-full max-w-xs text-left"
          />
        </div>
      </Reveal>
    </Section>
  );
}
