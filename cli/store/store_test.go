package store

import (
	"math"
	"testing"
	"time"

	"github.com/gog1withme/AgentOps/schema"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("AGENTOPS_DATA_DIR", dir)
	st, err := Open()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st
}

func TestMigrateAndRoundtrip(t *testing.T) {
	st := openTestStore(t)
	sessionID := "sess-roundtrip"
	now := time.Now().UTC().Truncate(time.Millisecond)

	e := &schema.Event{
		ID:           NewEventID(),
		SessionID:    sessionID,
		Timestamp:    now,
		Source:       "test",
		Type:         schema.EventShellCmd,
		ShellCommand: "echo hello",
		CostUSD:      0.42,
	}
	if err := st.WriteEvent(e); err != nil {
		t.Fatal(err)
	}

	got, err := st.GetEvent(e.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ShellCommand != e.ShellCommand {
		t.Fatalf("shell command mismatch: %q vs %q", got.ShellCommand, e.ShellCommand)
	}

	cost, tools, llms, err := st.SessionSpend(sessionID, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(cost-0.42) > 1e-6 {
		t.Fatalf("expected cost 0.42, got %v", cost)
	}
	if tools != 0 || llms != 0 {
		t.Fatalf("expected zero tool/llm counts, got tools=%d llms=%d", tools, llms)
	}
}

func TestBackendName(t *testing.T) {
	name := BackendName()
	if name != "sqlite" && name != "duckdb" {
		t.Fatalf("unexpected backend name %q", name)
	}
}
