"use client";

import { useEffect, useState } from "react";
import { AnimatePresence, motion, useReducedMotion } from "motion/react";
import { IconSend } from "@/components/icons";
import { HarnessLogo } from "@/components/brand/harness-logo";
import { PersonAvatar } from "@/components/ui/person-avatar";
import { LiveDot } from "@/components/ui/status-pill";
import { LiveSessionRow } from "./live-session-row";
import { liveSessionScript, liveSessionViewers } from "./live-session-script";

const STEP_MS = 1500;
const RESTART_MS = 4200;
const SEEDED_ROWS = 3;

function useScriptCursor(enabled: boolean) {
  const [visible, setVisible] = useState(
    enabled ? SEEDED_ROWS : liveSessionScript.length,
  );

  useEffect(() => {
    if (!enabled) return;

    const done = visible >= liveSessionScript.length;
    const timer = setTimeout(
      () => setVisible(done ? SEEDED_ROWS : visible + 1),
      done ? RESTART_MS : STEP_MS,
    );

    return () => clearTimeout(timer);
  }, [enabled, visible]);

  return visible;
}

export function LiveSession() {
  const prefersReducedMotion = useReducedMotion();
  const visible = useScriptCursor(!prefersReducedMotion);
  const rows = liveSessionScript.slice(0, visible);

  return (
    <div className="overflow-hidden rounded-xl border border-border/70 bg-card/40 shadow-[0_24px_70px_-40px_oklch(0_0_0/0.9)] backdrop-blur-sm">
      <div className="flex items-center gap-2.5 border-border/60 border-b bg-background/60 px-3.5 py-2.5">
        <HarnessLogo
          harness="claude-code"
          size={13}
          className="shrink-0 text-muted-foreground"
        />
        <span className="truncate font-medium text-xs text-foreground">
          acme/storefront
        </span>
        <span className="inline-flex items-center gap-1.5 rounded-full bg-live/12 px-2 py-0.5 font-medium text-2xs text-live">
          <LiveDot />
          Live
        </span>
        <div className="ml-auto flex -space-x-1.5">
          {liveSessionViewers.map((name) => (
            <PersonAvatar
              key={name}
              name={name}
              className="size-5 text-[9px] ring-2 ring-card"
            />
          ))}
        </div>
      </div>

      <div className="flex min-h-[264px] flex-col justify-end gap-3 px-3.5 py-4 sm:min-h-[280px]">
        <AnimatePresence initial={false} mode="popLayout">
          {rows.map((row, index) => (
            <motion.div
              key={`${index}-${row.kind}`}
              layout
              initial={prefersReducedMotion ? false : { opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0 }}
              transition={{ duration: 0.32, ease: [0.16, 1, 0.3, 1] }}
            >
              <LiveSessionRow row={row} />
            </motion.div>
          ))}
        </AnimatePresence>
      </div>

      <div className="flex items-center gap-2 border-border/60 border-t bg-background/60 px-3.5 py-2.5">
        <span className="rounded-md bg-secondary/60 px-2 py-1 font-mono text-2xs text-muted-foreground">
          to agent
        </span>
        <span className="flex-1 truncate text-xs text-muted-foreground/50">
          Send the agent a message…
        </span>
        <IconSend size={13} className="shrink-0 text-muted-foreground/50" />
      </div>
    </div>
  );
}
