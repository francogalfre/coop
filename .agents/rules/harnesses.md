# Harness integration facts

Everything in this file must be verified against a real session before it is
relied on. When you verify something, update this file with the date.

## Claude Code

- Hooks receive one line of JSON on stdin and answer via exit code and stdout
  JSON. HTTP hooks receive the same JSON as a POST body and answer in the
  response body. HTTP hooks are the primary ingest path for coop.
- Common envelope on every event: `session_id`, `transcript_path`, `cwd`,
  `hook_event_name`, plus `permission_mode` on most.
- `PreToolUse` adds `tool_name`, `tool_input`. `PostToolUse` adds
  `tool_output`.
- Exit code 2 on `PreToolUse` blocks the call and feeds stderr back to the
  agent. This is coop's veto primitive.
- `additionalContext` returned from a hook is coop's steering primitive.
- Known gotchas:
  - Session hooks do not run in `--print` mode.
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

- Plugin-based: TypeScript modules in `.opencode/plugins/` exporting an async
  function that returns event handlers.

## Anything else

Wrap the pty. Degraded event fidelity, still useful.
