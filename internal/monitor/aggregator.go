package monitor

import (
	"fmt"
	"sort"
	"time"
)

// Aggregate computes Stats from all raw usage records.
func Aggregate(records []UsageRecord) Stats {
	now := time.Now().UTC()
	todayMidnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	weekAgo := now.Add(-7 * 24 * time.Hour)
	monthAgo := now.Add(-30 * 24 * time.Hour)

	s := Stats{
		UpdatedAt: now,
	}

	// Per-model accumulators: model -> {cost, tokens}
	type modelAcc struct {
		cost   float64
		tokens int64
	}
	modelMap := make(map[string]*modelAcc)

	// Per-project accumulators: project -> {cost, sessions set, tokens}
	type projectAcc struct {
		cost     float64
		tokens   int64
		sessions map[string]struct{}
	}
	projectMap := make(map[string]*projectAcc)

	// Daily accumulators: date string -> {cost, tokens}
	type dailyAcc struct {
		cost   float64
		tokens int64
	}
	dailyMap := make(map[string]*dailyAcc)

	// Session accumulators: sessionID -> sessionAcc
	type sessionAcc struct {
		project    string
		modelCounts map[string]int
		cost       float64
		tokens     int64
		minTime    time.Time
		maxTime    time.Time
	}
	sessionMap := make(map[string]*sessionAcc)

	// Track unique session IDs per time window
	todaySessions := make(map[string]struct{})
	weekSessions := make(map[string]struct{})
	allSessions := make(map[string]struct{})

	for _, r := range records {
		tokens := int64(r.InputTokens + r.OutputTokens + r.CacheWriteTokens + r.CacheReadTokens)

		// Global totals
		s.TotalCostUSD += r.CostUSD
		s.TotalInputTokens += int64(r.InputTokens)
		s.TotalOutputTokens += int64(r.OutputTokens)
		s.TotalCacheWrite += int64(r.CacheWriteTokens)
		s.TotalCacheRead += int64(r.CacheReadTokens)

		// Time-window costs
		if !r.Timestamp.IsZero() {
			if r.Timestamp.After(todayMidnight) || r.Timestamp.Equal(todayMidnight) {
				s.TodayCostUSD += r.CostUSD
				todaySessions[r.SessionID] = struct{}{}
			}
			if r.Timestamp.After(weekAgo) {
				s.WeekCostUSD += r.CostUSD
				weekSessions[r.SessionID] = struct{}{}
			}
			if r.Timestamp.After(monthAgo) {
				s.MonthCostUSD += r.CostUSD
			}
		}

		// All sessions
		allSessions[r.SessionID] = struct{}{}

		// Per-model
		if _, ok := modelMap[r.Model]; !ok {
			modelMap[r.Model] = &modelAcc{}
		}
		modelMap[r.Model].cost += r.CostUSD
		modelMap[r.Model].tokens += tokens

		// Per-project
		if _, ok := projectMap[r.Project]; !ok {
			projectMap[r.Project] = &projectAcc{sessions: make(map[string]struct{})}
		}
		projectMap[r.Project].cost += r.CostUSD
		projectMap[r.Project].tokens += tokens
		projectMap[r.Project].sessions[r.SessionID] = struct{}{}

		// Daily (last 30 days only)
		if !r.Timestamp.IsZero() && r.Timestamp.After(monthAgo) {
			day := r.Timestamp.UTC().Format("2006-01-02")
			if _, ok := dailyMap[day]; !ok {
				dailyMap[day] = &dailyAcc{}
			}
			dailyMap[day].cost += r.CostUSD
			dailyMap[day].tokens += tokens
		}

		// Per-session
		if _, ok := sessionMap[r.SessionID]; !ok {
			sessionMap[r.SessionID] = &sessionAcc{
				modelCounts: make(map[string]int),
				minTime:     r.Timestamp,
				maxTime:     r.Timestamp,
			}
		}
		sa := sessionMap[r.SessionID]
		sa.cost += r.CostUSD
		sa.tokens += tokens
		sa.modelCounts[r.Model]++
		sa.project = r.Project // last seen project
		if !r.Timestamp.IsZero() {
			if r.Timestamp.Before(sa.minTime) || sa.minTime.IsZero() {
				sa.minTime = r.Timestamp
			}
			if r.Timestamp.After(sa.maxTime) {
				sa.maxTime = r.Timestamp
			}
		}
	}

	// Session counts
	s.Sessions = len(allSessions)
	s.TodaySessions = len(todaySessions)
	s.WeekSessions = len(weekSessions)

	// Cache hit rate
	denom := s.TotalInputTokens + s.TotalCacheRead
	if denom > 0 {
		s.CacheHitRate = float64(s.TotalCacheRead) / float64(denom) * 100
	}

	// ByModel: sort by cost desc, compute percent
	for model, acc := range modelMap {
		pct := 0.0
		if s.TotalCostUSD > 0 {
			pct = acc.cost / s.TotalCostUSD * 100
		}
		s.ByModel = append(s.ByModel, ModelStat{
			Model:   model,
			CostUSD: acc.cost,
			Tokens:  acc.tokens,
			Percent: pct,
		})
	}
	sort.Slice(s.ByModel, func(i, j int) bool {
		return s.ByModel[i].CostUSD > s.ByModel[j].CostUSD
	})

	// ByProject: sort by cost desc, top 10
	for project, acc := range projectMap {
		s.ByProject = append(s.ByProject, ProjectStat{
			Project:  project,
			CostUSD:  acc.cost,
			Sessions: len(acc.sessions),
			Tokens:   acc.tokens,
		})
	}
	sort.Slice(s.ByProject, func(i, j int) bool {
		return s.ByProject[i].CostUSD > s.ByProject[j].CostUSD
	})
	if len(s.ByProject) > 10 {
		s.ByProject = s.ByProject[:10]
	}

	// DailyTotals: last 30 days, one entry per day, sorted ascending
	for i := 29; i >= 0; i-- {
		day := now.Add(-time.Duration(i) * 24 * time.Hour).Format("2006-01-02")
		cost := 0.0
		var tokens int64
		if acc, ok := dailyMap[day]; ok {
			cost = acc.cost
			tokens = acc.tokens
		}
		s.DailyTotals = append(s.DailyTotals, DailyStat{
			Date:    day,
			CostUSD: cost,
			Tokens:  tokens,
		})
	}

	// RecentSessions: last 20 unique sessions sorted by maxTime desc
	type sessionEntry struct {
		id  string
		acc *sessionAcc
	}
	var sessionEntries []sessionEntry
	for id, acc := range sessionMap {
		sessionEntries = append(sessionEntries, sessionEntry{id: id, acc: acc})
	}
	sort.Slice(sessionEntries, func(i, j int) bool {
		return sessionEntries[i].acc.maxTime.After(sessionEntries[j].acc.maxTime)
	})
	if len(sessionEntries) > 20 {
		sessionEntries = sessionEntries[:20]
	}
	for _, se := range sessionEntries {
		acc := se.acc
		// find most common model
		topModel := ""
		topCount := 0
		for m, cnt := range acc.modelCounts {
			if cnt > topCount {
				topCount = cnt
				topModel = m
			}
		}
		s.RecentSessions = append(s.RecentSessions, SessionStat{
			SessionID:  se.id,
			Project:    acc.project,
			Model:      topModel,
			CostUSD:    acc.cost,
			Tokens:     acc.tokens,
			StartedAt:  acc.minTime,
			StartedRel: relativeTime(acc.minTime, now),
		})
	}

	return s
}

// relativeTime returns a human-readable relative time string.
func relativeTime(t, now time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	d := now.Sub(t)
	if d < 0 {
		d = -d
	}
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		mins := int(d.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case d < 24*time.Hour:
		hrs := int(d.Hours())
		if hrs == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hrs)
	case d < 7*24*time.Hour:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "yesterday"
		}
		return fmt.Sprintf("%d days ago", days)
	case d < 30*24*time.Hour:
		weeks := int(d.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	default:
		months := int(d.Hours() / 24 / 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	}
}
