"use client";

import { useEffect, useState } from "react";
import type { PresenceState } from "@/lib/relay/presenceState";

export function useTypingNames(
  presence: PresenceState,
  excludeName: string,
  windowMs: number,
): string[] {
  const [names, setNames] = useState<string[]>([]);

  useEffect(() => {
    let timer: ReturnType<typeof setTimeout> | null = null;

    function recompute() {
      const now = Date.now();
      const active: string[] = [];
      let nextExpiry = Infinity;

      for (const [name, entry] of Object.entries(presence)) {
        if (!entry.active || name === excludeName) continue;
        const expiresAt = entry.at + windowMs;
        if (expiresAt <= now) continue;
        active.push(name);
        nextExpiry = Math.min(nextExpiry, expiresAt);
      }

      setNames(active);
      if (nextExpiry !== Infinity) timer = setTimeout(recompute, nextExpiry - now);
    }

    recompute();

    return () => {
      if (timer) clearTimeout(timer);
    };
  }, [presence, excludeName, windowMs]);

  return names;
}
