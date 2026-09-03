"use client";

import { motion } from "motion/react";
import { IconAgent, IconPeople } from "@/components/icons";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";

export type SendTarget = "team" | "agent";

export function TargetToggle({
  target,
  onChange,
  disabled,
  agentDisabledReason,
}: {
  target: SendTarget;
  onChange: (target: SendTarget) => void;
  disabled?: boolean;
  agentDisabledReason?: string;
}) {
  const agentDisabled = disabled || Boolean(agentDisabledReason);

  const control = (
    <div className="relative inline-flex rounded-lg bg-secondary/50 p-0.5 text-2xs">
      <motion.span
        layout
        transition={{ type: "spring", stiffness: 550, damping: 33 }}
        className={cn(
          "absolute inset-y-0.5 w-[calc(50%-2px)] rounded-md bg-card",
          target === "agent" ? "right-0.5" : "left-0.5",
        )}
      />
      <button
        type="button"
        onClick={() => onChange("team")}
        disabled={disabled}
        aria-pressed={target === "team"}
        className={cn(
          "relative z-10 flex items-center gap-1.5 rounded-md px-2.5 py-1.5 transition-colors disabled:cursor-not-allowed",
          target === "team" ? "text-foreground" : "text-muted-foreground hover:text-foreground",
        )}
      >
        <IconPeople size={12} />
        Team
      </button>
      <button
        type="button"
        onClick={() => !agentDisabled && onChange("agent")}
        disabled={agentDisabled}
        aria-pressed={target === "agent"}
        className={cn(
          "relative z-10 flex items-center gap-1.5 rounded-md px-2.5 py-1.5 transition-colors disabled:cursor-not-allowed",
          target === "agent" ? "text-human" : "text-muted-foreground hover:text-foreground",
          agentDisabled && "opacity-60",
        )}
      >
        <IconAgent size={12} />
        Agent
      </button>
    </div>
  );

  if (!agentDisabledReason) return control;

  return (
    <Tooltip>
      <TooltipTrigger asChild>{control}</TooltipTrigger>
      <TooltipContent side="top">{agentDisabledReason}</TooltipContent>
    </Tooltip>
  );
}
