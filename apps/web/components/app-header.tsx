"use client";

import Link from "next/link";
import type { Route } from "next";
import { signOut, useSession } from "@/lib/auth/auth-client";
import { IconLogout } from "@/components/icons";
import { Button } from "@/components/ui/button";
import { initials, tintFor } from "@/lib/format";

export function AppHeader() {
  const { data } = useSession();
  const user = data?.user;

  return (
    <header className="sticky top-0 z-30 border-border/70 border-b bg-background/80 backdrop-blur-md">
      <div className="mx-auto flex h-14 max-w-5xl items-center gap-3 px-5">
        <Link
          href={"/" as Route}
          className="font-display font-semibold text-[16px] text-foreground tracking-tight"
        >
          coop
        </Link>

        <div className="ml-auto flex items-center gap-2.5">
          {user ? (
            <>
              <div className="flex items-center gap-2">
                <span
                  className="grid size-7 place-items-center rounded-full font-medium text-[11px] text-background"
                  style={{ background: tintFor(user.name ?? "?") }}
                >
                  {initials(user.name ?? "?")}
                </span>
                <span className="hidden text-[13px] text-muted-foreground sm:block">{user.name}</span>
              </div>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => void signOut()}
                className="h-8 gap-1.5 px-2 text-[12.5px] text-muted-foreground"
              >
                <IconLogout size={14} />
                <span className="hidden sm:inline">Sign out</span>
              </Button>
            </>
          ) : (
            <Button asChild size="sm" className="h-8 text-[12.5px]">
              <Link href={"/login" as Route}>Sign in</Link>
            </Button>
          )}
        </div>
      </div>
    </header>
  );
}
