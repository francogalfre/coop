import { describe, expect, it } from "vitest";
import { actor } from "./actor.js";
import { LIMITS } from "./limits.js";
import { redactedText } from "./redacted-text.js";

describe("actor", () => {
  it("accepts a valid fixture", () => {
    const result = actor.safeParse({ id: "u_123", display_name: "Alex" });
    expect(result.success).toBe(true);
  });

  it("rejects an empty id", () => {
    const result = actor.safeParse({ id: "", display_name: "Alex" });
    expect(result.success).toBe(false);
  });
});

describe("redactedText", () => {
  it("accepts a valid fixture", () => {
    const result = redactedText.safeParse({
      text: "hello [redacted]",
      redactions: 1,
      truncated: false,
    });
    expect(result.success).toBe(true);
  });

  it("rejects negative redactions", () => {
    const result = redactedText.safeParse({
      text: "hello",
      redactions: -1,
      truncated: false,
    });
    expect(result.success).toBe(false);
  });

  it("accepts text exactly at the limit", () => {
    const result = redactedText.safeParse({
      text: "a".repeat(LIMITS.text),
      redactions: 0,
      truncated: false,
    });
    expect(result.success).toBe(true);
  });

  it("rejects text one character over the limit", () => {
    const result = redactedText.safeParse({
      text: "a".repeat(LIMITS.text + 1),
      redactions: 0,
      truncated: true,
    });
    expect(result.success).toBe(false);
  });
});
