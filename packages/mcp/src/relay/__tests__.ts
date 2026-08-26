import { afterEach, describe, expect, it, vi } from "vitest";
import type { Config } from "../config/config.js";
import { fetchActiveSessions, fetchPresence } from "./client.js";
import { RelayUnreachableError } from "./types.js";

const baseConfig: Config = { relayUrl: "http://localhost:8787", repo: "/repo" };

function jsonResponse(body: unknown, ok = true, status = 200): Response {
  return {
    ok,
    status,
    json: async () => body,
    text: async () => JSON.stringify(body),
  } as Response;
}

describe("fetchPresence", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("builds the query string and maps snake_case to camelCase", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      jsonResponse({
        repo: "/repo",
        window_seconds: 900,
        paths: [
          {
            path: "src/foo.ts",
            signals: [
              { session_id: "sess-a", owner: "Alice", mode: "write", at: "2026-08-24T10:00:00Z", active: true },
            ],
          },
        ],
      }),
    );
    vi.stubGlobal("fetch", fetchMock);

    const result = await fetchPresence(baseConfig, ["src/foo.ts"], 60);

    expect(fetchMock).toHaveBeenCalledTimes(1);
    const url = fetchMock.mock.calls[0]![0] as URL;
    expect(url.pathname).toBe("/v1/presence");
    expect(url.searchParams.get("repo")).toBe("/repo");
    expect(url.searchParams.get("paths")).toBe("src/foo.ts");
    expect(url.searchParams.get("window_seconds")).toBe("60");

    expect(result).toEqual({
      repo: "/repo",
      windowSeconds: 900,
      paths: [
        {
          path: "src/foo.ts",
          signals: [
            {
              path: "src/foo.ts",
              sessionId: "sess-a",
              owner: "Alice",
              mode: "write",
              at: "2026-08-24T10:00:00Z",
              active: true,
            },
          ],
        },
      ],
    });
  });

  it("omits window_seconds from the query string when not given", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      jsonResponse({ repo: "/repo", window_seconds: 900, paths: [] }),
    );
    vi.stubGlobal("fetch", fetchMock);

    await fetchPresence(baseConfig, ["a.ts"]);

    const url = fetchMock.mock.calls[0]![0] as URL;
    expect(url.searchParams.has("window_seconds")).toBe(false);
  });

  it("sends the CLI credential as a bearer token when configured", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      jsonResponse({ repo: "/repo", window_seconds: 900, paths: [] }),
    );
    vi.stubGlobal("fetch", fetchMock);

    await fetchPresence({ ...baseConfig, cliCredential: "cli-token-abc" }, ["a.ts"]);

    const init = fetchMock.mock.calls[0]![1] as RequestInit;
    expect((init.headers as Record<string, string>).Authorization).toBe("Bearer cli-token-abc");
  });

  it("omits the Authorization header when no CLI credential is configured", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      jsonResponse({ repo: "/repo", window_seconds: 900, paths: [] }),
    );
    vi.stubGlobal("fetch", fetchMock);

    await fetchPresence(baseConfig, ["a.ts"]);

    const init = fetchMock.mock.calls[0]![1] as RequestInit | undefined;
    expect(init?.headers).toBeUndefined();
  });

  it("throws RelayUnreachableError on a non-2xx response", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse({ error: "boom" }, false, 500)));
    await expect(fetchPresence(baseConfig, ["a.ts"])).rejects.toThrow(RelayUnreachableError);
  });

  it("throws RelayUnreachableError when fetch throws", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("ECONNREFUSED")));
    await expect(fetchPresence(baseConfig, ["a.ts"])).rejects.toThrow(RelayUnreachableError);
  });
});

describe("fetchActiveSessions", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("builds the query string and maps snake_case to camelCase", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      jsonResponse({
        repo: "/repo",
        sessions: [{ session_id: "sess-a", owner: "Alice", started_at: "2026-08-24T09:50:00Z", active: true }],
      }),
    );
    vi.stubGlobal("fetch", fetchMock);

    const result = await fetchActiveSessions(baseConfig);

    const url = fetchMock.mock.calls[0]![0] as URL;
    expect(url.pathname).toBe("/v1/sessions");
    expect(url.searchParams.get("repo")).toBe("/repo");
    expect(result).toEqual([{ sessionId: "sess-a", owner: "Alice", startedAt: "2026-08-24T09:50:00Z", active: true }]);
  });

  it("throws RelayUnreachableError on a non-2xx response", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse({ error: "boom" }, false, 503)));
    await expect(fetchActiveSessions(baseConfig)).rejects.toThrow(RelayUnreachableError);
  });
});
