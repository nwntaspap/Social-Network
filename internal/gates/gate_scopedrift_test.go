package gates

import (
	"strings"
	"testing"
)

func TestScopeDriftGate_Run(t *testing.T) {
	oldExec := ExecCommand
	defer func() { ExecCommand = oldExec }()
	ExecCommand = mockExecCommand

	t.Run("on main", func(t *testing.T) {
		t.Setenv("MOCK_REV_MAIN", "1")
		g := &ScopeDriftGate{}
		res := g.Run()
		if res.Status != "PASS" {
			t.Errorf("expected scope-drift PASS on main, got: %s", res.Status)
		}
	})

	t.Run("on branch", func(t *testing.T) {
		t.Setenv("MOCK_REV_MAIN", "0")
		g := &ScopeDriftGate{}
		res := g.Run()
		if res.Status != "PASS" || !strings.Contains(res.Message, "changed") {
			t.Errorf("expected scope-drift PASS on branch with log, got: %s (%s)", res.Status, res.Message)
		}
	})
}

func TestScopeDriftGate_Error(t *testing.T) {
	oldExec := ExecCommand
	defer func() { ExecCommand = oldExec }()
	ExecCommand = mockExecCommand

	g := &ScopeDriftGate{}
	t.Setenv("MOCK_FAIL", "1")
	res := g.Run()
	if res.Status != "PASS" || !strings.Contains(res.Message, "no changes") {
		t.Errorf("expected PASS with no changes message on error, got: %s (%s)", res.Status, res.Message)
	}
}
