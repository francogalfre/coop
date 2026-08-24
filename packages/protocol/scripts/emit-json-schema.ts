import { writeFileSync } from "node:fs";
import { eventsJsonSchema } from "../dist/json-schema.js";

const outPath = new URL("../dist/events.schema.json", import.meta.url);
writeFileSync(outPath, JSON.stringify(eventsJsonSchema(), null, 2) + "\n");
console.log(`wrote ${outPath.pathname}`);
