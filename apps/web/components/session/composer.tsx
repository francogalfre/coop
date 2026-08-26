"use client";

import { useEffect, useRef, useState } from "react";
import { AnimatePresence, motion } from "motion/react";
import { toast } from "sonner";
import { IconAgent, IconSend } from "@/components/icons";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

const TYPING_IDLE_MS = 2000;

export function Composer({
  displayName,
  disabled,
  typingNames,
  onSend,
  onTypingChange,
}: {
  displayName: string;
  disabled?: boolean;
  typingNames: string[];
  onSend: (text: string, toAgent: boolean) => Promise<void>;
  onTypingChange: (active: boolean) => void;
}) {
  const [text, setText] = useState("");
  const [toAgent, setToAgent] = useState(true);
  const [sending, setSending] = useState(false);
  const typingRef = useRef(false);
  const idleTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const areaRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    return () => {
      if (idleTimer.current) clearTimeout(idleTimer.current);
    };
  }, []);

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
      <div className="h-5">
        <AnimatePresence>
          {typingNames.length > 0 && (
            <motion.p
              initial={{ opacity: 0, y: 4 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: 4 }}
              className="flex items-center gap-1.5 text-[12px] text-muted-foreground"
            >
              <span className="flex gap-0.5">
                {[0, 1, 2].map((i) => (
                  <motion.span
                    key={i}
                    className="size-1 rounded-full bg-muted-foreground/70"
                    animate={{ opacity: [0.3, 1, 0.3] }}
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
        <textarea
          ref={areaRef}
          rows={1}
          value={text}
          disabled={disabled}
          onChange={(e) => handleChange(e.target.value)}
          onBlur={stopTyping}
          onKeyDown={(e) => {
            if (e.key === "Enter" && !e.shiftKey) {
              e.preventDefault();
              void submit();
            }
          }}
          placeholder={disabled ? "This session has ended" : `Message as ${displayName}…`}
          className="w-full resize-none bg-transparent px-3.5 pt-3 pb-1 text-[13.5px] text-foreground leading-relaxed outline-none placeholder:text-muted-foreground/60 disabled:cursor-not-allowed"
        />

        <div className="flex items-center justify-between gap-2 px-2.5 pb-2.5">
          <button
            type="button"
            onClick={() => setToAgent((v) => !v)}
            disabled={disabled}
            aria-pressed={toAgent}
            className={cn(
              "inline-flex items-center gap-1.5 rounded-lg px-2 py-1.5 text-[12px] transition-all",
              toAgent
                ? "bg-human/15 text-human"
                : "text-muted-foreground hover:bg-secondary hover:text-foreground",
            )}
          >
            <IconAgent size={13} />
            {toAgent ? "Sending to agent" : "Team only"}
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

          <Button
            size="sm"
            onClick={() => void submit()}
            disabled={disabled || sending || !text.trim()}
            className="h-8 gap-1.5 rounded-lg px-3 text-[12.5px]"
          >
            <IconSend size={13} />
            Send
          </Button>
        </div>
      </div>

      <p className="mt-1.5 px-1 text-[11px] text-muted-foreground/60">
        {toAgent
          ? "The agent sees this attributed to you — never as a system instruction."
          : "Only teammates watching this session will see this."}
      </p>
      </div>
    </div>
  );
}
