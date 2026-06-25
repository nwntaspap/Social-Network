/*
LintGate runs Go linter tools (golangci-lint, staticcheck, go vet) to ensure code hygiene.
*/

package gates

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LintGate runs linters.
type LintGate struct {
	MaxLines int
	Dirs     []string
}

func (g *LintGate) Name() string { return "lint" }

//nolint:nestif
func (g *LintGate) Run() Result {
	maxLines := g.MaxLines
	if maxLines <= 0 {
		maxLines = 300 // Default limit
	}

	dirs := g.Dirs
	if len(dirs) == 0 {
		dirs = NewDirs
	}

	what := "Go source code files static analysis (linting) and file line length limits check"
	why := "to maintain code clean syntax rules, catch common bugs early, and restrict maximum file sizes for improved readability"

	var errors []string
	failFiles, warnings, walkErr := g.checkLineLimits(maxLines, dirs)
	if walkErr != nil {
		errors = append(errors, fmt.Sprintf("failed to scan line limits: %v", walkErr))
	} else if len(failFiles) > 0 {
		errors = append(errors, fmt.Sprintf("files exceeding maximum line limit (%d by >20%%): %s", maxLines, strings.Join(failFiles, ", ")))
	}

	if toolAvailable("golangci-lint") {
		// golangci-lint uses directories with /... suffix
		var lintDirs []string
		for _, dir := range dirs {
			lintDirs = append(lintDirs, dir+"/...")
		}
		args := append([]string{"run", "--timeout=5m"}, lintDirs...)
		// #nosec G204
		cmd := ExecCommand("golangci-lint", args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			errors = append(errors, fmt.Sprintf("golangci-lint error: %v (output: %q)", err, string(out)))
		}
	} else {
		// Fallbacks: staticcheck and go vet use package paths (NewPkgs)
		if toolAvailable("staticcheck") {
			args := append([]string{}, NewPkgs...)
			// #nosec G204
			cmd := ExecCommand("staticcheck", args...)
			if out, err := cmd.CombinedOutput(); err != nil {
				errors = append(errors, fmt.Sprintf("staticcheck error: %v (output: %q)", err, string(out)))
			}
		}

		args := append([]string{"vet"}, NewPkgs...)
		// #nosec G204
		cmd := ExecCommand("go", args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			errors = append(errors, fmt.Sprintf("go vet error: %v (output: %q)", err, string(out)))
		}
	}

	if len(errors) > 0 {
		return Result{
			Gate:    g.Name(),
			Status:  "FAIL",
			Message: fmt.Sprintf("checked: %s | why: %s | status: FAIL - %s | debug: run 'golangci-lint run' to review lint warnings manually", what, why, strings.Join(errors, "; ")),
		}
	}

	suffix := "golangci-lint"
	if !toolAvailable("golangci-lint") {
		if toolAvailable("staticcheck") {
			suffix = "staticcheck + go vet"
		} else {
			suffix = "go vet"
		}
	}

	warnStr := ""
	if len(warnings) > 0 {
		warnStr = " | WARNING: " + strings.Join(warnings, ", ") + " exceeds line limit by >10%"
	}

	return Result{
		Gate:    g.Name(),
		Status:  "PASS",
		Message: fmt.Sprintf("checked: %s | why: %s | status: OK - linters passed successfully (verified via %s)%s", what, why, suffix, warnStr),
	}
}

func (g *LintGate) checkLineLimits(maxLines int, dirs []string) ([]string, []string, error) {
	var warnings []string
	var failFiles []string

	for _, dir := range dirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
				count, err := countLines(path)
				if err != nil {
					return err
				}
				if count > int(float64(maxLines)*1.2) {
					failFiles = append(failFiles, fmt.Sprintf("%s (%d lines)", path, count))
				} else if count > int(float64(maxLines)*1.1) {
					warnings = append(warnings, fmt.Sprintf("%s (%d lines)", path, count))
				}
			}
			return nil
		})
		if err != nil {
			return nil, nil, err
		}
	}
	return failFiles, warnings, nil
}

func countLines(path string) (int, error) {
	// #nosec G304
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		count++
	}
	return count, scanner.Err()
}
