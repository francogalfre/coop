# Architecture

## The pieces

```
apps/relay/         Go     — HTTP hook ingest, WebSocket fanout, permissions, Postgres
apps/web/           Next.js — the session viewer
packages/protocol/  TS      — event schemas (Zod), shared by relay and web
packages/cli/       Go      — coop attach / run / detach, harness adapters
packages/mcp/       TS      — presence + conflict MCP server, runs next to the agent
```

## Event flow

```
agent (your machine)
  │  harness fires a hook
  ▼
coop CLI ──► redact ──► POST /v1/events ──► relay ──► Postgres
                                             │
                                             ▼  WebSocket
                                          browsers

browser ── steer text ──► relay ── mailbox ──► CLI ── injected into the harness
```

The CLI polls the relay for steering and delivers it the way each harness
expects: `additionalContext` on a hook response for Claude Code,
`client.tui.appendPrompt` for opencode, `pi.sendUserMessage(text, {deliverAs:
"steer"})` for pi. Blocking a tool call is `exit 2` on `PreToolUse` for Claude
Code.

## The relay is single-instance

The relay keeps several things in memory — the event store, the steering
mailbox, presence, takeover state, the pty hub, pending steer/permission
requests. It must run as exactly one replica and never be scaled horizontally.
The provided compose files enforce this.

Postgres holds the durable state: accounts, projects, sessions, and the full
event log for replay.

## Trust boundaries

- **Redaction is on send, not receive.** By the time an event reaches the
  relay it is already safe to store and broadcast. Filtering on the relay
  would be too late — the bytes have left the machine.
- **Steering text is untrusted input.** It is always attributed
  (`[name via coop]`), never framed as a system message, and can't grant
  permissions. Approving a tool call is a separate, explicit action.
- **Session identity is the better-auth user id.** The CLI resolves it once at
  login and stamps every event with it, so the web and relay agree on who owns
  a session.

## The event schema is a contract

`packages/protocol` is the single source of truth for event shapes. Changing
it means changing the relay and the web app in the same pull request — a field
is never added to one side only. The CLI writes events by hand against the
same schema.

Harness hook payloads change between releases, so the schema is verified
against real output (`scripts/probe.sh`), never assumed. See
`.agents/rules/` for the verified behavior of each harness.

## Sessions

A session is created by the first `session.start` event and ends on
`session.end`. Because `coop attach` on some harnesses has no reliable
session-end signal, the relay also sweeps sessions that have been idle for 30
minutes and marks them ended, and `coop detach` ends the session it left open.
