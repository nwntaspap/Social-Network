/*
MigrationsGate validates database schema migration files under the db/migrations directory (Gate #7).
It verifies that every '.up.sql' file has a matching '.down.sql' file, and parses SQL
content to check for invalid delimiters (e.g., statements ending with ':' instead of ';').
*/

package gates

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// MigrationsGate validates database migration files (Gate #7).
type MigrationsGate struct {
	MigrationDir string // defaults to "db/migrations"
}

func (g *MigrationsGate) Name() string { return "migrations" }

func (g *MigrationsGate) Run() Result {
	dir := g.MigrationDir
	if dir == "" {
		dir = "db/migrations"
	}

	what := "database migrations naming integrity (matching up/down scripts) and statement delimiter syntax"
	why := "to guarantee database schema changes are fully reversible and use semicolon separators instead of invalid colons"

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return Result{
			Gate:    g.Name(),
			Status:  "PASS",
			Message: fmt.Sprintf("checked: %s | why: %s | status: OK - no migration directory exists (pre-migration or dynamic storage only)", what, why),
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return Result{
			Gate:    g.Name(),
			Status:  "FAIL",
			Message: fmt.Sprintf("checked: %s | why: %s | status: FAIL - cannot read %s: %v | debug: verify migration path and directory permissions", what, why, dir, err),
		}
	}

	var errors []string
	// Collect .up.sql files and verify matching .down.sql exists
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}
		base := strings.TrimSuffix(name, ".up.sql")
		downFile := base + ".down.sql"
		if _, err := os.Stat(filepath.Join(dir, downFile)); os.IsNotExist(err) {
			errors = append(errors, "missing down migration for "+name)
		}
	}

	// Check for bad delimiters (colon-terminated statements instead of semicolons)
	badDelimiter := regexp.MustCompile(`^[^/-].*:\s*$`)
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		// #nosec G304
		content, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		for i, line := range strings.Split(string(content), "\n") {
			if badDelimiter.MatchString(line) {
				errors = append(errors, fmt.Sprintf("%s:%d bad delimiter (use ';' not ':')", e.Name(), i+1))
				break // one hit per file is enough
			}
		}
	}

	if len(errors) > 0 {
		return Result{
			Gate:    g.Name(),
			Status:  "FAIL",
			Message: fmt.Sprintf("checked: %s | why: %s | status: FAIL - %s | debug: run 'ls %s/' and check delimiter style inside scripts", what, why, strings.Join(errors, "; "), dir),
		}
	}
	return Result{
		Gate:    g.Name(),
		Status:  "PASS",
		Message: fmt.Sprintf("checked: %s | why: %s | status: OK - all migrations are paired with down scripts and use valid semicolon delimiters", what, why),
	}
}
