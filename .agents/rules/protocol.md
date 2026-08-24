# Protocol

`packages/protocol` is the single source of truth for every event that
crosses a boundary: CLI → relay, relay → web, both directions. Schemas are
Zod, declared once, and shared — never re-declared on the consumer side.

## The contract

1. Changing `packages/protocol` means changing the relay and the web app in
   the same change. Never add a field to one side only — an event the relay
   accepts but the viewer can't render is a bug shipped to two places.
2. Validate at every boundary. The relay parses ingest against the schema;
   unknown or malformed events are rejected, not best-effort forwarded.
3. Evolve additively. A new optional field or a new event type is safe.
   Renaming, removing, or changing a field's meaning is a breaking change —
   version the event instead of editing it silently.

## Naming

- Event names are dot-separated and past tense: `tool.blocked`. Follow the
  names already in the package; don't invent a parallel convention.
- When a field carries harness data, the field name comes from verified
  hook output (see `harnesses.md`). Never coin a field name because it
  "seems right" — if it wasn't in a captured payload, it doesn't exist.
