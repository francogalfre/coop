import { KNOWN_EVENT_TYPES } from "@coop/protocol";
import type {
  Event,
  SessionStart,
  SessionEnd,
  AgentTurnStart,
  AgentText,
  AgentTurnEnd,
  ToolCall,
  ToolResult,
  ToolBlocked,
  FileTouched,
  PermissionRequested,
  PermissionResolved,
  HumanJoin,
  HumanLeave,
  HumanSteer,
  HumanTakeover,
  HumanPrompt,
  UnknownEvent,
} from "@coop/protocol";

type KnownEvent =
  | SessionStart
  | SessionEnd
  | AgentTurnStart
  | AgentText
  | AgentTurnEnd
  | ToolCall
  | ToolResult
  | ToolBlocked
  | FileTouched
  | PermissionRequested
  | PermissionResolved
  | HumanJoin
  | HumanLeave
  | HumanSteer
  | HumanTakeover
  | HumanPrompt;

function isKnownEvent(event: Event): event is KnownEvent {
  return (KNOWN_EVENT_TYPES as readonly string[]).includes(event.type);
}

function withRedactionNote(text: { text: string; redactions: number; truncated: boolean }): string {
  const notes: string[] = [];
  if (text.redactions > 0) notes.push(`${text.redactions} redacted`);
  if (text.truncated) notes.push("truncated");
  return notes.length > 0 ? `${text.text} (${notes.join(", ")})` : text.text;
}

function summarizeSession(event: SessionStart | SessionEnd): string {
  if (event.type === "session.start") {
    return `${event.owner.display_name} started a ${event.harness} session in ${event.cwd}`;
  }
  return `session ended${event.reason ? ` (${event.reason})` : ""}`;
}

function summarizeAgent(event: AgentTurnStart | AgentText | AgentTurnEnd): string {
  switch (event.type) {
    case "agent.turn_start":
      return "agent turn started";
    case "agent.text":
      return withRedactionNote(event.text);
    case "agent.turn_end":
      return `agent turn ended${event.duration_ms !== undefined ? ` (${event.duration_ms}ms)` : ""}`;
  }
}

function summarizeTool(event: ToolCall | ToolResult | ToolBlocked): string {
  switch (event.type) {
    case "tool.call":
      return `${event.tool_name} called: ${withRedactionNote(event.input)}`;
    case "tool.result":
      return `${event.tool_name} ${event.ok === false ? "failed" : "returned"}: ${withRedactionNote(event.output)}`;
    case "tool.blocked":
      return `${event.tool_name} blocked by ${event.blocked_by.display_name}: ${withRedactionNote(event.reason)}`;
  }
}

function summarizeFile(event: FileTouched): string {
  return `${event.mode === "write" ? "wrote" : "read"} ${event.path}`;
}

function summarizePermission(event: PermissionRequested | PermissionResolved): string {
  switch (event.type) {
    case "permission.requested":
      return `permission requested for ${event.tool_name}`;
    case "permission.resolved":
      return `${event.decision} by ${event.resolved_by.display_name}${event.reason ? `: ${withRedactionNote(event.reason)}` : ""}`;
  }
}

function summarizeHuman(event: HumanJoin | HumanLeave | HumanSteer | HumanTakeover | HumanPrompt): string {
  switch (event.type) {
    case "human.join":
      return `${event.actor.display_name} joined`;
    case "human.leave":
      return `${event.actor.display_name} left`;
    case "human.steer":
      return `${event.actor.display_name} steered: "${withRedactionNote(event.text)}"`;
    case "human.takeover":
      return `${event.actor.display_name} took over`;
    case "human.prompt":
      return `prompt: "${withRedactionNote(event.text)}"`;
  }
}

function summarizeUnknown(event: UnknownEvent): string {
  return `unrecognized event: ${event.type}`;
}

function summarize(event: Event): string {
  if (!isKnownEvent(event)) return summarizeUnknown(event);

  switch (event.type) {
    case "session.start":
    case "session.end":
      return summarizeSession(event);
    case "agent.turn_start":
    case "agent.text":
    case "agent.turn_end":
      return summarizeAgent(event);
    case "tool.call":
    case "tool.result":
    case "tool.blocked":
      return summarizeTool(event);
    case "file.touched":
      return summarizeFile(event);
    case "permission.requested":
    case "permission.resolved":
      return summarizePermission(event);
    case "human.join":
    case "human.leave":
    case "human.steer":
    case "human.takeover":
    case "human.prompt":
      return summarizeHuman(event);
  }
}

function familyOf(type: string): string {
  return type.split(".")[0] ?? "unknown";
}

const FAMILY_COLORS: Record<string, string> = {
  session: "#6b7280",
  agent: "#2563eb",
  tool: "#7c3aed",
  file: "#059669",
  permission: "#d97706",
  human: "#dc2626",
};

function formatTime(ts: string): string {
  const date = new Date(ts);
  return Number.isNaN(date.getTime()) ? ts : date.toLocaleTimeString();
}

export function EventFeed({ events }: { events: Event[] }) {
  if (events.length === 0) {
    return <p>No events yet.</p>;
  }

  return (
    <ol style={{ listStyle: "none", padding: 0, margin: 0 }}>
      {events.map((event) => {
        const family = familyOf(event.type);
        const color = FAMILY_COLORS[family] ?? "#374151";
        return (
          <li
            key={`${event.session_id}-${event.seq}`}
            style={{ padding: "0.4rem 0", borderBottom: "1px solid #e5e7eb", fontSize: "0.9rem" }}
          >
            <span style={{ fontFamily: "monospace", fontSize: "0.75rem", color: "#6b7280" }}>
              {formatTime(event.ts)}
            </span>{" "}
            <span
              style={{
                fontSize: "0.7rem",
                fontWeight: 600,
                color: "white",
                background: color,
                borderRadius: "3px",
                padding: "0.05rem 0.35rem",
              }}
            >
              {event.type}
            </span>{" "}
            <span>{summarize(event)}</span>
          </li>
        );
      })}
    </ol>
  );
}
