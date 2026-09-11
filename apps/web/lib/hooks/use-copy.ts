"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { toast } from "sonner";

export function useCopy(resetAfterMs = 1800) {
  const [copied, setCopied] = useState(false);
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    return () => {
      if (timer.current) clearTimeout(timer.current);
    };
  }, []);

  const copy = useCallback(
    async (value: string) => {
      try {
        await navigator.clipboard.writeText(value);
        setCopied(true);
        if (timer.current) clearTimeout(timer.current);
        timer.current = setTimeout(() => setCopied(false), resetAfterMs);
      } catch {
        toast.error("Could not reach the clipboard. Copy it manually.");
      }
    },
    [resetAfterMs],
  );

  return { copied, copy };
}
