# Overview

coop makes an agent session something a team shares. This page explains the
concepts — projects, sessions, events, steering — before the mechanics. For
how the pieces are wired together, see [architecture.md](architecture.md);
for commands, see [cli.md](cli.md).

## Projects

A project is the unit of membership. Every session belongs to exactly one
project (`coop attach --project=<slug>`), and the people who can watch,
message, and steer a session are the project's members. Projects live in the
web app: you create one, invite teammates (invite links are valid for seven
days), and every member sees the project's sessions and notes. Members have
two roles, `owner` and `member`; the session owner is a separate thing,
below.

Projects also carry a shared **context** — a block of text the team can edit.
When you attach, the current project context is delivered to your session as
an attributed message, so the agent starts with the team's standing
instructions.

## Sessions

A session is one run of one coding agent, observed from the outside. It is
born when the agent fires its first `SessionStart` hook (or, under `coop
run`, when the wrapped process starts) and ends on `session.end`, when you
Ctrl-C the attach, or when the relay sweeps it after 30 minutes of
inactivity.

The person who started the session is its **owner**. Ownership matters: only
the owner can change the session mode, resolve a permission prompt, approve
steering requests in restricted mode, or pull pending steering messages out
of the mailbox. Everyone else on the project can watch, message, steer, ask
and answer questions, and claim takeover.

Sessions run in one of two modes:

- **auto** — steering from any member reaches the agent directly.
- **restricted** — steering from anyone but the owner becomes a pending
  request; the owner approves or denies each one. Only the owner can switch
  modes.

## Events

Everything the team sees is an event: a tool call and its result, a file
touched (with diff hunks), the agent's text, a human joining, a steer
arriving, a permission being requested or resolved. Events are structured
JSON defined by a single schema package (`packages/protocol`) — the CLI
writes them, the relay stores and fans them out, the web app renders them.
Unknown event types degrade gracefully, so a newer CLI talking to an older
relay keeps working.

Two things every event carries: an **envelope** (`v`, `session_id`, `seq`,
`ts`) and **redaction counts** — any text field reports how many secrets
were masked in it before it left the sender's machine.

## Steering

Steering is how the team talks into a session that someone else's terminal
owns. There are four distinct mechanisms, with different trust properties:

- **Message (steer)** — free text, up to 4096 characters, attributed. It
  reaches the agent as `[Alice via coop] ...` — context from a named person,
  never framed as a system instruction, and it can never approve a tool call
  or change a permission. The agent picks it up on its next hook (attach) or
  within about a second and a half (run).
- **Harness command** — an allowlisted slash command (`model`, `compact`,
  `clear`, `context`, `cost`, `status`), delivered only in `coop run` mode,
  typed into the pty exactly as a human would type it. This is keystroke
  injection, which is why the allowlist is the whole contract — the CLI
  re-validates it locally rather than trusting the relay.
- **Takeover** — a member claims the session; the owner's terminal shows who
  took over, and in run mode the claimant's keystrokes are routed to the
  agent while everyone watches the same pty output. Only one person can hold
  takeover at a time, and only the claimant or the session owner can release
  it. In attach mode with Claude Code, takeover also vetoes tool calls
  (`PreToolUse` deny) until released.
- **Ask the team** — the agent itself can ask a question (through the MCP
  server's `ask_team` tool); it appears in the web app, any member answers,
  and the answer is delivered back to the agent as attributed text.

## The trust boundary

Three rules hold the whole design together:

1. **Secrets never leave the machine.** Redaction runs in the CLI, on every
   text field of every event and on pty output, before anything is sent. The
   relay never sees the original bytes; filtering there would be too late.
   Patterns cover private keys, connection strings, `Authorization` and
   `Cookie` headers, bearer tokens, known vendor API keys, `key = value`
   assignments, and JWTs — all replaced with `[redacted]`.
2. **Steering is untrusted input.** It is always attributed, never disguised
   as a system message, and carries no authority. Approving a tool call is a
   separate, explicit action that only the session owner can take.
3. **The owner stays in control.** Mode changes, permission resolutions,
   steering approval in restricted mode, and takeover release are
   owner-only. Everyone else's influence is limited to the four mechanisms
   above, exactly as attributed.

## Two ways to capture a session

- **`coop attach`** — you keep your normal terminal and TUI. coop installs
  each detected harness's native hook or plugin, which forwards events to a
  local server the CLI runs on `127.0.0.1:8788`. Steering is delivered
  through each harness's own injection primitive (Claude Code: hook
  `additionalContext`; opencode: `client.tui.appendPrompt`; pi:
  `pi.sendUserMessage(text, { deliverAs: "steer" })`).
- **`coop run -- <cmd>`** — coop wraps the agent in a pseudo-terminal. Same
  hook wiring, plus coop owns stdin, so steering and commands are plain
  keystrokes, delivered immediately, for any harness — including ones with no
  adapter, which get pty-only capture with reduced event detail.

Both modes are first-class; attach exists because people want their own
terminal, run exists because coop cannot type into a terminal it doesn't
own.

## Multiple sessions in one directory

Claude Code supports several concurrent `coop attach` sessions in the same
directory — each attach keeps its own hook entries and receives its own
events. opencode and pi support only one: they load a single plugin file,
and a second attach would silently steal the first one's events, so the CLI
refuses it outright. See [cli.md](cli.md#multiple-sessions-in-one-directory).
