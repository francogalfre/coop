const DEFAULT_HTTP_URL = "http://localhost:8787";

function toWsUrl(httpUrl: string): string {
  return httpUrl.replace(/^http/, "ws");
}

const httpUrl = process.env.NEXT_PUBLIC_COOP_RELAY_URL ?? DEFAULT_HTTP_URL;

export const relayConfig = {
  httpUrl,
  wsUrl: toWsUrl(httpUrl),
  repo: process.env.NEXT_PUBLIC_COOP_REPO ?? "",
};
