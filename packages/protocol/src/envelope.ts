import { z } from "zod";
import { LIMITS } from "./shared/limits.js";

export const PROTOCOL_VERSION = 1;

export const envelopeFields = {
  v: z.literal(PROTOCOL_VERSION),
  session_id: z.string().min(1).max(LIMITS.session_id),
  seq: z.int().nonnegative(),
  ts: z.iso.datetime(),
};

export const envelope = z.object({ ...envelopeFields, type: z.string().min(1) });

export type Envelope = z.infer<typeof envelope>;
