# CLI reference

Every `coop` subcommand, flag, and environment variable, derived from the
flag parsing in `packages/cli/cmd/coop`. The overall usage line:

```text
coop attach [--harness=<name>] --project=<slug>
coop run   [--harness=<name>] --project=<slug> -- <cmd> [args...]
coop detach [dir]
coop login
coop version
```

## `coop attach`

Keep your own terminal. coop installs each detected harness's native hook or
plugin, which forwards events to a local server `coop attach` runs, and on
to the relay. Start your agent after attaching; press Ctrl-C to stop.

```text
coop attach [--harness=<name>] --project=<slug>
```

| Flag | Meaning |
| --- | --- |
| `--project=<slug>` | required. The session must belong to a coop project; the slug is the one shown in the web app. Implies you have run `coop login`. |
| `--harness=<name>` | install hooks for only this harness instead of auto-detecting. Accepted names: `claude-code`, `opencode`, `pi` (and `other`, the no-hooks fallback). Required when more than one harness is detected. |

What attach does, in order: detects harnesses by looking for the `claude`,
`opencode`, and `pi` binaries on your `PATH`; starts the hook server on
`COOP_HOOK_ADDR` (default `127.0.0.1:8788` — it fails rather than silently
picking another port, since the installed hooks point at this address);
writes the harness entries; records the attach in `.coop/sessions.json` in
the working directory so a second attach can detect it; and prints the
session id, relay URL, and a share link of the form
`http://localhost:3000/sessions/<id>?token=<token>`.

On Ctrl-C it removes exactly its own harness entries, unregisters, posts
`session.end` to the relay, and clears its attach state. If attach is killed
hard, `coop detach` does the cleanup.

What gets installed, per harness:

| Harness | File | Notes |
| --- | --- | --- |
| claude-code | `.claude/settings.local.json` | HTTP hook entries (`SessionStart`, `SessionEnd`, `UserPromptSubmit`, `PreToolUse`, `PostToolUse`, `Stop`), 5s timeout, tagged `source: coop`. Existing hooks are preserved; coop removes only its own entries. |
| opencode | `.opencode/plugin/coop.js` | plugin file, `{{BASE_URL}}` pointing at the hook server |
| pi | `.pi/extensions/coop.ts` | extension file. pi requires you to trust the project interactively before extensions load — attach prints a reminder. |

Steering delivery in attach mode goes through each harness's own primitive:
Claude Code gets `additionalContext` on the hook response (only on
`PreToolUse`, `PostToolUse`, and `UserPromptSubmit` — the events Claude Code
accepts it on); opencode's plugin calls `client.tui.appendPrompt`; pi's
extension calls `pi.sendUserMessage(text, { deliverAs: "steer" })`. Takeover
vetoes tool calls only on Claude Code (`PreToolUse` deny). Harness commands
are never delivered in attach mode — there is no pty to type into — and the
CLI drops them.

## `coop run`

Wrap the agent in a pseudo-terminal. coop owns stdin, so steering and
harness commands are plain keystrokes, delivered within ~1.5 seconds, for any
harness — including ones with no adapter, which get pty-only capture with
reduced event detail.

```text
coop run [--harness=<name>] --project=<slug> -- <cmd> [args...]
```

Flags are the same as `attach`. If the harness is not given, it is inferred
from the command's binary name. With no command at all, `coop run` runs
`claude`:

```bash
coop run --project=my-project                # wraps claude
coop run --project=my-project -- opencode    # wraps opencode
```

The wrapped terminal is mirrored to the relay (redacted) and shown live in
the web app. When someone claims takeover, their keystrokes are routed into
the pty; in a pty there is no hook-deny mechanism, so takeover is advisory
there except for the notice the owner sees.

## `coop detach`

Ends any session a killed attach left open in a directory and removes every
coop harness entry from it (Claude Code hook entries tagged `coop`, the
opencode plugin, the pi extension).

```bash
coop detach           # the current directory
coop detach ~/myrepo  # another directory
```

Attach state is stored per working directory under
`~/.config/coop/attached/`; detach reads it, posts `session.end` for each
recorded session, and then strips the harness files.

## `coop login`

Authenticates you once using GitHub's device flow. Prints a code, you
enter it at the verification URL, and the resulting token is exchanged with
the relay. The relay stores only a hash; the credential is saved at
`~/.config/coop/credentials.json` (mode 0600). `--project` on attach or run
requires this to have been run.

`coop login` needs `COOP_GITHUB_CLIENT_ID` set in the environment (the same
public client id the web app's OAuth app uses). To log out, `POST
/v1/auth/cli/revoke` with the credential, or remove the credentials file.

## `coop version`

Prints `coop <version>`. `--version` and `-v` are aliases.

## Environment variables

| Variable | Used by | Default | Meaning |
| --- | --- | --- | --- |
| `COOP_RELAY_URL` | all | `http://localhost:8787` | where the relay is |
| `COOP_HOOK_ADDR` | attach, run | `127.0.0.1:8788` | address for the local hook server. Two attaches on one machine need different values. |
| `COOP_SESSION_ID` | attach, run | generated (`sess-…`) | pin the session id (e.g. to reattach to a known session) |
| `COOP_SESSION_TOKEN` | attach | generated | token embedded in the printed share link |
| `COOP_DISABLE_STREAM_DIFFS` | attach, run | unset (diffs on) | set to a truthy value to stop streaming diff hunks with `file.touched` events |
| `COOP_GITHUB_CLIENT_ID` | login | — | public client id of the GitHub OAuth app |

## Files coop writes

| Path | What |
| --- | --- |
| `~/.config/coop/credentials.json` | CLI credential (0600) |
| `~/.config/coop/attached/<hash>.json` | per-directory attach records, for detach |
| `<dir>/.coop/sessions.json` | active attaches in this directory (conflict detection) |
| `<dir>/.claude/settings.local.json` | Claude Code hook entries (merged, other hooks preserved) |
| `<dir>/.opencode/plugin/coop.js`, `<dir>/.pi/extensions/coop.ts` | opencode / pi wiring |

## Multiple sessions in one directory

Claude Code supports several concurrent `coop attach` sessions in one
directory: each attach keeps its own hook entries, and each receives its own
events. opencode and pi allow only one — they load a single plugin file, so
a second attach would silently steal the first one's events; the CLI
detects the active attach in `.coop/sessions.json` and refuses:

```text
coop attach: opencode already has an active coop session in this directory
(session sess-…, pid 1234); opencode supports only one coop session per
directory, stop it first (Ctrl-C, then run `coop detach` to clean up)
```

The hook address is also part of the deal: it does not fall back to another
port, so a second attach in the same directory fails loudly instead of
leaving the first one's hooks pointing at a dead server. If you run multiple
attaches in one directory, give each its own `COOP_HOOK_ADDR`.

## The MCP server

`packages/mcp` builds a stdio MCP server (`coop-mcp`) meant to run next to
the agent. It exposes five tools: `ask_team` (ask the team a question and
wait for an answer), `check_conflicts` (files other active sessions are
touching), `list_active_sessions`, `share_note`, and `list_project_notes`.
It is configured with the same environment: `COOP_RELAY_URL`,
`COOP_CLI_CREDENTIAL` (or it reads the stored login credential),
`COOP_SESSION_ID`, `COOP_PROJECT`, and `COOP_REPO` (defaults to the git
origin of the current directory).
