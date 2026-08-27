import { z } from "zod";
import { envelopeFields } from "../envelope.js";
import { actor } from "../shared/actor.js";

export const sessionModeChanged = z.object({
  ...envelopeFields,
  type: z.literal("session.mode_changed"),
  mode: z.enum(["auto", "restricted"]),
  changed_by: actor,
});
export type SessionModeChanged = z.infer<typeof sessionModeChanged>;
