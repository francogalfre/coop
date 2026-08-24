import { z } from "zod";
import { envelope } from "./envelope.js";
import { EVENT_SCHEMAS, type Event } from "./event.js";
import { unknownEvent } from "./events/unknown.js";
import type { Result } from "./shared/result.js";

export type ParseIssue = { path: string; message: string };
export type ParseError = {
  code: "invalid_envelope" | "invalid_event";
  message: string;
  type?: string;
  issues: ParseIssue[];
};

type EventSchema = { safeParse: (input: unknown) => z.ZodSafeParseResult<unknown> };

export function parseEvent(input: unknown): Result<Event, ParseError> {
  try {
    const env = envelope.safeParse(input);
    if (!env.success) {
      return {
        ok: false,
        error: {
          code: "invalid_envelope",
          message: "envelope failed validation",
          issues: env.error.issues.map((i) => ({ path: i.path.join("."), message: i.message })),
        },
      };
    }

    const schema = (EVENT_SCHEMAS as Record<string, EventSchema>)[env.data.type];
    if (schema) {
      const result = schema.safeParse(input);
      if (!result.success) {
        return {
          ok: false,
          error: {
            code: "invalid_event",
            type: env.data.type,
            message: `"${env.data.type}" failed validation`,
            issues: result.error.issues.map((i) => ({ path: i.path.join("."), message: i.message })),
          },
        };
      }
      return { ok: true, value: result.data as Event };
    }

    const fallback = unknownEvent.safeParse(input);
    if (!fallback.success) {
      return {
        ok: false,
        error: {
          code: "invalid_envelope",
          message: "envelope failed validation",
          issues: fallback.error.issues.map((i) => ({ path: i.path.join("."), message: i.message })),
        },
      };
    }
    return { ok: true, value: fallback.data as Event };
  } catch (err) {
    return {
      ok: false,
      error: { code: "invalid_envelope", message: err instanceof Error ? err.message : "unknown error", issues: [] },
    };
  }
}
