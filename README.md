# coop

Multiplayer sessions for coding agents. Share a live session, steer it together.

> Early. Nothing works yet.

One person runs a coding agent. Teammates open a link and watch it work in real
time, send it a message, block a tool call before it runs, or take the session
over. Agents in the same repo see what other sessions are touching, so they
stop overwriting each other.

Works by hooking into the harness, not by scraping the terminal.

## Status

| | |
|---|---|
| protocol | in progress |
| relay (Go) | not started |
| cli | not started |
| viewer | not started |
| MCP server | not started |

## Development

```
./scripts/skills.sh   # agent tooling
./scripts/probe.sh    # capture real hook payloads before coding against them
```

See AGENTS.md.
