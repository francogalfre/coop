import { z } from "zod";
import { envelopeFields } from "../envelope.js";
import { redactedText } from "../shared/redacted-text.js";
import { actor } from "../shared/actor.js";

export const humanJoin = z.object({ ...envelopeFields, type: z.literal("human.join"), actor });
export type HumanJoin = z.infer<typeof humanJoin>;

export const humanLeave = z.object({ ...envelopeFields, type: z.literal("human.leave"), actor });
export type HumanLeave = z.infer<typeof humanLeave>;

export const humanSteer = z.object({
  ...envelopeFields,
  type: z.literal("human.steer"),
  actor,
  text: redactedText,
});
export type HumanSteer = z.infer<typeof humanSteer>;

export const humanTakeover = z.object({ ...envelopeFields, type: z.literal("human.takeover"), actor });
export type HumanTakeover = z.infer<typeof humanTakeover>;
