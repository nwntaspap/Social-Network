/*
Package gates handles registration and orchestration of gate checks.
It defines the main interfaces, structures, and helper functions for executing
the validation pipeline and formatting the report output into structured JSON.
*/
package gates

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
)

// ExecCommand is the function used to create exec.Cmd instances.
// Override in tests to mock subprocess calls.
var ExecCommand = exec.Command

var lookPath = exec.LookPath

// toolAvailable checks if a binary is in PATH.
func toolAvailable(name string) bool {
	_, err := lookPath(name)
	return err == nil
}

// Gate is the interface every check must implement.
type Gate interface {
	Name() string
	Run() Result
}

// Result is the outcome of a single gate check.
type Result struct {
	Gate    string `json:"gate"`
	Status  string `json:"status"` // PASS, FAIL, SKIP
	Message string `json:"message"`
}

// Report is the aggregated output of all gates.
type Report struct {
	Overall string   `json:"overall"`
	Gates   []Result `json:"gates"`
}

// Runner holds registered gates and executes them.
type Runner struct {
	gates    []Gate
	OnResult func(Result)
}

// NewRunner creates a runner with all gates registered.
func NewRunner() *Runner {
	return &Runner{}
}

// Register adds a gate to the runner.
func (r *Runner) Register(g Gate) {
	r.gates = append(r.gates, g)
}

// RunAll executes all registered gates and returns a report.
func (r *Runner) RunAll() Report {
	report := Report{Overall: "PASS"}
	for _, g := range r.gates {
		result := g.Run()
		if result.Status == "FAIL" {
			report.Overall = "FAIL"
		}
		report.Gates = append(report.Gates, result)
		if r.OnResult != nil {
			r.OnResult(result)
		}
	}
	return report
}

// RunOne executes a single named gate.
func (r *Runner) RunOne(name string) (Result, error) {
	for _, g := range r.gates {
		if g.Name() == name {
			return g.Run(), nil
		}
	}
	return Result{}, fmt.Errorf("gate %q not found", name)
}

// WriteJSON writes the report as JSON to the writer.
func WriteJSON(w io.Writer, report Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}
