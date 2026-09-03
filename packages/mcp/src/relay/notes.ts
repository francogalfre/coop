import type { Config } from "../config/config.js";
import { getJson, postJson } from "./request.js";
import type { ProjectNote } from "./types.js";

type RawProjectNote = {
  id: string;
  author_id: string;
  author_display_name: string;
  author_avatar_url?: string;
  source: string;
  session_id?: string;
  text: string;
  created_at: string;
};

function fromRaw(raw: RawProjectNote): ProjectNote {
  return {
    id: raw.id,
    authorDisplayName: raw.author_display_name,
    source: raw.source === "agent" ? "agent" : "human",
    sessionId: raw.session_id,
    text: raw.text,
    createdAt: raw.created_at,
  };
}

export async function shareNote(config: Config, project: string, text: string): Promise<ProjectNote> {
  const url = new URL(`/v1/projects/${encodeURIComponent(project)}/notes`, config.relayUrl);
  const raw = (await postJson(url, config, { text, source: "agent", session_id: config.sessionId })) as {
    note: RawProjectNote;
  };

  return fromRaw(raw.note);
}

export async function listNotes(config: Config, project: string, limit?: number): Promise<ProjectNote[]> {
  const params = limit ? `?limit=${limit}` : "";
  const url = new URL(`/v1/projects/${encodeURIComponent(project)}/notes${params}`, config.relayUrl);
  const raw = (await getJson(url, config)) as { notes: RawProjectNote[] };

  return raw.notes.map(fromRaw);
}
