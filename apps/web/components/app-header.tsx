"use client";

import Link from "next/link";
import type { Route } from "next";
import { signOut, useSession } from "@/lib/auth/auth-client";
import { IconLogout } from "@/components/icons";
import { Mark } from "@/components/mark";
import { Button } from "@/components/ui/button";
import { PersonAvatar } from "@/components/person-avatar";

export function AppHeader() {
  const { data } = useSession();
  const user = data?.user;

  return (
    <header className="sticky top-0 z-30 border-border/70 border-b bg-background/80 backdrop-blur-md">
      <div className="mx-auto flex h-16 max-w-5xl items-center gap-4 px-6">
        <Link
          href={"/" as Route}
          className="flex items-center gap-2 font-display font-semibold text-[16px] text-foreground tracking-tight"
        >
          <Mark size={22} />
          coop
        </Link>

        <Link
          href={"/docs" as Route}
          className="text-muted-foreground text-sm transition-colors hover:text-foreground"
        >
          Docs
        </Link>

        <div className="ml-auto flex items-center gap-3">
          {user ? (
            <>
              <div className="flex items-center gap-2.5">
                <PersonAvatar
                  name={user.name ?? "?"}
                  avatarUrl={user.image ?? undefined}
                  className="size-7 shrink-0 text-[11px]"
                />
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
              <Link href={"/" as Route}>Sign in</Link>
            </Button>
          )}
        </div>
      </div>
    </header>
  );
}
