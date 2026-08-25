import { describe, expect, it } from "vitest";
import { envelope } from "./envelope.js";
import { parseEvent } from "./parse.js";
import { eventsJsonSchema } from "./json-schema.js";
import { KNOWN_EVENT_TYPES } from "./event.js";

const validEnvelope = {
  v: 1,
  session_id: "sess_123",
  seq: 0,
  ts: "2026-08-24T12:00:00.000Z",
  type: "tool.blocked",
};

describe("envelope", () => {
  it("accepts a valid fixture", () => {
    const result = envelope.safeParse(validEnvelope);
    expect(result.success).toBe(true);
  });

  it("rejects the wrong protocol version", () => {
    const result = envelope.safeParse({ ...validEnvelope, v: 2 });
    expect(result.success).toBe(false);
  });

  it("rejects a negative seq", () => {
    const result = envelope.safeParse({ ...validEnvelope, seq: -1 });
    expect(result.success).toBe(false);
  });

  it("rejects a non-ISO ts", () => {
    const result = envelope.safeParse({ ...validEnvelope, ts: "not-a-date" });
    expect(result.success).toBe(false);
  });

  it("rejects a missing session_id", () => {
    const { session_id: _session_id, ...withoutSessionId } = validEnvelope;
    const result = envelope.safeParse(withoutSessionId);
    expect(result.success).toBe(false);
  });
});

const validSessionStart = {
  v: 1,
  session_id: "sess-1",
  seq: 0,
  ts: new Date().toISOString(),
  type: "session.start",
  harness: "claude-code",
  cwd: "/repo",
  owner: { id: "u1", display_name: "Franco" },
};

describe("parseEvent", () => {
  it("parses a valid session.start event", () => {
    const result = parseEvent(validSessionStart);
    expect(result.ok).toBe(true);
  });

  it("rejects a garbage envelope", () => {
    const result = parseEvent({ v: 2, seq: -1, type: "session.start" });
    expect(result.ok).toBe(false);
    if (!result.ok) {
      expect(result.error.code).toBe("invalid_envelope");
    }
  });

  it("rejects a known type with a missing required field, not via the unknown fallback", () => {
    const result = parseEvent({
      v: 1,
      session_id: "sess-1",
      seq: 0,
      ts: new Date().toISOString(),
      type: "tool.call",
      input: { text: "ls", redactions: 0, truncated: false },
    });
    expect(result.ok).toBe(false);
    if (!result.ok) {
      expect(result.error.code).toBe("invalid_event");
      expect(result.error.type).toBe("tool.call");
    }
  });

  it("parses an unrecognized type via the unknown variant", () => {
    const result = parseEvent({
      v: 1,
      session_id: "sess-1",
      seq: 0,
      ts: new Date().toISOString(),
      type: "future.event",
      extra: "field",
    });
    expect(result.ok).toBe(true);
  });

  it.each([null, undefined, "a string", [], {}])("never throws on %p", (input) => {
    expect(() => parseEvent(input)).not.toThrow();
    expect(parseEvent(input).ok).toBe(false);
  });

  it("parses a valid human.prompt event", () => {
    const result = parseEvent({
      v: 1,
      session_id: "sess-1",
      seq: 0,
      ts: new Date().toISOString(),
      type: "human.prompt",
      text: { text: "fix the login bug", redactions: 0, truncated: false },
    });
    expect(result.ok).toBe(true);
  });

  it("rejects a human.prompt event missing text", () => {
    const result = parseEvent({
      v: 1,
      session_id: "sess-1",
      seq: 0,
      ts: new Date().toISOString(),
      type: "human.prompt",
    });
    expect(result.ok).toBe(false);
    if (!result.ok) {
      expect(result.error.code).toBe("invalid_event");
      expect(result.error.type).toBe("human.prompt");
    }
  });
});

describe("eventsJsonSchema", () => {
  it("has one union entry per known event type", () => {
    const schema = eventsJsonSchema() as { oneOf?: unknown[]; anyOf?: unknown[] };
    const union = schema.oneOf ?? schema.anyOf;
    expect(union).toBeDefined();
    expect(union).toHaveLength(KNOWN_EVENT_TYPES.length);
    expect(KNOWN_EVENT_TYPES).toHaveLength(16);
  });

  it("stringifies without throwing", () => {
    expect(() => JSON.stringify(eventsJsonSchema())).not.toThrow();
  });
});
