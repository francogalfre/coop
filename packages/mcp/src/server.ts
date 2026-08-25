import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import type { Config } from "./config/config.js";
import { registerCheckConflicts } from "./tools/check-conflicts.js";
import { registerListActiveSessions } from "./tools/list-active-sessions.js";

export function createServer(config: Config): McpServer {
  const server = new McpServer({ name: "coop-mcp", version: "0.0.0" });

  registerCheckConflicts(server, config);
  registerListActiveSessions(server, config);

  return server;
}
