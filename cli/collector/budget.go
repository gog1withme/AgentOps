package collector

import (
	"os"
	"strconv"
	"time"

	"github.com/gog1withme/AgentOps/cli/internal/config"
	"github.com/gog1withme/AgentOps/cli/internal/platform"
	"github.com/gog1withme/AgentOps/cli/store"
)

type BudgetViolation struct {
	Rule    string  `json:"rule"`
	Current float64 `json:"current"`
	Limit   float64 `json:"limit"`
	Action  string  `json:"action"`
	Pct     float64 `json:"pct,omitempty"`
}

type BudgetChecker struct {
	store  *store.Store
	cfg    *config.Config
	since  time.Time
	warned map[string]bool
}

func NewBudgetChecker(st *store.Store, cfg *config.Config) *BudgetChecker {
	action := cfg.Budget.Action
	if v := os.Getenv("AGENTOPS_BUDGET_ACTION"); v != "" {
		action = v
	}
	if action == "" {
		action = "alert"
	}
	cfg.Budget.Action = action
	if v := os.Getenv("AGENTOPS_BUDGET_COST"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.Budget.CostLimitUSD = f
		}
	}
	return &BudgetChecker{
		store:  st,
		cfg:    cfg,
		since:  time.Now().Add(-24 * time.Hour),
		warned: make(map[string]bool),
	}
}

func (b *BudgetChecker) Check(sessionID string) *BudgetViolation {
	costLimit := b.cfg.Budget.CostLimitUSD
	toolLimit := b.cfg.Budget.ToolCallLimit
	llmLimit := b.cfg.Budget.LLMCallLimit
	if costLimit <= 0 && toolLimit <= 0 && llmLimit <= 0 {
		return nil
	}
	cost, tools, llms, err := b.store.SessionSpend(sessionID, b.since)
	if err != nil {
		return nil
	}
	action := b.cfg.Budget.Action
	if action == "" {
		action = "alert"
	}
	checks := []struct {
		rule    string
		current float64
		limit   float64
	}{
		{"cost_limit", cost, costLimit},
		{"tool_call_limit", float64(tools), float64(toolLimit)},
		{"llm_call_limit", float64(llms), float64(llmLimit)},
	}
	for _, c := range checks {
		if c.limit <= 0 {
			continue
		}
		pct := c.current / c.limit * 100
		if pct >= 80 && pct < 100 {
			key := sessionID + ":" + c.rule + ":warn"
			if !b.warned[key] {
				b.warned[key] = true
				return &BudgetViolation{Rule: c.rule, Current: c.current, Limit: c.limit, Action: "alert", Pct: pct}
			}
		}
		if pct >= 100 {
			return &BudgetViolation{Rule: c.rule, Current: c.current, Limit: c.limit, Action: action, Pct: pct}
		}
	}
	return nil
}

func (b *BudgetChecker) Pause(pid int) error {
	return platform.PauseProcess(pid)
}

func (b *BudgetChecker) Kill(pid int) error {
	return platform.KillProcess(pid)
}

func (b *BudgetChecker) Alert(v *BudgetViolation) {
	_ = v
}
