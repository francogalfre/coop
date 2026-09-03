import { z } from "zod";

export const capabilities = z.object({
  steer: z.boolean(),
  block: z.boolean(),
  commands: z.boolean(),
  pty: z.boolean(),
});

export type Capabilities = z.infer<typeof capabilities>;
