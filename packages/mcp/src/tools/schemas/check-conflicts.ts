import { z } from "zod";

export const checkConflictsInputSchema = {
  paths: z.array(z.string().min(1)).min(1).max(50),
  window_seconds: z.number().int().positive().max(3600).optional(),
};
