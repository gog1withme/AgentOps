//go:build windows

package store

import (
	"database/sql"
	"path/filepath"

	"github.com/gog1withme/AgentOps/cli/internal/paths"
	_ "modernc.org/sqlite"
)

func openDriver() (*sql.DB, error) {
	dbPath := filepath.Join(paths.DataDir(), "events.db")
	dsn := dbPath + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	return db, nil
}
