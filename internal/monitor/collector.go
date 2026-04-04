package monitor

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type rawEntry struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	SessionID string `json:"sessionId"`
	CWD       string `json:"cwd"`
	Message   *struct {
		Model string `json:"model"`
		Usage *struct {
			InputTokens              int `json:"input_tokens"`
			OutputTokens             int `json:"output_tokens"`
			CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
			CacheReadInputTokens     int `json:"cache_read_input_tokens"`
			ServerToolUse            *struct {
				WebSearchRequests int `json:"web_search_requests"`
			} `json:"server_tool_use"`
		} `json:"usage"`
	} `json:"message"`
}

// CollectRecords walks claudeDir/projects/**/*.jsonl and returns all assistant usage records.
func CollectRecords(claudeDir string) ([]UsageRecord, error) {
	projectsDir := filepath.Join(claudeDir, "projects")
	var records []UsageRecord

	err := filepath.WalkDir(projectsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".jsonl") {
			return nil
		}
		recs, _ := parseJSONL(path)
		records = append(records, recs...)
		return nil
	})
	return records, err
}

func parseJSONL(path string) ([]UsageRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var records []UsageRecord
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 4*1024*1024), 4*1024*1024)
	for sc.Scan() {
		var e rawEntry
		if err := json.Unmarshal(sc.Bytes(), &e); err != nil {
			continue
		}
		if e.Type != "assistant" || e.Message == nil || e.Message.Usage == nil {
			continue
		}
		u := e.Message.Usage
		ts, _ := time.Parse(time.RFC3339, e.Timestamp)
		if ts.IsZero() {
			ts, _ = time.Parse(time.RFC3339Nano, e.Timestamp)
		}
		project := projectName(e.CWD)
		searches := 0
		if u.ServerToolUse != nil {
			searches = u.ServerToolUse.WebSearchRequests
		}
		r := UsageRecord{
			Timestamp:        ts,
			SessionID:        e.SessionID,
			Model:            e.Message.Model,
			Project:          project,
			CWD:              e.CWD,
			InputTokens:      u.InputTokens,
			OutputTokens:     u.OutputTokens,
			CacheWriteTokens: u.CacheCreationInputTokens,
			CacheReadTokens:  u.CacheReadInputTokens,
			WebSearches:      searches,
		}
		r.CostUSD = CostForRecord(&r)
		records = append(records, r)
	}
	return records, sc.Err()
}

func projectName(cwd string) string {
	if cwd == "" {
		return "unknown"
	}
	return filepath.Base(cwd)
}
