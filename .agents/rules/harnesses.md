# Harness integration facts

Everything in this file must be verified against a real session before it is
relied on. When you verify something, update this file with the date.

## Claude Code

*Verified: 2026-08-25, against code.claude.com/docs/en/hooks AND a live
capture via `scripts/probe.sh` (command hooks) plus a manual HTTP-hook probe
(same payloads, `{"type":"http",...}` entries pointing at a local Python
server) run against `claude -p` in a scratch dir. Facts below marked
"(live)" were observed directly; the rest come from the doc.*

- (live) Common envelope fields actually observed: `session_id`,
  `transcript_path`, `cwd`, `hook_event_name`, `prompt_id` (absent on
  `SessionStart`, present on every event once the first prompt has been
  submitted), `permission_mode`, `effort.level`. `agent_id`/`agent_type`
  were not observed (no subagent was spawned in the probe) — still treat
  those as plausible-but-unverified for subagent runs.
- (live) `SessionStart` adds `source` (observed value: `"startup"`).
  `UserPromptSubmit` adds `prompt`. `PreToolUse`/`PostToolUse` add
  `tool_name`, `tool_input`, `tool_use_id`; `PostToolUse` additionally adds
  `tool_response` (shape is tool-specific, e.g. `{stdout,stderr,interrupted,
  isImage,noOutputExpected}` for Bash, `{filePath,oldString,newString,
  originalFile,structuredPatch,userModified,replaceAll}` for Edit) and
  `duration_ms`. `Stop` adds `stop_hook_active`, `last_assistant_message`,
  `background_tasks`, `session_crons`. `SessionEnd` adds `reason` (observed
  value: `"other"` on a normal `-p` exit).
- (live) **`additionalContext` is only safe on `PreToolUse`, `PostToolUse`,
  and `UserPromptSubmit`.** Returning `{"hookSpecificOutput":
  {"hookEventName":"SessionEnd","additionalContext":"..."}}` from a
  `SessionEnd` hook produced a visible "Hook JSON output validation failed"
  error. Returning the same shape from a `Stop` hook did not error, but
  caused Claude Code to re-invoke the `Stop` hook repeatedly (9 times in one
  run, until an outer timeout killed the process) instead of terminating —
  almost certainly the same continuation mechanism `Stop`'s own
  `decision:"block"` field uses. **coop's hook server must never attach
  `additionalContext` to `SessionStart`/`SessionEnd`/`Stop` responses** —
  reply `{}` for those regardless of pending steering text, and only surface
  steering on `PreToolUse`/`PostToolUse`/`UserPromptSubmit`.
- (live) The HTTP hook mechanism itself matches the doc: a `{"type":"http",
  "url":...,"timeout":...}` entry gets POSTed the same JSON body a command
  hook receives on stdin, and a 200 response with a JSON body is parsed as
  hook output the same way command-hook stdout is.

- Command hooks receive one line of JSON on stdin and answer via exit code
  and stdout JSON.
- **HTTP hooks are natively supported and are coop's ingest path.** A hook
  entry can be `{"type": "http", "url": "...", "timeout": <seconds>,
  "headers": {...}}` — Claude Code POSTs the same JSON payload a command
  hook would get on stdin as the request body, and reads the response body
  as the JSON output. 2xx + JSON body = parsed as output; 2xx + empty body =
  success/no-op; non-2xx, connection failure, or timeout = non-blocking,
  execution continues. This means `coop attach` does **not** need to spawn a
  subprocess per hook event — it registers HTTP hook entries pointing at a
  local server `coop attach` itself runs, which redacts and forwards to the
  relay, then answers Claude Code directly. (Exit-code-2 blocking, below, is
  a command-hook-only mechanism; HTTP hooks control the decision purely via
  the JSON response body's `permissionDecision`.)
- Common envelope fields on every event: `session_id`, `prompt_id` (UUID,
  absent until first user input), `transcript_path`, `cwd`,
  `permission_mode` (`default|plan|acceptEdits|auto|dontAsk|bypassPermissions`),
  `effort.level`, `hook_event_name`. Subagent runs add `agent_id`,
  `agent_type`.
- `PreToolUse`/`PostToolUse`/`PostToolUseFailure`/`PermissionRequest`/`PermissionDenied`
  add `tool_name`, `tool_input`, `tool_use_id`. `Stop`/`SubagentStop` add
  `last_assistant_message`.
- Exit code 2 on `PreToolUse` (command hooks only) blocks the call and feeds
  stderr back to the agent. This is coop's veto primitive for command hooks;
  out of scope for the current phase (attach mode uses HTTP hooks).
- Steering is delivered as `hookSpecificOutput.additionalContext` in the
  JSON response body:
  ```json
  { "hookSpecificOutput": { "hookEventName": "PreToolUse",
      "additionalContext": "..." } }
  ```
  The exact list of which hook events accept `additionalContext` was not
  fully enumerated in the fetched doc excerpt (page truncated at the
  "Decision control" table) — confirm which events actually apply it before
  relying on one outside `PreToolUse`/`PostToolUse`/`UserPromptSubmit`.
- Known gotchas:
  - Whether `SessionStart`/`SessionEnd` fire in `-p` (print/non-interactive)
    mode was not conclusively documented in the fetched excerpt — `Setup` is
    the event explicitly designed for `-p`/CI one-time init. Confirm with a
    live capture before depending on `SessionStart`/`SessionEnd` firing in
    `-p` mode.
  - On `--continue` / `--resume`, previously injected text is replayed rather
    than the hook re-running, so anything time-sensitive goes stale.
  - Out-of-band, system-looking injected text can trip prompt-injection
    defenses and get surfaced to the user instead of used as context.
- There is no supported way to inject a message into a running interactive
  TUI session from outside. That is why `run` mode owns a pty.

## Codex CLI

- Hooks are a near-direct port: same JSON-on-stdin protocol, same exit codes,
  same `additionalContext` and `hookSpecificOutput` shapes.
- Smaller lifecycle: `SessionStart`, `UserPromptSubmit`, `PreToolUse`,
  `PermissionRequest`, `PostToolUse`, `Stop`.
- Command hooks only — no HTTP hooks. Codex needs a local shim that POSTs to
  the relay.

## OpenCode

*Verified: 2026-08-25, against the opencode 1.18.22 binary and the
`@opencode-ai/plugin`/`@opencode-ai/sdk` type declarations actually installed
at `~/.config/opencode/node_modules/`, plus the live config schema fetched
from `https://opencode.ai/config.json`. This supersedes an earlier,
unverified note that described a `.opencode/plugins/` shell-hook system —
that does not exist.*

- In-process TypeScript/JS plugin, not a subprocess hook. Auto-loaded from
  `.opencode/plugin/*.js` (or `.opencode/plugin/<name>/index.js`) — no
  `opencode.json` entry required for a local file. `opencode.json` also
  supports a top-level `"plugin"` array (npm module names or
  `[name, options]` tuples) for published plugins.
- The plugin module exports an async factory:
  `export const X = async ({client, directory, worktree}) => ({ ...hooks })`.
  Relevant hooks for coop: `event` (catch-all bus subscriber, receives every
  `{type, properties}` on the bus), `tool.execute.before` /
  `tool.execute.after` (input `{tool, sessionID, callID}`, output object is
  mutated in place — `before`'s output has `.args`, `after`'s has
  `.title`/`.output`/`.metadata`).
- Event bus names (verified from `@opencode-ai/sdk`'s generated `Event`
  union): `session.created/updated/deleted/idle/error/status/compacted/diff`,
  `message.updated/removed`, `message.part.updated/removed`,
  `tool.execute.before/after` are hook names not bus events; the bus's
  `file.edited` is the free file-touched signal — no path inference needed,
  unlike Claude Code where it's derived from `tool_input.file_path`.
- **Confirmed schema has no `hooks` key and no `experimental.hook`.** An
  older description of shell-command hooks under
  `experimental.hook.file_edited`/`session_completed` is stale — the current
  schema's only `experimental` keys are `disable_paste_summary`,
  `batch_tool`, `openTelemetry`, `primary_tools`, `continue_loop_on_deny`,
  `mcp_timeout`, `policies`. Do not target it.
- Steering: `client.tui.appendPrompt({body:{text}})`, a method on the SDK
  client injected into the plugin factory as `client`. No polling needed —
  call it directly from the `event` handler when a steer message arrives.
- Veto: `permission.ask` hook, input is a `Permission` object, output
  `{status: "ask" | "deny" | "allow"}`.
- `opencode serve --port <n> --hostname 127.0.0.1` also exposes `GET /event`
  (SSE, same bus) and `GET /doc` (live OpenAPI spec) — a viable
  external-process integration path if the in-process plugin route is ever
  insufficient, not used by coop today.

## Pi

*Verified: 2026-08-25, against `pi` 0.84.2 installed on this machine and its
bundled docs at
`~/.nvm/versions/node/v24.3.0/lib/node_modules/@earendil-works/pi-coding-agent/docs/{extensions,rpc}.md`.*

- In-process TypeScript extension, auto-discovered from
  `.pi/extensions/*.ts` or `.pi/extensions/*/index.ts` (project-local — pi
  requires the user to interactively trust the project before project-local
  extensions load; `coop attach` should say so when pi is detected). Loaded
  via jiti, no build step.
- `export default function(pi: ExtensionAPI) { pi.on("session_start", ...); pi.on("tool_call", ...); pi.on("tool_result", ...); pi.on("turn_end", ...); pi.on("agent_end", ...); pi.on("session_shutdown", ...); }`.
  `tool_call` fires before execution and **can block**: return
  `{block: true, reason}`; `event.input` is mutable and tool-specific
  (`read` → `{path,...}`, `edit`/`write` similar).
- Steering: `pi.sendUserMessage(text, {deliverAs: "steer"})` — delivered
  after the current turn's tool calls finish, before the next LLM call. No
  polling, no mailbox race — the cleanest steering primitive of any harness
  coop supports.
- Also has an `rpc` run mode (`pi --mode rpc`, JSONL over stdin/stdout,
  `{"type":"prompt","streamingBehavior":"steer"}`) — a possible alternative
  external-process integration, not used by coop's extension-based adapter
  today.

## Codex CLI (doc-sourced, not yet verified live)

- Not installed on this development machine as of 2026-08-25 — the section
  above (near-direct port of Claude Code's hook protocol) comes from
  `learn.chatgpt.com/docs/hooks` and the `codex-rs/hooks` source tree, not a
  live capture. Config is `~/.codex/hooks.json` or `config.toml`
  (`[[hooks.PreToolUse]]` tables); events `SessionStart`, `SessionEnd`,
  `UserPromptSubmit`, `PreToolUse`, `PermissionRequest`, `PostToolUse`,
  `PreCompact`, `PostCompact`, `SubagentStart`, `SubagentStop`, `Stop`.
  Command hooks only, same `additionalContext`/`hookSpecificOutput` shape as
  Claude Code. **Do not build the Codex adapter from this alone** — run
  `scripts/probe.sh codex` against a real install first, per non-negotiable
  #4.

## Gemini CLI (doc-sourced, not yet verified live)

- Not installed on this development machine as of 2026-08-25. Hooks are
  gated behind `"enableHooks": true` and `"enableMessageBusIntegration": true`
  in `settings.json`, suggesting pre-GA status — config shape and
  `hookSpecificOutput.additionalContext`/`llm_response`/`tool_input`
  overrides are documented at
  `github.com/google-gemini/gemini-cli/blob/main/docs/hooks/reference.md`
  but unverified against a live run. Same rule: probe before building.

## Anything else

Wrap the pty. Degraded event fidelity, still useful.
