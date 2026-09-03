import type { AgentText, AgentTurnEnd, AgentTurnStart } from "@coop/protocol";
import type { TimelineState } from "./state";
import { readRedacted } from "./redacted";

type TurnEvent = AgentTurnStart | AgentText | AgentTurnEnd;

export function reduceTurn(state: TimelineState, event: TurnEvent): void {
  const key = `${event.type}-${event.seq}`;

  switch (event.type) {
    case "agent.turn_start": {
      state.items.push({ kind: "turn-start", key, seq: event.seq, ts: event.ts, turnId: event.turn_id });
      state.sawTurnEnd = false;

      for (const index of state.pendingSeen) {
        const item = state.items[index];
        if (item?.kind === "message") state.items[index] = { ...item, delivery: "seen" };
      }
      state.pendingSeen.clear();
      return;
    }

    case "agent.text": {
      const redacted = readRedacted(event.text);
      if (redacted.text.trim()) {
        state.items.push({
          kind: "agent-text",
          key,
          seq: event.seq,
          ts: event.ts,
          text: redacted.text,
          redactions: redacted.redactions,
          truncated: redacted.truncated,
        });
      }
      return;
    }

    case "agent.turn_end": {
      state.sawTurnEnd = true;
      state.items.push({
        kind: "turn-end",
        key,
        seq: event.seq,
        ts: event.ts,
        turnId: event.turn_id,
        durationMs: event.duration_ms,
        usage: event.usage
          ? {
              inputTokens: event.usage.input_tokens,
              outputTokens: event.usage.output_tokens,
              cacheCreationInputTokens: event.usage.cache_creation_input_tokens,
              cacheReadInputTokens: event.usage.cache_read_input_tokens,
              costUsd: event.usage.cost_usd,
            }
          : undefined,
      });
      return;
    }
  }
}
