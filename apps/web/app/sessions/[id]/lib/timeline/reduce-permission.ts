import type { PermissionRequested, PermissionResolved, SteerRequested, SteerResolved } from "@coop/protocol";
import type { PermissionItem, SteerRequestItem } from "../../types";
import type { TimelineState } from "./state";
import { readRedacted, textOf } from "./redacted";

type PermissionEvent = PermissionRequested | PermissionResolved | SteerRequested | SteerResolved;

export function reducePermission(state: TimelineState, event: PermissionEvent): void {
  const key = `${event.type}-${event.seq}`;

  switch (event.type) {
    case "permission.requested": {
      const input = readRedacted(event.input);
      const item: PermissionItem = {
        kind: "permission",
        key,
        seq: event.seq,
        ts: event.ts,
        requestId: event.request_id,
        toolName: event.tool_name,
        input: input.text,
        inputRedactions: input.redactions,
        inputTruncated: input.truncated,
        status: "pending",
      };
      state.items.push(item);
      state.permissionItemIndex.set(event.request_id, state.items.length - 1);
      return;
    }

    case "permission.resolved": {
      const targetIndex = state.permissionItemIndex.get(event.request_id);
      if (targetIndex === undefined) return;
      const target = state.items[targetIndex] as PermissionItem;
      state.items[targetIndex] = {
        ...target,
        status: event.decision === "allow" ? "allowed" : "denied",
        resolvedBy: event.resolved_by.display_name,
        reason: event.reason ? textOf(event.reason) : undefined,
      };
      return;
    }

    case "steer.requested": {
      const item: SteerRequestItem = {
        kind: "steer-request",
        key,
        seq: event.seq,
        ts: event.ts,
        requestId: event.request_id,
        author: event.actor.display_name,
        authorId: event.actor.id,
        authorAvatarUrl: event.actor.avatar_url,
        text: textOf(event.text),
        status: "pending",
      };
      state.items.push(item);
      state.steerItemIndex.set(event.request_id, state.items.length - 1);
      return;
    }

    case "steer.resolved": {
      const targetIndex = state.steerItemIndex.get(event.request_id);
      if (targetIndex === undefined) return;
      const target = state.items[targetIndex] as SteerRequestItem;
      state.items[targetIndex] = {
        ...target,
        status: event.decision === "allow" ? "allowed" : "denied",
        resolvedBy: event.resolved_by.display_name,
      };
      return;
    }
  }
}
