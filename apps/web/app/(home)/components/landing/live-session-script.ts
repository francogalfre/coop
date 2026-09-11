export type LiveRow =
  | { kind: "agent"; text: string }
  | { kind: "tool"; name: string; detail: string }
  | { kind: "file"; path: string; added: number; removed: number }
  | { kind: "human"; who: string; text: string }
  | { kind: "blocked"; name: string; detail: string };

export const liveSessionScript: LiveRow[] = [
  { kind: "agent", text: "Reading the checkout flow to find where the retry loop starts." },
  { kind: "tool", name: "grep", detail: "retryPayment --glob '*.ts'" },
  { kind: "agent", text: "Found it. `retryPayment` recurses without a ceiling on 5xx." },
  { kind: "file", path: "src/payments/retry.ts", added: 14, removed: 6 },
  {
    kind: "human",
    who: "Mara",
    text: "Cap it at 3 and make the delay exponential — legal needs the audit log too.",
  },
  { kind: "agent", text: "Capping at 3 with exponential backoff, and logging each attempt." },
  { kind: "file", path: "src/payments/retry.ts", added: 9, removed: 2 },
  { kind: "blocked", name: "bash", detail: "npm run deploy:staging" },
];

export const liveSessionViewers = ["Mara", "Tom", "Ana"];
