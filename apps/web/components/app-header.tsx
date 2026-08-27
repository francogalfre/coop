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
      <div className="mx-auto flex h-16 max-w-5xl items-center gap-4 px-6">
        <Link
          href={"/" as Route}
          className="font-display font-semibold text-[16px] text-foreground tracking-tight"
        >
          coop
        </Link>

        <div className="ml-auto flex items-center gap-3">
          {user ? (
            <>
              <div className="flex items-center gap-2.5">
                {user.image ? (
                  <img
                    src={user.image}
                    alt=""
                    referrerPolicy="no-referrer"
                    className="size-7 shrink-0 rounded-full border border-border/60 object-cover"
                  />
                ) : (
                  <span
                    className="grid size-7 shrink-0 place-items-center rounded-full font-medium text-[11px] text-background"
                    style={{ background: tintFor(user.name ?? "?") }}
                  >
                    {initials(user.name ?? "?")}
                  </span>
                )}
                <span className="hidden text-sm text-muted-foreground sm:block">{user.name}</span>
              </div>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => void signOut()}
                className="h-9 gap-1.5 px-2.5 text-xs text-muted-foreground"
              >
                <IconLogout size={14} />
                <span className="hidden sm:inline">Sign out</span>
              </Button>
            </>
          ) : (
            <Button asChild size="sm" className="h-9 text-xs">
              <Link href={"/login" as Route}>Sign in</Link>
            </Button>
          )}
        </div>
      </div>
    </header>
  );
}
