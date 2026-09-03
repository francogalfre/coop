import type { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import type { Config } from "../config/config.js";
import { askTeam, awaitAnswer } from "../relay/questions.js";
import { RelayUnreachableError } from "../relay/types.js";
import { askTeamInputSchema } from "./schemas/ask-team.js";

const defaultTimeoutSeconds = 300;

const description = [
  "Asks the humans watching this coop session a question and waits for one of them to answer.",
  "Use it when a decision is genuinely someone else's to make — an ambiguous requirement, a",
  "product call, domain knowledge you do not have, or a choice that is expensive to undo.",
  "It blocks until someone answers or the timeout elapses; on timeout it returns without an",
  "answer and you should decide for yourself and say what you assumed.",
].join(" ");

function textResult(text: string, structured?: Record<string, unknown>) {
  return { content: [{ type: "text" as const, text }], ...(structured ? { structuredContent: structured } : {}) };
}

export function registerAskTeam(server: McpServer, config: Config): void {
  server.registerTool(
    "ask_team",
    { title: "Ask the team", description, inputSchema: askTeamInputSchema },
    async ({ question, options, timeout_seconds }) => {
      if (!config.sessionId) {
        return {
          isError: true,
          content: [
            {
              type: "text" as const,
              text: "ask_team is unavailable: no coop session. Start one with `coop attach` or `coop run`.",
            },
          ],
        };
      }

      const timeout = timeout_seconds ?? defaultTimeoutSeconds;

      try {
        const questionId = await askTeam(config, config.sessionId, question, options);
        const answer = await awaitAnswer(config, config.sessionId, questionId, timeout);

        if (!answer) {
          return textResult(
            `Nobody answered within ${timeout}s. Decide for yourself and state the assumption you made.`,
            { question_id: questionId, answered: false },
          );
        }

        return textResult(`${answer.author}: ${answer.text}`, {
          question_id: questionId,
          answered: true,
          author: answer.author,
          answer: answer.text,
        });
      } catch (error) {
        if (error instanceof RelayUnreachableError) {
          return { isError: true, content: [{ type: "text" as const, text: `ask_team failed: ${error.message}` }] };
        }

        throw error;
      }
    },
  );
}
