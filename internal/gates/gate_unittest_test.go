package gates

import (
	"testing"
)

func TestUnitTestGate_Run(t *testing.T) {
	oldExec := ExecCommand
	oldLook := lookPath
	defer func() {
		ExecCommand = oldExec
		lookPath = oldLook
	}()
	ExecCommand = mockExecCommand

	g := &UnitTestGate{}

	// 1. Tests PASS
	t.Setenv("MOCK_FAIL", "0")
	res := g.Run()
	if res.Status != "PASS" {
		t.Errorf("expected unit tests PASS, got: %s (%s)", res.Status, res.Message)
	}

	// 2. Tests FAIL
	t.Setenv("MOCK_FAIL", "1")
	res = g.Run()
	if res.Status != "FAIL" {
		t.Errorf("expected unit tests FAIL, got: %s (%s)", res.Status, res.Message)
	}
}
