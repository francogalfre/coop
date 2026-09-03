import { z } from "zod";

export const askTeamInputSchema = {
  question: z.string().min(1).max(2000),
  options: z.array(z.string().min(1).max(200)).min(2).max(6).optional(),
  timeout_seconds: z.number().int().positive().max(1800).optional(),
};
