# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

Primary design audience right now: **YC partners and other evaluators meeting coop
for the first time**, usually through a screenshot, a Loom, or a two-minute live
demo. They are technical but have no context on the product, and they decide
whether the multiplayer idea is real within seconds.

Primary product users: **small engineering teams (2–10)** who already run coding
agents (Claude Code, opencode, pi) daily. One person runs an agent; teammates
watch it work, message it, and redirect it.

## Product Purpose

coop makes a coding-agent session multiplayer. One person runs an agent on their
machine; teammates open a link and watch it work live, send it messages, and
steer it — instead of the agent's work being trapped in one person's terminal.

Success for the evaluator audience: a stranger sees a session view and
immediately understands that several named humans and an agent are working in one
shared, live place.

## Positioning

coop observes agents through each harness's **native hook system**, not by
scraping terminal output. That yields structured, semantic events — tool calls,
file touches, results, turn boundaries — with a stable typed schema
(`packages/protocol`). A competitor mirroring a terminal cannot reconstruct this,
and it is what makes presence, conflict detection, and audit possible at all.

Redaction happens in the CLI before any bytes leave the user's machine, so the
relay is treated as untrusted infrastructure.

## Operating Context

- The agent runs on the user's own machine, in their own terminal.
- `coop attach` installs the harness's native hooks and leaves the user's TUI
  alone. `coop run -- <cmd>` wraps the agent in a pty instead.
- Teammates arrive in a browser via a link, after signing in with GitHub.
- Work is organized into **projects**: a project has members, and every session
  belongs to one. Membership, not link-knowledge, grants access.
- Sessions are long-running and often unattended — an evaluator or teammate may
  open one mid-flight or after it ended.

## Capabilities and Constraints

Confirmed and working today:

- Multi-harness event capture (Claude Code, opencode, pi), auto-detected.
- Live event streaming: CLI → relay → browser over WebSocket.
- GitHub auth in two forms: browser login (Better Auth) and `coop login`
  (device flow) for the CLI. Both resolve to one canonical user id.
- Projects, membership, invites; sessions and events persisted to Postgres.
- Messages sent to a running agent, attributed to their sender.
- Live presence: who is watching, who is typing, whether the agent is working.

Durable constraints that future work must not break:

- **Secrets never leave the machine.** Redaction is CLI-side, pre-send.
- **Never parse ANSI to understand what an agent did.** The pty is a keyboard,
  not a data source.
- **Steering text is untrusted user input** — always attributed to a person,
  never framed as a system instruction.
- The event schema is a contract shared by relay and web; both change together.
- In `coop attach` mode there is no pty at all, so any design that assumes a raw
  terminal stream excludes the primary capture mode.

Decided this session: the web renders the **structured event stream**, not a raw
terminal mirror. Human messages and agent events share one timeline, and each
message carries a per-message choice of whether it is sent to the agent or is
teammate-only discussion.

## Brand Commitments

The name is **coop**, lowercase. No logo, palette, or typography is committed —
the visual world is open.

## Evidence on Hand

Real, working, and demonstrable (verified end-to-end this session against a live
Postgres and real GitHub OAuth):

- Real agent sessions streaming real tool calls into the browser.
- Real GitHub sign-in producing real user records.
- Real project creation, invites, membership gating, and event persistence.

Not real, and must not be implied: customer logos, testimonials, usage numbers,
pricing, uptime claims, or any team beyond the single author.

## Product Principles

1. **Show the work, not the terminal.** The value is a legible account of what an
   agent did, readable by someone who did not run it.
2. **Presence is the product.** If a viewer cannot tell who else is here and what
   they are doing, coop is just a log viewer.
3. **Access is membership, never link-knowledge.** A session belongs to a
   project; the project decides who may watch or steer.
4. **The machine's secrets stay on the machine.** Any feature that would ship raw
   bytes off the user's box must pass redaction first or not ship.
5. **A stranger should understand it in one screen.** The session view is the
   product's entire argument.

## Accessibility & Inclusion

No product-specific standard has been established yet. Baseline expectation:
keyboard-reachable controls, visible focus, and text contrast that survives a
projector and a compressed screen recording — the actual conditions of the demo.
