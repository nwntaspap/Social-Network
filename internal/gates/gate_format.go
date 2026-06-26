/*
FormatGate validates Go code formatting (gofumpt and goimports).
If formatting violations are found, it directs the user to run 'make format'.
*/

package gates

import (
	"fmt"
	"strings"
)

// FormatGate checks Go code formatting.
type FormatGate struct{}

func (g *FormatGate) Name() string { return "format" }

func (g *FormatGate) Run() Result {
	what := "Go source code files formatting conventions (using gofumpt and goimports)"
	why := "to enforce a consistent codebase layout, structure, and import sorting styling rules"

	var errors []string
	var runDetails []string

	// Determine check tool (gofumpt or fallback to gofmt)
	tool := "gofumpt"
	failMsg := "gofumpt found unformatted files"
	detail := "gofumpt"
	if !toolAvailable("gofumpt") {
		tool = "gofmt"
		failMsg = "gofmt found unformatted files"
		detail = "gofmt (fallback)"
	}

	args := append([]string{"-l"}, NewDirs...)
	if msg, err := runFormatTool(tool, args, failMsg); err != nil {
		errors = append(errors, err.Error()+".")
	} else if msg != "" {
		errors = append(errors, msg)
	}
	runDetails = append(runDetails, detail)

	// Check goimports formatting
	if toolAvailable("goimports") {
		argsImports := append([]string{"-l", "-local", "social-network"}, NewDirs...)
		failMsgImports := "goimports found import issues"
		if msg, err := runFormatTool("goimports", argsImports, failMsgImports); err != nil {
			errors = append(errors, err.Error()+".")
		} else if msg != "" {
			errors = append(errors, msg)
		}
		runDetails = append(runDetails, "goimports")
	}

	if len(errors) > 0 {
		return Result{
			Gate:    g.Name(),
			Status:  "FAIL",
			Message: fmt.Sprintf("checked: %s | why: %s | status: FAIL - %s | debug: run 'make format' to format all unstyled Go source files", what, why, strings.Join(errors, "; ")),
		}
	}

	return Result{
		Gate:    g.Name(),
		Status:  "PASS",
		Message: fmt.Sprintf("checked: %s | why: %s | status: OK - all files are correctly styled (verified via %s)", what, why, strings.Join(runDetails, " + ")),
	}
}

func runFormatTool(name string, args []string, failMsg string) (string, error) {
	// #nosec G204
	cmd := ExecCommand(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s error: %w (output: %q)", name, err, string(out))
	}
	if len(strings.TrimSpace(string(out))) > 0 {
		return failMsg, nil
	}
	return "", nil
}
