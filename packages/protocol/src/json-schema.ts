import { z } from "zod";
import { knownEvent } from "./event.js";

export function eventsJsonSchema(): Record<string, unknown> {
  const schema = z.toJSONSchema(knownEvent, { target: "draft-2020-12", io: "input" });
  return {
    $id: "https://coop.dev/schemas/events.json",
    title: "coop protocol v1 events",
    ...schema,
  };
}
