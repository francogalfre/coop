import type { Event } from "@coop/protocol";
import { describe, expect, it } from "vitest";
import { buildTimeline } from "../build-timeline";
import { reduceTimeline } from "./reduce";
import { createTimelineState } from "./state";

const base = { v: 1 as const, session_id: "s_1" };
const owner = { id: "u_owner", display_name: "Franco" };
const teammate = { id: "u_team", display_name: "Ada" };
const redacted = (text: string, redactions = 0, truncated = false) => ({ text, redactions, truncated });

const events: Event[] = [
  { ...base, type: "session.start", seq: 0, ts: "2026-08-24T15:00:00.000Z", harness: "claude-code", cwd: "/repo", owner, mode: "auto" },
  { ...base, type: "agent.turn_start", seq: 1, ts: "2026-08-24T15:00:01.000Z", turn_id: "t1" },
  { ...base, type: "tool.call", seq: 2, ts: "2026-08-24T15:00:02.000Z", tool_name: "Bash", input: redacted('{"command":"ls"}'), tool_use_id: "tu_1" },
  { ...base, type: "tool.result", seq: 3, ts: "2026-08-24T15:00:03.000Z", tool_name: "Bash", output: redacted("a.txt\nb.txt", 1, false), tool_use_id: "tu_1", ok: true },
  { ...base, type: "file.touched", seq: 4, ts: "2026-08-24T15:00:04.000Z", path: "a.txt", mode: "read", tool_use_id: "tu_1" },
  { ...base, type: "agent.text", seq: 5, ts: "2026-08-24T15:00:05.000Z", text: redacted("Listed the files.") },
  { ...base, type: "tool.blocked", seq: 6, ts: "2026-08-24T15:00:06.000Z", tool_name: "Write", blocked_by: owner, reason: redacted("restricted mode") },
  { ...base, type: "permission.requested", seq: 7, ts: "2026-08-24T15:00:07.000Z", request_id: "p1", tool_name: "Write", input: redacted('{"file_path":"a.txt"}') },
  { ...base, type: "permission.resolved", seq: 8, ts: "2026-08-24T15:00:08.000Z", request_id: "p1", decision: "allow", resolved_by: owner },
  { ...base, type: "agent.asked_team", seq: 9, ts: "2026-08-24T15:00:09.000Z", question_id: "q1", text: redacted("Which branch?"), options: ["main", "dev"] },
  { ...base, type: "human.answered", seq: 10, ts: "2026-08-24T15:00:10.000Z", question_id: "q1", actor: teammate, text: redacted("main") },
  { ...base, type: "human.command", seq: 11, ts: "2026-08-24T15:00:11.000Z", actor: owner, command: "model", args: "sonnet" },
  { ...base, type: "human.steer", seq: 12, ts: "2026-08-24T15:00:12.000Z", actor: teammate, text: redacted("focus on tests"), steer_id: "st1" },
  { ...base, type: "steer.delivered", seq: 13, ts: "2026-08-24T15:00:13.000Z", steer_id: "st1" },
  { ...base, type: "agent.turn_end", seq: 14, ts: "2026-08-24T15:00:14.000Z", turn_id: "t1", duration_ms: 2500, usage: { input_tokens: 100, output_tokens: 50, cost_usd: 0.01 } },
  { ...base, type: "agent.turn_start", seq: 15, ts: "2026-08-24T15:00:15.000Z", turn_id: "t2" },
  { ...base, type: "session.end", seq: 16, ts: "2026-08-24T15:00:16.000Z" },
];

describe("buildTimeline / reduceTimeline", () => {
  it("folds every known event into an item without throwing", () => {
    const timeline = buildTimeline(events);
    expect(timeline.items.length).toBeGreaterThan(0);
    expect(timeline.meta.harness).toBe("claude-code");
  });

  it("marks the delivered steer message as seen once the next turn starts", () => {
    const timeline = buildTimeline(events);
    const message = timeline.items.find((item) => item.kind === "message" && item.steerId === "st1");
    expect(message?.kind === "message" && message.delivery).toBe("seen");
  });

  it("produces the same result whether folded all at once or in chunks", () => {
    const full = buildTimeline(events);

    let incremental = createTimelineState();
    for (const event of events) incremental = reduceTimeline(incremental, event);

    expect(incremental.items).toEqual(full.items);
    expect(incremental.meta).toEqual(full.meta);
    expect(incremental.openTools > 0 || !incremental.sawTurnEnd).toEqual(full.agentBusy);
  });

  it("carries state correctly across separately-arriving batches", () => {
    const full = buildTimeline(events);

    const mid = Math.floor(events.length / 2);
    let batched = createTimelineState();
    for (const event of events.slice(0, mid)) batched = reduceTimeline(batched, event);
    for (const event of events.slice(mid)) batched = reduceTimeline(batched, event);

    expect(batched.items).toEqual(full.items);
    expect(batched.meta).toEqual(full.meta);
  });

  it("never crashes on an unbalanced turn_end with no preceding turn_start", () => {
    const midSessionEvents: Event[] = [
      { ...base, type: "agent.turn_end", seq: 100, ts: "2026-08-24T15:01:00.000Z", turn_id: "orphan" },
      { ...base, type: "agent.text", seq: 101, ts: "2026-08-24T15:01:01.000Z", text: redacted("still here") },
    ];
    expect(() => buildTimeline(midSessionEvents)).not.toThrow();
  });
});
