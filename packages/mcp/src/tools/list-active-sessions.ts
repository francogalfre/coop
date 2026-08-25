import type { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import type { Config } from "../config/config.js";
import { fetchActiveSessions } from "../relay/client.js";
import type { SessionSummary } from "../relay/types.js";
import { RelayUnreachableError } from "../relay/types.js";

export function registerListActiveSessions(server: McpServer, config: Config): void {
  server.registerTool(
    "list_active_sessions",
    {
      title: "List active sessions",
      description: "Lists sessions currently active in this repo, per the relay's presence registry.",
    },
    async () => {
      let sessions: SessionSummary[];

      try {
        sessions = await fetchActiveSessions(config);
      } catch (error) {
        if (error instanceof RelayUnreachableError) {
          return {
            isError: true,
            content: [{ type: "text" as const, text: `list_active_sessions failed: ${error.message}` }],
          };
        }

        throw error;
      }

      return {
        content: [{ type: "text" as const, text: `${sessions.length} active session(s).` }],
        structuredContent: { repo: config.repo, sessions },
      };
    },
  );
}
