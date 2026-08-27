"use client";

import { useState } from "react";
import { motion } from "motion/react";
import { toast } from "sonner";
import { relayApi, RelayError, type Project } from "@/lib/relay/api";
import { IconPlus, IconSpinner } from "@/components/icons";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

function slugify(value: string) {
  return value
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9\s-]/g, "")
    .replace(/\s+/g, "-")
    .slice(0, 40);
}

export function CreateProject({ onCreated }: { onCreated: (p: Project) => void }) {
  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");
  const [busy, setBusy] = useState(false);

  async function create() {
    const slug = slugify(name);
    if (!name.trim() || !slug) return;

    setBusy(true);
    try {
      onCreated(await relayApi.createProject(name.trim(), slug));
    } catch (error) {
      toast.error(
        error instanceof RelayError && error.status === 500
          ? "That slug may already be taken."
          : "Could not create the project.",
      );
    } finally {
      setBusy(false);
    }
  }

  if (!open) {
    return (
      <Button
        variant="secondary"
        onClick={() => setOpen(true)}
        className="h-10 w-full gap-2 rounded-xl border border-border border-dashed bg-transparent text-[13.5px] text-muted-foreground hover:bg-card hover:text-foreground"
      >
        <IconPlus size={15} />
        New project
      </Button>
    );
  }

  return (
    <motion.div
      initial={{ opacity: 0, height: 0 }}
      animate={{ opacity: 1, height: "auto" }}
      className="overflow-hidden rounded-xl border border-border bg-card/50 p-4"
    >
      <Input
        autoFocus
        value={name}
        placeholder="Project name"
        onChange={(e) => setName(e.target.value)}
        onKeyDown={(e) => e.key === "Enter" && void create()}
        className="h-10 border-0 bg-transparent px-3 text-md shadow-none focus-visible:ring-0"
      />
      <div className="mt-3 flex items-center gap-2.5 border-border/60 border-t pt-3">
        <span className="flex-1 truncate px-3 font-mono text-2xs text-muted-foreground">
          {name.trim() ? `/projects/${slugify(name)}` : "…"}
        </span>
        <Button variant="ghost" size="sm" onClick={() => setOpen(false)} className="h-9 text-xs">
          Cancel
        </Button>
        <Button
          size="sm"
          onClick={() => void create()}
          disabled={busy || !name.trim()}
          className="h-9 gap-1.5 text-xs"
        >
          {busy && <IconSpinner size={13} className="animate-spin" />}
          Create
        </Button>
      </div>
    </motion.div>
  );
}
