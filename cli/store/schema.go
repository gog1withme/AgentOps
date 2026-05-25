package store

const schemaSQL = `
CREATE TABLE IF NOT EXISTS events (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    run_id TEXT,
    timestamp TEXT NOT NULL,
    source TEXT NOT NULL,
    type TEXT NOT NULL,
    model TEXT,
    prompt_hash TEXT,
    prompt_tokens INTEGER DEFAULT 0,
    output_tokens INTEGER DEFAULT 0,
    efficiency_score REAL DEFAULT 0,
    cost_usd REAL DEFAULT 0,
    duration_ms INTEGER DEFAULT 0,
    tool_name TEXT,
    tool_input TEXT,
    tool_output TEXT,
    file_path TEXT,
    file_diff TEXT,
    shell_command TEXT,
    shell_exit_code INTEGER,
    mcp_server TEXT,
    mcp_latency_ms INTEGER DEFAULT 0,
    error TEXT,
    metadata TEXT
);

CREATE TABLE IF NOT EXISTS prompts (
    hash TEXT PRIMARY KEY,
    content TEXT NOT NULL,
    first_seen TEXT NOT NULL,
    last_seen TEXT NOT NULL,
    session_count INTEGER DEFAULT 1,
    avg_cost_usd REAL DEFAULT 0,
    avg_prompt_tokens INTEGER DEFAULT 0,
    avg_efficiency_score REAL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS attribution (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    run_id TEXT,
    timestamp TEXT NOT NULL,
    file_path TEXT NOT NULL,
    agent TEXT NOT NULL,
    lines_added INTEGER DEFAULT 0,
    lines_removed INTEGER DEFAULT 0,
    diff TEXT,
    event_id TEXT
);

CREATE TABLE IF NOT EXISTS mcp_servers (
    name TEXT PRIMARY KEY,
    url TEXT,
    first_seen TEXT,
    last_seen TEXT,
    total_calls INTEGER DEFAULT 0,
    error_count INTEGER DEFAULT 0,
    avg_latency_ms REAL DEFAULT 0,
    p95_latency_ms REAL DEFAULT 0,
    avg_response_bytes INTEGER DEFAULT 0
);

CREATE TABLE IF NOT EXISTS budgets (
    session_id TEXT PRIMARY KEY,
    cost_limit_usd REAL,
    tool_call_limit INTEGER,
    llm_call_limit INTEGER,
    action TEXT DEFAULT 'alert',
    created_at TEXT,
    triggered_at TEXT
);

CREATE TABLE IF NOT EXISTS snapshots (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    timestamp TEXT NOT NULL,
    trigger TEXT NOT NULL,
    snapshot_path TEXT NOT NULL,
    file_count INTEGER DEFAULT 0,
    size_bytes INTEGER DEFAULT 0
);

CREATE TABLE IF NOT EXISTS alerts (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    rule TEXT NOT NULL,
    severity TEXT NOT NULL,
    message TEXT NOT NULL,
    timestamp TEXT NOT NULL,
    dismissed INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_events_session ON events(session_id);
CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp);
CREATE INDEX IF NOT EXISTS idx_events_type ON events(type);
CREATE INDEX IF NOT EXISTS idx_events_prompt ON events(prompt_hash);
CREATE INDEX IF NOT EXISTS idx_attribution_file ON attribution(file_path);
CREATE INDEX IF NOT EXISTS idx_attribution_sess ON attribution(session_id);
CREATE INDEX IF NOT EXISTS idx_snapshots_session ON snapshots(session_id);
`

func (s *Store) migrate() error {
	_, err := s.db.Exec(schemaSQL)
	return err
}
