import type { AgentAskedTeam, HumanAnswered, HumanCommand } from "@coop/protocol";
import type { CommandItem, QuestionItem } from "../../types";
import type { TimelineState } from "./state";
import { readRedacted, textOf } from "./redacted";

type QuestionEvent = AgentAskedTeam | HumanAnswered | HumanCommand;

export function reduceQuestion(state: TimelineState, event: QuestionEvent): void {
  const key = `${event.type}-${event.seq}`;

  switch (event.type) {
    case "agent.asked_team": {
      const text = readRedacted(event.text);
      const item: QuestionItem = {
        kind: "question",
        key,
        seq: event.seq,
        ts: event.ts,
        questionId: event.question_id,
        text: text.text,
        textRedactions: text.redactions,
        textTruncated: text.truncated,
        options: event.options,
        status: "open",
      };
      state.items.push(item);
      state.questionItemIndex.set(event.question_id, state.items.length - 1);
      return;
    }

    case "human.answered": {
      const targetIndex = state.questionItemIndex.get(event.question_id);
      if (targetIndex === undefined) return;
      const target = state.items[targetIndex] as QuestionItem;
      state.items[targetIndex] = {
        ...target,
        status: "answered",
        answeredBy: event.actor.display_name,
        answerText: textOf(event.text),
      };
      return;
    }

    case "human.command": {
      const item: CommandItem = {
        kind: "command",
        key,
        seq: event.seq,
        ts: event.ts,
        author: event.actor.display_name,
        authorAvatarUrl: event.actor.avatar_url,
        command: event.command,
        args: event.args,
      };
      state.items.push(item);
      return;
    }
  }
}
