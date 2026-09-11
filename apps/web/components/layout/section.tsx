import type { ReactNode } from "react";
import { cn } from "@/lib/cn";
import { Crosshair } from "@/components/layout/crosshair";

export function Section({
  children,
  className,
  id,
  ruled = true,
}: {
  children: ReactNode;
  className?: string;
  id?: string;
  ruled?: boolean;
}) {
  return (
    <section
      id={id}
      className={cn(
        "relative px-6 py-16 sm:px-10 sm:py-24",
        ruled && "border-border/60 border-t",
        className,
      )}
    >
      {ruled && (
        <>
          <Crosshair className="-top-1 -left-1" />
          <Crosshair className="-top-1 -right-1" />
        </>
      )}
      {children}
    </section>
  );
}
