/*
ScopeDriftGate detects and reports the number of files changed between the
current branch and the base branch, serving as an advisory gate to monitor
unexpected scope drift during ticket implementation.
*/

package gates

import (
	"fmt"
	"strings"
)

// ScopeDriftGate detects scope drift — advisory only (Gate #10).
type ScopeDriftGate struct{}

func (g *ScopeDriftGate) Name() string { return "scope-drift" }

func (g *ScopeDriftGate) Run() Result {
	branch := GitBranch()
	what := "count and names of modified files in the current branch against the base branch"
	why := "to advise developers on potential scope drift (unplanned changes) and encourage surgical edits"

	if branch == "main" || branch == "HEAD" {
		return Result{
			Gate:    g.Name(),
			Status:  "PASS",
			Message: fmt.Sprintf("checked: %s | why: %s | status: OK - on main or detached HEAD", what, why),
		}
	}

	base := FindBaseBranch()
	files, err := GitDiffFiles(base)
	if err != nil || len(files) == 0 {
		return Result{
			Gate:    g.Name(),
			Status:  "PASS",
			Message: fmt.Sprintf("checked: %s | why: %s | status: OK - no changes from %s", what, why, base),
		}
	}

	// Advisory: always PASS, but report file count
	return Result{
		Gate:    g.Name(),
		Status:  "PASS",
		Message: fmt.Sprintf("checked: %s | why: %s | status: OK - %d files changed (review for scope drift): %s", what, why, len(files), strings.Join(files[:min(len(files), 5)], ", ")),
	}
}
