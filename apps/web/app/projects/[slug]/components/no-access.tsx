import Link from "next/link";
import type { Route } from "next";
import { IconAlert } from "@/components/icons";
import { Button } from "@/components/ui/button";

export function NoAccess() {
  return (
    <div className="mx-auto max-w-sm space-y-3 py-24 text-center">
      <span className="mx-auto grid size-11 place-items-center rounded-xl border border-border bg-card text-muted-foreground">
        <IconAlert size={19} />
      </span>
      <h1 className="font-display font-semibold text-[17px] text-foreground">
        You don&apos;t have access to this project
      </h1>
      <p className="text-[13.5px] text-muted-foreground leading-relaxed">
        Ask a member to send you an invite link.
      </p>
      <Button asChild variant="secondary" size="sm" className="mt-1">
        <Link href={"/" as Route}>Back to projects</Link>
      </Button>
    </div>
  );
}
