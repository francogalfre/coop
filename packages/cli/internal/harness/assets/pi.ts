const BASE_URL = "{{BASE_URL}}";

type SteerReply = { steer?: string } | null;

type PiEvent = Record<string, unknown>;
type PiContext = { cwd?: string };

interface PiExtensionAPI {
  on(name: string, handler: (event: PiEvent, ctx: PiContext) => Promise<void>): void;
  sendUserMessage(text: string, opts: { deliverAs: "steer" }): void;
}

async function post(event: string, body: unknown): Promise<SteerReply> {
  try {
    const res = await fetch(`${BASE_URL}/hook/pi/${event}`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify(body),
    });
    if (!res.ok) return null;
    return (await res.json().catch(() => null)) as SteerReply;
  } catch {
    return null;
  }
}

export default function (pi: PiExtensionAPI) {
  const maybeSteer = async (reply: SteerReply) => {
    if (reply && typeof reply.steer === "string" && reply.steer) {
      pi.sendUserMessage(reply.steer, { deliverAs: "steer" });
    }
  };

  pi.on("session_start", async (event, ctx) => {
    await maybeSteer(await post("session_start", { ...event, cwd: ctx?.cwd }));
  });
  pi.on("tool_call", async (event) => {
    const reply = await post("tool_call", {
      toolName: event.toolName,
      toolCallId: event.toolCallId,
      input: event.input,
    });
    await maybeSteer(reply);
  });
  pi.on("tool_result", async (event) => {
    await maybeSteer(await post("tool_result", event));
  });
  pi.on("turn_end", async (event) => {
    await maybeSteer(await post("turn_end", event));
  });
  pi.on("agent_end", async (event) => {
    await maybeSteer(await post("agent_end", event));
  });
  pi.on("session_shutdown", async (event) => {
    await post("session_shutdown", event);
  });
}
