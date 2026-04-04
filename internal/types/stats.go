// Package types holds shared data types used by both the monitor and views packages.
package types

import "time"

// UsageRecord is one parsed assistant message with token usage.
type UsageRecord struct {
	Timestamp        time.Time
	SessionID        string
	Model            string
	Project          string // last segment of cwd path
	CWD              string
	InputTokens      int
	OutputTokens     int
	CacheWriteTokens int
	CacheReadTokens  int
	WebSearches      int
	CostUSD          float64
}

// Stats is the aggregated dashboard data computed from all records.
type Stats struct {
	UpdatedAt time.Time

	TotalCostUSD  float64
	TodayCostUSD  float64
	WeekCostUSD   float64
	MonthCostUSD  float64

	TotalInputTokens  int64
	TotalOutputTokens int64
	TotalCacheWrite   int64
	TotalCacheRead    int64

	CacheHitRate float64 // cache_read / (input + cache_read) * 100

	Sessions      int
	TodaySessions int
	WeekSessions  int

	ByModel     []ModelStat
	ByProject   []ProjectStat
	DailyTotals []DailyStat // last 30 days

	// Recent sessions for table display
	RecentSessions []SessionStat
}

// ModelStat holds per-model aggregated data.
type ModelStat struct {
	Model   string
	CostUSD float64
	Tokens  int64
	Percent float64
}

// ProjectStat holds per-project aggregated data.
type ProjectStat struct {
	Project  string
	CostUSD  float64
	Sessions int
	Tokens   int64
}

// DailyStat holds daily aggregated data.
type DailyStat struct {
	Date    string  // "2006-01-02"
	CostUSD float64
	Tokens  int64
}

// SessionStat holds per-session aggregated data for display.
type SessionStat struct {
	SessionID  string
	Project    string
	Model      string
	CostUSD    float64
	Tokens     int64
	StartedAt  time.Time
	StartedRel string
}
