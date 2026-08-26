"use client";

import { useState } from "react";
import { motion } from "motion/react";
import { toast } from "sonner";
import { relayApi, RelayError } from "@/lib/relay/api";
import { IconCheck, IconCopy, IconLink, IconSpinner } from "@/components/icons";
import { Button } from "@/components/ui/button";

export function InviteButton({ slug }: { slug: string }) {
  const [url, setUrl] = useState("");
  const [busy, setBusy] = useState(false);
  const [copied, setCopied] = useState(false);

  async function createInvite() {
    setBusy(true);
    try {
      const { token } = await relayApi.createInvite(slug);
      setUrl(`${window.location.origin}/projects/${slug}/invite/${token}`);
    } catch (error) {
      toast.error(
        error instanceof RelayError && error.isMissing
          ? "Only the project owner can invite people."
          : "Could not create an invite.",
      );
    } finally {
      setBusy(false);
    }
  }

  async function copy() {
    try {
      await navigator.clipboard.writeText(url);
      setCopied(true);
      setTimeout(() => setCopied(false), 1800);
    } catch {
      toast.error("Could not copy — select the link manually.");
    }
  }

  if (!url) {
    return (
      <Button
        variant="secondary"
        size="sm"
        onClick={() => void createInvite()}
        disabled={busy}
        className="h-9 gap-1.5 text-[13px]"
      >
        {busy ? <IconSpinner size={14} className="animate-spin" /> : <IconLink size={14} />}
        Invite teammate
      </Button>
    );
  }

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.97 }}
      animate={{ opacity: 1, scale: 1 }}
      className="flex items-center gap-1.5 rounded-lg border border-border bg-card p-1 pl-3"
    >
      <span className="max-w-[240px] truncate font-mono text-[11.5px] text-muted-foreground">
        {url.replace(/^https?:\/\//, "")}
      </span>
      <Button size="sm" onClick={() => void copy()} className="h-7 gap-1.5 text-[12px]">
        {copied ? <IconCheck size={12} /> : <IconCopy size={12} />}
        {copied ? "Copied" : "Copy"}
      </Button>
    </motion.div>
  );
}
