package server

import (
	"strconv"
	"time"

	"github.com/gog1withme/AgentOps/cli/contextanalysis"
	"github.com/gofiber/fiber/v2"
)

func (s *Server) contextAnalysis(c *fiber.Ctx) error {
	sessionID := c.Query("session_id", s.collector.Config().SessionID)
	limit := 200
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	since := time.Now().Add(-24 * time.Hour)
	if v := c.Query("since"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			since = t
		}
	}

	var analysis *contextanalysis.Analysis
	var err error
	if sessionID != "" {
		analysis, err = contextanalysis.AnalyzeSession(s.store, sessionID, limit)
	} else {
		analysis, err = contextanalysis.AnalyzeRecent(s.store, "", since, limit)
	}
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(analysis)
}
