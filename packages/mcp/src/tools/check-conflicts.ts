import type { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import type { Config } from "../config/config.js";
import { fetchPresence } from "../relay/client.js";
import type { PresenceResponse } from "../relay/types.js";
import { RelayUnreachableError } from "../relay/types.js";
import { checkConflictsInputSchema } from "./schemas/check-conflicts.js";

type Conflict = {
  path: string;
  session_id: string;
  owner?: string;
  mode: string;
  seconds_ago: number;
  active: boolean;
};

export function registerCheckConflicts(server: McpServer, config: Config): void {
  server.registerTool(
    "check_conflicts",
    {
      title: "Check for conflicting edits",
      description: "Checks whether other active sessions in this repo are currently writing to the given paths.",
      inputSchema: checkConflictsInputSchema,
    },
    async ({ paths, window_seconds }) => {
      let presence: PresenceResponse;

      try {
        presence = await fetchPresence(config, paths, window_seconds);
      } catch (error) {
        if (error instanceof RelayUnreachableError) {
          return {
            isError: true,
            content: [{ type: "text" as const, text: `check_conflicts failed: ${error.message}` }],
          };
        }

        throw error;
      }

      const conflicts: Conflict[] = [];
      const clear: string[] = [];
      const now = Date.now();

      for (const entry of presence.paths) {
        const pathConflicts = entry.signals.filter(
          (signal) => signal.sessionId !== config.sessionId && signal.mode === "write" && signal.active,
        );

        if (pathConflicts.length === 0) {
          clear.push(entry.path);
          continue;
        }

        for (const signal of pathConflicts) {
          conflicts.push({
            path: entry.path,
            session_id: signal.sessionId,
            owner: signal.owner,
            mode: signal.mode,
            seconds_ago: Math.max(0, Math.round((now - Date.parse(signal.at)) / 1000)),
            active: signal.active,
          });
        }
      }

      const summary = conflicts.length === 0 ? "All clear." : `${conflicts.length} conflict(s) found.`;

      return {
        content: [{ type: "text" as const, text: summary }],
        structuredContent: {
          repo: presence.repo,
          window_seconds: presence.windowSeconds,
          conflicts,
          clear,
        },
      };
    },
  );
}
