import type { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import type { Config } from "../config/config.js";
import { listNotes } from "../relay/notes.js";
import { RelayUnreachableError } from "../relay/types.js";
import { requireProject } from "./require-project.js";
import { listProjectNotesInputSchema } from "./schemas/share-note.js";

const description = [
  "Reads recent notes other agents and teammates have left for this project, newest first — findings,",
  "gotchas, or context posted since (or before) this session started. Call it when starting work, or when",
  "something feels like it might have already been figured out by someone else.",
].join(" ");

export function registerListProjectNotes(server: McpServer, config: Config): void {
  server.registerTool(
    "list_project_notes",
    { title: "List project notes", description, inputSchema: listProjectNotesInputSchema },
    async ({ limit }) => {
      const project = requireProject(config);
      if (!project.ok) {
        return {
          isError: true,
          content: [{ type: "text" as const, text: `list_project_notes failed: ${project.message}` }],
        };
      }

      try {
        const notes = await listNotes(config, project.project, limit);
        const text = notes.length === 0 ? "No notes yet." : `${notes.length} note(s).`;
        return { content: [{ type: "text" as const, text }], structuredContent: { notes } };
      } catch (error) {
        if (error instanceof RelayUnreachableError) {
          return {
            isError: true,
            content: [{ type: "text" as const, text: `list_project_notes failed: ${error.message}` }],
          };
        }

        throw error;
      }
    },
  );
}
