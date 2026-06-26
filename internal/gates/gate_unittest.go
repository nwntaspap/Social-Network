/*
UnitTestGate runs Go unit tests (go test -race <new_packages>) to ensure code correctness.
*/

package gates

import "fmt"

// UnitTestGate runs the unit test suite.
type UnitTestGate struct{}

func (g *UnitTestGate) Name() string { return "go-test" }

func (g *UnitTestGate) Run() Result {
	what := "backend unit tests execution with the race detector"
	why := "to ensure backend logical correctness and check for concurrent data race conditions"

	args := append([]string{"test", "-race"}, NewPkgs...)
	// #nosec G204
	cmd := ExecCommand("go", args...)
	if _, err := cmd.CombinedOutput(); err != nil {
		return Result{
			Gate:    g.Name(),
			Status:  "FAIL",
			Message: fmt.Sprintf("checked: %s | why: %s | status: FAIL - backend unit tests failed | debug: run 'go test -race ./...' to identify and fix failures", what, why),
		}
	}

	return Result{
		Gate:    g.Name(),
		Status:  "PASS",
		Message: fmt.Sprintf("checked: %s | why: %s | status: OK - all backend unit tests passed successfully", what, why),
	}
}
