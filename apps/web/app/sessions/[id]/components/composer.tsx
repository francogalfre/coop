"use client";

import { useEffect, useId, useRef, useState } from "react";
import { AnimatePresence, motion, useReducedMotion } from "motion/react";
import { toast } from "sonner";
import { IconAgent, IconClose, IconReply, IconSend, IconSpinner } from "@/components/icons";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

const TYPING_IDLE_MS = 2000;

export function Composer({
  displayName,
  disabled,
  typingNames,
  takeoverHeldBy,
  onSend,
  onTypingChange,
  replyingToSeq,
  onDismissReply,
}: {
  displayName: string;
  disabled?: boolean;
  typingNames: string[];
  takeoverHeldBy?: string;
  onSend: (text: string, toAgent: boolean) => Promise<void>;
  onTypingChange: (active: boolean) => void;
  replyingToSeq?: number;
  onDismissReply?: () => void;
}) {
  const [text, setText] = useState("");
  const [toAgent, setToAgent] = useState(true);
  const agentToggleDisabled = disabled || Boolean(takeoverHeldBy) || replyingToSeq !== undefined;
  const [sending, setSending] = useState(false);
  const typingRef = useRef(false);
  const idleTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const areaRef = useRef<HTMLTextAreaElement>(null);
  const textareaId = useId();
  const toggleDescriptionId = useId();
  const prefersReducedMotion = useReducedMotion();

  useEffect(() => {
    return () => {
      if (idleTimer.current) clearTimeout(idleTimer.current);
    };
  }, []);

  useEffect(() => {
    if (agentToggleDisabled) setToAgent(false);
  }, [agentToggleDisabled]);

  useEffect(() => {
    const el = areaRef.current;
    if (!el) return;
    el.style.height = "0px";
    el.style.height = `${Math.min(el.scrollHeight, 180)}px`;
  }, [text]);

  function stopTyping() {
    if (idleTimer.current) {
      clearTimeout(idleTimer.current);
      idleTimer.current = null;
    }
    if (typingRef.current) {
      typingRef.current = false;
      onTypingChange(false);
    }
  }

  function handleChange(value: string) {
    setText(value);
    if (!typingRef.current) {
      typingRef.current = true;
      onTypingChange(true);
    }
    if (idleTimer.current) clearTimeout(idleTimer.current);
    idleTimer.current = setTimeout(stopTyping, TYPING_IDLE_MS);
  }

  async function submit() {
    const body = text.trim();
    if (!body || sending || disabled) return;

    setSending(true);
    try {
      await onSend(body, toAgent);
      setText("");
      stopTyping();
    } catch {
      toast.error("Message failed to send.");
    } finally {
      setSending(false);
    }
  }

  return (
    <div className="shrink-0 border-border/70 border-t bg-card/40 backdrop-blur-sm">
      <div className="mx-auto max-w-3xl px-4 pt-2 pb-3 sm:px-6">
      <AnimatePresence>
        {replyingToSeq !== undefined && (
          <motion.div
            initial={{ opacity: 0, y: 4 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: 4 }}
            transition={{ duration: 0.2, ease: [0.16, 1, 0.3, 1] }}
            className="mb-1.5 flex items-center gap-1.5 text-xs text-human"
          >
            <IconReply size={12} />
            <span>replying to step {replyingToSeq}</span>
            <button
              type="button"
              onClick={onDismissReply}
              className="ml-0.5 grid size-4 place-items-center rounded text-human/70 hover:bg-human/15 hover:text-human"
            >
              <IconClose size={10} />
            </button>
          </motion.div>
        )}
      </AnimatePresence>
      <div className="h-5">
        <AnimatePresence>
          {typingNames.length > 0 && (
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
              {typingNames.join(", ")} {typingNames.length === 1 ? "is" : "are"} typing
            </motion.p>
          )}
        </AnimatePresence>
      </div>

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
            if (e.key === "Enter" && !e.shiftKey) {
              e.preventDefault();
              void submit();
            }
          }}
          placeholder={disabled ? "This session has ended" : `Message as ${displayName}…`}
          className="w-full resize-none bg-transparent px-3.5 pt-3 pb-1 text-sm text-foreground leading-relaxed outline-none placeholder:text-muted-foreground/60 disabled:cursor-not-allowed"
        />

        <div className="flex items-center justify-between gap-2 px-2.5 pb-2.5">
          <button
            type="button"
            onClick={() => setToAgent((v) => !v)}
            disabled={agentToggleDisabled}
            role="switch"
            aria-checked={toAgent}
            aria-describedby={takeoverHeldBy ? toggleDescriptionId : undefined}
            title={takeoverHeldBy ? `${takeoverHeldBy} has taken over — the agent is paused` : undefined}
            className={cn(
              "inline-flex items-center gap-1.5 rounded-lg px-2 py-1.5 text-xs transition-all disabled:cursor-not-allowed disabled:opacity-60",
              toAgent
                ? "bg-human/15 text-human"
                : "text-muted-foreground hover:bg-secondary hover:text-foreground",
            )}
          >
            <IconAgent size={13} />
            {takeoverHeldBy ? `${takeoverHeldBy} has taken over` : toAgent ? "Sending to agent" : "Team only"}
            <span
              className={cn(
                "relative ml-0.5 h-3.5 w-6 rounded-full transition-colors",
                toAgent ? "bg-human/50" : "bg-secondary",
              )}
            >
              <motion.span
                layout
                transition={{ type: "spring", stiffness: 600, damping: 34 }}
                className={cn(
                  "absolute top-0.5 size-2.5 rounded-full",
                  toAgent ? "left-3 bg-human" : "left-0.5 bg-muted-foreground",
                )}
              />
            </span>
          </button>
          {takeoverHeldBy && (
            <span id={toggleDescriptionId} className="sr-only">
              {`${takeoverHeldBy} has taken over — the agent is paused`}
            </span>
          )}

          <Button
            size="sm"
            onClick={() => void submit()}
            disabled={disabled || sending || !text.trim()}
            className="h-8 gap-1.5 rounded-lg px-3 text-xs"
          >
            {sending ? <IconSpinner size={13} className="animate-spin" /> : <IconSend size={13} />}
            {sending ? "Sending…" : "Send"}
          </Button>
        </div>
      </div>

      <p className="mt-1.5 px-1 text-2xs text-muted-foreground/60">
        {toAgent
          ? "The agent sees this attributed to you — never as a system instruction."
          : "Only teammates watching this session will see this."}
      </p>
      </div>
    </div>
  );
}
