import { z } from "zod";

export const usage = z.object({
  input_tokens: z.int().nonnegative().optional(),
  output_tokens: z.int().nonnegative().optional(),
  cache_creation_input_tokens: z.int().nonnegative().optional(),
  cache_read_input_tokens: z.int().nonnegative().optional(),
  cost_usd: z.number().nonnegative().optional(),
});

export type Usage = z.infer<typeof usage>;
