import type { HumanMessage, HumanPrompt, HumanSteer, SteerDelivered, SteerDropped } from "@coop/protocol";
import type { MessageItem } from "../../types";
import type { TimelineState } from "./state";
import { STEER_ECHO } from "./known-event";
import { textOf } from "./redacted";

type MessageEvent = HumanSteer | HumanPrompt | HumanMessage | SteerDelivered | SteerDropped;

export function reduceMessage(state: TimelineState, event: MessageEvent): void {
  const key = `${event.type}-${event.seq}`;

  switch (event.type) {
    case "human.steer": {
      if (event.request_id && state.steerItemIndex.has(event.request_id)) {
        state.sawTurnEnd = false;
        return;
      }

      const item: MessageItem = {
        kind: "message",
        key,
        seq: event.seq,
        ts: event.ts,
        author: event.actor.display_name,
        authorAvatarUrl: event.actor.avatar_url,
        text: textOf(event.text),
        toAgent: true,
        clientId: event.client_id,
        steerId: event.steer_id,
        delivery: event.steer_id ? "queued" : undefined,
        projectContextVersion: event.project_context_version,
      };
      state.items.push(item);
      if (event.steer_id) state.steerMessageIndex.set(event.steer_id, state.items.length - 1);
      state.sawTurnEnd = false;
      return;
    }

    case "human.prompt": {
      const text = textOf(event.text);
      if (!STEER_ECHO.test(text)) {
        state.items.push({
          kind: "message",
          key,
          seq: event.seq,
          ts: event.ts,
          author: state.meta.owner?.name ?? "someone",
          authorAvatarUrl: state.meta.owner?.avatarUrl,
          text,
          toAgent: true,
        });
      }
      state.sawTurnEnd = false;
      return;
    }

    case "human.message": {
      state.items.push({
        kind: "message",
        key,
        seq: event.seq,
        ts: event.ts,
        author: event.actor.display_name,
        authorAvatarUrl: event.actor.avatar_url,
        text: textOf(event.text),
        toAgent: false,
        anchorSeq: event.anchor_seq,
        clientId: event.client_id,
      });
      return;
    }

    case "steer.delivered": {
      const index = state.steerMessageIndex.get(event.steer_id);
      if (index === undefined) return;
      const item = state.items[index];
      if (item?.kind === "message") {
        state.items[index] = { ...item, delivery: "delivered" };
        state.pendingSeen.add(index);
      }
      return;
    }

    case "steer.dropped": {
      const index = state.steerMessageIndex.get(event.steer_id);
      if (index === undefined) return;
      const item = state.items[index];
      if (item?.kind === "message") state.items[index] = { ...item, delivery: "dropped" };
      state.pendingSeen.delete(index);
      return;
    }
  }
}
