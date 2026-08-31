import { describe, expect, it } from "vitest";
import { sessionStart, sessionEnd } from "./session.js";
import { agentTurnStart, agentText, agentTurnEnd } from "./agent.js";
import { toolCall, toolResult, toolBlocked } from "./tool.js";
import { fileTouched } from "./file.js";
import { permissionRequested, permissionResolved } from "./permission.js";
import { humanJoin, humanLeave, humanSteer, humanTakeover, humanPrompt, humanMessage } from "./human.js";
import { steerRequested, steerResolved } from "./steer.js";
import { sessionModeChanged } from "./session-mode.js";
import { unknownEvent } from "./unknown.js";

const base = { v: 1 as const, session_id: "s_1", seq: 0, ts: "2026-08-24T15:31:07.812Z" };
const anActor = { id: "u_1", display_name: "Franco" };

describe("sessionStart", () => {
  it("accepts a minimal valid fixture", () => {
    const result = sessionStart.safeParse({
      ...base,
      type: "session.start",
      harness: "claude-code",
      cwd: "/home/user/project",
      owner: { id: "u_1", display_name: "Alice" },
    });
    expect(result.success).toBe(true);
  });

  it("accepts a fixture with optional fields", () => {
    const result = sessionStart.safeParse({
      ...base,
      type: "session.start",
      harness: "codex",
      cwd: "/home/user/project",
      owner: { id: "u_1", display_name: "Alice" },
      harness_version: "1.2.3",
      permission_mode: "default",
    });
    expect(result.success).toBe(true);
  });

  it("rejects the wrong type literal", () => {
    const result = sessionStart.safeParse({
      ...base,
      type: "session.end",
      harness: "claude-code",
      cwd: "/home/user/project",
      owner: { id: "u_1", display_name: "Alice" },
    });
    expect(result.success).toBe(false);
  });

  it("defaults mode to auto when omitted", () => {
    const result = sessionStart.safeParse({
      ...base,
      type: "session.start",
      harness: "claude-code",
      cwd: "/home/user/project",
      owner: { id: "u_1", display_name: "Alice" },
    });
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.mode).toBe("auto");
    }
  });

  it("accepts an explicit restricted mode", () => {
    const result = sessionStart.safeParse({
      ...base,
      type: "session.start",
      harness: "claude-code",
      cwd: "/home/user/project",
      owner: { id: "u_1", display_name: "Alice" },
      mode: "restricted",
    });
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.mode).toBe("restricted");
    }
  });

  it("rejects an invalid mode", () => {
    const result = sessionStart.safeParse({
      ...base,
      type: "session.start",
      harness: "claude-code",
      cwd: "/home/user/project",
      owner: { id: "u_1", display_name: "Alice" },
      mode: "locked",
    });
    expect(result.success).toBe(false);
  });
});

describe("sessionEnd", () => {
  it("accepts a minimal valid fixture", () => {
    const result = sessionEnd.safeParse({ ...base, type: "session.end" });
    expect(result.success).toBe(true);
  });

  it("accepts a fixture with optional fields", () => {
    const result = sessionEnd.safeParse({ ...base, type: "session.end", reason: "completed" });
    expect(result.success).toBe(true);
  });

  it("rejects an invalid reason", () => {
    const result = sessionEnd.safeParse({ ...base, type: "session.end", reason: "timeout" });
    expect(result.success).toBe(false);
  });
});

describe("agentTurnStart", () => {
  it("accepts a minimal valid fixture", () => {
    const result = agentTurnStart.safeParse({ ...base, type: "agent.turn_start" });
    expect(result.success).toBe(true);
  });

  it("accepts a fixture with optional fields", () => {
    const result = agentTurnStart.safeParse({ ...base, type: "agent.turn_start", turn_id: "t_1" });
    expect(result.success).toBe(true);
  });

  it("rejects the wrong type literal", () => {
    const result = agentTurnStart.safeParse({ ...base, type: "agent.turn_end" });
    expect(result.success).toBe(false);
  });
});

describe("agentText", () => {
  it("accepts a minimal valid fixture", () => {
    const result = agentText.safeParse({
      ...base,
      type: "agent.text",
      text: { text: "hello", redactions: 0, truncated: false },
    });
    expect(result.success).toBe(true);
  });

  it("accepts a fixture with optional fields", () => {
    const result = agentText.safeParse({
      ...base,
      type: "agent.text",
      text: { text: "hello", redactions: 0, truncated: false },
      turn_id: "t_1",
    });
    expect(result.success).toBe(true);
  });

  it("rejects a missing text field", () => {
    const result = agentText.safeParse({ ...base, type: "agent.text" });
    expect(result.success).toBe(false);
  });
});

describe("agentTurnEnd", () => {
  it("accepts a minimal valid fixture", () => {
    const result = agentTurnEnd.safeParse({ ...base, type: "agent.turn_end" });
    expect(result.success).toBe(true);
  });

  it("accepts a fixture with optional fields", () => {
    const result = agentTurnEnd.safeParse({
      ...base,
      type: "agent.turn_end",
      turn_id: "t_1",
      usage: { input_tokens: 100, output_tokens: 50, cost_usd: 0.01 },
      duration_ms: 1500,
    });
    expect(result.success).toBe(true);
  });

  it("rejects a negative duration_ms", () => {
    const result = agentTurnEnd.safeParse({ ...base, type: "agent.turn_end", duration_ms: -1 });
    expect(result.success).toBe(false);
  });
});

describe("toolCall", () => {
  it("parses a minimal valid fixture", () => {
    const result = toolCall.safeParse({
      ...base,
      type: "tool.call",
      tool_name: "Bash",
      input: { text: "ls -la", redactions: 0, truncated: false },
    });
    expect(result.success).toBe(true);
  });

  it("parses a valid fixture exercising optional fields", () => {
    const result = toolCall.safeParse({
      ...base,
      type: "tool.call",
      tool_name: "Bash",
      input: { text: "ls -la", redactions: 0, truncated: false },
      tool_use_id: "tu_123",
    });
    expect(result.success).toBe(true);
  });

  it("rejects a fixture missing required fields", () => {
    const result = toolCall.safeParse({
      ...base,
      type: "tool.call",
      input: { text: "ls -la", redactions: 0, truncated: false },
    });
    expect(result.success).toBe(false);
  });
});

describe("toolResult", () => {
  it("parses a minimal valid fixture", () => {
    const result = toolResult.safeParse({
      ...base,
      type: "tool.result",
      tool_name: "Bash",
      output: { text: "done", redactions: 0, truncated: false },
    });
    expect(result.success).toBe(true);
  });

  it("parses a valid fixture exercising optional fields", () => {
    const result = toolResult.safeParse({
      ...base,
      type: "tool.result",
      tool_name: "Bash",
      output: { text: "done", redactions: 0, truncated: false },
      tool_use_id: "tu_123",
      ok: true,
    });
    expect(result.success).toBe(true);
  });

  it("rejects a fixture missing required fields", () => {
    const result = toolResult.safeParse({ ...base, type: "tool.result", tool_name: "Bash" });
    expect(result.success).toBe(false);
  });
});

describe("toolBlocked", () => {
  it("parses a minimal valid fixture", () => {
    const result = toolBlocked.safeParse({
      ...base,
      type: "tool.blocked",
      tool_name: "Bash",
      blocked_by: anActor,
      reason: { text: "policy violation", redactions: 0, truncated: false },
    });
    expect(result.success).toBe(true);
  });

  it("parses a valid fixture exercising optional fields", () => {
    const result = toolBlocked.safeParse({
      ...base,
      type: "tool.blocked",
      tool_name: "Bash",
      blocked_by: anActor,
      reason: { text: "policy violation", redactions: 0, truncated: false },
      tool_use_id: "tu_123",
    });
    expect(result.success).toBe(true);
  });

  it("rejects a fixture missing blocked_by", () => {
    const result = toolBlocked.safeParse({
      ...base,
      type: "tool.blocked",
      tool_name: "Bash",
      reason: { text: "policy violation", redactions: 0, truncated: false },
    });
    expect(result.success).toBe(false);
  });
});

describe("fileTouched", () => {
  it("parses a minimal valid fixture", () => {
    const result = fileTouched.safeParse({ ...base, type: "file.touched", path: "src/index.ts", mode: "read" });
    expect(result.success).toBe(true);
  });

  it("parses a valid fixture exercising optional fields", () => {
    const result = fileTouched.safeParse({
      ...base,
      type: "file.touched",
      path: "src/index.ts",
      mode: "write",
      tool_use_id: "tu_123",
    });
    expect(result.success).toBe(true);
  });

  it("rejects a fixture with an invalid mode", () => {
    const result = fileTouched.safeParse({ ...base, type: "file.touched", path: "src/index.ts", mode: "execute" });
    expect(result.success).toBe(false);
  });
});

describe("permissionRequested", () => {
  it("parses a minimal valid fixture", () => {
    const fixture = {
      ...base,
      type: "permission.requested",
      request_id: "req_1",
      tool_name: "Bash",
      input: { text: "ls -la", redactions: 0, truncated: false },
    };
    expect(permissionRequested.safeParse(fixture).success).toBe(true);
  });

  it("parses a valid fixture with optional permission_mode set", () => {
    const fixture = {
      ...base,
      type: "permission.requested",
      request_id: "req_1",
      tool_name: "Bash",
      input: { text: "ls -la", redactions: 0, truncated: false },
      permission_mode: "acceptEdits",
    };
    expect(permissionRequested.safeParse(fixture).success).toBe(true);
  });

  it("rejects a fixture missing required tool_name", () => {
    const fixture = {
      ...base,
      type: "permission.requested",
      request_id: "req_1",
      input: { text: "ls -la", redactions: 0, truncated: false },
    };
    expect(permissionRequested.safeParse(fixture).success).toBe(false);
  });
});

describe("permissionResolved", () => {
  it("parses a minimal valid fixture", () => {
    const fixture = {
      ...base,
      type: "permission.resolved",
      request_id: "req_1",
      decision: "allow",
      resolved_by: anActor,
    };
    expect(permissionResolved.safeParse(fixture).success).toBe(true);
  });

  it("parses a valid fixture with optional reason set", () => {
    const fixture = {
      ...base,
      type: "permission.resolved",
      request_id: "req_1",
      decision: "deny",
      resolved_by: anActor,
      reason: { text: "not safe", redactions: 0, truncated: false },
    };
    expect(permissionResolved.safeParse(fixture).success).toBe(true);
  });

  it("rejects a fixture omitting resolved_by", () => {
    const fixture = { ...base, type: "permission.resolved", request_id: "req_1", decision: "allow" };
    expect(permissionResolved.safeParse(fixture).success).toBe(false);
  });
});

describe("humanJoin", () => {
  it("parses a minimal valid fixture", () => {
    const fixture = { ...base, type: "human.join", actor: anActor };
    expect(humanJoin.safeParse(fixture).success).toBe(true);
  });

  it("rejects a fixture missing actor", () => {
    const fixture = { ...base, type: "human.join" };
    expect(humanJoin.safeParse(fixture).success).toBe(false);
  });
});

describe("humanLeave", () => {
  it("parses a minimal valid fixture", () => {
    const fixture = { ...base, type: "human.leave", actor: anActor };
    expect(humanLeave.safeParse(fixture).success).toBe(true);
  });

  it("rejects a fixture missing actor", () => {
    const fixture = { ...base, type: "human.leave" };
    expect(humanLeave.safeParse(fixture).success).toBe(false);
  });
});

describe("humanSteer", () => {
  it("parses a minimal valid fixture", () => {
    const fixture = {
      ...base,
      type: "human.steer",
      actor: anActor,
      text: { text: "try a different approach", redactions: 0, truncated: false },
    };
    expect(humanSteer.safeParse(fixture).success).toBe(true);
  });

  it("rejects a fixture omitting actor", () => {
    const fixture = {
      ...base,
      type: "human.steer",
      text: { text: "try a different approach", redactions: 0, truncated: false },
    };
    expect(humanSteer.safeParse(fixture).success).toBe(false);
  });
});

describe("humanMessage", () => {
  it("parses a minimal valid fixture", () => {
    const fixture = {
      ...base,
      type: "human.message",
      actor: anActor,
      text: { text: "worth a look before this lands", redactions: 0, truncated: false },
    };
    expect(humanMessage.safeParse(fixture).success).toBe(true);
  });

  it("parses a fixture anchored to a past event", () => {
    const fixture = {
      ...base,
      type: "human.message",
      actor: anActor,
      text: { text: "why this tool call?", redactions: 0, truncated: false },
      anchor_seq: 482,
    };
    const result = humanMessage.safeParse(fixture);
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.anchor_seq).toBe(482);
    }
  });

  it("rejects a fixture missing text", () => {
    const fixture = { ...base, type: "human.message", actor: anActor };
    expect(humanMessage.safeParse(fixture).success).toBe(false);
  });

  it("rejects a negative anchor_seq", () => {
    const fixture = {
      ...base,
      type: "human.message",
      actor: anActor,
      text: { text: "why this tool call?", redactions: 0, truncated: false },
      anchor_seq: -1,
    };
    expect(humanMessage.safeParse(fixture).success).toBe(false);
  });
});

describe("humanTakeover", () => {
  it("parses a minimal valid fixture claiming takeover", () => {
    const fixture = { ...base, type: "human.takeover", actor: anActor, active: true };
    expect(humanTakeover.safeParse(fixture).success).toBe(true);
  });

  it("parses a fixture releasing takeover", () => {
    const fixture = { ...base, type: "human.takeover", actor: anActor, active: false };
    expect(humanTakeover.safeParse(fixture).success).toBe(true);
  });

  it("rejects a fixture missing actor", () => {
    const fixture = { ...base, type: "human.takeover", active: true };
    expect(humanTakeover.safeParse(fixture).success).toBe(false);
  });

  it("rejects a fixture missing active", () => {
    const fixture = { ...base, type: "human.takeover", actor: anActor };
    expect(humanTakeover.safeParse(fixture).success).toBe(false);
  });
});

describe("humanPrompt", () => {
  it("parses a minimal valid fixture", () => {
    const fixture = {
      ...base,
      type: "human.prompt",
      text: { text: "fix the login bug", redactions: 0, truncated: false },
    };
    expect(humanPrompt.safeParse(fixture).success).toBe(true);
  });

  it("rejects a fixture missing text", () => {
    const fixture = { ...base, type: "human.prompt" };
    expect(humanPrompt.safeParse(fixture).success).toBe(false);
  });

  it("strips an unrecognized extra field", () => {
    const fixture = {
      ...base,
      type: "human.prompt",
      text: { text: "fix the login bug", redactions: 0, truncated: false },
      surprise_field: 123,
    };
    const result = humanPrompt.safeParse(fixture);
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data).not.toHaveProperty("surprise_field");
    }
  });
});

describe("steerRequested", () => {
  it("parses a minimal valid fixture", () => {
    const fixture = {
      ...base,
      type: "steer.requested",
      request_id: "req_1",
      actor: anActor,
      text: { text: "try a different approach", redactions: 0, truncated: false },
    };
    expect(steerRequested.safeParse(fixture).success).toBe(true);
  });

  it("rejects a fixture missing actor", () => {
    const fixture = {
      ...base,
      type: "steer.requested",
      request_id: "req_1",
      text: { text: "try a different approach", redactions: 0, truncated: false },
    };
    expect(steerRequested.safeParse(fixture).success).toBe(false);
  });
});

describe("steerResolved", () => {
  it("parses a minimal valid fixture", () => {
    const fixture = {
      ...base,
      type: "steer.resolved",
      request_id: "req_1",
      decision: "allow",
      resolved_by: anActor,
    };
    expect(steerResolved.safeParse(fixture).success).toBe(true);
  });

  it("rejects a fixture omitting resolved_by", () => {
    const fixture = { ...base, type: "steer.resolved", request_id: "req_1", decision: "deny" };
    expect(steerResolved.safeParse(fixture).success).toBe(false);
  });
});

describe("sessionModeChanged", () => {
  it("parses a minimal valid fixture", () => {
    const fixture = { ...base, type: "session.mode_changed", mode: "restricted", changed_by: anActor };
    expect(sessionModeChanged.safeParse(fixture).success).toBe(true);
  });

  it("rejects an invalid mode", () => {
    const fixture = { ...base, type: "session.mode_changed", mode: "locked", changed_by: anActor };
    expect(sessionModeChanged.safeParse(fixture).success).toBe(false);
  });

  it("rejects a fixture missing changed_by", () => {
    const fixture = { ...base, type: "session.mode_changed", mode: "restricted" };
    expect(sessionModeChanged.safeParse(fixture).success).toBe(false);
  });
});

describe("unknownEvent", () => {
  it("parses a minimal valid fixture", () => {
    const fixture = { ...base, type: "future.event" };
    expect(unknownEvent.safeParse(fixture).success).toBe(true);
  });

  it("parses (and keeps) an arbitrary extra key, proving the schema is loose", () => {
    const fixture = { ...base, type: "future.event", surprise_field: 123 };
    const result = unknownEvent.safeParse(fixture);
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.surprise_field).toBe(123);
    }
  });

  it("rejects a fixture missing required type", () => {
    const fixture = { ...base };
    expect(unknownEvent.safeParse(fixture).success).toBe(false);
  });
});
