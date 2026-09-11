"use client";

import { IconCheck, IconCopy } from "@/components/icons";
import { Button } from "@/components/ui/button";
import { useCopy } from "@/lib/hooks/use-copy";
import { cn } from "@/lib/cn";

export function CopyButton({
  value,
  label = "Copy",
  copiedLabel = "Copied",
  showLabel = false,
  className,
}: {
  value: string;
  label?: string;
  copiedLabel?: string;
  showLabel?: boolean;
  className?: string;
}) {
  const { copied, copy } = useCopy();

  return (
    <Button
      type="button"
      variant="ghost"
      size={showLabel ? "sm" : "icon-sm"}
      onClick={() => void copy(value)}
      aria-label={copied ? copiedLabel : label}
      className={cn("text-muted-foreground hover:text-foreground", className)}
    >
      {copied ? <IconCheck size={14} className="text-live" /> : <IconCopy size={14} />}
      {showLabel && <span className="text-xs">{copied ? copiedLabel : label}</span>}
    </Button>
  );
}
