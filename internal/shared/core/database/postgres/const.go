package postgres

import "time"

const (
	DefaultMaxConns int32 = 25
	DefaultMinConns int32 = 5
)

const (
	DefaultMaxConnLifetime   = time.Hour
	DefaultMaxConnIdleTime   = 15 * time.Minute
	DefaultHealthCheckPeriod = time.Minute
)
