# Contributing

## Setup

You need [Bun](https://bun.sh), [Go](https://go.dev) (the version pinned in
each module's `go.mod`), and a Postgres database.

```bash
bun install
cp apps/web/.env.example apps/web/.env.local   # fill in DATABASE_URL, GitHub OAuth, secrets
bun run dev                                    # relay on :8787, web on http://localhost:3000
```

`bun run dev` runs `scripts/dev.sh`, which reads every variable from
`apps/web/.env.local` and refuses to start without it. Build the CLI when
you work on it:

```bash
cd packages/cli && go build -o coop ./cmd/coop
```

## Commands

```bash
bun run dev        # relay + web (scripts/dev.sh)
bun run test       # go test ./... in apps/relay and packages/cli (scripts/go-test.sh), then vitest in every workspace
bun run lint       # oxlint
bun run typecheck  # tsc --noEmit via turbo
bun run build      # turbo run build
./scripts/probe.sh # dump real Claude Code hook payloads to verify against
```

CI (`.github/workflows/ci.yml`) runs exactly these — oxlint, typecheck,
build, tests, plus `gofmt` and `go vet` with `-race` for both Go modules —
against a Postgres 16 service, using `TEST_DATABASE_ADMIN_URL`. If it is
green locally, it is green in CI.

Some relay and CLI database tests need a local Postgres reachable as
`postgres://postgres:password@localhost:5432/postgres`; set
`TEST_DATABASE_ADMIN_URL` if yours differs.

## Layout

```text
apps/relay/         Go — HTTP ingest, WebSocket fanout, steering, permissions, Postgres (Ent)
apps/web/           Next.js — session viewer, projects, auth (better-auth, Drizzle)
packages/protocol/  TypeScript — Zod event schemas shared by relay and web
packages/cli/       Go — coop attach / run / detach / login, harness adapters, redaction
packages/mcp/       TypeScript — MCP server (ask the team, conflicts, notes)
scripts/            dev.sh, go-test.sh, probe.sh
.agents/            context.md, conventions.md, and rules/ — read rules before the area they cover
```

How the pieces talk to each other is in [architecture.md](architecture.md).

## Conventions

Read `.agents/conventions.md` in full before your first change. The ones
that bite hardest:

- **Commits:** `feature(scope): imperative summary` — subject line only, no
  body, no attribution. Scope is `relay`, `cli`, `protocol`, `web`, or `mcp`.
- **Files:** hard 300-line limit. One exported concept per file; no
  `utils.ts` / `helpers.go`; split by responsibility, not line count.
  Generated code (`internal/db/ent/`, `apps/web/database/migrations/`) is
  exempt — do not hand-reorganize it.
- **Comments:** none by default. The only allowed comment is one line
  explaining a genuinely non-obvious *why*. Before finishing, grep the files
  you touched and remove stray comments.
- **Go:** stdlib first, `gofmt`, errors wrapped with context, never
  swallowed. **TypeScript:** strict mode, no `any`, Zod for anything
  crossing a boundary.
- **No new dependencies without asking**, especially in Go.

## Tests before fixes

Reproduce a bug in a test before fixing it, and take fixtures from real
sessions rather than hand-written payloads. `.agents/rules/` records the
verified behavior of each harness — if a hook field or payload shape is not
in there or in `scripts/probe.sh` output, it does not exist yet, so do not
invent it.

## Adding a harness adapter

Adapters live in `packages/cli/internal/harness/<name>/` and implement the
`harness.Adapter` interface: `Name`, `Binary`, `IsFallback`,
`SupportsMultiAttach`, `Detect`, `Install`, `RemoveAll`. Register the adapter
in `packages/cli/cmd/coop/harnesses.go`, then:

1. Verify the harness's real hook payload shape first (`scripts/probe.sh`
   or the harness's own docs) — field names change between releases.
2. Write a translator in `packages/cli/internal/hooks/` that maps the
   harness's payload onto the protocol events.
3. Decide whether the harness supports multiple concurrent attaches in one
   directory (`SupportsMultiAttach`) — only if its wiring is per-session,
   like Claude Code's hook entries, not a single shared plugin file.
4. Record what you confirmed in `.agents/rules/harnesses.md`.

Harnesses with no adapter fall back to pty-only wrapping under `coop run`
(reduced event detail, no attach mode).
