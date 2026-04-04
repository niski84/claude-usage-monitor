# claude-usage-monitor

**Purpose**: Parse Claude Code CLI JSONL session files and display token usage and cost statistics.

**Port**: 8098

**Stack**: Go + Echo + Templ + HTMX + Alpine + Tailwind 4 (Smash Deck pattern, dark-first)

## Key Paths

| Path | Description |
|---|---|
| `cmd/claude-usage-monitor/main.go` | Entry point, graceful shutdown |
| `internal/monitor/collector.go` | Walks `~/.claude/projects/**/*.jsonl`, parses records |
| `internal/monitor/pricing.go` | Hardcoded pricing table, `CostForRecord()` |
| `internal/monitor/aggregator.go` | `Aggregate()` — computes all Stats fields |
| `internal/monitor/service.go` | Background poll loop (5 min), `Stats()` accessor |
| `internal/monitor/http.go` | Echo routes: `/`, `/dashboard`, `/api/stats`, `/api/health`, `/api/refresh` |
| `internal/monitor/views/` | Templ components (dashboard UI) |
| `web/embed.go` | Embeds `web/monitor/static/` into binary |

## Data Source

`~/.claude/projects/**/*.jsonl` — each line is a JSON event. Only `type == "assistant"` lines with `message.usage` are parsed.

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8098` | HTTP listen port |
| `CLAUDE_DIR` | `~/.claude` | Root Claude config directory |
| `DATA_DIR` | `data` | Local data directory |
