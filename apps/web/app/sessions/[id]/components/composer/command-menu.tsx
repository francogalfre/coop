"use client";

import { Popover as PopoverPrimitive } from "radix-ui";
import type { Capabilities } from "@coop/protocol";
import { cn } from "@/lib/utils";
import type { Command } from "./lib/commands";

export function CommandMenu({
  commands,
  activeIndex,
  isOwner,
  capabilities,
  onSelect,
}: {
  commands: Command[];
  activeIndex: number;
  isOwner: boolean;
  capabilities: Capabilities | undefined;
  onSelect: (command: Command) => void;
}) {
  return (
    <PopoverPrimitive.Portal>
      <PopoverPrimitive.Content
        side="top"
        align="start"
        sideOffset={8}
        onOpenAutoFocus={(e) => e.preventDefault()}
        onCloseAutoFocus={(e) => e.preventDefault()}
        className="z-50 w-[24rem] max-w-[calc(100vw-2rem)] overflow-hidden rounded-lg border border-border bg-popover p-1 shadow-lg"
      >
        {commands.length === 0 ? (
          <p className="px-2.5 py-2 text-xs text-muted-foreground">no matching commands</p>
        ) : (
          <ul role="listbox" aria-label="Slash commands" className="flex max-h-72 flex-col gap-0.5 overflow-y-auto">
            {commands.map((command, index) => {
              const availability = command.available(capabilities, isOwner);
              const disabled = availability !== true;
              return (
                <li
                  key={command.name}
                  role="option"
                  aria-selected={index === activeIndex}
                  aria-disabled={disabled}
                  onMouseDown={(e) => {
                    e.preventDefault();
                    if (!disabled) onSelect(command);
                  }}
                  className={cn(
                    "flex cursor-pointer flex-col gap-0.5 rounded-md px-3 py-2",
                    index === activeIndex && "bg-secondary",
                    disabled && "cursor-not-allowed opacity-50",
                  )}
                >
                  <div className="flex items-center gap-1.5">
                    <span className="font-mono text-xs font-medium text-foreground">/{command.name}</span>
                    {command.ownerOnly && (
                      <span className="rounded bg-secondary px-1 py-px text-3xs text-muted-foreground">owner</span>
                    )}
                  </div>
                  <span className="text-2xs text-muted-foreground">{command.description}</span>
                  {disabled && <span className="text-3xs text-destructive/80">{availability}</span>}
                </li>
              );
            })}
          </ul>
        )}
      </PopoverPrimitive.Content>
    </PopoverPrimitive.Portal>
  );
}
