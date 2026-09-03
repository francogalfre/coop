import { execFileSync } from "node:child_process";
import { basename } from "node:path";

// Must stay byte-identical to packages/cli/internal/repoid.Detect: the relay
// keys its file-touch registry by whatever the CLI reported, so a different
// normalization here silently returns "no conflicts" for every path.
export function detectRepo(dir: string): string {
  const remote = git(dir, ["config", "--get", "remote.origin.url"]);
  if (remote) return normalizeRemote(remote);

  const toplevel = git(dir, ["rev-parse", "--show-toplevel"]);
  if (toplevel) return basename(toplevel);

  return basename(dir);
}

function git(dir: string, args: string[]): string {
  try {
    return execFileSync("git", ["-C", dir, ...args], { encoding: "utf8", stdio: ["ignore", "pipe", "ignore"] }).trim();
  } catch {
    return "";
  }
}

function normalizeRemote(url: string): string {
  let normalized = url;

  for (const prefix of ["git@", "https://", "http://"]) {
    if (normalized.startsWith(prefix)) {
      normalized = normalized.slice(prefix.length);
      break;
    }
  }

  const colon = normalized.indexOf(":");
  if (colon !== -1 && !normalized.slice(0, colon).includes("/")) {
    normalized = `${normalized.slice(0, colon)}/${normalized.slice(colon + 1)}`;
  }

  return normalized.endsWith(".git") ? normalized.slice(0, -".git".length) : normalized;
}
