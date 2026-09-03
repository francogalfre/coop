import type { SessionMeta, TimelineItem } from "../../types";

export type TimelineState = {
  items: TimelineItem[];
  meta: SessionMeta;
  agentBusy: boolean;
  openTools: number;
  sawTurnEnd: boolean;
  toolItemIndex: Map<string, number>;
  todoItemIndex: number | null;
  steerItemIndex: Map<string, number>;
  permissionItemIndex: Map<string, number>;
  questionItemIndex: Map<string, number>;
  steerMessageIndex: Map<string, number>;
  pendingSeen: Set<number>;
};

export function createTimelineState(): TimelineState {
  return {
    items: [],
    meta: {},
    agentBusy: false,
    openTools: 0,
    sawTurnEnd: false,
    toolItemIndex: new Map(),
    todoItemIndex: null,
    steerItemIndex: new Map(),
    permissionItemIndex: new Map(),
    questionItemIndex: new Map(),
    steerMessageIndex: new Map(),
    pendingSeen: new Set(),
  };
}

export function cloneTimelineState(state: TimelineState): TimelineState {
  return {
    ...state,
    items: [...state.items],
    meta: { ...state.meta },
    toolItemIndex: new Map(state.toolItemIndex),
    steerItemIndex: new Map(state.steerItemIndex),
    permissionItemIndex: new Map(state.permissionItemIndex),
    questionItemIndex: new Map(state.questionItemIndex),
    steerMessageIndex: new Map(state.steerMessageIndex),
    pendingSeen: new Set(state.pendingSeen),
  };
}
