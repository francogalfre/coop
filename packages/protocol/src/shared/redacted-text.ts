import { z } from "zod";
import { LIMITS } from "./limits.js";

export const redactedText = z.object({
  text: z.string().max(LIMITS.text),
  redactions: z.int().nonnegative(),
  truncated: z.boolean(),
});

export type RedactedText = z.infer<typeof redactedText>;
