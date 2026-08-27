export type ToolStatus = "running" | "ok" | "failed";

export type TouchedFile = {
  path: string;
  mode: "read" | "write";
};

export type ToolItem = {
  kind: "tool";
  key: string;
  ts: string;
  toolName: string;
  input: string;
  output?: string;
  status: ToolStatus;
  files: TouchedFile[];
};

export type AgentTextItem = {
  kind: "agent-text";
  key: string;
  ts: string;
  text: string;
};

export type MessageItem = {
  kind: "message";
  key: string;
  ts: string;
  author: string;
  text: string;
  toAgent: boolean;
};

export type NoticeTone = "start" | "end" | "join" | "leave" | "turn" | "takeover";

export type NoticeItem = {
  kind: "notice";
  key: string;
  ts: string;
  tone: NoticeTone;
  text: string;
};

export type SteerRequestStatus = "pending" | "allowed" | "denied";

export type SteerRequestItem = {
  kind: "steer-request";
  key: string;
  ts: string;
  requestId: string;
  author: string;
  authorId: string;
  text: string;
  status: SteerRequestStatus;
  resolvedBy?: string;
};

export type TimelineItem = ToolItem | AgentTextItem | MessageItem | NoticeItem | SteerRequestItem;

export type TakeoverState = {
  active: boolean;
  by?: string;
};

export type SessionOwner = {
  id: string;
  name: string;
};

export type SessionMeta = {
  harness?: string;
  owner?: SessionOwner;
  repo?: string;
  cwd?: string;
  startedAt?: string;
  endedAt?: string;
  takeover?: TakeoverState;
  mode?: "auto" | "restricted";
};

export type BuiltTimeline = {
  items: TimelineItem[];
  meta: SessionMeta;
  agentBusy: boolean;
};
