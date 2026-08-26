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

export type NoticeTone = "start" | "end" | "join" | "leave" | "turn";

export type NoticeItem = {
  kind: "notice";
  key: string;
  ts: string;
  tone: NoticeTone;
  text: string;
};

export type TimelineItem = ToolItem | AgentTextItem | MessageItem | NoticeItem;

export type SessionMeta = {
  harness?: string;
  owner?: string;
  repo?: string;
  cwd?: string;
  startedAt?: string;
  endedAt?: string;
};

export type BuiltTimeline = {
  items: TimelineItem[];
  meta: SessionMeta;
  agentBusy: boolean;
};
