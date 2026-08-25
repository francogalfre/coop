import { z } from "zod";
import { sessionStart, sessionEnd } from "./events/session.js";
import { agentTurnStart, agentText, agentTurnEnd } from "./events/agent.js";
import { toolCall, toolResult, toolBlocked } from "./events/tool.js";
import { fileTouched } from "./events/file.js";
import { permissionRequested, permissionResolved } from "./events/permission.js";
import { humanJoin, humanLeave, humanSteer, humanTakeover, humanPrompt } from "./events/human.js";
import { unknownEvent } from "./events/unknown.js";

export const EVENT_SCHEMAS = {
  "session.start": sessionStart,
  "session.end": sessionEnd,
  "agent.turn_start": agentTurnStart,
  "agent.text": agentText,
  "agent.turn_end": agentTurnEnd,
  "tool.call": toolCall,
  "tool.result": toolResult,
  "tool.blocked": toolBlocked,
  "file.touched": fileTouched,
  "permission.requested": permissionRequested,
  "permission.resolved": permissionResolved,
  "human.join": humanJoin,
  "human.leave": humanLeave,
  "human.steer": humanSteer,
  "human.takeover": humanTakeover,
  "human.prompt": humanPrompt,
} as const;

export const KNOWN_EVENT_TYPES = Object.keys(EVENT_SCHEMAS) as (keyof typeof EVENT_SCHEMAS)[];

export const knownEvent = z.discriminatedUnion("type", [
  sessionStart, sessionEnd,
  agentTurnStart, agentText, agentTurnEnd,
  toolCall, toolResult, toolBlocked,
  fileTouched,
  permissionRequested, permissionResolved,
  humanJoin, humanLeave, humanSteer, humanTakeover, humanPrompt,
]);

export const event = z.union([knownEvent, unknownEvent]);
export type Event = z.infer<typeof event>;
