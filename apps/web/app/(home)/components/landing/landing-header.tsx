"use client";

import Link from "next/link";
import type { Route } from "next";
import { IconGithub } from "@/components/icons";
import { Mark } from "@/components/brand/mark";
import { Button } from "@/components/ui/button";

export function LandingHeader({ onSignIn }: { onSignIn: () => void }) {
  return (
    <header className="sticky top-0 z-40 border-border/60 border-b bg-background/85 backdrop-blur-md">
      <div className="flex h-14 items-center gap-6 px-6 sm:px-10">
        <Link
          href={"/" as Route}
          className="flex items-center gap-2 font-display font-semibold text-[15px] tracking-tight"
        >
          <Mark size={20} />
          coop
        </Link>

        <nav className="ml-auto flex items-center gap-1.5">
          <Button asChild variant="ghost" size="sm" className="text-muted-foreground text-xs">
            <Link href={"/docs" as Route}>Docs</Link>
          </Button>
          <Button size="sm" onClick={onSignIn} className="gap-1.5 text-xs">
            <IconGithub size={13} />
            Sign in
          </Button>
        </nav>
      </div>
    </header>
  );
}
