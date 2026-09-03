import { z } from "zod";
import { LIMITS } from "./limits.js";

export const actor = z.object({
  id: z.string().min(1).max(LIMITS.actor_id),
  display_name: z.string().min(1).max(LIMITS.display_name),
  avatar_url: z.url().max(LIMITS.avatar_url).optional(),
});

export type Actor = z.infer<typeof actor>;
