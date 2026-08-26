const BASE_URL = "{{BASE_URL}}";

async function post(event, body) {
  try {
    const res = await fetch(`${BASE_URL}/hook/opencode/${event}`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify(body),
    });
    if (!res.ok) return null;
    return await res.json().catch(() => null);
  } catch {
    return null;
  }
}

export const CoopPlugin = async ({ client, directory, worktree }) => {
  const maybeSteer = async (reply) => {
    if (reply && typeof reply.steer === "string" && reply.steer) {
      await client.tui.appendPrompt({ body: { text: reply.steer } });
    }
  };

  return {
    event: async ({ event }) => {
      const reply = await post(event.type, { directory, worktree, event });
      await maybeSteer(reply);
    },
    "tool.execute.before": async (input, output) => {
      const reply = await post("tool.execute.before", {
        directory,
        input,
        args: output.args,
      });
      await maybeSteer(reply);
    },
    "tool.execute.after": async (input, output) => {
      const reply = await post("tool.execute.after", {
        directory,
        input,
        title: output.title,
        out: output.output,
        metadata: output.metadata,
      });
      await maybeSteer(reply);
    },
  };
};
