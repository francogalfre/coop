import { readFileSync } from "node:fs";
import { homedir } from "node:os";
import { join } from "node:path";

// Written by `coop login` (packages/cli/internal/config/credentials.go). The
// MCP server is spawned by the agent, not by a shell, so it rarely inherits
// COOP_CLI_CREDENTIAL and would otherwise call the relay unauthenticated.
export function readStoredCredential(): string | undefined {
  try {
    const raw = readFileSync(join(homedir(), ".config", "coop", "credentials.json"), "utf8");
    const parsed: unknown = JSON.parse(raw);

    if (typeof parsed !== "object" || parsed === null) return undefined;

    const token = (parsed as { token?: unknown }).token;

    return typeof token === "string" && token.length > 0 ? token : undefined;
  } catch {
    return undefined;
  }
}
