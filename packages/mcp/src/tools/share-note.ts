import type { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import type { Config } from "../config/config.js";
import { shareNote } from "../relay/notes.js";
import { RelayUnreachableError } from "../relay/types.js";
import { requireProject } from "./require-project.js";
import { shareNoteInputSchema } from "./schemas/share-note.js";

const description = [
  "Leaves a short note for the rest of the team's agents working in this project — a finding, a gotcha,",
  "or context worth surfacing now rather than waiting for someone to ask. Other sessions, including ones",
  "already running, can read it with list_project_notes.",
].join(" ");

export function registerShareNote(server: McpServer, config: Config): void {
  server.registerTool(
    "share_note",
    { title: "Share a note with the team", description, inputSchema: shareNoteInputSchema },
    async ({ text }) => {
      const project = requireProject(config);
      if (!project.ok) {
        return { isError: true, content: [{ type: "text" as const, text: `share_note failed: ${project.message}` }] };
      }

      try {
        const note = await shareNote(config, project.project, text);
        return { content: [{ type: "text" as const, text: "Shared." }], structuredContent: { note } };
      } catch (error) {
        if (error instanceof RelayUnreachableError) {
          return { isError: true, content: [{ type: "text" as const, text: `share_note failed: ${error.message}` }] };
        }

        throw error;
      }
    },
  );
}
