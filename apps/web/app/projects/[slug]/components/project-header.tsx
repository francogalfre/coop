import Link from "next/link";
import type { Route } from "next";
import { IconFolder } from "@/components/icons";
import { Reveal } from "@/components/ui/reveal";
import { InviteButton } from "./invite-button";

export function ProjectHeader({ slug }: { slug: string }) {
  return (
    <Reveal className="mb-8 flex flex-wrap items-end justify-between gap-3">
      <div>
        <div className="flex items-center gap-2 text-muted-foreground text-xs">
          <IconFolder size={13} />
          <Link href={"/" as Route} className="transition-colors hover:text-foreground">
            projects
          </Link>
          <span className="text-muted-foreground/40">/</span>
        </div>
        <h1 className="mt-1.5 font-display font-semibold text-[26px] text-foreground tracking-tight">
          {slug}
        </h1>
      </div>
      <InviteButton slug={slug} />
    </Reveal>
  );
}
