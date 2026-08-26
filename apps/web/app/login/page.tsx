"use client";

import { motion } from "motion/react";
import { signIn } from "@/lib/auth/auth-client";
import { IconAgent, IconGithub } from "@/components/icons";
import { Button } from "@/components/ui/button";

export default function LoginPage() {
  return (
    <main className="mx-auto flex min-h-dvh max-w-md flex-col items-center justify-center px-6 text-center">
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.4, ease: [0.16, 1, 0.3, 1] }}
        className="w-full space-y-5"
      >
        <span className="mx-auto grid size-12 place-items-center rounded-xl border border-border bg-card text-foreground">
          <IconAgent size={21} />
        </span>

        <div className="space-y-1.5">
          <h1 className="font-display font-semibold text-[24px] text-foreground tracking-tight">
            Sign in to coop
          </h1>
          <p className="text-[13.5px] text-muted-foreground leading-relaxed">
            Watch and steer live agent sessions with your team.
          </p>
        </div>

        <Button
          size="lg"
          onClick={() => void signIn.social({ provider: "github" })}
          className="h-11 w-full gap-2 rounded-xl text-[14px]"
        >
          <IconGithub size={17} />
          Continue with GitHub
        </Button>
      </motion.div>
    </main>
  );
}
