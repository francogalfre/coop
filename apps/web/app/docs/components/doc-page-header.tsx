import type { ReactNode } from "react";

export function DocPageHeader({
  eyebrow,
  title,
  intro,
}: {
  eyebrow: string;
  title: string;
  intro: ReactNode;
}) {
  return (
    <div className="mb-10 max-w-2xl">
      <p className="font-mono text-2xs text-muted-foreground uppercase tracking-wide">{eyebrow}</p>
      <h1 className="mt-2 text-balance font-display font-semibold text-[32px] text-foreground leading-tight tracking-tight">
        {title}
      </h1>
      <p className="mt-3 text-[15px] text-muted-foreground leading-relaxed">{intro}</p>
    </div>
  );
}
