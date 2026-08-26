export type Config = { relayUrl: string; repo: string; sessionId?: string; cliCredential?: string };

export function loadConfig(env: NodeJS.ProcessEnv = process.env): Config {
  return {
    relayUrl: env.COOP_RELAY_URL ?? "http://localhost:8787",
    repo: env.COOP_REPO ?? process.cwd(),
    sessionId: env.COOP_SESSION_ID,
    cliCredential: env.COOP_CLI_CREDENTIAL,
  };
}
