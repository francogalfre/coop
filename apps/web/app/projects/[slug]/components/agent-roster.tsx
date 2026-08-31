"use client";

import { useState } from "react";
import Link from "next/link";
import type { Route } from "next";
import { motion } from "motion/react";
import { toast } from "sonner";
import { relayApi, RelayError, type Agent } from "@/lib/relay/api";
import { StatusPill } from "@/components/status-pill";
import { IconAgent, IconSend, IconSpinner } from "@/components/icons";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

function statusTone(status: Agent["status"]): "live" | "ended" | "neutral" {
  if (status === "online") return "live";
  if (status === "idle") return "neutral";
  return "ended";
}

function statusLabel(status: Agent["status"]): string {
  if (status === "online") return "Live";
  if (status === "idle") return "Idle";
  return "Offline";
}

function AgentComposer({ slug, agent }: { slug: string; agent: Agent }) {
  const [text, setText] = useState("");
  const [starting, setStarting] = useState(false);

  async function start() {
    const body = text.trim();
    if (!body || starting) return;

    setStarting(true);
    try {
      await relayApi.messageAgent(slug, agent.name, body);
      setText("");
    } catch (error) {
      toast.error(
        error instanceof RelayError && error.status === 503
          ? `${agent.name} isn't running right now.`
          : `Could not message ${agent.name}.`,
      );
      setStarting(false);
      return;
    }
    setStarting(false);
  }

  return (
    <div className="mt-3 flex items-center gap-2 border-border/60 border-t pt-3">
      <Input
        autoFocus
        value={text}
        placeholder={`Message ${agent.name}…`}
        disabled={starting}
        onChange={(e) => setText(e.target.value)}
        onKeyDown={(e) => e.key === "Enter" && void start()}
        className="h-9 text-sm"
      />
      <Button
        size="sm"
        onClick={() => void start()}
        disabled={starting || !text.trim()}
        className="h-9 gap-1.5 text-xs"
      >
        {starting ? <IconSpinner size={13} className="animate-spin" /> : <IconSend size={13} />}
        {starting ? "Starting…" : "Start"}
      </Button>
    </div>
  );
}

function AgentCard({ slug, agent, index }: { slug: string; agent: Agent; index: number }) {
  const [composerOpen, setComposerOpen] = useState(false);

  const body = (
    <div className="flex items-start gap-4">
      <span
        className={`grid size-10 shrink-0 place-items-center rounded-lg border ${
          agent.status === "online"
            ? "border-live/30 bg-live/10 text-live"
            : agent.status === "idle"
              ? "border-border bg-secondary/50 text-foreground"
              : "border-border bg-secondary/30 text-muted-foreground/60"
        }`}
      >
        <IconAgent size={17} />
      </span>

      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <p
            className={`truncate font-display font-medium text-md ${
              agent.status === "offline" ? "text-muted-foreground" : "text-foreground"
            }`}
          >
            {agent.display_name}
          </p>
        </div>
        <p className="mt-1 truncate font-mono text-xs text-muted-foreground">{agent.name}</p>
        <p className="mt-2.5 text-[13px] text-muted-foreground leading-relaxed">
          {agent.status === "offline"
            ? `No one's running ${agent.name} right now. Start it with coop serve --agent=${agent.name} --project=${slug}.`
            : agent.status === "idle"
              ? "Idle — send it a message to start a task."
              : "Running a task now."}
        </p>
      </div>

      <StatusPill tone={statusTone(agent.status)}>{statusLabel(agent.status)}</StatusPill>
    </div>
  );

  return (
    <motion.li
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3, delay: index * 0.04, ease: [0.16, 1, 0.3, 1] }}
    >
      {agent.status === "online" && agent.current_session_id ? (
        <Link
          href={`/sessions/${agent.current_session_id}?from=${slug}` as Route}
          className="group block rounded-xl border border-border bg-card/50 p-5 transition-all hover:border-border/80 hover:bg-card"
        >
          {body}
        </Link>
      ) : (
        <div
          className={`rounded-xl border border-border p-5 ${
            agent.status === "idle" ? "cursor-pointer bg-card/50 transition-all hover:border-border/80 hover:bg-card" : "bg-card/20"
          }`}
          role={agent.status === "idle" ? "button" : undefined}
          tabIndex={agent.status === "idle" ? 0 : undefined}
          aria-expanded={agent.status === "idle" ? composerOpen : undefined}
          onClick={agent.status === "idle" ? () => setComposerOpen((v) => !v) : undefined}
          onKeyDown={
            agent.status === "idle"
              ? (e) => {
                  if (e.key === "Enter" || e.key === " ") {
                    e.preventDefault();
                    setComposerOpen((v) => !v);
                  }
                }
              : undefined
          }
        >
          {body}
          {agent.status === "idle" && composerOpen && (
            <div onClick={(e) => e.stopPropagation()}>
              <AgentComposer slug={slug} agent={agent} />
            </div>
          )}
        </div>
      )}
    </motion.li>
  );
}

export function AgentRoster({ slug, agents }: { slug: string; agents: Agent[] }) {
  return (
    <ul className="space-y-2.5">
      {agents.map((agent, i) => (
        <AgentCard key={agent.id} slug={slug} agent={agent} index={i} />
      ))}
    </ul>
  );
}
