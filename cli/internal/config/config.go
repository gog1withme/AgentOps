package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gog1withme/AgentOps/cli/internal/paths"
	"github.com/oklog/ulid/v2"
)

type BudgetConfig struct {
	CostLimitUSD  float64 `json:"cost_limit_usd"`
	ToolCallLimit int     `json:"tool_call_limit"`
	LLMCallLimit  int     `json:"llm_call_limit"`
	Action        string  `json:"action"`
}

type Config struct {
	Version           string       `json:"version"`
	SessionID         string       `json:"session_id"`
	DetectedTools     []string     `json:"detected_tools"`
	ProxyPort         int          `json:"proxy_port"`
	CollectorPort     int          `json:"collector_port"`
	HooksInstalled    bool         `json:"hooks_installed"`
	DataDir           string       `json:"data_dir"`
	SnapshotsDir      string       `json:"snapshots_dir"`
	ScrubPatternsFile string       `json:"scrub_patterns_file"`
	Budget            BudgetConfig `json:"budget"`
	WorkDir           string       `json:"work_dir,omitempty"`
}

func Default() *Config {
	return &Config{
		Version:           "1.0.0",
		SessionID:         ulid.Make().String(),
		ProxyPort:         4318,
		CollectorPort:     4317,
		DataDir:           paths.DataDir(),
		SnapshotsDir:      paths.SnapshotsDir(),
		ScrubPatternsFile: paths.ScrubPatternsPath(),
		Budget: BudgetConfig{
			Action: "alert",
		},
	}
}

func Load() (*Config, error) {
	path := paths.ConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return nil, err
	}
	cfg := Default()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func Save(cfg *Config) error {
	if err := paths.EnsureDirs(); err != nil {
		return err
	}
	if cfg.SessionID == "" {
		cfg.SessionID = ulid.Make().String()
	}
	cfg.DataDir = paths.DataDir()
	cfg.SnapshotsDir = paths.SnapshotsDir()
	cfg.ScrubPatternsFile = paths.ScrubPatternsPath()
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(paths.ConfigPath(), data, 0o644)
}

func Port() int {
	if v := os.Getenv("AGENTOPS_PORT"); v != "" {
		var p int
		if _, err := fmt.Sscanf(v, "%d", &p); err == nil && p > 0 {
			return p
		}
	}
	cfg, err := Load()
	if err == nil && cfg.ProxyPort > 0 {
		return cfg.ProxyPort
	}
	return 4318
}

func SessionStart() time.Time {
	return time.Now().Add(-24 * time.Hour)
}
