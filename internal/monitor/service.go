package monitor

import (
	"log"
	"sync"
	"time"
)

// Service holds config, the latest Stats snapshot, and runs a background poll goroutine.
type Service struct {
	cfg   Config
	mu    sync.RWMutex
	stats Stats
	log   *log.Logger
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
	s.mu.Lock()
	s.stats = stats
	s.mu.Unlock()
	s.log.Printf("refreshed: %d records, total cost $%.4f", len(records), stats.TotalCostUSD)
}

// Stats returns a copy of the latest aggregated stats (safe for concurrent use).
func (s *Service) Stats() Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stats
}
