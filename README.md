# claude-usage-monitor

Local web dashboard that parses Claude Code session files and shows token usage and cost.

## What It Does

Claude Code writes one JSONL file per session under `~/.claude/projects/`. This service walks those files, extracts the `assistant` events that include token usage, multiplies through a hardcoded pricing table, and surfaces the totals on a small web dashboard.

A background goroutine re-reads the JSONL tree every five minutes (configurable) and keeps the latest snapshot in memory. The dashboard is server-rendered with Templ, with HTMX/Alpine for interactions, and the same data is exposed at `/api/stats` for scripting. There is no database; the JSONL files on disk are the source of truth.

Useful for keeping an eye on per-day spend, spotting runaway sessions, and seeing which projects are eating the most tokens.

## Tech Stack

- Go (single binary, embedded static assets)
- Echo HTTP framework
- Templ for server-rendered components
- HTMX + Alpine.js + Tailwind 4 (Smash Deck pattern, dark theme)

## Installation

```bash
git clone https://github.com/niski84/claude-usage-monitor.git
cd claude-usage-monitor
templ generate ./internal/monitor/views/
go build -o claude-usage-monitor ./cmd/claude-usage-monitor/
./claude-usage-monitor
```

Then open `http://127.0.0.1:8098/`. The repo's `scripts/reload.sh` does the same and tails the log.

## Configuration

Environment variables:

- `PORT` (default `8098`) - HTTP listen port
- `CLAUDE_DIR` (default `~/.claude`) - root Claude config directory
- `DATA_DIR` (default `data`) - local data directory

Endpoints:

- `GET /dashboard` - HTML dashboard
- `GET /api/stats` - current snapshot as JSON
- `GET /api/health` - liveness probe
- `POST /api/refresh` - force an immediate re-read of the JSONL tree

## License

MIT
