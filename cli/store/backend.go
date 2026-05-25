package store

import "runtime"

// BackendName returns the active store driver for this platform.
func BackendName() string {
	if runtime.GOOS == "windows" {
		return "sqlite"
	}
	return "duckdb"
}
