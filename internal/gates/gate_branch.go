/*
BranchGate validates the current Git branch name against the pattern
'<username>/<ticket-ID>-<detail>' and checks that all commit messages
on the branch follow the Conventional Commits format.
*/

package gates

import (
	"fmt"
	"regexp"
	"strings"
)

// BranchGate validates branch naming and conventional commits (Gate #9).
type BranchGate struct{}

func (g *BranchGate) Name() string { return "branch" }

var (
	branchPattern = regexp.MustCompile(`^(epapamic|ekaramet|dkotsi|geoikonomou|smichail)/[A-Za-z0-9-]+-[A-Za-z0-9-]+$`)
	commitPattern = regexp.MustCompile(`^(feat|fix|test|refactor|chore|docs|style|perf|ci|build|revert)(?:\((user|topic|follow|group|event|chat|notification|oauth|core|platform|comment|docs|dev|gates)\))?!?:`)
)

func (g *BranchGate) Run() Result {
	branch := GitBranch()
	what := "git branch name conformity and conventional commits history on the active branch"
	why := "to associate developer changes with ticket IDs (e.g. ekaramet/S1-BE-05-db-factory) and ensure clean structured commits"

	if branch == "main" || branch == "HEAD" {
		return Result{
			Gate:    g.Name(),
			Status:  "PASS",
			Message: fmt.Sprintf("checked: %s | why: %s | status: OK - on main or detached HEAD", what, why),
		}
	}

	var errors []string

	if !branchPattern.MatchString(branch) {
		errors = append(errors, fmt.Sprintf("branch '%s' doesn't match '<username>/<ticket-ID>-<detail>'", branch))
	}

	base := FindBaseBranch()
	commits, err := GitLog(base)
	if err == nil {
		for _, msg := range commits {
			if strings.HasPrefix(msg, "Merge ") {
				continue // skip auto-generated merge commits
			}
			if !commitPattern.MatchString(msg) {
				errors = append(errors, fmt.Sprintf("non-conventional commit: '%s'", msg))
			}
		}
	}

	if len(errors) > 0 {
		return Result{
			Gate:    g.Name(),
			Status:  "FAIL",
			Message: fmt.Sprintf("checked: %s | why: %s | status: FAIL - %s | debug: rename branch using 'git branch -m <new-name>' and fix commit history using 'git commit --amend' or interactive rebase", what, why, strings.Join(errors, "; ")),
		}
	}
	return Result{
		Gate:    g.Name(),
		Status:  "PASS",
		Message: fmt.Sprintf("checked: %s | why: %s | status: OK - branch name matches '<username>/<ticket-ID>-<detail>' and commit messages follow Conventional Commits", what, why),
	}
}
