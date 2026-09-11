# Architecture

How relay, web, CLI, protocol, and MCP fit together. Concepts live in
[overview.md](overview.md); commands in [cli.md](cli.md).

## The pieces

| Piece | Language | What it is |
| --- | --- | --- |
| `apps/relay` | Go | HTTP ingest, WebSocket fanout, steering mailbox, permissions, Postgres persistence |
| `apps/web` | Next.js | The session viewer: timeline, steering UI, projects, auth (better-auth) |
| `packages/protocol` | TypeScript | Zod event schemas shared by relay and web — the product's contract |
| `packages/cli` | Go | `coop attach` / `run` / `detach` / `login`, harness adapters, redaction |
| `packages/mcp` | TypeScript | MCP server (`coop-mcp`) the agent can call: ask the team, check conflicts, notes |

## Event lifecycle: harness hook to browser

```text
coding agent (your machine)
  │ harness fires a hook / plugin event
  ▼
coop CLI — local hook server on 127.0.0.1:8788
  │  translates the harness payload into a protocol event
  │  redacts every text field, stamps the envelope (v, session_id, seq, ts)
  ▼
POST /v1/events  (Authorization: Bearer <cli credential>, X-Coop-Project: <slug>)
  ▼
relay
  │  authenticates the CLI credential, checks session ownership
  │  persists the event to Postgres with a per-session seq
  │  pushes it to every live subscriber
  ▼
GET /v1/sessions/{id}/stream  (WebSocket)
  │  backfills up to the last 200 events from Postgres first
  ▼
browsers of every project member watching the session
```

The CLI never parses terminal output to figure out what happened. Each
harness adapter (`packages/cli/internal/harness/<name>/`) installs that
harness's native mechanism — HTTP hook entries in
`.claude/settings.local.json` for Claude Code, a plugin at
`.opencode/plugin/coop.js` for opencode, an extension at
`.pi/extensions/coop.ts` for pi — and each POSTs to the CLI's local hook
server at `/hook/{harness}/{event}`. The server translates, redacts, and
forwards to the relay. Events are queued and posted asynchronously; a full
queue drops the oldest event rather than ever blocking a hook response.

Under `coop run` there is an additional stream: the relay also hosts the pty
itself. The CLI connects as the pty *source* over
`GET /v1/sessions/{id}/pty` (owner credential required), streams redacted
terminal output, and receives viewer keystrokes when takeover is active.

## Steering lifecycle: browser to agent

```text
browser ── POST /v1/sessions/{id}/steer ──► relay mailbox
                                             │ (restricted mode + non-owner:
                                             │  becomes a pending request the
                                             │  owner must approve)
                                             ▼
coop CLI ── GET /v1/sessions/{id}/steer (owner) ── take-once, on each hook (attach)
                                             or every 1.5s (run)
  ▼ delivered the way this harness expects
attach:  Claude Code → additionalContext on the hook response
         opencode    → client.tui.appendPrompt
         pi         → pi.sendUserMessage(text, { deliverAs: "steer" })
run:     written as keystrokes into the pty
  ▼
relay records steer.delivered (or steer.dropped) so everyone sees the receipt
```

Messages are attributed (`[Alice via coop] ...`) and can never grant
permissions. Harness commands (`/model`, `/compact`, ...) are
allowlist-checked on the relay *and* re-validated in the CLI, and only
delivered in run mode — attach mode has no pty to type into, so the CLI
drops them.

Takeover is a relay-side claim (`POST /v1/sessions/{id}/takeover`): one
holder at a time, released by the claimant or the session owner. In run
mode, viewer keystrokes are routed to the claimant's pty through the pty
hub; in attach mode with Claude Code, takeover additionally answers
`PreToolUse` hooks with a deny until released.

The agent can also ask questions through the MCP server: `ask_team` posts
to the relay's question registry, the web app shows the question, a member
answers it, and the MCP server (long-polling) returns the attributed answer
to the agent.

## Authentication

Three identities meet at the relay:

- **Browser sessions.** The web app authenticates with GitHub through
  better-auth. The relay does not run its own browser login: it reads the
  better-auth session cookie and verifies it by calling the web app's
  internal endpoint (`/api/internal/session/verify`, authenticated with
  `COOP_INTERNAL_SECRET`), caching the result for 20 seconds.
- **CLI credentials.** `coop login` runs GitHub's device flow and exchanges
  the token at `POST /v1/auth/cli/exchange`. The relay stores only a hash;
  the CLI keeps the token in `~/.config/coop/credentials.json`. Credentials
  can be revoked (`POST /v1/auth/cli/revoke`).
- **Access control.** Every session route requires being a member of the
  session's project; a smaller set (mode changes, steer resolution,
  permission resolution, reading the steer mailbox) requires being the
  session owner. Unknown users get a 404, not a 403 — existence is not
  leaked.

CORS and WebSocket origins are restricted to `COOP_WEB_ORIGINS`.

## Redaction boundary

Redaction happens in the CLI, before an event is sent — never on the relay,
where the bytes have already left the machine. The same patterns are applied
to pty output (line-buffered, so a secret split across writes is still
caught). Every redacted text field carries a count of what was masked, so
viewers can see that something was hidden even though they cannot see what.

## The event schema is a contract

`packages/protocol` is the single source of truth for event shapes. Adding a
field means updating the relay and the web app in the same pull request — a
field is never added to one side only. Known event types:

- session: `session.start`, `session.end`, `session.mode_changed`
- agent: `agent.turn_start`, `agent.text`, `agent.turn_end`,
  `agent.asked_team`
- tool: `tool.call`, `tool.result`, `tool.blocked`
- files: `file.touched` (mode read/write, optional diff hunks)
- permissions: `permission.requested`, `permission.resolved`
- humans: `human.join`, `human.leave`, `human.steer`, `human.takeover`,
  `human.prompt`, `human.message`, `human.command`, `human.answered`
- steering receipts: `steer.requested`, `steer.resolved`,
  `steer.delivered`, `steer.dropped`

Unknown types parse as a loose `unknown` event, so a newer CLI does not
break an older deployment. Harness hook payloads themselves change between
releases; they are verified against real output (`scripts/probe.sh`), never
assumed.

## The relay is single-instance

The relay keeps live state in memory — the event store, steering mailbox,
presence and viewer hubs, takeover, pty, steer-request and question
registries. It must run as exactly one replica and never be scaled
horizontally; the compose files say so in comments. Postgres holds the
durable state: users' CLI credentials, projects, members, invites, sessions,
steer requests, takeover state, notes, and the full event log (per-session
sequence numbers) for replay and backfill.

The relay migrates its own schema on boot (Ent `Schema.Create`). Sessions
idle for 30 minutes are swept and ended every 5 minutes, so a crashed attach
does not leave a zombie session.

## HTTP surface

The relay's routes, grouped by who may call them:

| Route | Who | What |
| --- | --- | --- |
| `GET /healthz` | anyone | liveness + database ping |
| `POST /v1/events` | CLI credential | event ingest (rate-limited) |
| `POST /v1/auth/cli/exchange` / `revoke` | GitHub token / CLI credential | login / logout |
| `GET /v1/presence`, `GET /v1/sessions` | any identity | live sessions and file activity |
| `GET /v1/sessions/{id}/stream` | project member | event WebSocket |
| `GET /v1/sessions/{id}/pty` | member (owner connects as source) | pty WebSocket |
| `GET /v1/sessions/{id}/viewers` | project member | who is watching |
| `POST /v1/sessions/{id}/steer` | project member | send a steer message |
| `POST /v1/sessions/{id}/message` | project member | team chat message |
| `POST /v1/sessions/{id}/takeover` | project member | claim/release takeover |
| `POST /v1/sessions/{id}/command` | session owner | queue an allowlisted harness command |
| `POST /v1/sessions/{id}/questions` (+ `/answer`) | project member | ask the team (agent via MCP) / answer a question |
| `GET /v1/sessions/{id}/steer` (+ `/resolve`) | session owner | mailbox poll / restricted-mode approval |
| `POST /v1/sessions/{id}/mode` | session owner | auto ↔ restricted |
| `POST /v1/sessions/{id}/permissions/{id}/resolve` | session owner | allow/deny a permission prompt |
| `GET /v1/sessions/{id}/events` | project member | replay events |
| `POST/GET /v1/projects...` | any identity | projects, invites, members, notes, context, per-project agents |
