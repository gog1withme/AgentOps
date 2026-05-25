package store

import (
	"database/sql"

	"github.com/gog1withme/AgentOps/cli/internal/paths"
)

func openDB() (*sql.DB, error) {
	if err := paths.EnsureDirs(); err != nil {
		return nil, err
	}
	return openDriver()
}
