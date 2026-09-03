import type { Config } from "../config/config.js";
import { RelayUnreachableError } from "./types.js";

function headers(config: Config, body: boolean): Record<string, string> {
  const out: Record<string, string> = {};

  if (config.cliCredential) out.Authorization = `Bearer ${config.cliCredential}`;
  if (body) out["Content-Type"] = "application/json";

  return out;
}

async function send(url: URL, init: RequestInit): Promise<unknown> {
  let response: Response;

  try {
    response = await fetch(url, init);
  } catch (cause) {
    throw new RelayUnreachableError(
      `failed to reach relay at ${url}: ${cause instanceof Error ? cause.message : String(cause)}`,
    );
  }

  if (!response.ok) {
    const body = await response.text().catch(() => "");
    throw new RelayUnreachableError(`relay responded ${response.status} for ${url}: ${body}`);
  }

  return response.json();
}

export function getJson(url: URL, config: Config, signal?: AbortSignal): Promise<unknown> {
  return send(url, { headers: headers(config, false), signal });
}

export function postJson(url: URL, config: Config, body: unknown): Promise<unknown> {
  return send(url, { method: "POST", headers: headers(config, true), body: JSON.stringify(body) });
}
