import type { TimelineItem } from "../../types";

type TurnStart = Extract<TimelineItem, { kind: "turn-start" }>;
type TurnEnd = Extract<TimelineItem, { kind: "turn-end" }>;

export type TimelineGroup =
  | { kind: "item"; item: TimelineItem }
  | { kind: "turn"; key: string; start: TurnStart; end?: TurnEnd; items: TimelineItem[] };

export function groupTurns(items: TimelineItem[]): TimelineGroup[] {
  const groups: TimelineGroup[] = [];
  let open: Extract<TimelineGroup, { kind: "turn" }> | null = null;

  for (const item of items) {
    if (item.kind === "turn-start") {
      open = { kind: "turn", key: `turn-${item.key}`, start: item, items: [] };
      groups.push(open);
      continue;
    }

    if (item.kind === "turn-end") {
      const matches = open && (item.turnId === undefined || open.start.turnId === undefined || item.turnId === open.start.turnId);
      if (open && matches) {
        open.end = item;
        open = null;
      } else {
        groups.push({ kind: "item", item });
      }
      continue;
    }

    if (open) open.items.push(item);
    else groups.push({ kind: "item", item });
  }

  return groups;
}
