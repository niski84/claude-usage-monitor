// Package monitor provides JSONL parsing, aggregation, and HTTP serving
// for Claude Code CLI usage data. Shared data types live in internal/types.
package monitor

// Re-export types from internal/types for convenient use within this package.
// Callers outside this package should import internal/types directly.
import "github.com/niski84/claude-usage-monitor/internal/types"

// Type aliases so the rest of the monitor package can use short names.
type UsageRecord = types.UsageRecord
type Stats = types.Stats
type ModelStat = types.ModelStat
type ProjectStat = types.ProjectStat
type DailyStat = types.DailyStat
type SessionStat = types.SessionStat
