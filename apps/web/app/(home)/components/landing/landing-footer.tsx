import Link from "next/link";
import type { Route } from "next";
import { IconGithub } from "@/components/icons";
import { Mark } from "@/components/brand/mark";

const LINKS: { href: Route; label: string }[] = [
  { href: "/docs" as Route, label: "Docs" },
  { href: "/docs/how-it-works" as Route, label: "How it works" },
  { href: "/docs/cli" as Route, label: "CLI" },
];

export function LandingFooter() {
  return (
    <footer className="flex flex-col gap-5 border-border/60 border-t px-6 py-8 sm:flex-row sm:items-center sm:px-10">
      <div className="flex items-center gap-2 font-display font-semibold text-sm tracking-tight">
        <Mark size={18} />
        coop
      </div>

      <nav className="flex flex-wrap items-center gap-5 sm:ml-8">
        {LINKS.map((link) => (
          <Link
            key={link.href}
            href={link.href}
            className="text-muted-foreground text-xs transition-colors hover:text-foreground"
          >
            {link.label}
          </Link>
        ))}
        <a
          href="https://github.com/francogalfre/coop"
          target="_blank"
          rel="noreferrer"
          className="inline-flex items-center gap-1.5 text-muted-foreground text-xs transition-colors hover:text-foreground"
        >
          <IconGithub size={13} />
          GitHub
        </a>
      </nav>

      <p className="text-muted-foreground/60 text-xs sm:ml-auto">
        Multiplayer sessions for coding agents.
      </p>
    </footer>
  );
}
