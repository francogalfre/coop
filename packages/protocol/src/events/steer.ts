import { z } from "zod";
import { envelopeFields } from "../envelope.js";
import { redactedText } from "../shared/redacted-text.js";
import { actor } from "../shared/actor.js";
import { LIMITS } from "../shared/limits.js";

export const steerRequested = z.object({
  ...envelopeFields,
  type: z.literal("steer.requested"),
  request_id: z.string().min(1).max(LIMITS.request_id),
  actor,
  text: redactedText,
});
export type SteerRequested = z.infer<typeof steerRequested>;

export const steerResolved = z.object({
  ...envelopeFields,
  type: z.literal("steer.resolved"),
  request_id: z.string().min(1).max(LIMITS.request_id),
  decision: z.enum(["allow", "deny"]),
  resolved_by: actor,
});
export type SteerResolved = z.infer<typeof steerResolved>;
