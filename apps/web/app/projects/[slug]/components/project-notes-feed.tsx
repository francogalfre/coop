"use client";

import { useCallback, useRef, useState } from "react";
import { toast } from "sonner";
import { relayApi, RelayError, type ProjectNote } from "@/lib/relay/api";
import { useVisibilityPolling } from "@/lib/hooks/useVisibilityPolling";
import { IconAgent, IconMessage, IconSend, IconSpinner } from "@/components/icons";
import { PersonAvatar } from "@/components/person-avatar";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { relativeTime } from "@/lib/format";

function NoteRow({ note }: { note: ProjectNote }) {
  return (
    <li className="flex items-start gap-2.5">
      <PersonAvatar
        name={note.author_display_name}
        avatarUrl={note.author_avatar_url}
        className="size-6 shrink-0 text-3xs"
      />
      <div className="min-w-0 flex-1">
        <div className="flex flex-wrap items-center gap-1.5">
          <span className="font-medium text-xs text-foreground">{note.author_display_name}</span>
          {note.source === "agent" && (
            <span className="inline-flex items-center gap-1 rounded bg-agent/15 px-1.5 py-0.5 text-3xs text-agent">
              <IconAgent size={9} />
              agent
            </span>
          )}
          <span className="text-2xs text-muted-foreground/70">{relativeTime(note.created_at)}</span>
        </div>
        <p className="mt-0.5 whitespace-pre-wrap text-[13px] text-foreground/90 leading-relaxed">{note.text}</p>
      </div>
    </li>
  );
}

function NoteComposer({ slug, onPosted }: { slug: string; onPosted: (note: ProjectNote) => void }) {
  const [text, setText] = useState("");
  const [posting, setPosting] = useState(false);

  async function post() {
    const body = text.trim();
    if (!body || posting) return;

    setPosting(true);
    try {
      const { note } = await relayApi.postProjectNote(slug, body);
      onPosted(note);
      setText("");
    } catch {
      toast.error("Could not post note.");
    } finally {
      setPosting(false);
    }
  }

  return (
    <div className="mb-4 flex items-center gap-2">
      <Input
        value={text}
        placeholder="Share something worth knowing…"
        disabled={posting}
        onChange={(e) => setText(e.target.value)}
        onKeyDown={(e) => e.key === "Enter" && void post()}
        className="h-8 text-[13px]"
      />
      <Button
        size="icon-sm"
        onClick={() => void post()}
        disabled={posting || !text.trim()}
        aria-label="Post note"
      >
        {posting ? <IconSpinner size={13} className="animate-spin" /> : <IconSend size={13} />}
      </Button>
    </div>
  );
}

export function ProjectNotesFeed({ slug }: { slug: string }) {
  const [notes, setNotes] = useState<ProjectNote[] | null>(null);
  const [failed, setFailed] = useState(false);
  const latestLoadRef = useRef<symbol | undefined>(undefined);

  const load = useCallback(async () => {
    const token = Symbol();
    latestLoadRef.current = token;
    try {
      const data = await relayApi.listProjectNotes(slug);
      if (latestLoadRef.current !== token) return;
      setNotes(data.notes);
    } catch (error) {
      if (latestLoadRef.current !== token) return;
      if (!(error instanceof RelayError && error.isMissing)) setFailed(true);
    }
  }, [slug]);

  useVisibilityPolling(load, 8000);

  return (
    <section className="rounded-xl border border-border bg-card/50 p-5">
      <h2 className="mb-3 flex items-center gap-1.5 font-medium text-xs text-muted-foreground uppercase tracking-wider">
        <IconMessage size={13} />
        Notes
      </h2>

      <NoteComposer slug={slug} onPosted={(note) => setNotes((prev) => [note, ...(prev ?? [])])} />

      {failed ? (
        <p className="text-[13px] text-muted-foreground">Could not load notes.</p>
      ) : notes === null ? (
        <div className="space-y-3">
          {[0, 1].map((i) => (
            <div key={i} className="h-8 animate-pulse rounded-lg bg-accent" />
          ))}
        </div>
      ) : notes.length === 0 ? (
        <p className="text-[13px] text-muted-foreground leading-relaxed">
          No notes yet — the first teammate or agent to find something worth sharing starts this feed.
        </p>
      ) : (
        <ul className="space-y-3.5">
          {notes.map((note) => (
            <NoteRow key={note.id} note={note} />
          ))}
        </ul>
      )}
    </section>
  );
}
