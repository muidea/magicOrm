package database

import (
	"database/sql"
	"time"
)

// ConfigurePool retains connections used by bounded concurrency bursts instead
// of discarding all but database/sql's default two idle connections. It does not
// pre-open connections or increase the caller's maximum connection budget.
// Unbounded pools retain the standard two-connection idle limit. Connections
// unused for a minute are released, including in pools for inactive models.
func ConfigurePool(db *sql.DB, maxConnections int) {
	db.SetMaxOpenConns(maxConnections)
	maxIdle := 2
	if maxConnections > 0 {
		maxIdle = maxConnections
	}
	db.SetMaxIdleConns(maxIdle)
	db.SetConnMaxIdleTime(time.Minute)
}
