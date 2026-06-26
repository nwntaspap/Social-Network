package gates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrationsGate_NoDirPass(t *testing.T) {
	g := &MigrationsGate{MigrationDir: "/nonexistent"}
	result := g.Run()
	if result.Status != "PASS" {
		t.Errorf("expected PASS for missing dir, got %s: %s", result.Status, result.Message)
	}
}

func TestMigrationsGate_MissingDown(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "000001_init.up.sql"), []byte("CREATE TABLE t(id INT);"), 0o600); err != nil {
		t.Fatal(err)
	}

	g := &MigrationsGate{MigrationDir: dir}
	result := g.Run()
	if result.Status != "FAIL" {
		t.Errorf("expected FAIL for missing down migration, got %s: %s", result.Status, result.Message)
	}
}

func TestMigrationsGate_ValidPair(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "000001_init.up.sql"), []byte("CREATE TABLE t(id INT);"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "000001_init.down.sql"), []byte("DROP TABLE t;"), 0o600); err != nil {
		t.Fatal(err)
	}

	g := &MigrationsGate{MigrationDir: dir}
	result := g.Run()
	if result.Status != "PASS" {
		t.Errorf("expected PASS, got %s: %s", result.Status, result.Message)
	}
}

func TestMigrationsGate_BadDelimiter(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "000001_init.up.sql"), []byte("CREATE TABLE t(id INT);"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "000001_init.down.sql"), []byte("DROP TABLE t;"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "000002_add_col.up.sql"), []byte("ALTER TABLE t ADD COLUMN x INT:"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "000002_add_col.down.sql"), []byte("ALTER TABLE t DROP COLUMN x;"), 0o600); err != nil {
		t.Fatal(err)
	}

	g := &MigrationsGate{MigrationDir: dir}
	result := g.Run()
	if result.Status != "FAIL" {
		t.Errorf("expected FAIL for bad delimiter, got %s: %s", result.Status, result.Message)
	}
	if !strings.Contains(result.Message, "bad delimiter") {
		t.Errorf("expected bad delimiter error, got: %s", result.Message)
	}
}

func TestMigrationsGate_ReadDirError(t *testing.T) {
	// Pass a file path instead of a directory to force os.ReadDir failure
	dir := t.TempDir()
	filePath := filepath.Join(dir, "migrations.txt")
	if err := os.WriteFile(filePath, []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}
	g := &MigrationsGate{MigrationDir: filePath}
	res := g.Run()
	if res.Status != "FAIL" {
		t.Errorf("expected FAIL for file path as migration dir, got: %s (%s)", res.Status, res.Message)
	}
}
