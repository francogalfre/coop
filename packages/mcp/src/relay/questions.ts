import type { Config } from "../config/config.js";
import { getJson, postJson } from "./request.js";
import type { Answer } from "./types.js";

type RawAskResponse = { question_id: string };

type RawQuestionState = {
  status: "open" | "answered" | "expired";
  answer?: { text: string; actor: { display_name: string } };
};

const pollWindowSeconds = 25;

export async function askTeam(
  config: Config,
  sessionId: string,
  question: string,
  options?: string[],
): Promise<string> {
  const url = new URL(`/v1/sessions/${encodeURIComponent(sessionId)}/questions`, config.relayUrl);
  const raw = (await postJson(url, config, { text: question, options })) as RawAskResponse;

  return raw.question_id;
}

export async function awaitAnswer(
  config: Config,
  sessionId: string,
  questionId: string,
  timeoutSeconds: number,
): Promise<Answer | null> {
  const deadline = Date.now() + timeoutSeconds * 1000;

  while (Date.now() < deadline) {
    const remaining = Math.ceil((deadline - Date.now()) / 1000);
    const wait = Math.min(pollWindowSeconds, remaining);
    const url = new URL(
      `/v1/sessions/${encodeURIComponent(sessionId)}/questions/${encodeURIComponent(questionId)}?wait_seconds=${wait}`,
      config.relayUrl,
    );

    // eslint-disable-next-line no-await-in-loop -- each poll depends on the previous one's elapsed time, genuinely sequential
    const state = (await getJson(url, config)) as RawQuestionState;

    if (state.status === "answered" && state.answer) {
      return { text: state.answer.text, author: state.answer.actor.display_name };
    }

    if (state.status === "expired") return null;
  }

  return null;
}
