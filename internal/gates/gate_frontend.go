/*
FrontendGate runs frontend lint, format, typecheck, and test checks.
It checks for the presence of frontend-next or frontend directories, and runs
the corresponding package manager (bun) commands.
*/

package gates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FrontendGate runs frontend CI scripts.
type FrontendGate struct{}

func (g *FrontendGate) Name() string { return "frontend" }

func (g *FrontendGate) Run() Result {
	root := GitRepoRoot()
	if root == "" {
		root = "."
	}

	what := "frontend code quality validation checks (linting, formatting, typescript compiler typechecking, and tests)"
	why := "to guarantee that frontend codebase conforms to Next.js styling guidelines, typescript types compile without errors, and frontend tests pass"

	var dir string
	if _, err := os.Stat(filepath.Join(root, "frontend-next", "package.json")); err == nil {
		dir = filepath.Join(root, "frontend-next")
	} else if _, err := os.Stat(filepath.Join(root, "frontend", "package.json")); err == nil {
		dir = filepath.Join(root, "frontend")
	}

	if dir == "" {
		return Result{
			Gate:    g.Name(),
			Status:  "SKIP",
			Message: fmt.Sprintf("checked: %s | why: %s | status: SKIP - no frontend scaffolded yet", what, why),
		}
	}

	steps := []struct {
		name    string
		command []string
	}{
		{"lint", []string{"bun", "run", "lint"}},
		{"format:check", []string{"bun", "run", "format:check"}},
		{"typecheck", []string{"bun", "x", "tsc", "--noEmit"}},
		{"test", []string{"bun", "run", "test"}},
	}

	for _, step := range steps {
		// #nosec G204
		cmd := ExecCommand(step.command[0], step.command[1:]...)
		cmd.Dir = dir
		if _, err := cmd.CombinedOutput(); err != nil {
			return Result{
				Gate:    g.Name(),
				Status:  "FAIL",
				Message: fmt.Sprintf("checked: %s | why: %s | status: FAIL - frontend %s check failed | debug: run 'cd %s && %s' to view error logs and debug", what, why, step.name, dir, strings.Join(step.command, " ")),
			}
		}
	}

	return Result{
		Gate:    g.Name(),
		Status:  "PASS",
		Message: fmt.Sprintf("checked: %s | why: %s | status: OK - all frontend CI validation checks (lint, format, typecheck, tests) passed successfully", what, why),
	}
}
