import { z } from "zod";
import { envelopeFields } from "../envelope.js";
import { LIMITS } from "../shared/limits.js";

export const fileTouched = z.object({
  ...envelopeFields,
  type: z.literal("file.touched"),
  path: z.string().min(1).max(LIMITS.path),
  mode: z.enum(["read", "write"]),
  tool_use_id: z.string().max(LIMITS.tool_use_id).optional(),
});
export type FileTouched = z.infer<typeof fileTouched>;
