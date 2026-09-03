import type { HumanJoin, HumanLeave, HumanTakeover, SessionEnd, SessionModeChanged, SessionStart } from "@coop/protocol";
import type { TimelineState } from "./state";

type SessionEvent = SessionStart | SessionEnd | SessionModeChanged | HumanJoin | HumanLeave | HumanTakeover;

export function reduceSession(state: TimelineState, event: SessionEvent): void {
  const key = `${event.type}-${event.seq}`;

  switch (event.type) {
    case "session.start": {
      state.meta.harness = event.harness;
      state.meta.owner = event.owner
        ? { id: event.owner.id, name: event.owner.display_name, avatarUrl: event.owner.avatar_url }
        : undefined;
      state.meta.mode = event.mode ?? "auto";
      state.meta.repo = event.repo;
      state.meta.cwd = event.cwd;
      state.meta.startedAt = event.ts;
      state.meta.capabilities = event.capabilities;
      state.items.push({
        kind: "notice",
        key,
        seq: event.seq,
        ts: event.ts,
        tone: "start",
        text: `${event.owner?.display_name ?? "someone"} started a ${event.harness} session`,
      });
      return;
    }

    case "session.end": {
      state.meta.endedAt = event.ts;
      state.items.push({ kind: "notice", key, seq: event.seq, ts: event.ts, tone: "end", text: "Session ended" });
      return;
    }

    case "session.mode_changed": {
      state.meta.mode = event.mode;
      return;
    }

    case "human.takeover": {
      const by = event.actor.display_name;
      state.meta.takeover = { active: event.active, by: event.active ? by : undefined };
      state.items.push({
        kind: "notice",
        key,
        seq: event.seq,
        ts: event.ts,
        tone: "takeover",
        text: event.active ? `${by} took over this session` : `${by} released this session`,
      });
      return;
    }

    case "human.join":
    case "human.leave": {
      state.items.push({
        kind: "notice",
        key,
        seq: event.seq,
        ts: event.ts,
        tone: event.type === "human.join" ? "join" : "leave",
        text: `${event.actor.display_name} ${event.type === "human.join" ? "joined" : "left"}`,
      });
      return;
    }
  }
}
