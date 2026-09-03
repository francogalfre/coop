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

Release binaries are published to [GitHub Releases](../../releases) whenever
a `vX.Y.Z` tag is pushed — download one for your platform instead of building
from source if you just want to run `coop`. To build locally (needed for
`coop` development, e.g. `./scripts/probe.sh`):

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

# If attach exits uncleanly (crash, kill -9), clean up: removes leftover hook
# entries and ends the session that attach left open.
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
- Codex and Gemini CLI adapters aren't built yet — see
  `.agents/rules/harnesses.md`

## For Development

```
./scripts/probe.sh     # verify Claude Code hook payloads against your harness
./scripts/dev.sh       # start relay + web (also: bun run dev)
bun run test           # run all tests
bun run lint && bun run build
```

See `.agents/context.md` for the full vision and `.agents/conventions.md` for code style.

## Deploying

Deploy straight from GitHub with [Coolify](https://coolify.io): add this repo
as a **Docker Compose** resource. Coolify builds `web` and `relay` from
`docker-compose.yml` and proxies both through its own Traefik instance using
the `traefik.*` labels already in that file — there's no separate reverse
proxy to run. Postgres is external/managed (e.g. Neon) — bring your own
`DATABASE_URL`.

1. Point `COOP_DOMAIN`'s DNS A/AAAA record at your Coolify server.
2. Set the variables from `.env.production.example` (`COOP_DOMAIN`,
   `DATABASE_URL`, `COOP_GITHUB_CLIENT_ID`/`SECRET`, `COOP_INTERNAL_SECRET`,
   `BETTER_AUTH_SECRET`) as environment variables on the Coolify resource.
3. Deploy. Coolify provisions the TLS certificate on first request.

To run it anywhere else without Coolify, `docker compose up -d --build` still
works, but you'd need your own reverse proxy in front of it — the compose
file only carries Traefik labels, not a bundled proxy.

The relay holds several in-memory singletons (`Store`, `Mailbox`,
`PresenceHub`, `TakeoverRegistry`, `PtyHub`, `SteerRequestRegistry` — see
`apps/relay/cmd/relay/main.go`) and is strictly single-instance —
`docker-compose.yml` runs exactly one `relay` replica and must keep doing so.
