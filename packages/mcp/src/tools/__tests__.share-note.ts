import type { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { Config } from "../config/config.js";

vi.mock("../relay/notes.js", () => ({ shareNote: vi.fn(), listNotes: vi.fn() }));

import { listNotes, shareNote } from "../relay/notes.js";
import { RelayUnreachableError } from "../relay/types.js";
import { registerListProjectNotes } from "./list-project-notes.js";
import { registerShareNote } from "./share-note.js";

type ToolHandler = (...args: unknown[]) => Promise<{
  isError?: boolean;
  content: { type: "text"; text: string }[];
  structuredContent?: Record<string, unknown>;
}>;

function fakeServer() {
  const tools = new Map<string, ToolHandler>();
  const registerTool = vi.fn((name: string, _config: unknown, handler: ToolHandler) => {
    tools.set(name, handler);
  });
  return { server: { registerTool } as unknown as McpServer, tools };
}

const configWithProject: Config = { relayUrl: "http://localhost:8787", repo: "/repo", project: "acme" };
const configWithoutProject: Config = { relayUrl: "http://localhost:8787", repo: "/repo" };

describe("share_note", () => {
  beforeEach(() => {
    vi.mocked(shareNote).mockReset();
  });

  it("shares a note when a project is configured", async () => {
    const { server, tools } = fakeServer();
    registerShareNote(server, configWithProject);
    vi.mocked(shareNote).mockResolvedValue({
      id: "note-1",
      authorDisplayName: "coop-mcp",
      source: "agent",
      text: "found it",
      createdAt: "2026-09-03T10:00:00Z",
    });

    const result = await tools.get("share_note")!({ text: "found it" }, {});

    expect(result.isError).toBeUndefined();
    expect(shareNote).toHaveBeenCalledWith(configWithProject, "acme", "found it");
  });

  it("fails clearly when no project is configured", async () => {
    const { server, tools } = fakeServer();
    registerShareNote(server, configWithoutProject);

    const result = await tools.get("share_note")!({ text: "found it" }, {});

    expect(result.isError).toBe(true);
    expect(result.content[0]!.text).toContain("no project configured");
    expect(shareNote).not.toHaveBeenCalled();
  });

  it("soft-fails when the relay is unreachable", async () => {
    const { server, tools } = fakeServer();
    registerShareNote(server, configWithProject);
    vi.mocked(shareNote).mockRejectedValue(new RelayUnreachableError("connection refused"));

    const result = await tools.get("share_note")!({ text: "found it" }, {});

    expect(result.isError).toBe(true);
    expect(result.content[0]!.text).toContain("connection refused");
  });
});

describe("list_project_notes", () => {
  beforeEach(() => {
    vi.mocked(listNotes).mockReset();
  });

  it("lists notes when a project is configured", async () => {
    const { server, tools } = fakeServer();
    registerListProjectNotes(server, configWithProject);
    vi.mocked(listNotes).mockResolvedValue([
      { id: "note-1", authorDisplayName: "Mara", source: "human", text: "hi", createdAt: "2026-09-03T10:00:00Z" },
    ]);

    const result = await tools.get("list_project_notes")!({}, {});

    expect(result.content[0]!.text).toBe("1 note(s).");
    expect(listNotes).toHaveBeenCalledWith(configWithProject, "acme", undefined);
  });

  it("reports no notes without erroring", async () => {
    const { server, tools } = fakeServer();
    registerListProjectNotes(server, configWithProject);
    vi.mocked(listNotes).mockResolvedValue([]);

    const result = await tools.get("list_project_notes")!({}, {});

    expect(result.isError).toBeUndefined();
    expect(result.content[0]!.text).toBe("No notes yet.");
  });

  it("fails clearly when no project is configured", async () => {
    const { server, tools } = fakeServer();
    registerListProjectNotes(server, configWithoutProject);

    const result = await tools.get("list_project_notes")!({}, {});

    expect(result.isError).toBe(true);
    expect(listNotes).not.toHaveBeenCalled();
  });
});
