package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/niski84/claude-usage-monitor/internal/types"
)

// FetchRateLimits runs a minimal `claude --print` session with a temporary statusline
// command that writes the status JSON (including rate_limits) to a temp file.
// Returns zero-value RateLimits if unavailable.
func FetchRateLimits(claudeDir string) types.RateLimits {
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("claude_rl_%d.json", os.Getpid()))
	defer os.Remove(tmpFile)

	settingsPath := filepath.Join(claudeDir, "settings.json")

	// Read current settings
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return types.RateLimits{}
	}
	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		return types.RateLimits{}
	}

	// Inject statusline command
	origStatusLine, hadStatusLine := settings["statusLine"]
	settings["statusLine"] = map[string]any{
		"type":    "command",
		"command": fmt.Sprintf(`input=$(cat); echo "$input" > %s; echo ""`, tmpFile),
	}
	patched, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return types.RateLimits{}
	}
	if err := os.WriteFile(settingsPath, patched, 0644); err != nil {
		return types.RateLimits{}
	}

	// Restore settings when done
	defer func() {
		if hadStatusLine {
			settings["statusLine"] = origStatusLine
		} else {
			delete(settings, "statusLine")
		}
		restored, _ := json.MarshalIndent(settings, "", "  ")
		_ = os.WriteFile(settingsPath, restored, 0644)
	}()

	// Run a minimal claude session
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "claude", "--print", "--model", "haiku", "--dangerously-skip-permissions")
	cmd.Stdin = strings.NewReader("hi")
	_ = cmd.Run()

	// Read the captured status JSON
	out, err := os.ReadFile(tmpFile)
	if err != nil {
		return types.RateLimits{}
	}

	var status struct {
		RateLimits struct {
			FiveHour *struct {
				UsedPercentage float64 `json:"used_percentage"`
				ResetsAt       int64   `json:"resets_at"`
			} `json:"five_hour"`
			SevenDay *struct {
				UsedPercentage float64 `json:"used_percentage"`
				ResetsAt       int64   `json:"resets_at"`
			} `json:"seven_day"`
		} `json:"rate_limits"`
	}
	if err := json.Unmarshal(out, &status); err != nil {
		return types.RateLimits{}
	}

	rl := types.RateLimits{}
	if h := status.RateLimits.FiveHour; h != nil {
		rl.FiveHourPct = h.UsedPercentage
		rl.FiveHourResetsAt = time.Unix(h.ResetsAt, 0)
		rl.Available = true
	}
	if w := status.RateLimits.SevenDay; w != nil {
		rl.SevenDayPct = w.UsedPercentage
		rl.SevenDayResetsAt = time.Unix(w.ResetsAt, 0)
		rl.Available = true
	}
	return rl
}
