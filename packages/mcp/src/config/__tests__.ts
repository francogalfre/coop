import { describe, expect, it } from "vitest";
import { loadConfig } from "./config.js";

describe("loadConfig", () => {
  it("uses defaults when env vars are unset", () => {
    const config = loadConfig({});
    expect(config.relayUrl).toBe("http://localhost:8787");
    expect(config.repo).toBe(process.cwd());
    expect(config.sessionId).toBeUndefined();
  });

  it("honors env overrides", () => {
    const config = loadConfig({
      COOP_RELAY_URL: "http://relay.example:9000",
      COOP_REPO: "/some/repo",
      COOP_SESSION_ID: "sess-xyz",
    });
    expect(config).toEqual({
      relayUrl: "http://relay.example:9000",
      repo: "/some/repo",
      sessionId: "sess-xyz",
    });
  });
});
