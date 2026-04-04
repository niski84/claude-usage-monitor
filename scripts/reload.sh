#!/bin/bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BINARY="$PROJECT_DIR/claude-usage-monitor"

echo "=== Claude Usage Monitor reload ==="
echo "→ Stopping existing process..."
pkill -f "$BINARY" 2>/dev/null && sleep 1 || echo "  (none running)"

if [ -f "$PROJECT_DIR/.env" ]; then
    set -a; source "$PROJECT_DIR/.env"; set +a
fi

echo "→ Generating templ..."
cd "$PROJECT_DIR"
~/go/bin/templ generate ./internal/monitor/views/ 2>/dev/null || true

echo "→ Building..."
if ! go build -o "$BINARY" ./cmd/claude-usage-monitor/; then
    echo "✗ Build failed"
    exit 1
fi
echo "  Build OK"

PORT="${PORT:-8098}"
export PORT
echo "→ Starting on :${PORT}..."
nohup env PORT="$PORT" "$BINARY" > "$PROJECT_DIR/claude-usage-monitor.log" 2>&1 &
echo $! > "$PROJECT_DIR/claude-usage-monitor.pid"

for i in $(seq 1 30); do
    sleep 0.3
    if curl -fsS "http://127.0.0.1:${PORT}/api/health" >/dev/null 2>&1; then
        echo "✓ Claude Usage Monitor at http://127.0.0.1:${PORT}/"
        exit 0
    fi
done
echo "✗ Server did not respond — check claude-usage-monitor.log"
exit 1
