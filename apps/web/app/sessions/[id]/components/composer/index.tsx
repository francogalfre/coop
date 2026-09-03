"use client";

import { useEffect, useId, useRef } from "react";
import { Popover as PopoverPrimitive } from "radix-ui";
import type { Capabilities } from "@coop/protocol";
import { IconSend, IconSpinner } from "@/components/icons";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import type { PendingPatch, PendingSend } from "../../lib/usePendingSends";
import { CommandMenu } from "./command-menu";
import { QueueStrip } from "./queue-strip";
import { ReplyBanner } from "./reply-banner";
import { TargetToggle } from "./target-toggle";
import { TypingIndicator } from "./typing-indicator";
import { useComposerSubmit } from "./lib/useComposerSubmit";

export function Composer({
  sessionId,
  displayName,
  userAvatarUrl,
  isOwner,
  disabled,
  typingNames,
  takeoverHeldBy,
  heldByMe,
  agentBusy,
  agentState,
  capabilities,
  onTypingChange,
  replyingToSeq,
  onDismissReply,
  onPendingSend,
  onPendingUpdate,
}: {
  sessionId: string;
  displayName: string;
  userAvatarUrl?: string;
  isOwner: boolean;
  disabled?: boolean;
  typingNames: string[];
  takeoverHeldBy?: string;
  heldByMe?: boolean;
  agentBusy: boolean;
  agentState?: string;
  capabilities?: Capabilities;
  onTypingChange: (active: boolean) => void;
  replyingToSeq?: number;
  onDismissReply?: () => void;
  onPendingSend: (send: PendingSend) => void;
  onPendingUpdate: (clientId: string, patch: PendingPatch) => void;
}) {
  const areaRef = useRef<HTMLTextAreaElement>(null);
  const textareaId = useId();
  const toggleDescriptionId = useId();

  const {
    text,
    setText,
    target,
    setTarget,
    sending,
    queueDepth,
    handleChange,
    submit,
    chooseCommand,
    paletteOpen,
    paletteMatches,
    paletteIndex,
    setPaletteIndex,
    sendLabel,
    stopTyping,
  } = useComposerSubmit({
    sessionId,
    displayName,
    userAvatarUrl,
    isOwner,
    disabled,
    takeoverHeldBy,
    heldByMe,
    capabilities,
    agentBusy,
    replyingToSeq,
    onTypingChange,
    onDismissReply,
    onPendingSend,
    onPendingUpdate,
  });

  const agentDisabledReason = takeoverHeldBy
    ? `${takeoverHeldBy} has taken over — the agent is paused`
    : undefined;

  useEffect(() => {
    const el = areaRef.current;
    if (!el) return;
    el.style.height = "0px";
    el.style.height = `${Math.min(el.scrollHeight, 180)}px`;
  }, [text]);

  return (
    <div className="shrink-0 border-border/70 border-t bg-card/40 backdrop-blur-sm">
      <div className="mx-auto max-w-3xl px-4 pt-2 pb-3 sm:px-6">
        <ReplyBanner seq={replyingToSeq} onDismiss={onDismissReply} />
        <TypingIndicator names={typingNames} />

        <PopoverPrimitive.Root open={paletteOpen}>
          <PopoverPrimitive.Anchor asChild>
            <div
              className={cn(
                "rounded-xl border border-border bg-background/70 transition-colors",
                "focus-within:border-ring/60 focus-within:bg-background",
              )}
            >
              <label htmlFor={textareaId} className="sr-only">
                Message as {displayName}
              </label>
              <textarea
                ref={areaRef}
                id={textareaId}
                rows={1}
                value={text}
                disabled={disabled || sending}
                onChange={(e) => handleChange(e.target.value)}
                onBlur={stopTyping}
                onKeyDown={(e) => {
                  if (paletteOpen && paletteMatches.length > 0) {
                    if (e.key === "ArrowDown") {
                      e.preventDefault();
                      setPaletteIndex((i) => (i + 1) % paletteMatches.length);
                      return;
                    }
                    if (e.key === "ArrowUp") {
                      e.preventDefault();
                      setPaletteIndex((i) => (i - 1 + paletteMatches.length) % paletteMatches.length);
                      return;
                    }
                    if (e.key === "Enter" && !e.shiftKey) {
                      e.preventDefault();
                      const active = paletteMatches[paletteIndex];
                      if (active) void chooseCommand(active, () => areaRef.current?.focus());
                      return;
                    }
                  }
                  if (e.key === "Escape" && paletteOpen) {
                    e.preventDefault();
                    setText("");
                    return;
                  }
                  if (e.key === "Enter" && !e.shiftKey) {
                    e.preventDefault();
                    void submit();
                  }
                }}
                placeholder={disabled ? "This session has ended" : `Message as ${displayName}…`}
                className="w-full resize-none bg-transparent px-3.5 pt-3 pb-1.5 text-sm text-foreground leading-relaxed outline-none placeholder:text-muted-foreground/60 disabled:cursor-not-allowed"
              />

              <div className="flex items-center justify-between gap-2 px-2.5 pb-2.5">
                <div className="flex min-w-0 items-center gap-2">
                  <div aria-describedby={takeoverHeldBy ? toggleDescriptionId : undefined}>
                    <TargetToggle
                      target={target}
                      onChange={setTarget}
                      disabled={disabled}
                      agentDisabledReason={agentDisabledReason}
                    />
                  </div>
                  <QueueStrip agentBusy={agentBusy} agentState={agentState} queueDepth={queueDepth} />
                </div>
                {takeoverHeldBy && (
                  <span id={toggleDescriptionId} className="sr-only">
                    {agentDisabledReason}
                  </span>
                )}

                <Button
                  size="sm"
                  onClick={() => void submit()}
                  disabled={disabled || sending || !text.trim()}
                  className={cn(
                    "h-8 shrink-0 gap-1.5 rounded-lg px-3.5 text-xs font-medium",
                    sendLabel === "Queue" && "bg-agent text-white hover:bg-agent/90",
                  )}
                >
                  {sending ? <IconSpinner size={13} className="animate-spin" /> : <IconSend size={13} />}
                  {sendLabel}
                </Button>
              </div>
            </div>
          </PopoverPrimitive.Anchor>

          <CommandMenu
            commands={paletteMatches}
            activeIndex={paletteIndex}
            isOwner={isOwner}
            capabilities={capabilities}
            onSelect={(command) => void chooseCommand(command, () => areaRef.current?.focus())}
          />
        </PopoverPrimitive.Root>

        <p className="mt-1.5 flex items-center gap-1.5 px-1 text-2xs text-muted-foreground/60">
          <span>
            {target === "agent"
              ? "The agent sees this attributed to you — never as a system instruction."
              : "Only teammates watching this session will see this."}
          </span>
          <span className="text-muted-foreground/40">·</span>
          <kbd className="rounded bg-secondary/70 px-1 font-mono text-[0.95em] text-muted-foreground/80">/</kbd>
          <span>for commands</span>
        </p>
      </div>
    </div>
  );
}
