package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"social-network/internal/gates"
)

var noColor = os.Getenv("NO_COLOR") != ""

func ansi(code, s string) string {
	if noColor {
		return s
	}
	return code + s + "\033[0m"
}

var (
	green  = func(s string) string { return ansi("\033[32m", s) }
	red    = func(s string) string { return ansi("\033[31m", s) }
	yellow = func(s string) string { return ansi("\033[33m", s) }
	bold   = func(s string) string { return ansi("\033[1m", s) }
	dim    = func(s string) string { return ansi("\033[2m", s) }
)

func iconFor(status string) string {
	m := map[string]string{
		"PASS": "\u2705",
		"FAIL": "\u274C",
		"SKIP": "\u2796",
	}
	if noColor {
		m2 := map[string]string{
			"PASS": "[PASS]",
			"FAIL": "[FAIL]",
			"SKIP": "[SKIP]",
		}
		return m2[status]
	}
	return m[status]
}

func colorFor(status string) func(string) string {
	switch status {
	case "PASS":
		return green
	case "FAIL":
		return red
	case "SKIP":
		return yellow
	default:
		return func(s string) string { return s }
	}
}

func printHeader() {
	sep := dim(strings.Repeat("━", 48))
	fmt.Println()
	fmt.Printf("  %s\n", sep)
	fmt.Printf("  %s %s\n", bold(" 🔍  Review Gates"), dim("— Code quality verification"))
	fmt.Printf("  %s\n", sep)
	fmt.Println()
}

func main() {
	all := flag.Bool("all", false, "run all gates")
	gate := flag.String("gate", "", "run a specific gate by name")
	jsonOutput := flag.Bool("json", false, "output in JSON format")
	plainOutput := flag.Bool("plain", false, "plain ASCII output (for CI/lefthook)")
	flag.Parse()

	if *plainOutput {
		noColor = true
	}

	runner := gates.NewRunner()

	// Register all gates
	// infra → quality → tests → architecture → security → diff → frontend
	runner.Register(&gates.StackGate{})
	runner.Register(&gates.BranchGate{})
	runner.Register(&gates.FormatGate{})
	runner.Register(&gates.LintGate{MaxLines: 400})
	runner.Register(&gates.UnitTestGate{})
	runner.Register(&gates.CoverageGate{})
	runner.Register(&gates.LayoutGate{})
	runner.Register(&gates.BoundariesGate{})
	runner.Register(&gates.DAGGate{})
	runner.Register(&gates.TDDGate{})
	runner.Register(&gates.MigrationsGate{})
	runner.Register(&gates.SecurityGate{})
	runner.Register(&gates.ScopeDriftGate{})
	runner.Register(&gates.FrontendGate{})

	if *gate != "" {
		result, err := runner.RunOne(*gate)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(2)
		}
		report := gates.Report{Overall: result.Status, Gates: []gates.Result{result}}

		switch {
		case *jsonOutput:
			if err := gates.WriteJSON(os.Stdout, report); err != nil {
				fmt.Fprintf(os.Stderr, "error writing JSON: %v\n", err)
				os.Exit(2)
			}
		case *plainOutput:
			fmt.Print(plainHeader())
			fmt.Println(plainMessage(result))
			fmt.Print(plainSummary(report))
		default:
			printHeader()
			printResult(result)
			printSummary(report)
		}

		if result.Status == "FAIL" {
			os.Exit(1)
		}
		return
	}

	if *all || flag.NArg() == 0 {
		switch {
		case *jsonOutput:
			runner.OnResult = nil
		case *plainOutput:
			fmt.Print(plainHeader())
			runner.OnResult = func(result gates.Result) {
				fmt.Println(plainMessage(result))
			}
		default:
			printHeader()
			runner.OnResult = func(result gates.Result) {
				printResult(result)
			}
		}

		report := runner.RunAll()

		switch {
		case *jsonOutput:
			if err := gates.WriteJSON(os.Stdout, report); err != nil {
				fmt.Fprintf(os.Stderr, "error writing JSON: %v\n", err)
				os.Exit(2)
			}
		case *plainOutput:
			fmt.Print(plainSummary(report))
		default:
			printSummary(report)
		}

		if report.Overall == "FAIL" {
			os.Exit(1)
		}
		return
	}

	fmt.Fprintln(os.Stderr, "usage: gates --all | --gate=<name> [--json] [--plain]")
	os.Exit(2)
}

func highlightStatus(msg string, _ func(string) string) string {
	const (
		checkedPrefix = "checked: "
		whySep        = " | why: "
		statusSep     = " | status: "
		debugSep      = " | debug: "
	)

	statusIdx := strings.Index(msg, statusSep)
	if statusIdx < 0 {
		return msg
	}

	// Extract status findings (what was actually found)
	debugIdx := strings.LastIndex(msg, debugSep)
	var statusContent string
	if debugIdx > statusIdx {
		statusContent = strings.TrimSpace(msg[statusIdx+len(statusSep) : debugIdx])
	} else {
		statusContent = strings.TrimSpace(msg[statusIdx+len(statusSep):])
	}

	var b strings.Builder
	b.WriteString(bold("status: " + statusContent))

	// Merge checked + why into reason prose
	checkedIdx := strings.Index(msg, checkedPrefix)
	whyIdx := strings.Index(msg, whySep)
	if checkedIdx >= 0 && whyIdx >= 0 && whyIdx > checkedIdx && statusIdx > whyIdx {
		what := strings.TrimSpace(msg[checkedIdx+len(checkedPrefix) : whyIdx])
		why := strings.TrimSpace(msg[whyIdx+len(whySep) : statusIdx])
		what = strings.TrimRight(what, ".")
		b.WriteString(dim(" | reason: Checks " + what + " " + why))
	}

	if debugIdx >= 0 {
		b.WriteString(dim(msg[debugIdx:]))
	}

	return b.String()
}

func printResult(result gates.Result) {
	status := result.Status
	icon := iconFor(status)
	col := colorFor(status)
	display := highlightStatus(result.Message, col)

	fmt.Printf("  %s %-20s %s\n", icon, col(bold(result.Gate)), display)
}

func printSummary(report gates.Report) {
	var pass, fail, skip int
	for _, r := range report.Gates {
		switch r.Status {
		case "PASS":
			pass++
		case "FAIL":
			fail++
		case "SKIP":
			skip++
		}
	}
	total := pass + fail + skip

	fmt.Println()
	sep := dim(strings.Repeat("━", 48))
	fmt.Printf("  %s\n", sep)

	pLabel := iconFor("PASS") + " " + strconv.Itoa(pass)
	fLabel := iconFor("FAIL") + " " + strconv.Itoa(fail)
	sLabel := iconFor("SKIP") + " " + strconv.Itoa(skip)
	summary := bold(" Review Gates ") + dim("┃") +
		"  " + green(pLabel) +
		"  |  " + red(fLabel) +
		"  |  " + yellow(sLabel) +
		"  |  " + dim(fmt.Sprintf("%d total", total))

	fmt.Println("  " + summary)

	if report.Overall == "PASS" {
		fmt.Printf("  %s %s %s\n", dim("┃"), green(iconFor("PASS")), green(bold("All gates passed")))
	} else {
		fmt.Printf("  %s %s %s\n", dim("┃"), red(iconFor("FAIL")), red(bold("Some gates failed — review above")))
	}
	fmt.Printf("  %s\n", sep)
}

// ── Plain ASCII output mode (for CI / lefthook) ────────────────

func statusIconPlain(status string) string {
	switch status {
	case "PASS":
		return "[PASS]"
	case "FAIL":
		return "[FAIL]"
	case "SKIP":
		return "[SKIP]"
	default:
		return ""
	}
}

func plainMessage(result gates.Result) string {
	icon := statusIconPlain(result.Status)
	display := highlightStatus(result.Message, nil)
	return fmt.Sprintf("  %s %-20s %s", icon, result.Gate, display)
}

func plainHeader() string {
	return "\n  ------------------------------------------------\n   Review Gates - Code quality verification\n  ------------------------------------------------\n\n"
}

func plainSummary(report gates.Report) string {
	var pass, fail, skip int
	for _, r := range report.Gates {
		switch r.Status {
		case "PASS":
			pass++
		case "FAIL":
			fail++
		case "SKIP":
			skip++
		}
	}
	total := pass + fail + skip

	var b strings.Builder
	b.WriteString("\n  ------------------------------------------------\n")
	fmt.Fprintf(&b, "  Results: %d passed, %d failed, %d skipped (%d total)\n", pass, fail, skip, total)
	if report.Overall == "PASS" {
		b.WriteString("  All gates passed\n")
	} else {
		b.WriteString("  Some gates failed - see above\n")
	}
	b.WriteString("  ------------------------------------------------\n")
	return b.String()
}
