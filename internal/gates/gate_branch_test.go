package gates

import (
	"strings"
	"testing"
)

func TestBranchGate_Run(t *testing.T) {
	oldExec := ExecCommand
	defer func() { ExecCommand = oldExec }()
	ExecCommand = mockExecCommand

	t.Run("on main branch", func(t *testing.T) {
		t.Setenv("MOCK_REV_MAIN", "1")
		g := &BranchGate{}
		res := g.Run()
		if res.Status != "PASS" || !strings.Contains(res.Message, "on main or detached HEAD") {
			t.Errorf("expected PASS for main branch, got: %s (%s)", res.Status, res.Message)
		}
	})

	t.Run("conventional branch and commit", func(t *testing.T) {
		t.Setenv("MOCK_REV_MAIN", "0")
		t.Setenv("MOCK_COMMIT_FAIL", "0")
		g := &BranchGate{}
		res := g.Run()
		if res.Status != "PASS" {
			t.Errorf("expected branch check PASS, got: %s (%s)", res.Status, res.Message)
		}
	})

	t.Run("non-conventional commit", func(t *testing.T) {
		t.Setenv("MOCK_REV_MAIN", "0")
		t.Setenv("MOCK_COMMIT_FAIL", "1")
		g := &BranchGate{}
		res := g.Run()
		if res.Status != "FAIL" {
			t.Errorf("expected branch check FAIL for bad commit, got: %s", res.Status)
		}
	})

	t.Run("unapproved branch username", func(t *testing.T) {
		t.Setenv("MOCK_REV_MAIN", "0")
		t.Setenv("MOCK_REV_UNAPPROVED", "1")
		t.Setenv("MOCK_COMMIT_FAIL", "0")
		g := &BranchGate{}
		res := g.Run()
		if res.Status != "FAIL" || !strings.Contains(res.Message, "doesn't match '<username>") {
			t.Errorf("expected branch check FAIL for unapproved username, got: %s (%s)", res.Status, res.Message)
		}
	})
}

func TestBranchGate_CommitPattern(t *testing.T) {
	tests := []struct {
		msg   string
		match bool
	}{
		{"feat(user): add auth handler", true},
		{"fix(docs): resolve drift-report issues", true},
		{"refactor: migrate agent configurations", true},
		{"chore: hide all subagents except flowmaster", true},
		{"docs: replace all Biome references", true},
		{"feat!: breaking change", true},
		{"feat(user)!: breaking change with scope", true},
		{"feat(gates): add source file line limit checks to lint gate", true},
		{"invalid: commit msg", false},
		{"feat(invalid_scope): msg", false},
		{"Fixing code", false},
	}

	for _, tc := range tests {
		t.Run(tc.msg, func(t *testing.T) {
			matched := commitPattern.MatchString(tc.msg)
			if matched != tc.match {
				t.Errorf("expected match=%v for %q, got %v", tc.match, tc.msg, matched)
			}
		})
	}
}
