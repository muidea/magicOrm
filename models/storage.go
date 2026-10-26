package models

import "time"

// DatabaseConfig is the existing connection-description contract. It grants no
// SQL execution capability; concrete driver configuration stays internal.
type DatabaseConfig interface {
	Server() string
	Username() string
	Password() string
	Database() string
	Schema() string
	GetDsn() string
}

// SchemaIssue is catalog evidence, not an executable repair instruction.
type SchemaIssue struct {
	Table  string `json:"table"`
	Object string `json:"object,omitempty"`
	Kind   string `json:"kind"`
}

// SchemaReport retains the public diagnostic fields and JSON contract.
type SchemaReport struct {
	Matches bool          `json:"matches"`
	Tables  []string      `json:"tables"`
	Issues  []SchemaIssue `json:"issues"`
}

// PoolStats is a connection-budget observation, never a database handle.
// A zero maximum means unbounded, consistent with database/sql.
type PoolStats struct {
	MaxOpenConnections int
	OpenConnections    int
	InUse              int
	Idle               int
	WaitCount          int64
	WaitDuration       time.Duration
}
