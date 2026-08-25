# coop

Multiplayer sessions for coding agents. One person runs an agent; teammates
watch it live, steer it, and take it over.

## How it works (read this before touching code)

The agent runs on the user's machine. `coop` observes it through the harness's
hook system and streams structured events to a relay. The relay fans those
events out to browsers over WebSocket. Steering goes back the other way —
never by parsing terminal output.

Two capture modes, both required:

- **attach** — user keeps their normal TUI session. We install each detected
  harness's native hook/plugin mechanism, which POSTs (or forwards) events to
  a local server `coop attach` runs, which redacts and forwards to the relay.
  Steering delivery is harness-specific (Claude Code: `additionalContext` on
  a hook response; opencode: `client.tui.appendPrompt`; pi:
  `pi.sendUserMessage(text, {deliverAs:"steer"})`). Blocking is `exit 2` on
  `PreToolUse` for Claude Code; other harnesses have their own veto shape,
  out of scope for now.
- **run** — `coop run -- claude` wraps the agent in a pty. Same hook/plugin
  wiring, plus we own stdin, so steering is immediate for harnesses without a
  native injection primitive.

`packages/cli/internal/harness/` holds one adapter per harness (Claude Code,
opencode, pi today), each implementing detect/install/uninstall. Unsupported
harnesses fall back to pty-only wrapping with degraded event fidelity — see
`.agents/rules/harnesses.md`.

We never parse ANSI to understand what happened. The pty is a keyboard, not a
data source.

## Layout

```
apps/relay/         Go — HTTP hook ingest, WebSocket fanout, permissions
apps/web/           Next.js — the session viewer
packages/protocol/  TypeScript — event schemas (Zod), shared with web
packages/cli/       Go — coop attach / coop run, harness adapters
packages/mcp/       TypeScript — presence + conflict MCP server
.agents/            context, conventions, and rules/ — read these first
```

## Commands

<!-- FILL THESE IN as they become real. An unrunnable command here is worse
     than no command. -->

```
pnpm dev          # relay + web (scripts/dev.sh)
pnpm test         # go test ./... (both modules) && pnpm test
pnpm lint         # oxlint
pnpm typecheck    # tsc --noEmit
```

## Rules

Read the file that matches your task before starting:

- `.agents/rules/security.md` — redaction, steering trust boundaries
- `.agents/rules/protocol.md` — the event schema contract
- `.agents/rules/harnesses.md` — verified harness behavior (hook payloads)
- `.agents/rules/testing.md` — tests before the fix, fixtures from real sessions

## Non-negotiables

1. **Secrets never leave the machine.** Redaction happens in the CLI, before
   an event is sent. Filtering on the relay is too late. See
   `.agents/rules/security.md`.
2. **The event schema is a contract.** Changing `packages/protocol` means
   changing the relay and the web app in the same PR. Never add a field to
   one side only. See `.agents/rules/protocol.md`.
3. **Steering text is untrusted user input.** It is attributed, never framed
   as a system instruction. See `.agents/rules/security.md`.
4. **Verify harness behavior, don't assume it.** Hook field names and the
   transcript JSONL format change between releases. If you're unsure what a
   hook emits, run `scripts/probe.sh` and look. Do not invent field names.
5. **No new dependencies without asking.** Especially in Go — the stdlib
   covers most of what the relay needs.
6. **Tests before the fix.** Reproduce the bug in a test first. See
   `.agents/rules/testing.md`.

## Style

- Go: stdlib first. `gofmt`. Errors wrapped with context, never swallowed.
- TypeScript: strict mode, no `any`, Zod for anything crossing a boundary.
- Commits, file size, and function rules: `.agents/conventions.md`.
- Comments explain *why*, never *what*.

## When you're stuck

Say so and stop. A wrong guess about how a harness behaves costs more than a
question. Do not fabricate a hook event name, a JSON field, or a CLI flag —
if it isn't in `.agents/rules/harnesses.md` or verified output, it doesn't
exist yet.
