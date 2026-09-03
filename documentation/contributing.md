# Contributing

## Setup

You'll need [Bun](https://bun.sh), [Go](https://go.dev) (see
`packages/cli/go.mod` and `apps/relay/go.mod` for the version), and a Postgres
database.

```bash
bun install
cp apps/web/.env.example apps/web/.env.local   # fill in DATABASE_URL, GitHub OAuth, secrets
bun run dev                                     # relay + web on http://localhost:3000
```

Build the CLI when you're working on it:

```bash
cd packages/cli && go build -o coop ./cmd/coop
```

## Scripts

```bash
bun run dev            # relay + web (scripts/dev.sh)
bun run test           # go test ./... for both modules, then vitest
bun run lint           # oxlint
bun run typecheck      # tsc --noEmit
./scripts/probe.sh     # dump real Claude Code hook payloads to verify against
```

## Layout

See [architecture.md](architecture.md). Each package is independent; the web
app imports `@coop/protocol`'s built output, so run its build (or `bun run
dev`, which handles it) before typechecking the web app alone.

## Conventions

Read `.agents/conventions.md` before your first change. In short:

- Commits: `feature(scope): imperative summary`, subject line only, no body,
  no attribution. Scope is `relay`, `cli`, `protocol`, `web`, or `mcp`.
- 300-line hard limit per file; one exported concept per file; no
  `utils.go` / `helpers.ts`.
- Comments explain *why*, never *what* — and only when the why is
  non-obvious.
- Go: standard library first, `gofmt`, errors wrapped with context.
- TypeScript: strict mode, no `any`, Zod for anything crossing a boundary.

## Adding a harness adapter

Each adapter lives in `packages/cli/internal/harness/<name>/` and implements
detect / install / uninstall. Verify the harness's real hook payload shape
first — field names and the transcript format change between releases, so
don't invent them. Record what you confirmed in `.agents/rules/harnesses.md`.
Harnesses with no adapter fall back to pty-only wrapping.

## Tests

Reproduce a bug in a test before fixing it. Fixtures come from real sessions,
not hand-written payloads. See `.agents/rules/testing.md`.
