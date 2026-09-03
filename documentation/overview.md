# Overview

## The problem

AI agents are the most powerful tool a team has, and the one thing everyone
still uses alone. You open a session, type a prompt, get an answer in a window
only you can see. When you want a teammate involved, the best you can do is
send them a link to a transcript they can read but not touch.

Agents now run tasks that take hours or days. Work at that scale was never
meant to be done by one person watching a terminal.

## What coop does

coop makes an agent session something a team shares. One person runs the
agent; anyone on the project can open the session in a browser and:

- **Watch** every tool call, message, and file change as it happens.
- **Steer** it — send a message the agent picks up on its next step,
  attributed to you.
- **Take over** — pause the agent's next tool call and drive it yourself,
  then hand control back.

Agents also see which files other sessions are touching, so two people working
in the same repo don't collide.

## How it works

The agent runs on your machine as normal. coop observes it through the
harness's own hook system and streams structured events to a relay, which fans
them out to browsers over WebSockets. Steering travels back the same channel.

coop never reads the terminal to figure out what happened — the terminal is
treated as a keyboard, not a data source.

Two ways to capture a session, both supported:

- **attach** — you keep your normal terminal. `coop attach` installs each
  detected harness's native hook or plugin, which forwards events to a small
  local server and on to the relay.
- **run** — `coop run -- claude` wraps the agent in a pseudo-terminal. Same
  hook wiring, plus coop owns stdin, so steering is delivered instantly.

## What's supported

| Harness | attach | run |
| --- | --- | --- |
| Claude Code | ✅ native hooks | ✅ |
| opencode | ✅ plugin | ✅ |
| pi | ✅ plugin | ✅ |
| anything else | — | ✅ pty only, reduced event detail |

Codex and Gemini CLI adapters aren't built yet.

## Safety

- **Secrets never leave your machine.** Redaction happens in the CLI, before
  an event is sent — API keys, tokens, connection strings, `Authorization`
  headers, and private-key blocks are replaced with `[redacted]`.
- **Steering is attributed, never disguised.** A teammate's message reaches
  the agent as text from a named person, never as a system instruction, and it
  can't approve a tool call or change permissions.
- **The session owner is in control.** Only the person who started a session
  can approve steering in restricted mode, resolve permission prompts, or
  change the session mode.
