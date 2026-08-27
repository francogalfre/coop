"use client";

import { AnimatePresence, motion } from "motion/react";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { initials, tintFor } from "@/lib/format";
import { cn } from "@/lib/utils";

export function PresenceStack({
  names,
  max = 5,
  className,
}: {
  names: string[];
  max?: number;
  className?: string;
}) {
  const shown = names.slice(0, max);
  const overflow = names.length - shown.length;

  return (
    <div className={cn("flex items-center", className)}>
      <AnimatePresence initial={false}>
        {shown.map((name) => (
          <motion.div
            key={name}
            layout
            initial={{ opacity: 0, scale: 0.6, x: -6 }}
            animate={{ opacity: 1, scale: 1, x: 0 }}
            exit={{ opacity: 0, scale: 0.6, x: -6 }}
            transition={{ type: "spring", stiffness: 500, damping: 32 }}
            className="-ml-2 first:ml-0"
          >
            <Tooltip>
              <TooltipTrigger asChild>
                <button
                  type="button"
                  aria-label={name}
                  className="grid size-7 place-items-center rounded-full ring-2 ring-background font-medium text-2xs text-background select-none"
                  style={{ background: tintFor(name) }}
                >
                  {initials(name)}
                </button>
              </TooltipTrigger>
              <TooltipContent>{name}</TooltipContent>
            </Tooltip>
          </motion.div>
        ))}
      </AnimatePresence>

      {overflow > 0 && (
        <div className="-ml-2 grid size-7 place-items-center rounded-full bg-secondary text-2xs text-muted-foreground ring-2 ring-background">
          +{overflow}
        </div>
      )}
    </div>
  );
}
