import type { Event } from "@coop/protocol";
import { asKnown } from "./known-event";
import { reduceMessage } from "./reduce-message";
import { reducePermission } from "./reduce-permission";
import { reduceQuestion } from "./reduce-question";
import { reduceSession } from "./reduce-session";
import { reduceTool } from "./reduce-tool";
import { reduceTurn } from "./reduce-turn";
import { cloneTimelineState, type TimelineState } from "./state";

export function reduceTimeline(prev: TimelineState, raw: Event): TimelineState {
  const event = asKnown(raw);
  if (!event) return prev;

  const state = cloneTimelineState(prev);

  switch (event.type) {
    case "session.start":
    case "session.end":
    case "session.mode_changed":
    case "human.join":
    case "human.leave":
    case "human.takeover":
      reduceSession(state, event);
      break;

    case "tool.call":
    case "tool.result":
    case "tool.blocked":
    case "file.touched":
      reduceTool(state, event);
      break;

    case "agent.turn_start":
    case "agent.text":
    case "agent.turn_end":
      reduceTurn(state, event);
      break;

    case "human.steer":
    case "human.prompt":
    case "human.message":
    case "steer.delivered":
    case "steer.dropped":
      reduceMessage(state, event);
      break;

    case "permission.requested":
    case "permission.resolved":
    case "steer.requested":
    case "steer.resolved":
      reducePermission(state, event);
      break;

    case "agent.asked_team":
    case "human.answered":
    case "human.command":
      reduceQuestion(state, event);
      break;
  }

  return state;
}
