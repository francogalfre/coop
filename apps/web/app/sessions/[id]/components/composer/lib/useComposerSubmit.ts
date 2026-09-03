"use client";

import { useEffect, useRef, useState } from "react";
import { toast } from "sonner";
import type { Capabilities } from "@coop/protocol";
import { relayApi } from "@/lib/relay/api";
import type { PendingPatch, PendingSend } from "../../../lib/usePendingSends";
import type { SendTarget } from "../target-toggle";
import { COMMANDS, filterCommands, parseCommandInvocation, type Command, type CommandContext } from "./commands";

const TYPING_IDLE_MS = 2000;

export function useComposerSubmit({
  sessionId,
  displayName,
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
}: {
  sessionId: string;
  displayName: string;
  isOwner: boolean;
  disabled?: boolean;
  takeoverHeldBy?: string;
  heldByMe?: boolean;
  capabilities?: Capabilities;
  agentBusy: boolean;
  replyingToSeq?: number;
  onTypingChange: (active: boolean) => void;
  onDismissReply?: () => void;
  onPendingSend: (send: PendingSend) => void;
  onPendingUpdate: (clientId: string, patch: PendingPatch) => void;
}) {
  const [text, setText] = useState("");
  const [target, setTarget] = useState<SendTarget>("agent");
  const [sending, setSending] = useState(false);
  const [queueDepth, setQueueDepth] = useState<number | undefined>(undefined);
  const [paletteIndex, setPaletteIndex] = useState(0);
  const typingRef = useRef(false);
  const idleTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  const slashQuery = !text.includes("\n") && text.startsWith("/") ? text.slice(1).split(/\s/)[0] : null;
  const paletteOpen = slashQuery !== null;
  const paletteMatches = paletteOpen ? filterCommands(COMMANDS, slashQuery ?? "") : [];
  const sendLabel = sending ? "Sending…" : target === "agent" && agentBusy ? "Queue" : "Send";

  useEffect(() => {
    return () => {
      if (idleTimer.current) clearTimeout(idleTimer.current);
    };
  }, []);

  useEffect(() => {
    if (disabled || takeoverHeldBy) setTarget("team");
  }, [disabled, takeoverHeldBy]);

  useEffect(() => {
    setPaletteIndex(0);
  }, [slashQuery]);

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

  async function sendMessage(body: string, sendTarget: SendTarget) {
    const clientId = crypto.randomUUID();
    const toAgent = sendTarget === "agent";
    onPendingSend({ clientId, author: displayName, text: body, toAgent, anchorSeq: replyingToSeq });

    try {
      if (toAgent) {
        const result = await relayApi.steerAgent(sessionId, body, clientId, replyingToSeq);
        if (result.status === "accepted") {
          setQueueDepth(result.queued);
          onPendingUpdate(clientId, { queuePosition: result.queued });
        } else {
          toast("Waiting for the owner to approve.");
        }
      } else {
        await relayApi.sendTeamMessage(sessionId, body, clientId, replyingToSeq);
      }
      onDismissReply?.();
    } catch {
      onPendingUpdate(clientId, { delivery: "dropped" });
      toast.error("Message failed to send.");
    }
  }

  async function runCommand(command: Command, args: string) {
    const ctx: CommandContext = {
      args,
      isOwner,
      heldByMe: Boolean(heldByMe),
      takeoverHeldBy,
      setTarget,
      ask: (question) => sendMessage(question, "agent"),
      setTakeover: async (active) => {
        await relayApi.setTakeover(sessionId, active);
      },
      setMode: async (mode) => {
        await relayApi.setSessionMode(sessionId, mode);
      },
      runHarnessCommand: async (cmd, cmdArgs) => {
        await relayApi.runCommand(sessionId, cmd, cmdArgs);
      },
    };

    try {
      await command.run(ctx);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Command failed.");
    }
  }

  async function chooseCommand(command: Command, focusArea: () => void) {
    const availability = command.available(capabilities, isOwner);
    if (availability !== true) {
      toast.error(availability);
      return;
    }
    if (command.args) {
      setText(`/${command.name} `);
      focusArea();
      return;
    }
    setText("");
    await runCommand(command, "");
  }

  async function submit() {
    const body = text.trim();
    if (!body || sending || disabled) return;

    const invocation = parseCommandInvocation(body);
    const command = invocation ? COMMANDS.find((c) => c.name === invocation.name) : undefined;

    setSending(true);
    try {
      if (invocation && command) {
        const availability = command.available(capabilities, isOwner);
        if (availability !== true) {
          toast.error(availability);
          return;
        }
        setText("");
        stopTyping();
        await runCommand(command, invocation.args);
        return;
      }

      setText("");
      stopTyping();
      await sendMessage(body, target);
    } finally {
      setSending(false);
    }
  }

  return {
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
  };
}
