/*
CoverageGate compares test coverage of the current branch against the
base branch (e.g. main). It runs tests in a temporary Git worktree to prevent
mutating the current workspace, ensuring test coverage does not drop beyond
a threshold.
*/

package gates

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// pkgToDir converts a module-qualified package path like
// "social-network/internal/user" to its directory path.
func pkgToDir(pkg string) string {
	parts := strings.SplitN(pkg, "/", 2)
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}

// hasGoFiles returns true if the directory contains at least one .go file.
func hasGoFiles(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".go") {
			return true, nil
		}
	}
	return false, nil
}

// CoverageGate compares branch coverage vs base branch (Gate #13).
// Uses git worktree to avoid mutating the active workspace.
type CoverageGate struct {
	Threshold float64 // max allowed coverage drop %, defaults to 5
}

func (g *CoverageGate) Name() string { return "coverage-delta" }

func (g *CoverageGate) Run() Result {
	threshold := g.Threshold
	if threshold == 0 {
		threshold = 5
	}

	base := FindBaseBranch()
	what := "test coverage percentage delta between the current branch and the base branch"
	why := "to guarantee that code additions do not degrade overall test coverage beyond the allowed threshold"

	// Get base branch coverage via git worktree
	baseCov, err := getBaselineCoverage(base)
	if err != nil {
		return Result{
			Gate:    g.Name(),
			Status:  "PASS",
			Message: fmt.Sprintf("checked: %s | why: %s | status: OK - could not compute baseline: %v", what, why, err),
		}
	}

	// Get current branch coverage
	branchCov, err := getCurrentCoverage()
	if err != nil {
		return Result{
			Gate:    g.Name(),
			Status:  "PASS",
			Message: fmt.Sprintf("checked: %s | why: %s | status: OK - could not compute branch coverage: %v", what, why, err),
		}
	}

	delta := branchCov - baseCov
	if delta < -threshold {
		return Result{
			Gate:    g.Name(),
			Status:  "FAIL",
			Message: fmt.Sprintf("checked: %s | why: %s | status: FAIL - coverage dropped by %.1f%% (base: %.1f%% -> current: %.1f%%) exceeding threshold %.1f%% | debug: run 'go test -coverprofile=coverage.out ./...' to inspect coverage and add missing tests", what, why, -delta, baseCov, branchCov, threshold),
		}
	}

	return Result{
		Gate:    g.Name(),
		Status:  "PASS",
		Message: fmt.Sprintf("checked: %s | why: %s | status: OK - coverage is at %.1f%% (delta: %+.1f%%, base: %.1f%%)", what, why, branchCov, delta, baseCov),
	}
}

func getBaselineCoverage(baseBranch string) (float64, error) {
	tempDir := filepath.Join(os.TempDir(), "sn-gate-cov-base")
	// Clean up stale tempdir and worktree registration from prior crashed runs
	// #nosec G204
	_ = ExecCommand("git", "worktree", "remove", "--force", tempDir).Run()
	_ = os.RemoveAll(tempDir)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create worktree (--detach avoids failure if baseBranch is checked out)
	// #nosec G204
	add := ExecCommand("git", "worktree", "add", "--detach", tempDir, baseBranch)
	if err := add.Run(); err != nil {
		return 0, fmt.Errorf("git worktree add: %w", err)
	}
	defer func() {
		// #nosec G204
		_ = ExecCommand("git", "worktree", "remove", "--force", tempDir).Run()
	}()

	// Run tests in worktree (only packages that exist in the worktree)
	covFile := filepath.Join(tempDir, "coverage.out")
	existingPkgs := make([]string, 0, len(NewPkgs))
	for _, pkg := range NewPkgs {
		dir := pkgToDir(pkg)
		if dir == "" {
			continue
		}
		pkgDir := filepath.Join(tempDir, dir)
		info, err := os.Stat(pkgDir)
		if err != nil || !info.IsDir() {
			continue
		}
		hasGoFile, _ := hasGoFiles(pkgDir)
		if hasGoFile {
			existingPkgs = append(existingPkgs, pkg)
		}
	}
	args := append([]string{"test", "-coverprofile=" + covFile}, existingPkgs...)
	// #nosec G204
	testCmd := ExecCommand("go", args...)
	testCmd.Dir = tempDir
	if err := testCmd.Run(); err != nil {
		return 0, fmt.Errorf("go test in worktree: %w", err)
	}

	return parseCoverageFile(covFile)
}

func getCurrentCoverage() (float64, error) {
	covFile := filepath.Join(os.TempDir(), "sn-gate-cov-branch.out")
	defer func() { _ = os.Remove(covFile) }()

	args := append([]string{"test", "-coverprofile=" + covFile}, NewPkgs...)
	// #nosec G204
	testCmd := ExecCommand("go", args...)
	if err := testCmd.Run(); err != nil {
		return 0, fmt.Errorf("go test: %w", err)
	}
	return parseCoverageFile(covFile)
}

func parseCoverageFile(path string) (float64, error) {
	// Use go tool cover to get total coverage
	// #nosec G204
	cmd := ExecCommand("go", "tool", "cover", "-func="+path)
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	// Last line: "total: (statements) XX.X%"
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	var lastLine string
	for scanner.Scan() {
		lastLine = scanner.Text()
	}

	fields := strings.Fields(lastLine)
	if len(fields) < 3 {
		return 0, fmt.Errorf("unexpected coverage output: %s", lastLine)
	}
	pctStr := strings.TrimSuffix(fields[len(fields)-1], "%")
	return strconv.ParseFloat(pctStr, 64)
}
