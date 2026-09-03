import { describe, expect, it } from "vitest";
import { agentAskedTeam } from "./agent.js";
import { fileTouched } from "./file.js";
import { humanAnswered, humanCommand, humanMessage, humanSteer } from "./human.js";
import { sessionStart } from "./session.js";
import { steerDelivered, steerDropped } from "./steer.js";
import { parseEvent } from "../parse.js";
import { LIMITS } from "../shared/limits.js";

const base = { v: 1 as const, session_id: "s_1", seq: 0, ts: "2026-08-24T15:31:07.812Z" };
const anActor = { id: "u_1", display_name: "Franco" };
const someText = { text: "hello", redactions: 0, truncated: false };

const aHunk = {
  old_start: 1,
  old_lines: 2,
  new_start: 1,
  new_lines: 3,
  lines: [
    { text: "-const a = 1", redactions: 0, truncated: false },
    { text: "+const a = 2", redactions: 0, truncated: false },
  ],
};

describe("sessionStart capabilities", () => {
  it("accepts a session declaring its capabilities", () => {
    const result = sessionStart.safeParse({
      ...base,
      type: "session.start",
      harness: "claude-code",
      cwd: "/home/user/project",
      owner: anActor,
      capabilities: { steer: true, block: true, commands: false, pty: false },
    });
    expect(result.success).toBe(true);
  });

  it("stays valid when capabilities are absent, so older CLIs keep working", () => {
    const result = sessionStart.safeParse({
      ...base,
      type: "session.start",
      harness: "opencode",
      cwd: "/home/user/project",
      owner: anActor,
    });
    expect(result.success).toBe(true);
  });

  it("rejects a partial capabilities object", () => {
    const result = sessionStart.safeParse({
      ...base,
      type: "session.start",
      harness: "pi",
      cwd: "/home/user/project",
      owner: anActor,
      capabilities: { steer: true },
    });
    expect(result.success).toBe(false);
  });
});

describe("fileTouched diffs", () => {
  it("accepts a write carrying hunks and line counts", () => {
    const result = fileTouched.safeParse({
      ...base,
      type: "file.touched",
      path: "apps/web/lib/format.ts",
      mode: "write",
      added: 3,
      removed: 2,
      hunks: [aHunk],
    });
    expect(result.success).toBe(true);
  });

  it("still accepts a touch with no diff, for metadata-only projects", () => {
    const result = fileTouched.safeParse({
      ...base,
      type: "file.touched",
      path: "apps/web/lib/format.ts",
      mode: "read",
    });
    expect(result.success).toBe(true);
  });

  it("rejects more hunks than the limit", () => {
    const result = fileTouched.safeParse({
      ...base,
      type: "file.touched",
      path: "a.ts",
      mode: "write",
      hunks: Array.from({ length: LIMITS.diff_hunks + 1 }, () => aHunk),
    });
    expect(result.success).toBe(false);
  });

  it("rejects a hunk whose lines are bare strings rather than redacted text", () => {
    const result = fileTouched.safeParse({
      ...base,
      type: "file.touched",
      path: "a.ts",
      mode: "write",
      hunks: [{ ...aHunk, lines: ["-const a = 1"] }],
    });
    expect(result.success).toBe(false);
  });
});

describe("agentAskedTeam", () => {
  it("accepts an open question", () => {
    const result = agentAskedTeam.safeParse({
      ...base,
      type: "agent.asked_team",
      question_id: "q_1",
      text: someText,
    });
    expect(result.success).toBe(true);
  });

  it("accepts a question with options", () => {
    const result = agentAskedTeam.safeParse({
      ...base,
      type: "agent.asked_team",
      question_id: "q_1",
      text: someText,
      options: ["reuse the session table", "add a new one"],
    });
    expect(result.success).toBe(true);
  });

  it("rejects more options than a person can reasonably pick from", () => {
    const result = agentAskedTeam.safeParse({
      ...base,
      type: "agent.asked_team",
      question_id: "q_1",
      text: someText,
      options: Array.from({ length: LIMITS.question_options + 1 }, (_, i) => `option ${i}`),
    });
    expect(result.success).toBe(false);
  });

  it("rejects a missing question_id", () => {
    const result = agentAskedTeam.safeParse({ ...base, type: "agent.asked_team", text: someText });
    expect(result.success).toBe(false);
  });
});

describe("humanAnswered", () => {
  it("accepts an attributed answer", () => {
    const result = humanAnswered.safeParse({
      ...base,
      type: "human.answered",
      question_id: "q_1",
      actor: anActor,
      text: someText,
    });
    expect(result.success).toBe(true);
  });

  it("rejects an unattributed answer", () => {
    const result = humanAnswered.safeParse({
      ...base,
      type: "human.answered",
      question_id: "q_1",
      text: someText,
    });
    expect(result.success).toBe(false);
  });
});

describe("humanCommand", () => {
  it("accepts an allowlisted command with args", () => {
    const result = humanCommand.safeParse({
      ...base,
      type: "human.command",
      actor: anActor,
      command: "model",
      args: "sonnet",
    });
    expect(result.success).toBe(true);
  });

  it("rejects a command that is not on the allowlist", () => {
    const result = humanCommand.safeParse({
      ...base,
      type: "human.command",
      actor: anActor,
      command: "bash",
    });
    expect(result.success).toBe(false);
  });

  it("rejects a shell fragment smuggled in as a command", () => {
    const result = humanCommand.safeParse({
      ...base,
      type: "human.command",
      actor: anActor,
      command: "model; rm -rf /",
    });
    expect(result.success).toBe(false);
  });
});

describe("steer receipts", () => {
  it("accepts a delivery naming the hook that carried it", () => {
    const result = steerDelivered.safeParse({
      ...base,
      type: "steer.delivered",
      steer_id: "st_1",
      hook_event: "PreToolUse",
    });
    expect(result.success).toBe(true);
  });

  it("accepts a drop with a known reason", () => {
    const result = steerDropped.safeParse({
      ...base,
      type: "steer.dropped",
      steer_id: "st_1",
      reason: "queue_overflow",
    });
    expect(result.success).toBe(true);
  });

  it("rejects a drop with an invented reason", () => {
    const result = steerDropped.safeParse({
      ...base,
      type: "steer.dropped",
      steer_id: "st_1",
      reason: "because",
    });
    expect(result.success).toBe(false);
  });
});

describe("project context delivery", () => {
  it("accepts a steer carrying project_context_version", () => {
    const result = humanSteer.safeParse({
      ...base,
      type: "human.steer",
      actor: anActor,
      text: someText,
      project_context_version: 7,
    });
    expect(result.success).toBe(true);
  });

  it("stays valid without project_context_version, for an ordinary steer", () => {
    const result = humanSteer.safeParse({
      ...base,
      type: "human.steer",
      actor: anActor,
      text: someText,
    });
    expect(result.success).toBe(true);
  });

  it("rejects a negative version", () => {
    const result = humanSteer.safeParse({
      ...base,
      type: "human.steer",
      actor: anActor,
      text: someText,
      project_context_version: -1,
    });
    expect(result.success).toBe(false);
  });
});

describe("client-side correlation ids", () => {
  it("accepts a steer carrying steer_id and client_id", () => {
    const result = humanSteer.safeParse({
      ...base,
      type: "human.steer",
      actor: anActor,
      text: someText,
      steer_id: "st_1",
      client_id: "c_1",
    });
    expect(result.success).toBe(true);
  });

  it("accepts a team message carrying client_id", () => {
    const result = humanMessage.safeParse({
      ...base,
      type: "human.message",
      actor: anActor,
      text: someText,
      client_id: "c_1",
    });
    expect(result.success).toBe(true);
  });
});

describe("parseEvent over the new types", () => {
  const fixtures = [
    { ...base, type: "agent.asked_team", question_id: "q_1", text: someText },
    { ...base, type: "human.answered", question_id: "q_1", actor: anActor, text: someText },
    { ...base, type: "human.command", actor: anActor, command: "compact" },
    { ...base, type: "steer.delivered", steer_id: "st_1" },
    { ...base, type: "steer.dropped", steer_id: "st_1", reason: "session_ended" },
  ];

  for (const fixture of fixtures) {
    it(`routes ${fixture.type} to its schema`, () => {
      const result = parseEvent(fixture);
      expect(result.ok).toBe(true);
      if (result.ok) expect(result.value.type).toBe(fixture.type);
    });
  }

  it("reports a known type with a bad payload as invalid_event rather than degrading it", () => {
    const result = parseEvent({ ...base, type: "human.command", actor: anActor, command: "sudo" });
    expect(result.ok).toBe(false);
    if (!result.ok) expect(result.error.code).toBe("invalid_event");
  });

  it("still degrades an unregistered type to unknown", () => {
    const result = parseEvent({ ...base, type: "agent.invented_tomorrow", extra: 1 });
    expect(result.ok).toBe(true);
  });
});
