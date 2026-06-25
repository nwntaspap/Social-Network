/*
TDDGate enforces Test-Driven Development rules (Gate #6).
It checks every vertical slice feature package to verify that if Go source files
exist within the 'commands' directory, there are corresponding test files ('*_test.go')
present in the same directory.
*/

package gates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TDDGate verifies test files exist for feature code (Gate #6).
type TDDGate struct {
	InternalDir string // defaults to "internal"
}

func (g *TDDGate) Name() string { return "tdd" }

func (g *TDDGate) Run() Result {
	dir := g.InternalDir
	if dir == "" {
		dir = "internal"
	}

	what := "presence of corresponding test files in commands/ directories for active feature packages"
	why := "to enforce Test-Driven Development (TDD) rules ensuring every use case commands directory contains matching unit/integration tests"

	entries, err := os.ReadDir(dir)
	if err != nil {
		return Result{
			Gate:    g.Name(),
			Status:  "SKIP",
			Message: fmt.Sprintf("checked: %s | why: %s | status: SKIP - cannot read %s: %v", what, why, dir, err),
		}
	}

	var errors []string
	for _, e := range entries {
		if !e.IsDir() || !isFeatureSlice(dir, e.Name()) {
			continue
		}
		feature := e.Name()
		cmdDir := filepath.Join(dir, feature, "commands")
		errors = append(errors, checkTestCoverage(cmdDir)...)
	}

	if len(errors) > 0 {
		return Result{
			Gate:    g.Name(),
			Status:  "FAIL",
			Message: fmt.Sprintf("checked: %s | why: %s | status: FAIL - %s | debug: run 'ls %s/<feature>/commands/' and create matching _test.go files", what, why, strings.Join(errors, "; "), dir),
		}
	}
	return Result{
		Gate:    g.Name(),
		Status:  "PASS",
		Message: fmt.Sprintf("checked: %s | why: %s | status: OK - all active commands directories contain corresponding test files", what, why),
	}
}

// checkTestCoverage verifies a directory has test files if it has Go source files.
func checkTestCoverage(dir string) []string {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}

	hasGoFiles := false
	hasTestFiles := false

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, "_test.go") {
			hasTestFiles = true
		} else if strings.HasSuffix(name, ".go") {
			hasGoFiles = true
		}
	}

	if hasGoFiles && !hasTestFiles {
		return []string{dir + ": has Go files but no test files"}
	}
	return nil
}
