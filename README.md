# Claude Usage Monitor

A web dashboard for tracking and analyzing Claude API token usage and costs from Claude Code CLI sessions.

## Overview

Claude Usage Monitor parses Claude Code CLI session logs and displays real-time statistics about:
- Token consumption (input/output tokens)
- API costs
- Model usage patterns
- Historical trends

The dashboard updates automatically every 5 minutes with fresh data from your local Claude session files.

## Features

- **Real-time Dashboard**: Dark-themed web interface showing token usage and costs
- **Cost Breakdown**: Detailed pricing information for different Claude models
- **Session Analysis**: Track token consumption across all Claude Code sessions
- **Auto-refresh**: Background polling updates statistics every 5 minutes
- **Dark Theme**: First-class dark mode support for comfortable viewing

## Stack

- **Backend**: Go + Echo framework
- **Frontend**: Templ, HTMX, Alpine.js, Tailwind CSS 4
- **Data Source**: Claude Code CLI JSONL session logs

## Quick Start

### Prerequisites

- Go 1.24 or later
- Claude Code CLI with active sessions

### Installation

```bash
git clone https://github.com/niski84/claude-usage-monitor.git
cd claude-usage-monitor
go build -o claude-usage-monitor ./cmd/claude-usage-monitor
```

### Running

```bash
./claude-usage-monitor
```

The web dashboard will be available at `http://localhost:8098`

## Configuration

Configure the monitor using environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8098` | HTTP server port |
| `CLAUDE_DIR` | `~/.claude` | Claude configuration directory |
| `DATA_DIR` | `data` | Local data directory |

Example:
```bash
PORT=8080 CLAUDE_DIR=$HOME/.claude ./claude-usage-monitor
```

## Architecture

### Key Components

| Component | File | Description |
|-----------|------|-------------|
| Collector | `internal/monitor/collector.go` | Walks Claude session files and parses JSONL records |
| Pricing | `internal/monitor/pricing.go` | Maintains pricing table and cost calculations |
| Aggregator | `internal/monitor/aggregator.go` | Computes token usage and cost statistics |
| Service | `internal/monitor/service.go` | Background polling service (5 min intervals) |
| HTTP Server | `internal/monitor/http.go` | Echo routes for dashboard and API endpoints |
| Views | `internal/monitor/views/` | Templ components for the web UI |

### Data Flow

1. Collector reads `~/.claude/projects/**/*.jsonl` files
2. Parses assistant messages with token usage information
3. Aggregator computes statistics (total tokens, costs, per-model breakdown)
4. Service exposes stats via HTTP endpoints
5. Dashboard UI renders real-time data

## API Endpoints

- `GET /` — Dashboard UI
- `GET /api/stats` — JSON statistics
- `GET /api/health` — Health check
- `POST /api/refresh` — Force immediate stats refresh

## Development

### Building with Templ

This project uses Templ for server-side components. Generate templates before building:

```bash
go generate ./...
go build -o claude-usage-monitor ./cmd/claude-usage-monitor
```

### Running in Development

```bash
go run ./cmd/claude-usage-monitor
```

## Data Source

The monitor reads from `~/.claude/projects/**/*.jsonl`, where each line is a JSON event. Only `type == "assistant"` records with usage statistics are parsed and aggregated.

## License

[Add your license here]

## Contributing

Contributions welcome! Please feel free to submit issues and pull requests.
