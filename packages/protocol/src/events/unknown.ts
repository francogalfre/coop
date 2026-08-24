import { z } from "zod";
import { envelopeFields } from "../envelope.js";

export const unknownEvent = z.looseObject({
  ...envelopeFields,
  type: z.string().min(1),
});
export type UnknownEvent = z.infer<typeof unknownEvent>;
