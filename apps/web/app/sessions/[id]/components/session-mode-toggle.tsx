"use client";

import { useEffect, useState } from "react";
import { toast } from "sonner";
import { IconLock, IconUnlock } from "@/components/icons";
import { Button } from "@/components/ui/button";
import { MetaChip } from "@/components/status-pill";
import { relayApi } from "@/lib/relay/api";
import { cn } from "@/lib/utils";

const MODE_LABEL: Record<"auto" | "restricted", string> = {
  auto: "Auto",
  restricted: "Restricted",
};

const MODE_HINT: Record<"auto" | "restricted", string> = {
  auto: "Auto: teammates' messages reach the agent immediately.",
  restricted: "Restricted: your approval is required before a teammate's message reaches the agent.",
};

export function SessionModeToggle({
  sessionId,
  mode,
  isOwner,
}: {
  sessionId: string;
  mode: "auto" | "restricted";
  isOwner: boolean;
}) {
  const [pending, setPending] = useState<"auto" | "restricted" | null>(null);
  const current = pending ?? mode;

  useEffect(() => {
    if (pending !== null && mode === pending) setPending(null);
  }, [mode, pending]);

  if (!isOwner) {
    return (
      <span title={MODE_HINT[current]}>
        <MetaChip className="hidden sm:inline-flex">{MODE_LABEL[current]}</MetaChip>
      </span>
    );
  }

  async function toggle() {
    const next = current === "auto" ? "restricted" : "auto";
    setPending(next);
    try {
      await relayApi.setSessionMode(sessionId, next);
    } catch {
      setPending(null);
      toast.error("Failed to update session mode.");
    }
  }

  return (
    <Button
      size="sm"
      variant={current === "restricted" ? "default" : "outline"}
      onClick={() => void toggle()}
      title={MODE_HINT[current]}
      className={cn("h-7 gap-1.5 rounded-full px-2.5 text-xs")}
    >
      {current === "restricted" ? <IconLock size={12} /> : <IconUnlock size={12} />}
      {MODE_LABEL[current]}
    </Button>
  );
}
