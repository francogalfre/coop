import { IconLock, IconMessage, IconTerminal, IconUnlock } from "@/components/icons";
import { DocPageHeader } from "../components/doc-page-header";

const MODES = [
  {
    title: "attach — keep your own terminal",
    body: "You keep your normal TUI session running exactly as before. coop attach detects the harness in your working directory (Claude Code, opencode, or pi) and installs that harness's native hook or plugin mechanism, which POSTs events to a small local server coop attach runs. That server redacts secrets and forwards everything to the relay. Steering delivery is harness-specific: Claude Code receives it as additionalContext on a hook response, opencode via client.tui.appendPrompt, pi via pi.sendUserMessage(text, {deliverAs: \"steer\"}). Blocking a tool call is exit 2 on Claude Code's PreToolUse hook.",
  },
  {
    title: "run — coop drives the pty",
    body: "coop run -- claude wraps the agent in a pseudo-terminal that coop owns. It installs the same hook wiring as attach, but because coop also owns stdin, steering is immediate even for harnesses without a native injection primitive.",
  },
];

const ACTIONS = [
  {
    icon: IconTerminal,
    title: "Watch",
    body: "Every teammate with the link sees the same event stream, live, as the relay fans it out over WebSocket.",
  },
  {
    icon: IconMessage,
    title: "Redirect",
    body: "Send a message and choose whether it goes to the whole team or straight to the agent. Messages sent to the agent are attributed — the agent sees who sent it, never framed as a system instruction.",
  },
  {
    icon: IconUnlock,
    title: "Hand off",
    body: 'Click "Take over" on a live session and the agent pauses. The composer switches to team-only for everyone else until the person holding it clicks "Release."',
  },
];

export default function DocsHowItWorks() {
  return (
    <>
      <DocPageHeader
        eyebrow="How it Works"
        title="Two people, one terminal"
        intro="coop doesn't read your screen. It hooks into your agent's own harness and streams structured events out — and sends steering back through that same door."
      />

      <section>
        <h2 className="font-display font-semibold text-foreground text-lg">Two capture modes</h2>
        <p className="mt-2 max-w-2xl text-[14px] text-muted-foreground leading-relaxed">
          Both are fully supported today. Pick attach if you already live in your agent's terminal; pick run if
          you'd rather let coop launch it for you.
        </p>
        <div className="mt-4 grid gap-4 sm:grid-cols-2">
          {MODES.map((mode) => (
            <div key={mode.title} className="rounded-xl border border-border bg-card/50 p-5">
              <p className="font-display font-medium text-foreground text-md">{mode.title}</p>
              <p className="mt-2 text-[13px] text-muted-foreground leading-relaxed">{mode.body}</p>
            </div>
          ))}
        </div>
        <p className="mt-4 max-w-2xl text-[13px] text-muted-foreground leading-relaxed">
          Harnesses without an adapter fall back to plain pty wrapping, with reduced event detail — coop still
          sees keystrokes and output timing, just not structured tool calls.
        </p>
      </section>

      <section className="mt-10">
        <h2 className="font-display font-semibold text-foreground text-lg">The pty is a keyboard, not a data source</h2>
        <p className="mt-2 max-w-2xl text-[14px] text-muted-foreground leading-relaxed">
          coop never parses ANSI output to figure out what the agent did. Every event you see in the viewer — a
          tool call, a file edit, a message — comes from the harness's own hook payloads, never from scraping
          the terminal. That's also why blocking and steering flow through the hook/injection layer instead of
          typing into the pty on your behalf.
        </p>
      </section>

      <section className="mt-10">
        <h2 className="font-display font-semibold text-foreground text-lg">Watch, redirect, hand off</h2>
        <div className="mt-4 grid gap-3 sm:grid-cols-3">
          {ACTIONS.map(({ icon: Icon, title, body }) => (
            <div key={title} className="rounded-xl border border-border bg-card/50 p-5">
              <Icon size={18} className="text-muted-foreground" />
              <p className="mt-3 font-display font-medium text-foreground text-md">{title}</p>
              <p className="mt-1 text-[13px] text-muted-foreground leading-relaxed">{body}</p>
            </div>
          ))}
        </div>
      </section>

      <section className="mt-10">
        <h2 className="font-display font-semibold text-foreground text-lg">Auto vs. Restricted approval</h2>
        <p className="mt-2 max-w-2xl text-[14px] text-muted-foreground leading-relaxed">
          The session owner can flip a live session between two modes with the toggle in the session header:
        </p>
        <div className="mt-4 grid gap-4 sm:grid-cols-2">
          <div className="rounded-xl border border-border bg-card/50 p-5">
            <div className="flex items-center gap-2">
              <IconUnlock size={14} className="text-muted-foreground" />
              <p className="font-display font-medium text-foreground text-md">Auto</p>
            </div>
            <p className="mt-2 text-[13px] text-muted-foreground leading-relaxed">
              A teammate's message reaches the agent immediately.
            </p>
          </div>
          <div className="rounded-xl border border-border bg-card/50 p-5">
            <div className="flex items-center gap-2">
              <IconLock size={14} className="text-muted-foreground" />
              <p className="font-display font-medium text-foreground text-md">Restricted</p>
            </div>
            <p className="mt-2 text-[13px] text-muted-foreground leading-relaxed">
              A teammate's message waits in the timeline as a pending request, marked "awaiting approval," until
              the owner approves or denies it. Denied requests stay visible, grayed out, with who denied them.
            </p>
          </div>
        </div>
      </section>
    </>
  );
}
