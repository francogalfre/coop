import type { ReactNode } from "react";
import { cn } from "@/lib/cn";

export function Marquee({
  children,
  durationSeconds = 40,
  className,
}: {
  children: ReactNode;
  durationSeconds?: number;
  className?: string;
}) {
  return (
    <div
      className={cn(
        "group relative flex overflow-hidden [mask-image:linear-gradient(90deg,transparent,#000_12%,#000_88%,transparent)]",
        className,
      )}
    >
      <div
        className="flex w-max animate-marquee-x gap-10 group-hover:[animation-play-state:paused] motion-reduce:animate-none"
        style={{ "--marquee-duration": `${durationSeconds}s` } as React.CSSProperties}
      >
        <div className="flex shrink-0 gap-10">{children}</div>
        <div aria-hidden className="flex shrink-0 gap-10">
          {children}
        </div>
      </div>
    </div>
  );
}
