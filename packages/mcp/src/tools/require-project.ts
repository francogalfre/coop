import type { Config } from "../config/config.js";

export function requireProject(config: Config): { ok: true; project: string } | { ok: false; message: string } {
  if (!config.project) {
    return { ok: false, message: "no project configured — set COOP_PROJECT to the project slug" };
  }

  return { ok: true, project: config.project };
}
