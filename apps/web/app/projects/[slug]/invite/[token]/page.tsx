"use client";

import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import type { Route } from "next";
import { useParams, useRouter } from "next/navigation";
import { motion } from "motion/react";
import { signIn, useSession } from "@/lib/auth/auth-client";
import { relayApi } from "@/lib/relay/api";
import { IconAlert, IconGithub, IconPeople, IconSpinner } from "@/components/icons";
import { Button } from "@/components/ui/button";

type State = "checking" | "signed-out" | "joining" | "failed";

export default function InvitePage() {
  const params = useParams<{ slug: string; token: string }>();
  const router = useRouter();
  const { data, isPending } = useSession();

  const [state, setState] = useState<State>("checking");
  const attempted = useRef(false);

  useEffect(() => {
    if (isPending) return;

    if (!data?.user) {
      setState("signed-out");
      return;
    }

    if (attempted.current) return;
    attempted.current = true;
    setState("joining");

    relayApi
      .acceptInvite(params.token)
      .then((project) => router.replace(`/projects/${project.slug}` as Route))
      .catch(() => setState("failed"));
  }, [isPending, data, params.token, router]);

  return (
    <main className="mx-auto flex min-h-dvh max-w-md flex-col items-center justify-center px-6 text-center">
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.4, ease: [0.16, 1, 0.3, 1] }}
        className="w-full space-y-5"
      >
        <span className="mx-auto grid size-12 place-items-center rounded-xl border border-border bg-card text-foreground">
          {state === "failed" ? <IconAlert size={21} /> : <IconPeople size={21} />}
        </span>

        {state === "failed" ? (
          <>
            <div className="space-y-1.5">
              <h1 className="font-display font-semibold text-[22px] text-foreground tracking-tight">
                This invite isn&apos;t valid
              </h1>
              <p className="text-[13.5px] text-muted-foreground leading-relaxed">
                It may have expired, been revoked, or already been used. Ask for a fresh link.
              </p>
            </div>
            <Button asChild variant="secondary" className="w-full">
              <Link href={"/" as Route}>Go to my projects</Link>
            </Button>
          </>
        ) : state === "signed-out" ? (
          <>
            <div className="space-y-1.5">
              <h1 className="font-display font-semibold text-[22px] text-foreground tracking-tight">
                You&apos;ve been invited to{" "}
                <span className="font-mono text-[19px]">{params.slug}</span>
              </h1>
              <p className="text-[13.5px] text-muted-foreground leading-relaxed">
                Sign in with GitHub to join the project and see its live agent sessions.
              </p>
            </div>
            <Button
              size="lg"
              onClick={() =>
                void signIn.social({
                  provider: "github",
                  callbackURL: `/projects/${params.slug}/invite/${params.token}`,
                })
              }
              className="h-11 w-full gap-2 rounded-xl text-[14px]"
            >
              <IconGithub size={17} />
              Continue with GitHub
            </Button>
          </>
        ) : (
          <div className="flex items-center justify-center gap-2 text-[13.5px] text-muted-foreground">
            <IconSpinner size={15} className="animate-spin" />
            Joining {params.slug}…
          </div>
        )}
      </motion.div>
    </main>
  );
}
