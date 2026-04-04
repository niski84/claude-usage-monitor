package monitor

import (
	"log"
	"sync"
	"time"
)

const rateLimitCacheDuration = 30 * time.Minute

// Service holds config, the latest Stats snapshot, and runs a background poll goroutine.
type Service struct {
	cfg              Config
	mu               sync.RWMutex
	stats            Stats
	log              *log.Logger
	rlMu             sync.Mutex // guards rate limit fetch (one at a time)
	lastRLFetch      time.Time
}

// NewService creates a Service with the given config.
func NewService(cfg Config) *Service {
	return &Service{
		cfg: cfg,
		log: log.New(log.Writer(), "[claude-usage] ", log.LstdFlags),
	}
}

// Start performs an initial data load and launches the background poll loop.
func (s *Service) Start() {
	s.refresh()
	go s.RefreshRateLimits() // fetch live limits once on startup
	go func() {
		t := time.NewTicker(time.Duration(s.cfg.PollIntervalSec) * time.Second)
		for range t.C {
			s.refresh()
		}
	}()
}

// refresh re-reads all JSONL files and recomputes Stats.
func (s *Service) refresh() {
	records, err := CollectRecords(s.cfg.ClaudeDir)
	if err != nil {
		s.log.Printf("collect error: %v", err)
		return
	}
	stats := Aggregate(records)
	stats.AccountInfo = ReadAccountInfo(s.cfg.ClaudeDir)

	// Preserve existing rate limits (updated separately by RefreshRateLimits)
	s.mu.RLock()
	stats.RateLimits = s.stats.RateLimits
	s.mu.RUnlock()

	s.mu.Lock()
	s.stats = stats
	s.mu.Unlock()
	s.log.Printf("refreshed: %d records, total cost $%.4f", len(records), stats.TotalCostUSD)
}

// RefreshRateLimits fetches live rate limit percentages, at most once per 30 minutes.
// Concurrent calls block until the in-flight fetch completes, then return the cached value.
func (s *Service) RefreshRateLimits() {
	s.rlMu.Lock()
	defer s.rlMu.Unlock()

	if time.Since(s.lastRLFetch) < rateLimitCacheDuration {
		s.log.Printf("rate limits: using cached value (next fetch in %s)",
			(rateLimitCacheDuration - time.Since(s.lastRLFetch)).Round(time.Minute))
		return
	}

	s.log.Printf("fetching live rate limits...")
	rl := FetchRateLimits(s.cfg.ClaudeDir)
	s.lastRLFetch = time.Now()
	s.mu.Lock()
	s.stats.RateLimits = rl
	s.mu.Unlock()
	if rl.Available {
		s.log.Printf("rate limits: 5h=%.0f%% 7d=%.0f%%", rl.FiveHourPct, rl.SevenDayPct)
	}
}

// Stats returns a copy of the latest aggregated stats (safe for concurrent use).
func (s *Service) Stats() Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stats
}
