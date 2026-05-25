package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gog1withme/AgentOps/schema"
	"github.com/oklog/ulid/v2"
)

type Store struct {
	db    *sql.DB
	mu    sync.Mutex
	retry []schema.Event
}

func Open() (*Store, error) {
	db, err := openDB()
	if err != nil {
		return nil, err
	}
	s := &Store{db: db, retry: make([]schema.Event, 0, 1000)}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) DB() *sql.DB {
	return s.db
}

func (s *Store) WriteEvent(e *schema.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	meta, _ := json.Marshal(e.Metadata)
	_, err := s.db.Exec(`
		INSERT INTO events (
			id, session_id, run_id, timestamp, source, type, model, prompt_hash,
			prompt_tokens, output_tokens, efficiency_score, cost_usd, duration_ms,
			tool_name, tool_input, tool_output, file_path, file_diff, shell_command,
			shell_exit_code, mcp_server, mcp_latency_ms, error, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, e.ID, e.SessionID, e.RunID, e.Timestamp.Format(time.RFC3339Nano), e.Source, string(e.Type), e.Model, e.PromptHash,
		e.PromptTokens, e.OutputTokens, e.EfficiencyScore, e.CostUSD, e.DurationMS,
		e.ToolName, e.ToolInput, e.ToolOutput, e.FilePath, e.FileDiff, e.ShellCommand,
		e.ShellExitCode, e.MCPServer, e.MCPLatencyMS, e.Error, string(meta))
	if err != nil {
		if len(s.retry) < 1000 {
			s.retry = append(s.retry, *e)
		}
		return err
	}
	return s.flushRetryLocked()
}

func (s *Store) flushRetryLocked() error {
	if len(s.retry) == 0 {
		return nil
	}
	remaining := s.retry[:0]
	for _, e := range s.retry {
		meta, _ := json.Marshal(e.Metadata)
		_, err := s.db.Exec(`
			INSERT INTO events (
				id, session_id, run_id, timestamp, source, type, model, prompt_hash,
				prompt_tokens, output_tokens, efficiency_score, cost_usd, duration_ms,
				tool_name, tool_input, tool_output, file_path, file_diff, shell_command,
				shell_exit_code, mcp_server, mcp_latency_ms, error, metadata
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, e.ID, e.SessionID, e.RunID, e.Timestamp.Format(time.RFC3339Nano), e.Source, string(e.Type), e.Model, e.PromptHash,
			e.PromptTokens, e.OutputTokens, e.EfficiencyScore, e.CostUSD, e.DurationMS,
			e.ToolName, e.ToolInput, e.ToolOutput, e.FilePath, e.FileDiff, e.ShellCommand,
			e.ShellExitCode, e.MCPServer, e.MCPLatencyMS, e.Error, string(meta))
		if err != nil {
			remaining = append(remaining, e)
		}
	}
	s.retry = remaining
	return nil
}

func (s *Store) UpsertPrompt(p *schema.Prompt) error {
	_, err := s.db.Exec(`
		INSERT INTO prompts (hash, content, first_seen, last_seen, session_count, avg_cost_usd, avg_prompt_tokens, avg_efficiency_score)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(hash) DO UPDATE SET
			last_seen = excluded.last_seen,
			session_count = prompts.session_count + 1,
			avg_cost_usd = (prompts.avg_cost_usd + excluded.avg_cost_usd) / 2,
			avg_prompt_tokens = (prompts.avg_prompt_tokens + excluded.avg_prompt_tokens) / 2,
			avg_efficiency_score = (prompts.avg_efficiency_score + excluded.avg_efficiency_score) / 2
	`, p.Hash, p.Content, p.FirstSeen.Format(time.RFC3339Nano), p.LastSeen.Format(time.RFC3339Nano), p.SessionCount, p.AvgCostUSD, p.AvgPromptTokens, p.AvgEfficiencyScore)
	return err
}

func (s *Store) WriteAttribution(a *schema.Attribution) error {
	_, err := s.db.Exec(`
		INSERT INTO attribution (id, session_id, run_id, timestamp, file_path, agent, lines_added, lines_removed, diff, event_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, a.ID, a.SessionID, a.RunID, a.Timestamp.Format(time.RFC3339Nano), a.FilePath, a.Agent, a.LinesAdded, a.LinesRemoved, a.Diff, a.EventID)
	return err
}

func (s *Store) UpsertMCPServer(m *schema.MCPServer) error {
	_, err := s.db.Exec(`
		INSERT INTO mcp_servers (name, url, first_seen, last_seen, total_calls, error_count, avg_latency_ms, p95_latency_ms, avg_response_bytes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(name) DO UPDATE SET
			url = excluded.url,
			last_seen = excluded.last_seen,
			total_calls = mcp_servers.total_calls + 1,
			error_count = mcp_servers.error_count + excluded.error_count,
			avg_latency_ms = excluded.avg_latency_ms,
			p95_latency_ms = excluded.p95_latency_ms,
			avg_response_bytes = excluded.avg_response_bytes
	`, m.Name, m.URL, m.FirstSeen.Format(time.RFC3339Nano), m.LastSeen.Format(time.RFC3339Nano), m.TotalCalls, m.ErrorCount, m.AvgLatencyMS, m.P95LatencyMS, m.AvgResponseBytes)
	return err
}

func (s *Store) SaveBudget(b *schema.Budget) error {
	triggeredStr := ""
	if b.TriggeredAt != nil {
		triggeredStr = b.TriggeredAt.Format(time.RFC3339Nano)
	}
	_, err := s.db.Exec(`
		INSERT INTO budgets (session_id, cost_limit_usd, tool_call_limit, llm_call_limit, action, created_at, triggered_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(session_id) DO UPDATE SET
			cost_limit_usd = excluded.cost_limit_usd,
			tool_call_limit = excluded.tool_call_limit,
			llm_call_limit = excluded.llm_call_limit,
			action = excluded.action
	`, b.SessionID, b.CostLimitUSD, b.ToolCallLimit, b.LLMCallLimit, b.Action, b.CreatedAt.Format(time.RFC3339Nano), nullStr(triggeredStr))
	return err
}

func (s *Store) GetBudget(sessionID string) (*schema.Budget, error) {
	row := s.db.QueryRow(`
		SELECT session_id, cost_limit_usd, tool_call_limit, llm_call_limit, action, created_at, triggered_at
		FROM budgets WHERE session_id = ?
	`, sessionID)
	var b schema.Budget
	var triggered sql.NullString
	var createdStr string
	err := row.Scan(&b.SessionID, &b.CostLimitUSD, &b.ToolCallLimit, &b.LLMCallLimit, &b.Action, &createdStr, &triggered)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	b.CreatedAt, _ = parseTime(createdStr)
	b.TriggeredAt = scanNullTime(triggered)
	return &b, nil
}

func (s *Store) WriteSnapshot(snap *schema.Snapshot) error {
	_, err := s.db.Exec(`
		INSERT INTO snapshots (id, session_id, timestamp, trigger, snapshot_path, file_count, size_bytes)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, snap.ID, snap.SessionID, snap.Timestamp.Format(time.RFC3339Nano), snap.Trigger, snap.SnapshotPath, snap.FileCount, snap.SizeBytes)
	return err
}

func (s *Store) ListSnapshots(sessionID string) ([]schema.Snapshot, error) {
	rows, err := s.db.Query(`
		SELECT id, session_id, timestamp, trigger, snapshot_path, file_count, size_bytes
		FROM snapshots WHERE session_id = ? ORDER BY timestamp DESC
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]schema.Snapshot, 0)
	for rows.Next() {
		var snap schema.Snapshot
		var ts string
		if err := rows.Scan(&snap.ID, &snap.SessionID, &ts, &snap.Trigger, &snap.SnapshotPath, &snap.FileCount, &snap.SizeBytes); err != nil {
			return nil, err
		}
		snap.Timestamp, _ = parseTime(ts)
		out = append(out, snap)
	}
	return out, rows.Err()
}

func (s *Store) SessionSpend(sessionID string, since time.Time) (cost float64, toolCalls, llmCalls int, err error) {
	row := s.db.QueryRow(`
		SELECT COALESCE(SUM(cost_usd), 0),
			COALESCE(SUM(CASE WHEN type='tool_call' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN type='llm_call' THEN 1 ELSE 0 END), 0)
		FROM events WHERE session_id = ? AND timestamp > ?
	`, sessionID, since.Format(time.RFC3339Nano))
	err = row.Scan(&cost, &toolCalls, &llmCalls)
	return
}

func (s *Store) ListEvents(limit int, since time.Time, source, eventType string) ([]schema.Event, error) {
	query := `SELECT id, session_id, run_id, timestamp, source, type, model, prompt_hash,
		prompt_tokens, output_tokens, efficiency_score, cost_usd, duration_ms,
		tool_name, tool_input, tool_output, file_path, file_diff, shell_command,
		shell_exit_code, mcp_server, mcp_latency_ms, error, metadata
		FROM events WHERE timestamp > ?`
	args := []any{formatTime(since)}
	if source != "" {
		query += " AND source = ?"
		args = append(args, source)
	}
	if eventType != "" {
		query += " AND type = ?"
		args = append(args, eventType)
	}
	query += " ORDER BY timestamp DESC LIMIT ?"
	args = append(args, limit)
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEvents(rows)
}

func (s *Store) GetEvent(id string) (*schema.Event, error) {
	row := s.db.QueryRow(`
		SELECT id, session_id, run_id, timestamp, source, type, model, prompt_hash,
		prompt_tokens, output_tokens, efficiency_score, cost_usd, duration_ms,
		tool_name, tool_input, tool_output, file_path, file_diff, shell_command,
		shell_exit_code, mcp_server, mcp_latency_ms, error, metadata
		FROM events WHERE id = ?
	`, id)
	events, err := scanEventsRows(row)
	if err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return nil, sql.ErrNoRows
	}
	return &events[0], nil
}

func (s *Store) ReplaySession(sessionID string) ([]schema.Event, error) {
	rows, err := s.db.Query(`
		SELECT id, session_id, run_id, timestamp, source, type, model, prompt_hash,
		prompt_tokens, output_tokens, efficiency_score, cost_usd, duration_ms,
		tool_name, tool_input, tool_output, file_path, file_diff, shell_command,
		shell_exit_code, mcp_server, mcp_latency_ms, error, metadata
		FROM events WHERE session_id = ? ORDER BY timestamp ASC
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEvents(rows)
}

func (s *Store) ListSessions() ([]schema.SessionSummary, error) {
	rows, err := s.db.Query(`
		SELECT session_id, MIN(timestamp), MAX(timestamp), COUNT(*),
			COALESCE(SUM(cost_usd), 0), COALESCE(SUM(prompt_tokens + output_tokens), 0)
		FROM events GROUP BY session_id ORDER BY MAX(timestamp) DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]schema.SessionSummary, 0)
	for rows.Next() {
		var ss schema.SessionSummary
		var first, last string
		if err := rows.Scan(&ss.SessionID, &first, &last, &ss.EventCount, &ss.TotalCost, &ss.TotalTokens); err != nil {
			return nil, err
		}
		ss.FirstEvent, _ = parseTime(first)
		ss.LastEvent, _ = parseTime(last)
		out = append(out, ss)
	}
	return out, rows.Err()
}

func (s *Store) ListAttribution(filePath string) ([]schema.Attribution, error) {
	rows, err := s.db.Query(`
		SELECT id, session_id, run_id, timestamp, file_path, agent, lines_added, lines_removed, diff, event_id
		FROM attribution WHERE file_path = ? ORDER BY timestamp DESC
	`, filePath)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]schema.Attribution, 0)
	for rows.Next() {
		var a schema.Attribution
		var ts string
		if err := rows.Scan(&a.ID, &a.SessionID, &a.RunID, &ts, &a.FilePath, &a.Agent, &a.LinesAdded, &a.LinesRemoved, &a.Diff, &a.EventID); err != nil {
			return nil, err
		}
		a.Timestamp, _ = parseTime(ts)
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) ListAttributionBySession(sessionID string) ([]schema.Attribution, error) {
	rows, err := s.db.Query(`
		SELECT id, session_id, run_id, timestamp, file_path, agent, lines_added, lines_removed, diff, event_id
		FROM attribution WHERE session_id = ? ORDER BY timestamp DESC
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]schema.Attribution, 0)
	for rows.Next() {
		var a schema.Attribution
		var ts string
		if err := rows.Scan(&a.ID, &a.SessionID, &a.RunID, &ts, &a.FilePath, &a.Agent, &a.LinesAdded, &a.LinesRemoved, &a.Diff, &a.EventID); err != nil {
			return nil, err
		}
		a.Timestamp, _ = parseTime(ts)
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) ListMCPServers() ([]schema.MCPServer, error) {
	rows, err := s.db.Query(`
		SELECT name, url, first_seen, last_seen, total_calls, error_count, avg_latency_ms, p95_latency_ms, avg_response_bytes
		FROM mcp_servers ORDER BY last_seen DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]schema.MCPServer, 0)
	for rows.Next() {
		var m schema.MCPServer
		var first, last string
		if err := rows.Scan(&m.Name, &m.URL, &first, &last, &m.TotalCalls, &m.ErrorCount, &m.AvgLatencyMS, &m.P95LatencyMS, &m.AvgResponseBytes); err != nil {
			return nil, err
		}
		m.FirstSeen, _ = parseTime(first)
		m.LastSeen, _ = parseTime(last)
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Store) ListPrompts() ([]schema.Prompt, error) {
	rows, err := s.db.Query(`
		SELECT hash, content, first_seen, last_seen, session_count, avg_cost_usd, avg_prompt_tokens, avg_efficiency_score
		FROM prompts ORDER BY first_seen DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]schema.Prompt, 0)
	for rows.Next() {
		var p schema.Prompt
		var first, last string
		if err := rows.Scan(&p.Hash, &p.Content, &first, &last, &p.SessionCount, &p.AvgCostUSD, &p.AvgPromptTokens, &p.AvgEfficiencyScore); err != nil {
			return nil, err
		}
		p.FirstSeen, _ = parseTime(first)
		p.LastSeen, _ = parseTime(last)
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) GetPrompt(hash string) (*schema.Prompt, error) {
	row := s.db.QueryRow(`
		SELECT hash, content, first_seen, last_seen, session_count, avg_cost_usd, avg_prompt_tokens, avg_efficiency_score
		FROM prompts WHERE hash = ?
	`, hash)
	var p schema.Prompt
	var first, last string
	err := row.Scan(&p.Hash, &p.Content, &first, &last, &p.SessionCount, &p.AvgCostUSD, &p.AvgPromptTokens, &p.AvgEfficiencyScore)
	if err != nil {
		return nil, err
	}
	p.FirstSeen, _ = parseTime(first)
	p.LastSeen, _ = parseTime(last)
	return &p, nil
}

func (s *Store) CostAggregates(groupBy string) ([]map[string]any, error) {
	col := "model"
	switch groupBy {
	case "source":
		col = "source"
	case "run_id":
		col = "run_id"
	}
	query := fmt.Sprintf(`
		SELECT %s, SUM(cost_usd), SUM(prompt_tokens), SUM(output_tokens), COUNT(*)
		FROM events GROUP BY %s ORDER BY SUM(cost_usd) DESC
	`, col, col)
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var key string
		var cost float64
		var pt, ot, count int
		if err := rows.Scan(&key, &cost, &pt, &ot, &count); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"key":           key,
			"cost_usd":      cost,
			"prompt_tokens": pt,
			"output_tokens": ot,
			"count":         count,
		})
	}
	return out, rows.Err()
}

func (s *Store) EfficiencyTrend(sessionID string) ([]map[string]any, error) {
	rows, err := s.db.Query(`
		SELECT timestamp, efficiency_score, prompt_tokens
		FROM events WHERE session_id = ? AND type = 'llm_call' ORDER BY timestamp ASC
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var tsStr string
		var score float64
		var tokens int
		if err := rows.Scan(&tsStr, &score, &tokens); err != nil {
			return nil, err
		}
		ts, _ := parseTime(tsStr)
		out = append(out, map[string]any{
			"timestamp":        ts,
			"efficiency_score": score,
			"prompt_tokens":    tokens,
		})
	}
	return out, rows.Err()
}

func (s *Store) WriteAlert(a *schema.AlertRecord) error {
	_, err := s.db.Exec(`
		INSERT INTO alerts (id, session_id, rule, severity, message, timestamp, dismissed)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, a.ID, a.SessionID, a.Rule, a.Severity, a.Message, a.Timestamp.Format(time.RFC3339Nano), boolInt(a.Dismissed))
	return err
}

func (s *Store) ListAlerts(activeOnly bool) ([]schema.AlertRecord, error) {
	query := `SELECT id, session_id, rule, severity, message, timestamp, dismissed FROM alerts`
	if activeOnly {
		query += " WHERE dismissed = 0"
	}
	query += " ORDER BY timestamp DESC LIMIT 100"
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]schema.AlertRecord, 0)
	for rows.Next() {
		var a schema.AlertRecord
		var ts string
		var dismissed int
		if err := rows.Scan(&a.ID, &a.SessionID, &a.Rule, &a.Severity, &a.Message, &ts, &dismissed); err != nil {
			return nil, err
		}
		a.Timestamp, _ = parseTime(ts)
		a.Dismissed = dismissed != 0
		out = append(out, a)
	}
	return out, rows.Err()
}

func NewEventID() string {
	return ulid.Make().String()
}

func scanEvents(rows *sql.Rows) ([]schema.Event, error) {
	out := make([]schema.Event, 0)
	for rows.Next() {
		e, err := scanEventRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func scanEventRow(scanner interface {
	Scan(dest ...any) error
}) (schema.Event, error) {
	var e schema.Event
	var typ string
	var metaJSON sql.NullString
	var tsStr string
	err := scanner.Scan(
		&e.ID, &e.SessionID, &e.RunID, &tsStr, &e.Source, &typ, &e.Model, &e.PromptHash,
		&e.PromptTokens, &e.OutputTokens, &e.EfficiencyScore, &e.CostUSD, &e.DurationMS,
		&e.ToolName, &e.ToolInput, &e.ToolOutput, &e.FilePath, &e.FileDiff, &e.ShellCommand,
		&e.ShellExitCode, &e.MCPServer, &e.MCPLatencyMS, &e.Error, &metaJSON,
	)
	if err != nil {
		return e, err
	}
	e.Timestamp, _ = parseTime(tsStr)
	e.Type = schema.EventType(typ)
	if metaJSON.Valid && metaJSON.String != "" {
		_ = json.Unmarshal([]byte(metaJSON.String), &e.Metadata)
	}
	return e, nil
}

func scanEventsRows(row *sql.Row) ([]schema.Event, error) {
	e, err := scanEventRow(row)
	if err != nil {
		return nil, err
	}
	return []schema.Event{e}, nil
}
