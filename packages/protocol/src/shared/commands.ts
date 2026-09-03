import { z } from "zod";

// An allowlist, not a convenience: a harness command is typed into the session
// owner's own terminal, so free text here would be remote keystroke injection.
export const HARNESS_COMMANDS = ["model", "compact", "clear", "context", "cost", "status"] as const;

export const harnessCommand = z.enum(HARNESS_COMMANDS);

export type HarnessCommand = z.infer<typeof harnessCommand>;
