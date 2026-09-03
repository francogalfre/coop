import { z } from "zod";
import { envelopeFields } from "../envelope.js";
import { redactedText } from "../shared/redacted-text.js";
import { actor } from "../shared/actor.js";
import { harnessCommand } from "../shared/commands.js";
import { LIMITS } from "../shared/limits.js";

export const humanJoin = z.object({ ...envelopeFields, type: z.literal("human.join"), actor });
export type HumanJoin = z.infer<typeof humanJoin>;

export const humanLeave = z.object({ ...envelopeFields, type: z.literal("human.leave"), actor });
export type HumanLeave = z.infer<typeof humanLeave>;

export const humanSteer = z.object({
  ...envelopeFields,
  type: z.literal("human.steer"),
  actor,
  text: redactedText,
  request_id: z.string().min(1).optional(),
  steer_id: z.string().min(1).max(LIMITS.request_id).optional(),
  client_id: z.string().min(1).max(LIMITS.client_id).optional(),
  project_context_version: z.int().nonnegative().optional(),
});
export type HumanSteer = z.infer<typeof humanSteer>;

export const humanMessage = z.object({
  ...envelopeFields,
  type: z.literal("human.message"),
  actor,
  text: redactedText,
  anchor_seq: z.int().nonnegative().optional(),
  client_id: z.string().min(1).max(LIMITS.client_id).optional(),
});
export type HumanMessage = z.infer<typeof humanMessage>;

export const humanTakeover = z.object({
  ...envelopeFields,
  type: z.literal("human.takeover"),
  actor,
  active: z.boolean(),
});
export type HumanTakeover = z.infer<typeof humanTakeover>;

export const humanPrompt = z.object({
  ...envelopeFields,
  type: z.literal("human.prompt"),
  text: redactedText,
});
export type HumanPrompt = z.infer<typeof humanPrompt>;

export const humanCommand = z.object({
  ...envelopeFields,
  type: z.literal("human.command"),
  actor,
  command: harnessCommand,
  args: z.string().max(LIMITS.command_args).optional(),
});
export type HumanCommand = z.infer<typeof humanCommand>;

export const humanAnswered = z.object({
  ...envelopeFields,
  type: z.literal("human.answered"),
  question_id: z.string().min(1).max(LIMITS.request_id),
  actor,
  text: redactedText,
});
export type HumanAnswered = z.infer<typeof humanAnswered>;
