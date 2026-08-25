#!/usr/bin/env node

import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { loadConfig } from "./config/config.js";
import { createServer } from "./server.js";

try {
  const config = loadConfig();
  const server = createServer(config);
  await server.connect(new StdioServerTransport());
} catch (error) {
  process.stderr.write(`coop-mcp: fatal error: ${error instanceof Error ? error.message : String(error)}\n`);
  process.exit(1);
}
