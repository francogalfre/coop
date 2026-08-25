import type { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { Config } from "../config/config.js";

vi.mock("../relay/client.js", () => ({
  fetchPresence: vi.fn(),
  fetchActiveSessions: vi.fn(),
}));

import { fetchActiveSessions, fetchPresence } from "../relay/client.js";
import { RelayUnreachableError } from "../relay/types.js";
import { registerCheckConflicts } from "./check-conflicts.js";
import { registerListActiveSessions } from "./list-active-sessions.js";

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

const config: Config = { relayUrl: "http://localhost:8787", repo: "/repo", sessionId: "sess-self" };

describe("check_conflicts", () => {
  beforeEach(() => {
    vi.mocked(fetchPresence).mockReset();
  });

  it("reports clear when there are no signals", async () => {
    const { server, tools } = fakeServer();
    registerCheckConflicts(server, config);
    vi.mocked(fetchPresence).mockResolvedValue({
      repo: "/repo",
      windowSeconds: 900,
      paths: [{ path: "a.ts", signals: [] }],
    });

    const result = await tools.get("check_conflicts")!({ paths: ["a.ts"] }, {});

    expect(result.structuredContent).toEqual({
      repo: "/repo",
      window_seconds: 900,
      conflicts: [],
      clear: ["a.ts"],
    });
  });

  it("flags another active session's write as a conflict", async () => {
    const { server, tools } = fakeServer();
    registerCheckConflicts(server, config);
    vi.mocked(fetchPresence).mockResolvedValue({
      repo: "/repo",
      windowSeconds: 900,
      paths: [
        {
          path: "a.ts",
          signals: [
            {
              path: "a.ts",
              sessionId: "sess-other",
              owner: "Alice",
              mode: "write",
              at: new Date().toISOString(),
              active: true,
            },
          ],
        },
      ],
    });

    const result = await tools.get("check_conflicts")!({ paths: ["a.ts"] }, {});

    expect(result.structuredContent?.clear).toEqual([]);
    const conflicts = result.structuredContent?.conflicts as { path: string; session_id: string }[];
    expect(conflicts).toHaveLength(1);
    expect(conflicts[0]).toMatchObject({ path: "a.ts", session_id: "sess-other" });
  });

  it("treats a read-only touch as info, not a conflict", async () => {
    const { server, tools } = fakeServer();
    registerCheckConflicts(server, config);
    vi.mocked(fetchPresence).mockResolvedValue({
      repo: "/repo",
      windowSeconds: 900,
      paths: [
        {
          path: "a.ts",
          signals: [
            {
              path: "a.ts",
              sessionId: "sess-other",
              mode: "read",
              at: new Date().toISOString(),
              active: true,
            },
          ],
        },
      ],
    });

    const result = await tools.get("check_conflicts")!({ paths: ["a.ts"] }, {});

    expect(result.structuredContent).toMatchObject({ conflicts: [], clear: ["a.ts"] });
  });

  it("treats an ended session's write as info, not a conflict", async () => {
    const { server, tools } = fakeServer();
    registerCheckConflicts(server, config);
    vi.mocked(fetchPresence).mockResolvedValue({
      repo: "/repo",
      windowSeconds: 900,
      paths: [
        {
          path: "a.ts",
          signals: [
            {
              path: "a.ts",
              sessionId: "sess-other",
              mode: "write",
              at: new Date().toISOString(),
              active: false,
            },
          ],
        },
      ],
    });

    const result = await tools.get("check_conflicts")!({ paths: ["a.ts"] }, {});

    expect(result.structuredContent).toMatchObject({ conflicts: [], clear: ["a.ts"] });
  });

  it("excludes the caller's own session from conflicts", async () => {
    const { server, tools } = fakeServer();
    registerCheckConflicts(server, config);
    vi.mocked(fetchPresence).mockResolvedValue({
      repo: "/repo",
      windowSeconds: 900,
      paths: [
        {
          path: "a.ts",
          signals: [
            {
              path: "a.ts",
              sessionId: config.sessionId as string,
              mode: "write",
              at: new Date().toISOString(),
              active: true,
            },
          ],
        },
      ],
    });

    const result = await tools.get("check_conflicts")!({ paths: ["a.ts"] }, {});

    expect(result.structuredContent).toMatchObject({ conflicts: [], clear: ["a.ts"] });
  });

  it("returns a tool error instead of throwing when the relay is unreachable", async () => {
    const { server, tools } = fakeServer();
    registerCheckConflicts(server, config);
    vi.mocked(fetchPresence).mockRejectedValue(new RelayUnreachableError("relay down"));

    const result = await tools.get("check_conflicts")!({ paths: ["a.ts"] }, {});

    expect(result.isError).toBe(true);
    expect(result.content[0]?.text).toContain("relay down");
  });
});

describe("list_active_sessions", () => {
  beforeEach(() => {
    vi.mocked(fetchActiveSessions).mockReset();
  });

  it("passes through the relay's session list", async () => {
    const { server, tools } = fakeServer();
    registerListActiveSessions(server, config);
    const sessions = [{ sessionId: "sess-a", owner: "Alice", startedAt: "2026-08-24T09:50:00Z", active: true }];
    vi.mocked(fetchActiveSessions).mockResolvedValue(sessions);

    const result = await tools.get("list_active_sessions")!({});

    expect(result.structuredContent).toEqual({ repo: "/repo", sessions });
  });

  it("returns a tool error instead of throwing when the relay is unreachable", async () => {
    const { server, tools } = fakeServer();
    registerListActiveSessions(server, config);
    vi.mocked(fetchActiveSessions).mockRejectedValue(new RelayUnreachableError("relay down"));

    const result = await tools.get("list_active_sessions")!({});

    expect(result.isError).toBe(true);
    expect(result.content[0]?.text).toContain("relay down");
  });
});
