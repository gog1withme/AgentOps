package schema

import "time"

type EventType string

const (
	EventLLMCall     EventType = "llm_call"
	EventToolCall    EventType = "tool_call"
	EventFileEdit    EventType = "file_edit"
	EventShellCmd    EventType = "shell_command"
	EventMCPCall     EventType = "mcp_call"
	EventBudgetAlert EventType = "budget_alert"
	EventRestore     EventType = "restore"
	EventConflict    EventType = "agent_conflict"
	EventAlert       EventType = "alert"
)

type Event struct {
	ID              string            `json:"id"`
	SessionID       string            `json:"session_id"`
	RunID           string            `json:"run_id,omitempty"`
	Timestamp       time.Time         `json:"timestamp"`
	Source          string            `json:"source"`
	Type            EventType         `json:"type"`
	Model           string            `json:"model,omitempty"`
	PromptHash      string            `json:"prompt_hash,omitempty"`
	PromptTokens    int               `json:"prompt_tokens,omitempty"`
	OutputTokens    int               `json:"output_tokens,omitempty"`
	EfficiencyScore float64           `json:"efficiency_score,omitempty"`
	CostUSD         float64           `json:"cost_usd,omitempty"`
	DurationMS      int               `json:"duration_ms,omitempty"`
	ToolName        string            `json:"tool_name,omitempty"`
	ToolInput       string            `json:"tool_input,omitempty"`
	ToolOutput      string            `json:"tool_output,omitempty"`
	FilePath        string            `json:"file_path,omitempty"`
	FileDiff        string            `json:"file_diff,omitempty"`
	ShellCommand    string            `json:"shell_command,omitempty"`
	ShellExitCode   int               `json:"shell_exit_code,omitempty"`
	MCPServer       string            `json:"mcp_server,omitempty"`
	MCPLatencyMS    int               `json:"mcp_latency_ms,omitempty"`
	Error           string            `json:"error,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

type Prompt struct {
	Hash                string    `json:"hash"`
	Content             string    `json:"content"`
	FirstSeen           time.Time `json:"first_seen"`
	LastSeen            time.Time `json:"last_seen"`
	SessionCount        int       `json:"session_count"`
	AvgCostUSD          float64   `json:"avg_cost_usd"`
	AvgPromptTokens     int       `json:"avg_prompt_tokens"`
	AvgEfficiencyScore  float64   `json:"avg_efficiency_score"`
}

type Attribution struct {
	ID           string    `json:"id"`
	SessionID    string    `json:"session_id"`
	RunID        string    `json:"run_id,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
	FilePath     string    `json:"file_path"`
	Agent        string    `json:"agent"`
	LinesAdded   int       `json:"lines_added"`
	LinesRemoved int       `json:"lines_removed"`
	Diff         string    `json:"diff"`
	EventID      string    `json:"event_id"`
}

type MCPServer struct {
	Name              string    `json:"name"`
	URL               string    `json:"url,omitempty"`
	FirstSeen         time.Time `json:"first_seen"`
	LastSeen          time.Time `json:"last_seen"`
	TotalCalls        int       `json:"total_calls"`
	ErrorCount        int       `json:"error_count"`
	AvgLatencyMS      float64   `json:"avg_latency_ms"`
	P95LatencyMS      float64   `json:"p95_latency_ms"`
	AvgResponseBytes  int       `json:"avg_response_bytes"`
}

type Budget struct {
	SessionID      string     `json:"session_id"`
	CostLimitUSD   float64    `json:"cost_limit_usd"`
	ToolCallLimit  int        `json:"tool_call_limit"`
	LLMCallLimit   int        `json:"llm_call_limit"`
	Action         string     `json:"action"`
	CreatedAt      time.Time  `json:"created_at"`
	TriggeredAt    *time.Time `json:"triggered_at,omitempty"`
}

type Snapshot struct {
	ID           string    `json:"id"`
	SessionID    string    `json:"session_id"`
	Timestamp    time.Time `json:"timestamp"`
	Trigger      string    `json:"trigger"`
	SnapshotPath string    `json:"snapshot_path"`
	FileCount    int       `json:"file_count"`
	SizeBytes    int64     `json:"size_bytes"`
}

type SessionSummary struct {
	SessionID   string    `json:"session_id"`
	FirstEvent  time.Time `json:"first_event"`
	LastEvent   time.Time `json:"last_event"`
	EventCount  int       `json:"event_count"`
	TotalCost   float64   `json:"total_cost"`
	TotalTokens int       `json:"total_tokens"`
}

type AlertRecord struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Rule      string    `json:"rule"`
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Dismissed bool      `json:"dismissed"`
}
