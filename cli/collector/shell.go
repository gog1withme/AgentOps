package collector

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gog1withme/AgentOps/cli/store"
	"github.com/gog1withme/AgentOps/schema"
)

type ShellIngest struct {
	Command  string `json:"command"`
	ExitCode int    `json:"exit_code"`
	Source   string `json:"source"`
}

type ShellHandler struct {
	collector *Collector
}

func NewShellHandler(c *Collector) *ShellHandler {
	return &ShellHandler{collector: c}
}

func (h *ShellHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var payload ShellIngest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	source := payload.Source
	if source == "" {
		source = "shell"
	}
	h.collector.Ingest(schema.Event{
		ID:            store.NewEventID(),
		SessionID:     h.collector.cfg.SessionID,
		Timestamp:     time.Now(),
		Source:        source,
		Type:          schema.EventShellCmd,
		ShellCommand:  payload.Command,
		ShellExitCode: payload.ExitCode,
	})
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok":true}`))
}
