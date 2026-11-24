#!/usr/bin/env bash
# One command that runs every fast verification in the repo.
# Use it after making changes: `bash scripts/check.sh` (or `sh scripts/check.sh`).
set -euo pipefail

echo "==> JS lint (oxlint)"
bun run lint

echo "==> JS typecheck (tsc via turbo)"
bun run typecheck

echo "==> JS tests (vitest via turbo)"
bun run test

echo "==> Go checks"
for module in apps/relay packages/cli; do
    echo "--> $module"
    (
        cd "$module"
        test -z "$(gofmt -l .)" || { gofmt -l .; echo "gofmt found unformatted files in $module"; exit 1; }
        go vet ./...
        go test ./...
    )
done

echo "All checks passed."
