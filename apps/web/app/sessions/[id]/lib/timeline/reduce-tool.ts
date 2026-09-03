import type { FileTouched, ToolBlocked, ToolCall, ToolResult } from "@coop/protocol";
import type { ToolItem, ToolStatus, TouchedFile } from "../../types";
import type { TimelineState } from "./state";
import { readRedacted } from "./redacted";

type ToolEvent = ToolCall | ToolResult | ToolBlocked | FileTouched;

export function reduceTool(state: TimelineState, event: ToolEvent): void {
  const key = `${event.type}-${event.seq}`;

  switch (event.type) {
    case "tool.call": {
      const redacted = readRedacted(event.input);
      const isTodoWrite = event.tool_name.toLowerCase() === "todowrite";

      if (isTodoWrite && state.todoItemIndex !== null) {
        const targetIndex = state.todoItemIndex;
        const target = state.items[targetIndex] as ToolItem;
        state.items[targetIndex] = {
          ...target,
          ts: event.ts,
          input: redacted.text,
          inputRedactions: redacted.redactions,
          inputTruncated: redacted.truncated,
          output: undefined,
          outputRedactions: 0,
          outputTruncated: false,
          status: "running",
        };
        if (event.tool_use_id) state.toolItemIndex.set(event.tool_use_id, targetIndex);
        return;
      }

      const item: ToolItem = {
        kind: "tool",
        key,
        seq: event.seq,
        ts: event.ts,
        toolName: event.tool_name,
        input: redacted.text,
        inputRedactions: redacted.redactions,
        inputTruncated: redacted.truncated,
        outputRedactions: 0,
        outputTruncated: false,
        status: "running",
        files: [],
      };
      state.items.push(item);
      state.openTools += 1;
      state.sawTurnEnd = false;
      if (event.tool_use_id) state.toolItemIndex.set(event.tool_use_id, state.items.length - 1);
      if (isTodoWrite) state.todoItemIndex = state.items.length - 1;
      return;
    }

    case "tool.result": {
      const targetIndex = event.tool_use_id ? state.toolItemIndex.get(event.tool_use_id) : undefined;
      const target = targetIndex !== undefined ? (state.items[targetIndex] as ToolItem) : undefined;
      const redacted = readRedacted(event.output);
      const status: ToolStatus = event.ok === false ? "failed" : "ok";

      if (target && targetIndex !== undefined) {
        state.items[targetIndex] = {
          ...target,
          output: redacted.text,
          outputRedactions: redacted.redactions,
          outputTruncated: redacted.truncated,
          status,
          durationMs: event.duration_ms,
        };
        state.openTools = Math.max(0, state.openTools - 1);
      } else {
        state.items.push({
          kind: "tool",
          key,
          seq: event.seq,
          ts: event.ts,
          toolName: event.tool_name,
          input: "",
          inputRedactions: 0,
          inputTruncated: false,
          output: redacted.text,
          outputRedactions: redacted.redactions,
          outputTruncated: redacted.truncated,
          status,
          files: [],
          durationMs: event.duration_ms,
        });
      }
      return;
    }

    case "file.touched": {
      const targetIndex = event.tool_use_id ? state.toolItemIndex.get(event.tool_use_id) : undefined;
      const target = targetIndex !== undefined ? (state.items[targetIndex] as ToolItem) : undefined;
      const file: TouchedFile = {
        path: event.path,
        mode: event.mode,
        added: event.added,
        removed: event.removed,
        hunks: event.hunks,
      };

      if (target && targetIndex !== undefined) {
        state.items[targetIndex] = { ...target, files: [...target.files, file] };
      } else {
        state.items.push({
          kind: "tool",
          key,
          seq: event.seq,
          ts: event.ts,
          toolName: event.mode === "write" ? "Write" : "Read",
          input: event.path,
          inputRedactions: 0,
          inputTruncated: false,
          outputRedactions: 0,
          outputTruncated: false,
          status: "ok",
          files: [file],
        });
      }
      return;
    }

    case "tool.blocked": {
      const targetIndex = event.tool_use_id ? state.toolItemIndex.get(event.tool_use_id) : undefined;
      if (targetIndex !== undefined) {
        const target = state.items[targetIndex] as ToolItem;
        state.items[targetIndex] = { ...target, status: "failed" };
        state.openTools = Math.max(0, state.openTools - 1);
      }

      const reason = readRedacted(event.reason);
      state.items.push({
        kind: "blocked",
        key,
        seq: event.seq,
        ts: event.ts,
        toolName: event.tool_name,
        blockedBy: event.blocked_by.display_name,
        blockedByAvatarUrl: event.blocked_by.avatar_url,
        reason: reason.text,
        reasonRedactions: reason.redactions,
        reasonTruncated: reason.truncated,
      });
      return;
    }
  }
}
