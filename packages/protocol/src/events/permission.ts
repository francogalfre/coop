import { z } from "zod";
import { envelopeFields } from "../envelope.js";
import { redactedText } from "../shared/redacted-text.js";
import { actor } from "../shared/actor.js";
import { LIMITS } from "../shared/limits.js";

export const permissionRequested = z.object({
  ...envelopeFields,
  type: z.literal("permission.requested"),
  request_id: z.string().min(1).max(LIMITS.request_id),
  tool_name: z.string().min(1).max(LIMITS.tool_name),
  input: redactedText,
  permission_mode: z.string().max(64).optional(),
});
export type PermissionRequested = z.infer<typeof permissionRequested>;

export const permissionResolved = z.object({
  ...envelopeFields,
  type: z.literal("permission.resolved"),
  request_id: z.string().min(1).max(LIMITS.request_id),
  decision: z.enum(["allow", "deny"]),
  resolved_by: actor,
  reason: redactedText.optional(),
});
export type PermissionResolved = z.infer<typeof permissionResolved>;
