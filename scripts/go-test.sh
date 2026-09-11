#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

for module in apps/relay packages/cli; do
  echo "go test: $module"
  (cd "$REPO_ROOT/$module" && go test ./...)
done
