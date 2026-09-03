"use client";

import { AnimatePresence, motion, useReducedMotion } from "motion/react";

export function TypingIndicator({ names }: { names: string[] }) {
  const prefersReducedMotion = useReducedMotion();

  return (
    <div className="h-5">
      <AnimatePresence>
        {names.length > 0 && (
          <motion.p
            initial={{ opacity: 0, y: 4 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: 4 }}
            className="flex items-center gap-1.5 text-xs text-muted-foreground"
          >
            <span className="flex gap-0.5">
              {[0, 1, 2].map((i) => (
                <motion.span
                  key={i}
                  className="size-1 rounded-full bg-muted-foreground/70"
                  animate={prefersReducedMotion ? undefined : { opacity: [0.3, 1, 0.3] }}
                  transition={{ duration: 1.1, repeat: Infinity, delay: i * 0.18 }}
                />
              ))}
            </span>
            {names.join(", ")} {names.length === 1 ? "is" : "are"} typing
          </motion.p>
        )}
      </AnimatePresence>
    </div>
  );
}
