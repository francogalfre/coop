import type { Event } from "@coop/protocol";
import type { BuiltTimeline } from "../types";
import { reduceTimeline } from "./timeline/reduce";
import { createTimelineState } from "./timeline/state";

export function buildTimeline(events: Event[]): BuiltTimeline {
  let state = createTimelineState();
  for (const event of events) state = reduceTimeline(state, event);
  return { items: state.items, meta: state.meta, agentBusy: state.openTools > 0 || !state.sawTurnEnd };
}
