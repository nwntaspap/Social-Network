package database

import (
	"database/sql"
	"time"
)

// ResolveTime returns the time from a sql.NullTime, or fallback if NULL.
func ResolveTime(nt sql.NullTime, fallback time.Time) time.Time {
	if nt.Valid {
		return nt.Time
	}
	return fallback
}
