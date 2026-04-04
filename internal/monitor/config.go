package monitor

import (
	"os"
	"path/filepath"
)

// Config holds runtime configuration for the service.
type Config struct {
	Port            string
	ClaudeDir       string // default ~/.claude
	DataDir         string
	PollIntervalSec int
}

// LoadConfig reads config from environment variables with sensible defaults.
func LoadConfig() Config {
	home, _ := os.UserHomeDir()
	return Config{
		Port:            getenv("PORT", "8098"),
		ClaudeDir:       getenv("CLAUDE_DIR", filepath.Join(home, ".claude")),
		DataDir:         getenv("DATA_DIR", "data"),
		PollIntervalSec: 300, // 5 minutes
	}
}

func getenv(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}
