import { describe, expect, it } from "vitest";
import { type ConfigSources, loadConfig } from "./config.js";

const sources: ConfigSources = {
  detectRepo: () => "github.com/acme/widgets",
  readStoredCredential: () => "stored-token",
};

const noSources: ConfigSources = {
  detectRepo: () => "github.com/acme/widgets",
  readStoredCredential: () => undefined,
};

describe("loadConfig", () => {
  it("uses defaults when env vars are unset", () => {
    const config = loadConfig({}, sources);
    expect(config.relayUrl).toBe("http://localhost:8787");
    expect(config.sessionId).toBeUndefined();
  });

  it("honors env overrides", () => {
    const config = loadConfig(
      {
        COOP_RELAY_URL: "http://relay.example:9000",
        COOP_REPO: "/some/repo",
        COOP_SESSION_ID: "sess-xyz",
        COOP_CLI_CREDENTIAL: "cli-token-abc",
      },
      sources,
    );
    expect(config).toEqual({
      relayUrl: "http://relay.example:9000",
      repo: "/some/repo",
      sessionId: "sess-xyz",
      cliCredential: "cli-token-abc",
    });
  });

  it("derives repo the same way the CLI does when COOP_REPO is unset", () => {
    expect(loadConfig({}, sources).repo).toBe("github.com/acme/widgets");
  });

  it("falls back to the credential `coop login` stored on disk", () => {
    expect(loadConfig({}, sources).cliCredential).toBe("stored-token");
  });

  it("leaves the credential undefined when there is neither an env var nor a stored one", () => {
    expect(loadConfig({}, noSources).cliCredential).toBeUndefined();
  });
});
