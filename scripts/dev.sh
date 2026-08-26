#!/usr/bin/env bash
set -euo pipefail

trap 'kill 0 2>/dev/null || true' EXIT INT TERM

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

echo "🚀 coop development server"
echo "  relay: http://localhost:8787"
echo "  web:   http://localhost:3000"
echo ""
echo "Stopping any existing processes..."
pkill -f "go run.*relay|next dev" 2>/dev/null || true
sleep 1

echo "Starting relay and web..."
(cd apps/relay && go run ./cmd/relay) &
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
