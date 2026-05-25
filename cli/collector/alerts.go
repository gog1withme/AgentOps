package collector

import (
	"strings"
	"time"

	"github.com/gog1withme/AgentOps/cli/internal/config"
	"github.com/gog1withme/AgentOps/cli/store"
	"github.com/gog1withme/AgentOps/schema"
)

type AlertsEngine struct {
	store *store.Store
	cfg   *config.Config
	seen  map[string]time.Time
}

func NewAlertsEngine(st *store.Store, cfg *config.Config) *AlertsEngine {
	return &AlertsEngine{store: st, cfg: cfg, seen: make(map[string]time.Time)}
}

func (a *AlertsEngine) Check(e *schema.Event) *schema.AlertRecord {
	if e == nil {
		return nil
	}
	var alert *schema.AlertRecord
	if e.Type == schema.EventShellCmd {
		alert = a.checkDangerousShell(e)
	}
	if e.Type == schema.EventLLMCall {
		if e.PromptTokens > 100000 {
			alert = a.makeAlert(e.SessionID, "token_explosion", "warning", "Token explosion: prompt exceeds 100k tokens")
		}
		if e.EfficiencyScore > 0 && e.EfficiencyScore < 30 {
			alert = a.makeAlert(e.SessionID, "low_efficiency", "warning", "Low token efficiency on LLM call")
		}
	}
	if e.Type == schema.EventMCPCall && e.MCPLatencyMS > 2000 {
		alert = a.makeAlert(e.SessionID, "mcp_slow", "warning", "MCP server "+e.MCPServer+" latency > 2s")
	}
	if alert != nil {
		_ = a.store.WriteAlert(alert)
	}
	return alert
}

func (a *AlertsEngine) checkDangerousShell(e *schema.Event) *schema.AlertRecord {
	cmd := strings.ToLower(e.ShellCommand)
	dangerous := []string{"rm -rf /", "format c:", "mkfs", ":(){ :|:& };:", "dd if=/dev/zero"}
	for _, d := range dangerous {
		if strings.Contains(cmd, d) {
			return a.makeAlert(e.SessionID, "dangerous_shell", "critical", "Dangerous shell command detected")
		}
	}
	return nil
}

func (a *AlertsEngine) makeAlert(sessionID, rule, severity, message string) *schema.AlertRecord {
	key := sessionID + ":" + rule
	if t, ok := a.seen[key]; ok && time.Since(t) < 5*time.Minute {
		return nil
	}
	a.seen[key] = time.Now()
	return &schema.AlertRecord{
		ID:        store.NewEventID(),
		SessionID: sessionID,
		Rule:      rule,
		Severity:  severity,
		Message:   message,
		Timestamp: time.Now(),
	}
}
