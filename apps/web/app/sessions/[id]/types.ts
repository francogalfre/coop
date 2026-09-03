import type { Capabilities, DiffHunk } from "@coop/protocol";

export type ToolStatus = "running" | "ok" | "failed";

export type TouchedFile = {
  path: string;
  mode: "read" | "write";
  added?: number;
  removed?: number;
  hunks?: DiffHunk[];
};

export type ToolItem = {
  kind: "tool";
  key: string;
  seq: number;
  ts: string;
  toolName: string;
  input: string;
  inputRedactions: number;
  inputTruncated: boolean;
  output?: string;
  outputRedactions: number;
  outputTruncated: boolean;
  status: ToolStatus;
  files: TouchedFile[];
  durationMs?: number;
};

export type AgentTextItem = {
  kind: "agent-text";
  key: string;
  seq: number;
  ts: string;
  text: string;
  redactions: number;
  truncated: boolean;
};

export type DeliveryState = "sending" | "queued" | "delivered" | "seen" | "dropped";

export type MessageItem = {
  kind: "message";
  key: string;
  seq: number;
  ts: string;
  author: string;
  authorAvatarUrl?: string;
  text: string;
  toAgent: boolean;
  anchorSeq?: number;
  clientId?: string;
  steerId?: string;
  delivery?: DeliveryState;
  queuePosition?: number;
  projectContextVersion?: number;
};

export type NoticeTone = "start" | "end" | "join" | "leave" | "turn" | "takeover";

export type NoticeItem = {
  kind: "notice";
  key: string;
  seq: number;
  ts: string;
  tone: NoticeTone;
  text: string;
};

export type SteerRequestStatus = "pending" | "allowed" | "denied";

export type SteerRequestItem = {
  kind: "steer-request";
  key: string;
  seq: number;
  ts: string;
  requestId: string;
  author: string;
  authorId: string;
  authorAvatarUrl?: string;
  text: string;
  status: SteerRequestStatus;
  resolvedBy?: string;
};

export type PermissionStatus = "pending" | "allowed" | "denied";

export type PermissionItem = {
  kind: "permission";
  key: string;
  seq: number;
  ts: string;
  requestId: string;
  toolName: string;
  input: string;
  inputRedactions: number;
  inputTruncated: boolean;
  status: PermissionStatus;
  resolvedBy?: string;
  reason?: string;
};

export type BlockedItem = {
  kind: "blocked";
  key: string;
  seq: number;
  ts: string;
  toolName: string;
  blockedBy: string;
  blockedByAvatarUrl?: string;
  reason: string;
  reasonRedactions: number;
  reasonTruncated: boolean;
};

export type TurnStartItem = {
  kind: "turn-start";
  key: string;
  seq: number;
  ts: string;
  turnId?: string;
};

export type TurnUsage = {
  inputTokens?: number;
  outputTokens?: number;
  cacheCreationInputTokens?: number;
  cacheReadInputTokens?: number;
  costUsd?: number;
};

export type TurnEndItem = {
  kind: "turn-end";
  key: string;
  seq: number;
  ts: string;
  turnId?: string;
  durationMs?: number;
  usage?: TurnUsage;
};

export type QuestionStatus = "open" | "answered";

export type QuestionItem = {
  kind: "question";
  key: string;
  seq: number;
  ts: string;
  questionId: string;
  text: string;
  textRedactions: number;
  textTruncated: boolean;
  options?: string[];
  status: QuestionStatus;
  answeredBy?: string;
  answerText?: string;
};

export type CommandItem = {
  kind: "command";
  key: string;
  seq: number;
  ts: string;
  author: string;
  authorAvatarUrl?: string;
  command: string;
  args?: string;
};

export type TimelineItem =
  | ToolItem
  | AgentTextItem
  | MessageItem
  | NoticeItem
  | SteerRequestItem
  | PermissionItem
  | BlockedItem
  | TurnStartItem
  | TurnEndItem
  | QuestionItem
  | CommandItem;

export type TakeoverState = {
  active: boolean;
  by?: string;
};

export type SessionOwner = {
  id: string;
  name: string;
  avatarUrl?: string;
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
  capabilities?: Capabilities;
};

export type BuiltTimeline = {
  items: TimelineItem[];
  meta: SessionMeta;
  agentBusy: boolean;
};
