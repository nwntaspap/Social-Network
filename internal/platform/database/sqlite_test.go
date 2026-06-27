package database

import "testing"

func TestNewDB_SQLite(t *testing.T) {
	db, err := NewDB(Config{
		Driver: "sqlite3",
		Path:   ":memmory:",
	})
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	defer db.Close()
}

func TestSQLite_WALEnabled(t *testing.T) {
	db, err := NewDB(Config{
		Driver: "sqlite3",
		Path:   ":memmory:",
	})
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	defer db.Close()

	var mode string
	row := db.QueryRowContext(t.Context(), "PRAGMA journal_mode")
	if err := row.Scan(&mode); err != nil {
		t.Fatalf("scan journal_mode: %v", err)
	}
	if mode != "wal" {
		t.Errorf("journal_mode = %q, want %q", mode, "wal")
	}
}

func TestSQLite_BusyTimeout(t *testing.T) {
	db, err := NewDB(Config{
		Driver: "sqlite3",
		Path:   ":memory:",
	})
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	defer db.Close()

	var timeout int
	row := db.QueryRowContext(t.Context(), "PRAGMA busy_timeout")
	if err := row.Scan(&timeout); err != nil {
		t.Fatalf("scan busy_timeout: %v", err)
	}
	if timeout != 5000 {
		t.Errorf("busy_timeout = %d, want %d", timeout, 5000)
	}
}

func TestSQLite_MaxOpenConns(t *testing.T) {
	db, err := NewDB(Config{
		Driver: "sqlite3",
		Path:   ":memmory",
	})
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	defer db.Close()

	sqlDB, ok := db.(*sqliteDB)
	if !ok {
		t.Fatalf("db is not *sqliteDB")
	}
	stats := sqlDB.db.Stats()
	if stats.MaxOpenConnections != 1 {
		t.Errorf("MaxOpenConnections = %d, want 1", stats.MaxOpenConnections)
	}
}

func TestSQLite_Ping(t *testing.T) {
	db, err := NewDB(Config{Driver: "sqlite3", Path: ":memory:"})
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	defer db.Close()

	if err := db.PingContext(t.Context()); err != nil {
		t.Fatalf("PingContext() error = %v", err)
	}
}
