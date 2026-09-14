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
        <h2 className="text-balance text-editorial text-[34px] text-foreground leading-[1.04] sm:text-[46px]">
          Your next session is the one you were going to run anyway
        </h2>
        <p className="mx-auto mt-5 max-w-sm text-balance text-[15px] text-muted-foreground leading-relaxed">
          Only this time, your team is in the room.
        </p>

        <div className="mt-9 flex flex-col items-center gap-4">
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
