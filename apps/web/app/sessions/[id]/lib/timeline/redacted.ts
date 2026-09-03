export type RedactionCounts = { text: string; redactions: number; truncated: boolean };

export function readRedacted(value: unknown): RedactionCounts {
  if (typeof value === "string") return { text: value, redactions: 0, truncated: false };
  if (value && typeof value === "object" && "text" in value) {
    const record = value as { text: unknown; redactions?: unknown; truncated?: unknown };
    if (typeof record.text === "string") {
      return {
        text: record.text,
        redactions: typeof record.redactions === "number" ? record.redactions : 0,
        truncated: record.truncated === true,
      };
    }
  }
  return { text: "", redactions: 0, truncated: false };
}

export function textOf(value: unknown): string {
  return readRedacted(value).text;
}
