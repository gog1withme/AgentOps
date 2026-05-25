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
	return sql.Open("sqlite", dbPath)
}
