import { z } from "zod";
import { envelopeFields } from "../envelope.js";
import { redactedText } from "../shared/redacted-text.js";
import { usage } from "../shared/usage.js";
import { LIMITS } from "../shared/limits.js";

export const agentTurnStart = z.object({
  ...envelopeFields,
  type: z.literal("agent.turn_start"),
  turn_id: z.string().max(LIMITS.turn_id).optional(),
});
export type AgentTurnStart = z.infer<typeof agentTurnStart>;

export const agentText = z.object({
  ...envelopeFields,
  type: z.literal("agent.text"),
  text: redactedText,
  turn_id: z.string().max(LIMITS.turn_id).optional(),
});
export type AgentText = z.infer<typeof agentText>;

export const agentTurnEnd = z.object({
  ...envelopeFields,
  type: z.literal("agent.turn_end"),
  turn_id: z.string().max(LIMITS.turn_id).optional(),
  usage: usage.optional(),
  duration_ms: z.int().nonnegative().optional(),
});
export type AgentTurnEnd = z.infer<typeof agentTurnEnd>;

export const agentAskedTeam = z.object({
  ...envelopeFields,
  type: z.literal("agent.asked_team"),
  question_id: z.string().min(1).max(LIMITS.request_id),
  text: redactedText,
  options: z.array(z.string().min(1).max(LIMITS.question_option)).max(LIMITS.question_options).optional(),
  turn_id: z.string().max(LIMITS.turn_id).optional(),
});
export type AgentAskedTeam = z.infer<typeof agentAskedTeam>;
