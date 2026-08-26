#!/usr/bin/env bash
set -euo pipefail

trap 'kill 0 2>/dev/null || true' EXIT INT TERM

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

WEB_ENV="$REPO_ROOT/apps/web/.env.local"

if [ ! -f "$WEB_ENV" ]; then
  echo "error: $WEB_ENV not found. Copy apps/web/.env.example and fill it in first." >&2
  exit 1
fi

env_var() {
  grep "^$1=" "$WEB_ENV" | head -1 | cut -d '=' -f2-
}

echo "🚀 coop development server"
echo "  relay: http://localhost:8787"
echo "  web:   http://localhost:3000"
echo ""
echo "Stopping any existing processes..."
pkill -f "go run.*relay|next dev" 2>/dev/null || true
sleep 1

echo "Starting relay and web..."
(cd apps/relay && env \
  DATABASE_URL="$(env_var DATABASE_URL)" \
  COOP_INTERNAL_SECRET="$(env_var COOP_INTERNAL_SECRET)" \
  COOP_GITHUB_CLIENT_ID="$(env_var COOP_GITHUB_CLIENT_ID)" \
  COOP_GITHUB_CLIENT_SECRET="$(env_var COOP_GITHUB_CLIENT_SECRET)" \
  COOP_WEB_INTERNAL_URL="http://localhost:3000" \
  COOP_WEB_ORIGINS="http://localhost:3000" \
  go run ./cmd/relay) &
RELAY_PID=$!
sleep 1

(cd apps/web && bun run dev) &
WEB_PID=$!

echo ""
echo "✓ Relay running (PID $RELAY_PID)"
echo "✓ Web running (PID $WEB_PID)"
echo ""
echo "Press Ctrl-C to stop."
echo ""

wait
