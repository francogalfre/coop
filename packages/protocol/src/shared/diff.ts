import { z } from "zod";
import { redactedText } from "./redacted-text.js";
import { LIMITS } from "./limits.js";

export const diffHunk = z.object({
  old_start: z.int().nonnegative(),
  old_lines: z.int().nonnegative(),
  new_start: z.int().nonnegative(),
  new_lines: z.int().nonnegative(),
  lines: z.array(redactedText).max(LIMITS.diff_lines),
});

export type DiffHunk = z.infer<typeof diffHunk>;
