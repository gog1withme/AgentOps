package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gog1withme/AgentOps/cli/collector"
	"github.com/gog1withme/AgentOps/cli/hooks"
	"github.com/gog1withme/AgentOps/cli/snapshots"
	"github.com/gog1withme/AgentOps/cli/store"
	"github.com/gog1withme/AgentOps/schema"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/sergi/go-diff/diffmatchpatch"
)

type Server struct {
	app        *fiber.App
	collector  *collector.Collector
	store      *store.Store
	workDir    string
	snapMgr    *snapshots.Manager
	mcpServers []string
}

func New(c *collector.Collector, st *store.Store, workDir string, snapMgr *snapshots.Manager, dashboardDir string) *Server {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(cors.New())
	s := &Server{app: app, collector: c, store: st, workDir: workDir, snapMgr: snapMgr}
	s.routes()
	if dashboardDir != "" {
		if _, err := os.Stat(dashboardDir); err == nil {
			s.app.Use(dashboardStatic(dashboardDir))
		}
	}
	return s
}

func (s *Server) routes() {
	proxy := collector.NewProxy(s.collector, os.Getenv("OPENAI_UPSTREAM"))
	s.app.All("/proxy/*", adaptor.HTTPHandler(http.StripPrefix("/proxy", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "" || r.URL.Path == "/" {
			r.URL.Path = "/"
		}
		proxy.Handler(w, r)
	}))))

	upstreams := hooks.LoadMCPServerUpstreams()
	if len(upstreams) > 0 {
		mcpProxy := collector.NewMCPProxy(s.collector, upstreams)
		for name := range upstreams {
			n := name
			s.mcpServers = append(s.mcpServers, n)
			route := "/mcp/" + n
			handler := mcpProxy.Handler(n)
			s.app.All(route, adaptor.HTTPHandler(handler))
			s.app.All(route+"/*", adaptor.HTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handler(w, r)
			})))
		}
	}

	api := s.app.Group("/api")
	api.Get("/events", s.listEvents)
	api.Get("/events/:id", s.getEvent)
	api.Get("/sessions", s.listSessions)
	api.Get("/sessions/:id", s.getSession)
	api.Get("/sessions/:id/replay", s.replaySession)
	api.Get("/cost", s.cost)
	api.Get("/cost/efficiency", s.efficiency)
	api.Get("/alerts", s.alerts)
	api.Get("/stats", s.stats)
	api.Get("/budget", s.budget)
	api.Get("/blame/*", s.blame)
	api.Get("/mcp", s.listMCP)
	api.Get("/mcp/:name", s.getMCP)
	api.Get("/prompts", s.listPrompts)
	api.Get("/prompts/diff", s.promptDiff)
	api.Get("/prompts/:hash", s.getPrompt)
	api.Get("/snapshots", s.listSnapshots)
	api.Post("/restore", s.restore)
	api.Post("/ingest/shell", s.ingestShell)
	s.app.Get("/api/stream", s.stream)
}

func (s *Server) Listen(addr string) error {
	return s.app.Listen(addr)
}

func (s *Server) MCPServers() []string {
	return s.mcpServers
}

func (s *Server) listEvents(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "100"))
	since := time.Now().Add(-24 * time.Hour)
	if v := c.Query("since"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			since = t
		}
	}
	events, err := s.store.ListEvents(limit, since, c.Query("source"), c.Query("type"))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(events)
}

func (s *Server) getEvent(c *fiber.Ctx) error {
	e, err := s.store.GetEvent(c.Params("id"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}
	return c.JSON(e)
}

func (s *Server) listSessions(c *fiber.Ctx) error {
	sessions, err := s.store.ListSessions()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(sessions)
}

func (s *Server) getSession(c *fiber.Ctx) error {
	id := c.Params("id")
	events, err := s.store.ReplaySession(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"session_id": id, "events": events})
}

func (s *Server) replaySession(c *fiber.Ctx) error {
	events, err := s.store.ReplaySession(c.Params("id"))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(events)
}

func (s *Server) cost(c *fiber.Ctx) error {
	groupBy := c.Query("group_by", "model")
	data, err := s.store.CostAggregates(groupBy)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(data)
}

func (s *Server) efficiency(c *fiber.Ctx) error {
	sessionID := s.collector.Config().SessionID
	data, err := s.store.EfficiencyTrend(sessionID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(data)
}

func (s *Server) alerts(c *fiber.Ctx) error {
	alerts, err := s.store.ListAlerts(false)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(alerts)
}

func (s *Server) stats(c *fiber.Ctx) error {
	sessionID := s.collector.Config().SessionID
	cost, tools, llms, _ := s.store.SessionSpend(sessionID, time.Now().Add(-24*time.Hour))
	return c.JSON(fiber.Map{
		"session_id": sessionID,
		"cost_usd":   cost,
		"tool_calls": tools,
		"llm_calls":  llms,
	})
}

func (s *Server) budget(c *fiber.Ctx) error {
	cfg := s.collector.Config()
	sessionID := cfg.SessionID
	cost, tools, llms, _ := s.store.SessionSpend(sessionID, time.Now().Add(-24*time.Hour))
	return c.JSON(fiber.Map{
		"budget": cfg.Budget,
		"spend": fiber.Map{
			"cost_usd":   cost,
			"tool_calls": tools,
			"llm_calls":  llms,
		},
	})
}

func (s *Server) blame(c *fiber.Ctx) error {
	path := c.Params("*")
	path, _ = url.PathUnescape(path)
	rows, err := s.store.ListAttribution(path)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(rows)
}

func (s *Server) listMCP(c *fiber.Ctx) error {
	servers, err := s.store.ListMCPServers()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(servers)
}

func (s *Server) getMCP(c *fiber.Ctx) error {
	servers, err := s.store.ListMCPServers()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	name := c.Params("name")
	for _, srv := range servers {
		if srv.Name == name {
			return c.JSON(srv)
		}
	}
	return c.Status(404).JSON(fiber.Map{"error": "not found"})
}

func (s *Server) listPrompts(c *fiber.Ctx) error {
	prompts, err := s.store.ListPrompts()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(prompts)
}

func (s *Server) getPrompt(c *fiber.Ctx) error {
	p, err := s.store.GetPrompt(c.Params("hash"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}
	return c.JSON(p)
}

func (s *Server) promptDiff(c *fiber.Ctx) error {
	a, err := s.store.GetPrompt(c.Query("a"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "prompt a not found"})
	}
	b, err := s.store.GetPrompt(c.Query("b"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "prompt b not found"})
	}
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(a.Content, b.Content, false)
	return c.JSON(fiber.Map{"diff": dmp.DiffPrettyText(diffs)})
}

func (s *Server) listSnapshots(c *fiber.Ctx) error {
	sessionID := s.collector.Config().SessionID
	snaps, err := s.store.ListSnapshots(sessionID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(snaps)
}

func (s *Server) restore(c *fiber.Ctx) error {
	var req struct {
		SessionID  string `json:"session_id"`
		SnapshotID string `json:"snapshot_id"`
		DryRun     bool   `json:"dry_run"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	if req.SessionID == "" {
		req.SessionID = s.collector.Config().SessionID
	}
	err := snapshots.Restore(s.store, req.SessionID, req.SnapshotID, s.workDir, req.DryRun)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (s *Server) ingestShell(c *fiber.Ctx) error {
	var payload collector.ShellIngest
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	source := payload.Source
	if source == "" {
		source = "shell"
	}
	s.collector.Ingest(schema.Event{
		ID:            store.NewEventID(),
		SessionID:     s.collector.Config().SessionID,
		Timestamp:     time.Now(),
		Source:        source,
		Type:          schema.EventShellCmd,
		ShellCommand:  payload.Command,
		ShellExitCode: payload.ExitCode,
	})
	return c.JSON(fiber.Map{"ok": true})
}

func (s *Server) stream(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")

	ch := s.collector.Subscribe()

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer s.collector.Unsubscribe(ch)
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		write := func(format string, args ...any) bool {
			if _, err := fmt.Fprintf(w, format, args...); err != nil {
				return false
			}
			if err := w.Flush(); err != nil {
				return false
			}
			return true
		}

		for {
			select {
			case msg, ok := <-ch:
				if !ok {
					return
				}
				data, err := json.Marshal(msg)
				if err != nil {
					continue
				}
				if !write("data: %s\n\n", data) {
					return
				}
			case <-ticker.C:
				if !write(": ping\n\n") {
					return
				}
			}
		}
	})
	return nil
}

func DashboardOutDir() string {
	var candidates []string

	home, _ := os.UserHomeDir()
	if home != "" {
		candidates = append(candidates, filepath.Join(home, ".agentops", "dashboard", "out"))
	}

	if prefix := os.Getenv("HOMEBREW_PREFIX"); prefix != "" {
		candidates = append(candidates, filepath.Join(prefix, "share", "agentops", "dashboard", "out"))
	}

	wd, _ := os.Getwd()
	candidates = append(candidates,
		filepath.Join(wd, "dashboard", "out"),
		filepath.Join(wd, "..", "dashboard", "out"),
		filepath.Join(wd, "out"),
	)

	exe, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "dashboard", "out"),
			filepath.Join(exeDir, "..", "dashboard", "out"),
		)
	}

	for _, p := range candidates {
		if _, err := os.Stat(filepath.Join(p, "index.html")); err == nil {
			abs, _ := filepath.Abs(p)
			return abs
		}
	}
	return ""
}
