export function clockTime(ts: string): string {
  const d = new Date(ts);
  if (Number.isNaN(d.getTime())) return "--:--";
  return d.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", second: "2-digit", hour12: false });
}

export function relativeTime(ts: string): string {
  const d = new Date(ts).getTime();
  if (Number.isNaN(d)) return "";

  const seconds = Math.round((Date.now() - d) / 1000);
  if (seconds < 45) return "just now";
  const minutes = Math.round(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.round(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  return `${Math.round(hours / 24)}d ago`;
}

export function shortPath(path: string, keep = 2): string {
  const parts = path.split("/").filter(Boolean);
  if (parts.length <= keep) return parts.join("/");
  return "…/" + parts.slice(-keep).join("/");
}

export function basename(path: string): string {
  const parts = path.split("/").filter(Boolean);
  return parts[parts.length - 1] ?? path;
}

export function initials(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  const first = parts[0];
  if (!first) return "?";

  const last = parts.length > 1 ? parts[parts.length - 1] : undefined;
  if (!last) return first.slice(0, 2).toUpperCase();

  return `${first[0] ?? ""}${last[0] ?? ""}`.toUpperCase();
}

const AVATAR_TINTS: readonly [string, ...string[]] = [
  "oklch(0.704 0.148 258.3)",
  "oklch(0.769 0.152 88.4)",
  "oklch(0.702 0.163 305.1)",
  "oklch(0.75 0.13 195)",
  "oklch(0.723 0.183 148.5)",
  "oklch(0.72 0.16 25)",
];

export function tintFor(name: string): string {
  let hash = 0;
  for (let i = 0; i < name.length; i++) hash = (hash * 31 + name.charCodeAt(i)) | 0;
  return AVATAR_TINTS[Math.abs(hash) % AVATAR_TINTS.length] ?? AVATAR_TINTS[0];
}

export function prettyJson(raw: string): string {
  try {
    return JSON.stringify(JSON.parse(raw), null, 2);
  } catch {
    return raw;
  }
}

export function summarizeToolInput(toolName: string, rawInput: string): string {
  let parsed: unknown;
  try {
    parsed = JSON.parse(rawInput);
  } catch {
    return rawInput.slice(0, 120);
  }

  if (typeof parsed !== "object" || parsed === null) return String(parsed).slice(0, 120);

  const fields = parsed as Record<string, unknown>;
  const preferred = ["command", "file_path", "path", "pattern", "query", "url", "description"];

  for (const key of preferred) {
    const value = fields[key];
    if (typeof value === "string" && value) {
      return key.endsWith("path") ? shortPath(value, 3) : value.slice(0, 140);
    }
  }

  const first = Object.values(fields).find((v) => typeof v === "string" && v);
  return typeof first === "string" ? first.slice(0, 140) : toolName;
}
