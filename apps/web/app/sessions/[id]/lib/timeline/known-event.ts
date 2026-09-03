import { KNOWN_EVENT_TYPES, type Event, type knownEvent } from "@coop/protocol";
import type { z } from "zod";

export type KnownEvent = z.infer<typeof knownEvent>;

const KNOWN = new Set<string>(KNOWN_EVENT_TYPES);

export function asKnown(event: Event): KnownEvent | null {
  return KNOWN.has(event.type) ? (event as KnownEvent) : null;
}

// `coop run` steering is delivered by typing "[name via coop] text" into the
// harness's own input (packages/cli/internal/ptywrap/steer.go), so it comes
// back through UserPromptSubmit as a second, indistinguishable-looking
// human.prompt. The human.steer event already rendered that message once.
export const STEER_ECHO = /^\[.+ via coop\] /;
