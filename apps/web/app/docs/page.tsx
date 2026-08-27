import Link from "next/link";
import type { Route } from "next";
import { IconChevronRight } from "@/components/icons";
import { DocPageHeader } from "./components/doc-page-header";
import { CodeBlock } from "./components/code-block";

const STEPS = [
  {
    title: "Build the CLI",
    body: "There's no published binary yet — build coop from source. You'll need Go 1.26+.",
    code: "cd packages/cli\ngo build -o coop ./cmd/coop",
    note: "Run the resulting binary from wherever you cloned coop, e.g. ../coop/packages/cli/coop, or put it on your PATH.",
  },
  {
    title: "Log in",
    body: "coop login authenticates you with GitHub's device flow: it prints a URL and a code, waits for you to approve it in the browser, then exchanges the GitHub token for a coop credential and saves it locally.",
    code: "coop login",
    note: 'Prints something like: "Go to https://github.com/login/device and enter code: XXXX-XXXX", then "Logged in as <your-username>" once you approve it.',
  },
  {
    title: "Run your agent inside coop",
    body: "Every session belongs to a project. Wrap your agent's command with coop run and pass the project slug it belongs to:",
    code: "coop run --project=my-app -- claude",
    note: "Leave off -- <cmd> and coop run launches claude by default. coop owns stdin, so steering lands immediately no matter which harness you're running.",
  },
  {
    title: "Open the session",
    body: "Head to your project in the web app — your session shows up live within seconds of the agent starting.",
    code: "open http://localhost:3000/projects/my-app",
    note: "From there, open the session to watch its event stream, send it a message, or take over the keyboard.",
  },
];

export default function DocsGettingStarted() {
  return (
    <>
      <DocPageHeader
        eyebrow="Getting Started"
        title="Run your first shared session"
        intro="coop wraps the coding agent you already use so anyone on the team can watch it work, send it a message mid-task, or take over the keyboard — all from a link."
      />

      <ol className="space-y-4">
        {STEPS.map((step, i) => (
          <li key={step.title} className="rounded-xl border border-border bg-card/50 p-6">
            <div className="flex items-baseline gap-3">
              <span className="font-display font-semibold text-[13px] text-muted-foreground">
                {String(i + 1).padStart(2, "0")}
              </span>
              <h2 className="font-display font-semibold text-foreground text-lg">{step.title}</h2>
            </div>
            <p className="mt-2 text-[14px] text-muted-foreground leading-relaxed">{step.body}</p>
            <CodeBlock>{step.code}</CodeBlock>
            <p className="text-[13px] text-muted-foreground leading-relaxed">{step.note}</p>
          </li>
        ))}
      </ol>

      <div className="mt-8 flex flex-wrap gap-3">
        <Link
          href={"/docs/cli" as Route}
          className="inline-flex items-center gap-1 rounded-lg border border-border bg-card/50 px-3.5 py-2 text-[13px] text-foreground transition-colors hover:bg-card"
        >
          CLI reference
          <IconChevronRight size={13} />
        </Link>
        <Link
          href={"/docs/how-it-works" as Route}
          className="inline-flex items-center gap-1 rounded-lg border border-border bg-card/50 px-3.5 py-2 text-[13px] text-foreground transition-colors hover:bg-card"
        >
          How it works
          <IconChevronRight size={13} />
        </Link>
      </div>
    </>
  );
}
