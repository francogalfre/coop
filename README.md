<img src="assets/banner.svg" width="100%" alt="coop" />

<p align="center">Multiplayer sessions for coding agents.</p>

<p align="center">
  <a href="documentation/overview.md">Overview</a> ·
  <a href="documentation/architecture.md">Architecture</a> ·
  <a href="documentation/cli.md">CLI reference</a> ·
  <a href="documentation/deployment.md">Deployment</a> ·
  <a href="documentation/contributing.md">Contributing</a>
</p>

<br />

coop is for teams that run coding agents. One person starts a session with
Claude Code, opencode, or pi; `coop attach` wires the harness's own hook
system to a relay, and everyone else on the project opens the session in a
browser: they watch every tool call and file change as it happens, send the
agent attributed messages mid-task, answer its questions, or take over and
drive it themselves.

Nothing is scraped off the terminal. coop listens through each harness's
native hooks or plugins, and steering travels back through the same channel.
Secrets are redacted in the CLI before an event ever leaves the machine.

From the browser, teammates can:

- watch the session live — every tool call, result, file change, and message,
  with the file activity of every other session in the project
- steer the agent with attributed messages (`[Alice via coop] ...`), subject
  to the session's mode
- answer the agent's questions when it asks the team
- take over — claim the session and drive it while everyone watches

The session owner keeps the controls that matter: mode changes, permission
prompts, and steering approval in restricted mode are owner-only.

## Supported harnesses

| Harness | attach (keep your own terminal) | run (wrapped in a pty) | Concurrent attaches per directory |
| --------- | ------------------------------- | ---------------------- | -------------------------------- |
| Claude Code | yes — native hooks in `.claude/settings.local.json` | yes | several |
| opencode | yes — plugin in `.opencode/plugin/coop.js` | yes | one |
| pi | yes — extension in `.pi/extensions/coop.ts` | yes | one |
| anything else | no hooks detected | yes — pty only, reduced event detail | several |

opencode and pi load a single plugin/extension file per directory, so a
second attach would steal the first one's events; the CLI refuses it. Claude
Code keeps a separate hook entry per attach, so several can coexist. Details
in the [CLI reference](documentation/cli.md).

## Quick start

You need [Bun](https://bun.sh), [Go](https://go.dev), a Postgres database,
and a GitHub OAuth app (dev callback:
`http://localhost:3000/api/auth/callback/github`).

```bash
bun install
cp apps/web/.env.example apps/web/.env.local   # fill in DATABASE_URL, GitHub OAuth, secrets
bun run dev                                    # relay on :8787, web viewer on http://localhost:3000
```

In another terminal, build and log in with the CLI:

```bash
cd packages/cli
go build -o coop ./cmd/coop
export COOP_RELAY_URL=http://localhost:8787
export COOP_GITHUB_CLIENT_ID=<same client id as the web app>
./coop login
```

Open `http://localhost:3000`, sign in with GitHub, and create a project.
Then, in the repository where your agent works:

```bash
/path/to/packages/cli/coop attach --project=<project-slug>
```

Start your agent as usual. `coop attach` prints the session id and a link to
the live session; share it with your team. Press Ctrl-C to stop and clean
up.

## Documentation

- [Overview](documentation/overview.md) — the mental model: sessions,
  projects, events, steering, and the trust boundary.
- [Architecture](documentation/architecture.md) — how relay, web, CLI,
  protocol, and MCP fit together, from harness hook to browser.
- [CLI reference](documentation/cli.md) — every `coop` subcommand, flag, and
  environment variable.
- [Deployment](documentation/deployment.md) — production setup, env vars,
  database migrations, and the go-live checklist.
- [Contributing](documentation/contributing.md) — local setup, tests, and
  the conventions that matter.

<br />

<p align="center">Built with Go, Next.js, and Postgres — by <a href="https://francogalfre.site">Franco Galfré</a>.</p>
