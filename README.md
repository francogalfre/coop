# coop

Multiplayer sessions for coding agents. Watch a live agent work, redirect it mid-task, and hand off control — the way you'd work with any other teammate.

One person runs a coding agent. Teammates open a link and watch it work live, send it a message the agent sees mid-task, or take over completely and drive. Agents see what other sessions are touching, so they don't collide.

Works with Claude Code, opencode, and pi today — `coop attach` detects which
harness you're using and installs the right adapter automatically. Anything
else falls back to plain pty wrapping (`coop run -- <cmd>`), with reduced
event detail.

## Quick Start

Requires a Postgres connection string (`DATABASE_URL`) and a GitHub OAuth app
(`COOP_GITHUB_CLIENT_ID`/`COOP_GITHUB_CLIENT_SECRET`, plus `COOP_INTERNAL_SECRET`
and `BETTER_AUTH_SECRET`). Copy `apps/web/.env.example` to `apps/web/.env.local`
and fill in the values.

**Terminal 1: Start relay + web viewer**

```bash
# from the repo root
bun run dev
```

**Terminal 2: Build the CLI and use it**

```bash
cd packages/cli
go build -o coop ./cmd/coop

# Now use it in your project
cd /your/test/project
coop attach
# coop detects claude / opencode / pi automatically and installs hooks for each.
# In another terminal, start your agent as usual: claude, opencode, or pi

# Or let coop launch it directly and wrap the terminal:
coop run -- claude

# If attach exits uncleanly (crash, kill -9), remove any leftover hook entries:
coop detach
```

(Adjust `coop`'s path in the commands above, e.g. `../coop/packages/cli/coop`,
if you're not running it from `packages/cli`.)

**Open the browser**

```
http://localhost:3000
```

You'll see your live session. Send steering text from the input box at the bottom.

## What's Working

- Relay event streaming + WebSocket fanout
- Multi-harness capture: Claude Code, opencode, and pi adapters, auto-detected
  (`attach` mode: native hooks/plugins; `run` mode: pty wrapping + steering)
- Web viewer with live event feed, per-session harness label
- Secret redaction before events leave your machine
- Conflict detection between sessions (MCP tool)
- Codex and Gemini CLI hooks exist but aren't wired up yet — see
  `.agents/rules/harnesses.md`

## For Development

```
./scripts/probe.sh     # verify Claude Code hook payloads against your harness
./scripts/dev.sh       # start relay + web (also: bun run dev)
bun run test           # run all tests
bun run lint && bun run build
```

See `.agents/context.md` for the full vision and `.agents/conventions.md` for code style.
