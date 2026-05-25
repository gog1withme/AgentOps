//go:build !windows

package store

import (
	"database/sql"

	"github.com/gog1withme/AgentOps/cli/internal/paths"
	_ "github.com/duckdb/duckdb-go/v2"
)

func openDriver() (*sql.DB, error) {
	return sql.Open("duckdb", paths.DuckDBPath())
}
