import type { ReactNode } from "react";
import { cn } from "@/lib/cn";

export function Container({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) {
  return (
    <div className="bg-hatch">
      <div
        className={cn(
          "relative mx-auto min-h-dvh w-full max-w-[1120px] border-border bg-background lg:border-x",
          className,
        )}
      >
        {children}
      </div>
    </div>
  );
}
