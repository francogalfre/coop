import type { PresenceFrame } from "./presenceFrame";

export type PresenceState = Record<string, { active: boolean; at: number }>;

export function emptyPresenceState(): PresenceState {
  return {};
}

export function reducePresence(state: PresenceState, frame: PresenceFrame): PresenceState {
  if (frame.type !== "presence.typing") return state;
  return { ...state, [frame.actor.name]: { active: frame.active, at: Date.now() } };
}
