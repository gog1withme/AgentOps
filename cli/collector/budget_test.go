package collector

import (
	"testing"
	"time"

	"github.com/gog1withme/AgentOps/cli/internal/config"
	"github.com/gog1withme/AgentOps/cli/store"
	"github.com/gog1withme/AgentOps/schema"
)

func testStore(t *testing.T) *store.Store {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("AGENTOPS_DATA_DIR", dir)
	st, err := store.Open()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st
}

func writeCostEvent(t *testing.T, st *store.Store, sessionID string, cost float64) {
	t.Helper()
	err := st.WriteEvent(&schema.Event{
		ID:        store.NewEventID(),
		SessionID: sessionID,
		Timestamp: time.Now(),
		Source:    "test",
		Type:      schema.EventLLMCall,
		CostUSD:   cost,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestBudgetCheckerWarningAt80Percent(t *testing.T) {
	st := testStore(t)
	sessionID := "sess-warn"
	cfg := config.Default()
	cfg.SessionID = sessionID
	cfg.Budget.CostLimitUSD = 1.0
	cfg.Budget.Action = "alert"

	writeCostEvent(t, st, sessionID, 0.85)

	checker := NewBudgetChecker(st, cfg)
	v := checker.Check(sessionID)
	if v == nil {
		t.Fatal("expected 80% warning violation")
	}
	if v.Action != "alert" {
		t.Fatalf("expected alert action, got %q", v.Action)
	}
	if v.Pct < 80 || v.Pct >= 100 {
		t.Fatalf("expected pct in [80,100), got %v", v.Pct)
	}

	// Second check should not repeat the same warning.
	if checker.Check(sessionID) != nil {
		t.Fatal("expected no duplicate 80% warning")
	}
}

func TestBudgetCheckerLimitAt100Percent(t *testing.T) {
	st := testStore(t)
	sessionID := "sess-limit"
	cfg := config.Default()
	cfg.SessionID = sessionID
	cfg.Budget.CostLimitUSD = 1.0
	cfg.Budget.Action = "pause"

	writeCostEvent(t, st, sessionID, 1.05)

	checker := NewBudgetChecker(st, cfg)
	v := checker.Check(sessionID)
	if v == nil {
		t.Fatal("expected 100% violation")
	}
	if v.Action != "pause" {
		t.Fatalf("expected pause action, got %q", v.Action)
	}
	if v.Pct < 100 {
		t.Fatalf("expected pct >= 100, got %v", v.Pct)
	}
}

func TestBudgetCheckerNoLimits(t *testing.T) {
	st := testStore(t)
	cfg := config.Default()
	checker := NewBudgetChecker(st, cfg)
	if v := checker.Check("any"); v != nil {
		t.Fatalf("expected nil when no limits configured, got %+v", v)
	}
}
