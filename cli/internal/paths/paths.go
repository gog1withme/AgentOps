package paths

import (
	"os"
	"path/filepath"
	"strings"
)

func HomeDir() string {
	if h, err := os.UserHomeDir(); err == nil {
		return h
	}
	return "."
}

func AgentOpsDir() string {
	if v := os.Getenv("AGENTOPS_DATA_DIR"); v != "" {
		return filepath.Dir(v)
	}
	return filepath.Join(HomeDir(), ".agentops")
}

func DataDir() string {
	if v := os.Getenv("AGENTOPS_DATA_DIR"); v != "" {
		return v
	}
	return filepath.Join(AgentOpsDir(), "data")
}

func SnapshotsDir() string {
	return filepath.Join(AgentOpsDir(), "snapshots")
}

func ConfigPath() string {
	return filepath.Join(AgentOpsDir(), "config.json")
}

func ScrubPatternsPath() string {
	if v := os.Getenv("AGENTOPS_SCRUB_PATTERNS"); v != "" {
		return v
	}
	return filepath.Join(AgentOpsDir(), "scrub_patterns.txt")
}

func DuckDBPath() string {
	return filepath.Join(DataDir(), "events.duckdb")
}

func Expand(path string) string {
	if path == "" {
		return path
	}
	if strings.HasPrefix(path, "~") {
		return filepath.Join(HomeDir(), strings.TrimPrefix(path, "~"))
	}
	return path
}

func EnsureDirs() error {
	dirs := []string{AgentOpsDir(), DataDir(), SnapshotsDir()}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	return nil
}
