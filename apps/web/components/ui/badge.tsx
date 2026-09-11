import type { ReactNode } from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/cn";

const badgeVariants = cva(
  "inline-flex items-center gap-1.5 rounded-full border px-2.5 py-0.5 font-medium text-2xs tracking-wide",
  {
    variants: {
      tone: {
        neutral: "border-border/70 bg-secondary/50 text-muted-foreground",
        agent: "border-agent/25 bg-agent/10 text-agent",
        human: "border-human/25 bg-human/10 text-human",
        tool: "border-tool/25 bg-tool/10 text-tool",
        live: "border-live/25 bg-live/10 text-live",
      },
    },
    defaultVariants: { tone: "neutral" },
  },
);

export function Badge({
  children,
  tone,
  className,
}: { children: ReactNode; className?: string } & VariantProps<typeof badgeVariants>) {
  return <span className={cn(badgeVariants({ tone }), className)}>{children}</span>;
}
