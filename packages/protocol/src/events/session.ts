import { z } from "zod";
import { envelopeFields } from "../envelope.js";
import { actor } from "../shared/actor.js";
import { LIMITS } from "../shared/limits.js";

export const sessionStart = z.object({
  ...envelopeFields,
  type: z.literal("session.start"),
  harness: z.enum(["claude-code", "codex", "opencode", "pi", "amp", "other"]),
  cwd: z.string().min(1).max(LIMITS.path),
  repo: z.string().min(1).max(LIMITS.path).optional(),
  owner: actor,
  harness_version: z.string().max(64).optional(),
  permission_mode: z.string().max(64).optional(),
});
export type SessionStart = z.infer<typeof sessionStart>;

export const sessionEnd = z.object({
  ...envelopeFields,
  type: z.literal("session.end"),
  reason: z.enum(["completed", "cancelled", "error"]).optional(),
});
export type SessionEnd = z.infer<typeof sessionEnd>;
