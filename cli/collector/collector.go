package collector

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gog1withme/AgentOps/cli/internal/config"
	"github.com/gog1withme/AgentOps/cli/internal/platform"
	"github.com/gog1withme/AgentOps/cli/scrubber"
	"github.com/gog1withme/AgentOps/cli/store"
	"github.com/gog1withme/AgentOps/schema"
	"github.com/rs/zerolog/log"
)

type SSEMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type Collector struct {
	events    chan schema.Event
	store     *store.Store
	scrubber  *scrubber.Scrubber
	budget    *BudgetChecker
	alerts    *AlertsEngine
	cfg       *config.Config
	listeners []chan SSEMessage
	mu        sync.RWMutex
	done      chan struct{}
	onSnapshot func()
}

func New(st *store.Store, scrub *scrubber.Scrubber, cfg *config.Config) *Collector {
	c := &Collector{
		events:   make(chan schema.Event, 512),
		store:    st,
		scrubber: scrub,
		cfg:      cfg,
		done:     make(chan struct{}),
	}
	c.budget = NewBudgetChecker(st, cfg)
	c.alerts = NewAlertsEngine(st, cfg)
	go c.writeLoop()
	return c
}

func (c *Collector) SetSnapshotCallback(fn func()) {
	c.onSnapshot = fn
}

func (c *Collector) Ingest(e schema.Event) {
	select {
	case c.events <- e:
	default:
		log.Warn().Msg("event channel full, dropping event")
	}
}

func (c *Collector) Subscribe() chan SSEMessage {
	ch := make(chan SSEMessage, 64)
	c.mu.Lock()
	c.listeners = append(c.listeners, ch)
	c.mu.Unlock()
	return ch
}

func (c *Collector) Unsubscribe(ch chan SSEMessage) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, l := range c.listeners {
		if l == ch {
			c.listeners = append(c.listeners[:i], c.listeners[i+1:]...)
			return
		}
	}
}

func (c *Collector) broadcast(msg SSEMessage) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, ch := range c.listeners {
		select {
		case ch <- msg:
		default:
		}
	}
}

func (c *Collector) writeLoop() {
	for {
		select {
		case <-c.done:
			return
		case e := <-c.events:
			c.processEvent(&e)
		}
	}
}

func (c *Collector) processEvent(e *schema.Event) {
	var scrubbed *schema.Event
	if c.scrubber != nil {
		scrubbed = c.scrubber.ScrubEvent(e)
	} else {
		copy := *e
		scrubbed = &copy
	}
	if scrubbed.Type == schema.EventLLMCall && scrubbed.EfficiencyScore == 0 {
		prompt := scrubbed.ToolInput
		output := scrubbed.ToolOutput
		scrubbed.EfficiencyScore = ScoreEfficiency(prompt, output)
	}
	if err := c.store.WriteEvent(scrubbed); err != nil {
		log.Error().Err(err).Msg("failed to write event")
	}
	if scrubbed.Type == schema.EventFileEdit && c.onSnapshot != nil {
		c.onSnapshot()
	}
	if v := c.budget.Check(scrubbed.SessionID); v != nil {
		c.handleBudgetViolation(v, scrubbed.SessionID)
	}
	if alert := c.alerts.Check(scrubbed); alert != nil {
		c.broadcast(SSEMessage{Type: "alert", Data: alert})
	}
	c.broadcast(SSEMessage{Type: "event", Data: scrubbed})
}

func (c *Collector) handleBudgetViolation(v *BudgetViolation, sessionID string) {
	c.broadcast(SSEMessage{Type: "budget_hit", Data: v})
	alertEvent := schema.Event{
		ID:        store.NewEventID(),
		SessionID: sessionID,
		Timestamp: time.Now(),
		Source:    "agentops",
		Type:      schema.EventBudgetAlert,
		Error:     v.Rule + " limit reached",
		Metadata: map[string]string{
			"rule":    v.Rule,
			"current": formatFloat(v.Current),
			"limit":   formatFloat(v.Limit),
			"action":  v.Action,
		},
	}
	c.store.WriteEvent(&alertEvent)
	pid := platform.FindAgentPID("cursor")
	if pid == 0 {
		pid = platform.FindAgentPID("claude")
	}
	switch v.Action {
	case "pause":
		_ = c.budget.Pause(pid)
	case "kill":
		_ = c.budget.Kill(pid)
	}
}

func (c *Collector) Close() {
	select {
	case <-c.done:
		return
	default:
	}
	for {
		select {
		case e := <-c.events:
			c.processEvent(&e)
		default:
			close(c.done)
			return
		}
	}
}

func formatFloat(v float64) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func (c *Collector) Store() *store.Store {
	return c.store
}

func (c *Collector) Config() *config.Config {
	return c.cfg
}
