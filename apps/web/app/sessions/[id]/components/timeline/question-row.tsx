"use client";

import { useState } from "react";
import { toast } from "sonner";
import { IconPeople, IconSend } from "@/components/icons";
import { Button } from "@/components/ui/button";
import { relayConfig } from "@/lib/relay/config";
import type { TimelineItem } from "../../types";
import { Row } from "../timeline-row-shell";
import { RedactionChip } from "./redaction-chip";

// relayApi (lib/relay/api.ts) has no question-answer endpoint yet — this
// posts directly against the shape the relay is expected to expose, so the
// row still typechecks and works once that endpoint lands.
async function answerQuestion(sessionId: string, questionId: string, text: string) {
  const res = await fetch(
    `${relayConfig.httpUrl}/v1/sessions/${encodeURIComponent(sessionId)}/questions/${encodeURIComponent(questionId)}/answer`,
    {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ text }),
    },
  );
  if (!res.ok) throw new Error(`Request failed (${res.status}).`);
}

export function QuestionRow({
  item,
  sessionId,
  isOwner,
}: {
  item: Extract<TimelineItem, { kind: "question" }>;
  sessionId: string;
  isOwner: boolean;
}) {
  const [draft, setDraft] = useState("");
  const [answering, setAnswering] = useState(false);

  async function answer(text: string) {
    if (!text.trim()) return;
    setAnswering(true);
    try {
      await answerQuestion(sessionId, item.questionId, text);
      setDraft("");
    } catch {
      toast.error("Failed to send the answer.");
    } finally {
      setAnswering(false);
    }
  }

  return (
    <Row
      ts={item.ts}
      seq={item.seq}
      rail={
        <span className="relative z-10 grid size-6 place-items-center rounded-md border border-agent/30 bg-agent/10 text-agent">
          <IconPeople size={12} />
        </span>
      }
    >
      <div className="rounded-lg rounded-tl-sm border border-agent/25 bg-agent/[0.06] px-3 py-2">
        <div className="mb-0.5 flex items-center gap-2">
          <span className="font-medium text-xs text-agent">asked the team</span>
          {item.status === "answered" && (
            <span className="ml-auto text-3xs text-muted-foreground">answered by {item.answeredBy}</span>
          )}
        </div>
        <p className="whitespace-pre-wrap text-sm text-foreground/90 leading-relaxed">{item.text}</p>
        <RedactionChip redactions={item.textRedactions} truncated={item.textTruncated} />

        {item.status === "answered" ? (
          <p className="mt-1.5 whitespace-pre-wrap text-sm text-agent/90 leading-relaxed">“{item.answerText}”</p>
        ) : isOwner ? (
          <div className="mt-2 space-y-1.5">
            {item.options && item.options.length > 0 && (
              <div className="flex flex-wrap gap-1.5">
                {item.options.map((option) => (
                  <Button
                    key={option}
                    size="sm"
                    variant="outline"
                    disabled={answering}
                    onClick={() => void answer(option)}
                    className="h-6.5 rounded-md px-2 text-2xs"
                  >
                    {option}
                  </Button>
                ))}
              </div>
            )}
            <div className="flex gap-1.5">
              <input
                value={draft}
                onChange={(e) => setDraft(e.target.value)}
                placeholder="Type an answer…"
                disabled={answering}
                className="h-6.5 min-w-0 flex-1 rounded-md border border-border bg-background px-2 text-2xs text-foreground outline-none focus-visible:ring-[3px] focus-visible:ring-ring/50"
              />
              <Button
                size="sm"
                variant="outline"
                disabled={answering}
                onClick={() => void answer(draft)}
                className="h-6.5 gap-1 rounded-md px-2 text-2xs"
              >
                <IconSend size={11} />
              </Button>
            </div>
          </div>
        ) : null}
      </div>
    </Row>
  );
}
