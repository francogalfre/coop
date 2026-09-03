import { readStoredCredential } from "./credentials.js";
import { detectRepo } from "../repo/identity.js";

export type Config = { relayUrl: string; repo: string; sessionId?: string; cliCredential?: string };

export type ConfigSources = {
  detectRepo: (dir: string) => string;
  readStoredCredential: () => string | undefined;
};

const defaultSources: ConfigSources = { detectRepo, readStoredCredential };

export function loadConfig(env: NodeJS.ProcessEnv = process.env, sources: ConfigSources = defaultSources): Config {
  return {
    relayUrl: env.COOP_RELAY_URL ?? "http://localhost:8787",
    repo: env.COOP_REPO ?? sources.detectRepo(process.cwd()),
    sessionId: env.COOP_SESSION_ID,
    cliCredential: env.COOP_CLI_CREDENTIAL ?? sources.readStoredCredential(),
  };
}
