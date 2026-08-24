# Security

## Secrets never leave the machine

Redaction happens in the CLI, before an event is sent. The relay is
untrusted infrastructure: by the time an event reaches it, it must already
be safe to store and to broadcast to every connected browser.

- Filter on send, not on receive. Relay-side filtering is too late — the
  bytes have already left the machine.
- Redact by pattern, not by location: API keys, tokens, private key blocks,
  connection strings, `Authorization` headers, cookie values.
- Redact in every field, not just tool output. Secrets show up in prompts,
  file contents, diffs, and command arguments too.
- When unsure whether something is a secret, redact it. A false positive
  costs readability; a false negative costs a credential.
- Replacement text should make the hole visible (`[redacted]`), so viewers
  understand why content is missing instead of trusting a gap.

## Steering text is untrusted user input

Anything a teammate types into the viewer is user input, even though it
arrives through our own relay. Treat it like form data, not like a prompt.

- Always attribute. The agent must see who sent it (`[name via coop]`), so
  it can weigh the instruction and the session owner can audit the log.
- Never frame steering as a system instruction. It is delivered as context
  from a person — never with wording that impersonates the harness, the
  session owner, or a system message. (System-looking injected text also
  trips prompt-injection defenses; see `harnesses.md`.)
- Steering never changes permissions. A message cannot approve a tool call,
  grant access, or relax redaction. Approval is a separate explicit action
  in the relay's permission flow, not text in the event stream.
- The same rules apply to veto messages (stderr on `PreToolUse` exit 2):
  attributed, framed as coming from a person, no system voice.

## Logging

Relay and CLI logs only ever see the redacted event, never the raw one. If
you need a raw payload for debugging, keep it on the user's machine.
