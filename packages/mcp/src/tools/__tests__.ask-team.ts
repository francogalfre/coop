import type { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { Config } from "../config/config.js";

vi.mock("../relay/questions.js", () => ({ askTeam: vi.fn(), awaitAnswer: vi.fn() }));

import { askTeam, awaitAnswer } from "../relay/questions.js";
import { RelayUnreachableError } from "../relay/types.js";
import { registerAskTeam } from "./ask-team.js";

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

const config: Config = { relayUrl: "http://localhost:8787", repo: "/repo", sessionId: "sess-1" };

function invoke(cfg: Config = config) {
  const { server, tools } = fakeServer();
  registerAskTeam(server, cfg);
  return tools.get("ask_team")!;
}

describe("ask_team", () => {
  beforeEach(() => {
    vi.mocked(askTeam).mockReset();
    vi.mocked(awaitAnswer).mockReset();
  });

  it("returns the answer attributed to whoever gave it", async () => {
    vi.mocked(askTeam).mockResolvedValue("q_1");
    vi.mocked(awaitAnswer).mockResolvedValue({ text: "use the existing table", author: "Mara" });

    const result = await invoke()({ question: "new table or reuse?" }, {});

    expect(result.content[0]!.text).toBe("Mara: use the existing table");
    expect(result.structuredContent).toEqual({
      question_id: "q_1",
      answered: true,
      author: "Mara",
      answer: "use the existing table",
    });
  });

  it("forwards options to the relay", async () => {
    vi.mocked(askTeam).mockResolvedValue("q_1");
    vi.mocked(awaitAnswer).mockResolvedValue({ text: "reuse", author: "Mara" });

    await invoke()({ question: "which?", options: ["reuse", "new"] }, {});

    expect(askTeam).toHaveBeenCalledWith(config, "sess-1", "which?", ["reuse", "new"]);
  });

  it("tells the agent to decide for itself when nobody answers", async () => {
    vi.mocked(askTeam).mockResolvedValue("q_1");
    vi.mocked(awaitAnswer).mockResolvedValue(null);

    const result = await invoke()({ question: "anyone?", timeout_seconds: 60 }, {});

    expect(result.isError).toBeUndefined();
    expect(result.content[0]!.text).toContain("Nobody answered within 60s");
    expect(result.structuredContent).toEqual({ question_id: "q_1", answered: false });
  });

  it("is unavailable outside a coop session rather than hanging", async () => {
    const result = await invoke({ ...config, sessionId: undefined })({ question: "hi" }, {});

    expect(result.isError).toBe(true);
    expect(result.content[0]!.text).toContain("no coop session");
    expect(askTeam).not.toHaveBeenCalled();
  });

  it("soft-fails when the relay is unreachable", async () => {
    vi.mocked(askTeam).mockRejectedValue(new RelayUnreachableError("connection refused"));

    const result = await invoke()({ question: "hi" }, {});

    expect(result.isError).toBe(true);
    expect(result.content[0]!.text).toContain("connection refused");
  });
});
