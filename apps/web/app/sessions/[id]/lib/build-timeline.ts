import type { z } from "zod";
import { KNOWN_EVENT_TYPES, type Event, type knownEvent } from "@coop/protocol";
import type { BuiltTimeline, SessionMeta, TimelineItem, ToolItem } from "../types";

type KnownEvent = z.infer<typeof knownEvent>;

const KNOWN = new Set<string>(KNOWN_EVENT_TYPES);

function asKnown(event: Event): KnownEvent | null {
  return KNOWN.has(event.type) ? (event as KnownEvent) : null;
}

function textOf(value: unknown): string {
  if (typeof value === "string") return value;
  if (value && typeof value === "object" && "text" in value) {
    const inner = (value as { text: unknown }).text;
    if (typeof inner === "string") return inner;
  }
  return "";
}

export function buildTimeline(events: Event[]): BuiltTimeline {
  const items: TimelineItem[] = [];
  const meta: SessionMeta = {};
  const toolItemIndex = new Map<string, number>();
  let openTools = 0;
  let sawTurnEnd = false;

  for (const raw of events) {
    const event = asKnown(raw);
    if (!event) continue;

    const key = `${event.type}-${event.seq}`;

    switch (event.type) {
      case "session.start": {
        meta.harness = event.harness;
        meta.owner = event.owner?.display_name;
        meta.repo = "repo" in event ? (event.repo as string | undefined) : undefined;
        meta.cwd = event.cwd;
        meta.startedAt = event.ts;
        items.push({
          kind: "notice",
          key,
          ts: event.ts,
          tone: "start",
          text: `${event.owner?.display_name ?? "someone"} started a ${event.harness} session`,
        });
        break;
      }

      case "session.end": {
        meta.endedAt = event.ts;
        items.push({ kind: "notice", key, ts: event.ts, tone: "end", text: "Session ended" });
        break;
      }

      case "tool.call": {
        const item: ToolItem = {
          kind: "tool",
          key,
          ts: event.ts,
          toolName: event.tool_name,
          input: textOf(event.input),
          status: "running",
          files: [],
        };
        items.push(item);
        openTools += 1;
        sawTurnEnd = false;
        if (event.tool_use_id) toolItemIndex.set(event.tool_use_id, items.length - 1);
        break;
      }

      case "tool.result": {
        const targetIndex = event.tool_use_id ? toolItemIndex.get(event.tool_use_id) : undefined;
        const target = targetIndex !== undefined ? (items[targetIndex] as ToolItem) : undefined;
        const output = textOf(event.output);
        const status = event.ok === false ? "failed" : "ok";

        if (target && targetIndex !== undefined) {
          items[targetIndex] = { ...target, output, status };
          openTools = Math.max(0, openTools - 1);
        } else {
          items.push({
            kind: "tool",
            key,
            ts: event.ts,
            toolName: event.tool_name,
            input: "",
            output,
            status,
            files: [],
          });
        }
        break;
      }

      case "file.touched": {
        const targetIndex = event.tool_use_id ? toolItemIndex.get(event.tool_use_id) : undefined;
        const target = targetIndex !== undefined ? (items[targetIndex] as ToolItem) : undefined;
        if (target && targetIndex !== undefined) {
          items[targetIndex] = { ...target, files: [...target.files, { path: event.path, mode: event.mode }] };
        } else {
          items.push({
            kind: "tool",
            key,
            ts: event.ts,
            toolName: event.mode === "write" ? "Write" : "Read",
            input: event.path,
            status: "ok",
            files: [{ path: event.path, mode: event.mode }],
          });
        }
        break;
      }

      case "agent.text": {
        const text = textOf(event.text);
        if (text.trim()) items.push({ kind: "agent-text", key, ts: event.ts, text });
        break;
      }

      case "agent.turn_end": {
        sawTurnEnd = true;
        break;
      }

      case "human.steer": {
        items.push({
          kind: "message",
          key,
          ts: event.ts,
          author: event.actor?.display_name ?? "someone",
          text: textOf(event.text),
          toAgent: true,
        });
        sawTurnEnd = false;
        break;
      }

      case "human.prompt": {
        items.push({
          kind: "message",
          key,
          ts: event.ts,
          author: meta.owner ?? "someone",
          text: textOf(event.text),
          toAgent: true,
        });
        sawTurnEnd = false;
        break;
      }

      case "human.takeover": {
        const by = event.actor?.display_name ?? "someone";
        meta.takeover = { active: event.active, by: event.active ? by : undefined };
        items.push({
          kind: "notice",
          key,
          ts: event.ts,
          tone: "takeover",
          text: event.active ? `${by} took over this session` : `${by} released this session`,
        });
        break;
      }

      case "human.join":
      case "human.leave": {
        items.push({
          kind: "notice",
          key,
          ts: event.ts,
          tone: event.type === "human.join" ? "join" : "leave",
          text: `${event.actor?.display_name ?? "someone"} ${
            event.type === "human.join" ? "joined" : "left"
          }`,
        });
        break;
      }
    }
  }

  return { items, meta, agentBusy: openTools > 0 || !sawTurnEnd };
}
