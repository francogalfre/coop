# Testing

## Tests before the fix

Reproduce the bug in a failing test before changing code. The test is the
receipt that you understood the bug; a fix without one is a guess.

## Fixtures come from real sessions

Harness behavior is verified, not assumed (see `harnesses.md`). When a harness release changes
a payload, update the fixture and `harnesses.md`, not the assertion.

## Commands

- Go: `go test ./...`
- TypeScript: `bun run test`
