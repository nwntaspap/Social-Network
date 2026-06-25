package gates

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunnerRunAll(t *testing.T) {
	runner := NewRunner()
	runner.Register(&StackGate{GoModPath: createTempGoMod(t, "module social-network\n\ngo 1.25\n")})

	report := runner.RunAll()
	if report.Overall != "PASS" {
		t.Errorf("expected PASS overall, got %s", report.Overall)
	}
	if len(report.Gates) != 1 {
		t.Errorf("expected 1 gate result, got %d", len(report.Gates))
	}
}

func TestRunnerRunOne(t *testing.T) {
	runner := NewRunner()
	runner.Register(&StackGate{GoModPath: createTempGoMod(t, "module social-network\n\ngo 1.25\n")})

	result, err := runner.RunOne("stack")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "PASS" {
		t.Errorf("expected PASS, got %s", result.Status)
	}

	_, err = runner.RunOne("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent gate")
	}
}

func TestWriteJSON(t *testing.T) {
	report := Report{
		Overall: "PASS",
		Gates: []Result{
			{Gate: "stack", Status: "PASS", Message: "OK"},
		},
	}
	var buf bytes.Buffer
	err := WriteJSON(&buf, report)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), `"overall": "PASS"`) {
		t.Errorf("unexpected JSON: %s", buf.String())
	}
}

func TestRunnerOnResult(t *testing.T) {
	runner := NewRunner()
	runner.Register(&StackGate{GoModPath: createTempGoMod(t, "module social-network\n\ngo 1.25\n")})

	var called bool
	var calledWith Result
	runner.OnResult = func(res Result) {
		called = true
		calledWith = res
	}

	_ = runner.RunAll()
	if !called {
		t.Error("expected OnResult callback to be called")
	}
	if calledWith.Gate != "stack" {
		t.Errorf("expected callback with gate stack, got %s", calledWith.Gate)
	}
}
