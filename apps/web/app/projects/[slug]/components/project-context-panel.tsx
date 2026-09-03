"use client";

import { type ChangeEvent, useCallback, useRef, useState } from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { toast } from "sonner";
import { relayApi, RelayError, type ProjectContext } from "@/lib/relay/api";
import { useVisibilityPolling } from "@/lib/hooks/useVisibilityPolling";
import { IconEdit, IconFile } from "@/components/icons";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { relativeTime } from "@/lib/format";

function ContextEditor({
  draft,
  onDraftChange,
  onSave,
  onCancel,
}: {
  draft: string;
  onDraftChange: (value: string) => void;
  onSave: () => void;
  onCancel: () => void;
}) {
  const fileInputRef = useRef<HTMLInputElement>(null);

  async function loadFile(event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    event.target.value = "";
    if (!file) return;
    if (file.size > 1_000_000) {
      toast.error("File is too large (max 1 MB).");
      return;
    }
    try {
      onDraftChange(await file.text());
    } catch {
      toast.error("Couldn't read that file.");
    }
  }

  return (
    <div>
      <textarea
        autoFocus
        value={draft}
        onChange={(e) => onDraftChange(e.target.value)}
        rows={8}
        placeholder="What should every new session already know?"
        className="w-full resize-none rounded-lg border border-border bg-background/70 px-3 py-2.5 text-sm text-foreground leading-relaxed outline-none focus:border-ring/60"
      />
      <input
        ref={fileInputRef}
        type="file"
        accept=".md,.markdown,text/markdown,text/plain"
        onChange={(e) => void loadFile(e)}
        hidden
      />
      <div className="mt-2 flex justify-end gap-2">
        <Button
          variant="ghost"
          size="sm"
          onClick={() => fileInputRef.current?.click()}
          className="mr-auto h-7 text-xs"
        >
          <IconFile size={13} />
          Upload .md
        </Button>
        <Button variant="ghost" size="sm" onClick={onCancel} className="h-7 text-xs">
          Cancel
        </Button>
        <Button size="sm" onClick={onSave} className="h-7 text-xs">
          Save
        </Button>
      </div>
    </div>
  );
}

export function ProjectContextPanel({ slug }: { slug: string }) {
  const [context, setContext] = useState<ProjectContext | null>(null);
  const [failed, setFailed] = useState(false);
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState("");
  const editingRef = useRef(editing);
  editingRef.current = editing;
  const latestLoadRef = useRef<symbol | undefined>(undefined);

  const load = useCallback(async () => {
    if (editingRef.current) return;
    const token = Symbol();
    latestLoadRef.current = token;
    try {
      const data = await relayApi.getProjectContext(slug);
      if (latestLoadRef.current !== token) return;
      setContext(data);
    } catch (error) {
      if (latestLoadRef.current !== token) return;
      if (!(error instanceof RelayError && error.isMissing)) setFailed(true);
    }
  }, [slug]);

  useVisibilityPolling(load, 8000);

  function startEdit() {
    if (!context) return;
    setDraft(context.text);
    setEditing(true);
  }

  async function save() {
    if (!context) return;
    const previous = context;
    setContext({ ...context, text: draft, version: context.version + 1 });
    setEditing(false);
    try {
      const next = await relayApi.setProjectContext(slug, draft);
      setContext(next);
    } catch {
      setContext(previous);
      setEditing(true);
      toast.error("Could not save project context.");
    }
  }

  return (
    <section className="rounded-xl border border-border bg-card/50 p-5">
      <div className="mb-3 flex items-center justify-between gap-2">
        <h2 className="flex items-center gap-1.5 font-medium text-xs text-muted-foreground uppercase tracking-wider">
          <IconFile size={13} />
          Project context
        </h2>
        {context && !editing && (
          <Button
            variant="ghost"
            size="icon-xs"
            onClick={startEdit}
            aria-label="Edit project context"
            className="text-muted-foreground hover:text-foreground"
          >
            <IconEdit size={13} />
          </Button>
        )}
      </div>

      {failed ? (
        <p className="text-[13px] text-muted-foreground">Could not load project context.</p>
      ) : context === null ? (
        <Skeleton className="h-20 rounded-lg" />
      ) : editing ? (
        <ContextEditor draft={draft} onDraftChange={setDraft} onSave={() => void save()} onCancel={() => setEditing(false)} />
      ) : context.text.trim() === "" ? (
        <p className="text-[13px] text-muted-foreground leading-relaxed">
          No context yet — add one so every new session starts with what your team already knows.
        </p>
      ) : (
        <>
          <div className="prose-timeline max-w-none text-sm text-foreground/90 leading-relaxed">
            <ReactMarkdown remarkPlugins={[remarkGfm]}>{context.text}</ReactMarkdown>
          </div>
          <p className="mt-3 text-2xs text-muted-foreground/70">
            v{context.version}
            {context.updated_by && ` · edited by ${context.updated_by}`}
            {context.updated_at && ` · ${relativeTime(context.updated_at)}`}
          </p>
        </>
      )}
    </section>
  );
}
