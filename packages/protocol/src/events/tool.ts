import { z } from "zod";
import { envelopeFields } from "../envelope.js";
import { redactedText } from "../shared/redacted-text.js";
import { actor } from "../shared/actor.js";
import { LIMITS } from "../shared/limits.js";

export const toolCall = z.object({
  ...envelopeFields,
  type: z.literal("tool.call"),
  tool_name: z.string().min(1).max(LIMITS.tool_name),
  input: redactedText,
  tool_use_id: z.string().max(LIMITS.tool_use_id).optional(),
});
export type ToolCall = z.infer<typeof toolCall>;

export const toolResult = z.object({
  ...envelopeFields,
  type: z.literal("tool.result"),
  tool_name: z.string().min(1).max(LIMITS.tool_name),
  output: redactedText,
  tool_use_id: z.string().max(LIMITS.tool_use_id).optional(),
  ok: z.boolean().optional(),
});
export type ToolResult = z.infer<typeof toolResult>;

export const toolBlocked = z.object({
  ...envelopeFields,
  type: z.literal("tool.blocked"),
  tool_name: z.string().min(1).max(LIMITS.tool_name),
  blocked_by: actor,
  reason: redactedText,
  tool_use_id: z.string().max(LIMITS.tool_use_id).optional(),
});
export type ToolBlocked = z.infer<typeof toolBlocked>;
