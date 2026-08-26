# coop

Multiplayer sessions for coding agents. Share a live session, steer it together.

One person runs a coding agent. Teammates open a link and watch it work live, send it a message, or take over. Agents see what other sessions are touching, so they don't collide.

Works with Claude Code, opencode, and pi today — `coop attach` detects which
harness you're using and installs the right adapter automatically. Anything
else falls back to plain pty wrapping (`coop run -- <cmd>`), with reduced
event detail.

## Quick Start

**Terminal 1: Start relay + web viewer**

```bash
cd /home/francogalfre/Documentos/dev/coop
bun run dev
```

**Terminal 2: Build the CLI and use it**

```bash
cd /home/francogalfre/Documentos/dev/coop/packages/cli
go build -o coop ./cmd/coop

# Now use it in your project
cd /your/test/project
/home/francogalfre/Documentos/dev/coop/packages/cli/coop attach
# coop detects claude / opencode / pi automatically and installs hooks for each.
# In another terminal, start your agent as usual: claude, opencode, or pi

# Or let coop launch it directly and wrap the terminal:
/home/francogalfre/Documentos/dev/coop/packages/cli/coop run -- claude

# If attach exits uncleanly (crash, kill -9), remove any leftover hook entries:
/home/francogalfre/Documentos/dev/coop/packages/cli/coop detach
```

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
