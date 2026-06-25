package main

import (
	"io"
	"os"
	"strings"
	"testing"

	"social-network/internal/gates"
)

func TestHighlightStatus(t *testing.T) {
	saved := noColor
	noColor = true
	defer func() { noColor = saved }()

	tests := []struct {
		msg  string
		want string
	}{
		{msg: "no pipes", want: "no pipes"},
		{msg: "checked: formatting | why: consistency | status: OK - all good", want: "status: OK - all good | reason: Checks formatting consistency"},
		{msg: "checked: lint | why: quality | status: FAIL - violations | debug: fix it", want: "status: FAIL - violations | reason: Checks lint quality | debug: fix it"},
		{msg: "checked: formatting | status: OK - only status", want: "status: OK - only status"},
		{msg: "checked: lint | status: FAIL - error | debug: check it", want: "status: FAIL - error | debug: check it"},
	}
	noCol := func(s string) string { return s }
	for _, tt := range tests {
		got := highlightStatus(tt.msg, noCol)
		if got != tt.want {
			t.Errorf("highlightStatus(%q) = %q, want %q", tt.msg, got, tt.want)
		}
	}
}

func TestPlainMessage(t *testing.T) {
	r := gates.Result{
		Gate:    "format",
		Status:  "PASS",
		Message: "checked: Go formatting | why: consistency | status: OK - all good",
	}
	got := plainMessage(r)
	if !strings.Contains(got, "[PASS]") {
		t.Errorf("plainMessage should contain [PASS], got: %q", got)
	}
	if !strings.Contains(got, "status: OK") {
		t.Errorf("plainMessage should have status first, got: %q", got)
	}
	if !strings.Contains(got, "reason: Checks Go formatting consistency") {
		t.Errorf("plainMessage should contain merged reason, got: %q", got)
	}
	if !strings.Contains(got, "format") {
		t.Errorf("plainMessage should contain gate name, got: %q", got)
	}
}

func TestPlainMessage_FailWithDebug(t *testing.T) {
	r := gates.Result{
		Gate:    "lint",
		Status:  "FAIL",
		Message: "checked: lint checks | why: quality | status: FAIL - violations found | debug: run 'golangci-lint run'",
	}
	got := plainMessage(r)
	if !strings.Contains(got, "status: FAIL - violations found") {
		t.Errorf("plainMessage(FAIL) should show status first, got: %q", got)
	}
	if !strings.Contains(got, "debug:") {
		t.Errorf("plainMessage(FAIL) should contain debug, got: %q", got)
	}
}

func TestPlainMessage_NoStructuredFormat(t *testing.T) {
	r := gates.Result{
		Gate:    "custom",
		Status:  "PASS",
		Message: "plain unstructured message",
	}
	got := plainMessage(r)
	if !strings.Contains(got, "plain unstructured message") {
		t.Errorf("plainMessage(no-structure) should pass through original message, got: %q", got)
	}
}

func TestPlainSummary(t *testing.T) {
	report := gates.Report{
		Overall: "FAIL",
		Gates: []gates.Result{
			{Gate: "format", Status: "PASS", Message: "ok"},
			{Gate: "lint", Status: "FAIL", Message: "checked: lint | why: quality | status: FAIL - err | debug: fix it"},
			{Gate: "coverage", Status: "SKIP", Message: "ok"},
		},
	}

	got := plainSummary(report)
	if !strings.Contains(got, "1 passed") {
		t.Errorf("plainSummary should show pass count, got: %s", got)
	}
	if !strings.Contains(got, "1 failed") {
		t.Errorf("plainSummary should show fail count, got: %s", got)
	}
	if !strings.Contains(got, "1 skipped") {
		t.Errorf("plainSummary should show skip count, got: %s", got)
	}
	if !strings.Contains(got, "Some gates failed") {
		t.Errorf("plainSummary should show failure message, got: %s", got)
	}
}

func TestPlainSummary_AllPass(t *testing.T) {
	report := gates.Report{
		Overall: "PASS",
		Gates: []gates.Result{
			{Gate: "format", Status: "PASS", Message: "ok"},
			{Gate: "lint", Status: "PASS", Message: "ok"},
		},
	}
	got := plainSummary(report)
	if !strings.Contains(got, "All gates passed") {
		t.Errorf("plainSummary(all pass) should show success, got: %s", got)
	}
}

func TestPlainHeader(t *testing.T) {
	got := plainHeader()
	if !strings.Contains(got, "Review Gates") {
		t.Errorf("plainHeader should contain title, got: %s", got)
	}
	if strings.Contains(got, "\033") {
		t.Errorf("plainHeader should not contain ANSI codes, got: %s", got)
	}
	if strings.Contains(got, "━") {
		t.Errorf("plainHeader should not contain unicode box-drawing, got: %s", got)
	}
}

func TestStatusIconPlain(t *testing.T) {
	tests := []struct {
		status string
		want   string
	}{
		{"PASS", "[PASS]"},
		{"FAIL", "[FAIL]"},
		{"SKIP", "[SKIP]"},
		{"UNKNOWN", ""},
	}
	for _, tt := range tests {
		got := statusIconPlain(tt.status)
		if got != tt.want {
			t.Errorf("statusIconPlain(%q) = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestIconFor(t *testing.T) {
	saved := noColor
	noColor = true
	if got := iconFor("PASS"); got != "[PASS]" {
		t.Errorf("iconFor(PASS) under noColor = %q, want [PASS]", got)
	}
	if got := iconFor("FAIL"); got != "[FAIL]" {
		t.Errorf("iconFor(FAIL) under noColor = %q, want [FAIL]", got)
	}
	if got := iconFor("SKIP"); got != "[SKIP]" {
		t.Errorf("iconFor(SKIP) under noColor = %q, want [SKIP]", got)
	}

	noColor = false
	if got := iconFor("PASS"); got != "\u2705" {
		t.Errorf("iconFor(PASS) = %q, want \u2705", got)
	}
	noColor = saved
}

func TestColorFor(t *testing.T) {
	saved := noColor
	noColor = true

	fPass := colorFor("PASS")
	if got := fPass("text"); got != "text" {
		t.Errorf("colorFor(PASS) under noColor = %q, want text", got)
	}

	noColor = false
	fFail := colorFor("FAIL")
	if got := fFail("text"); !strings.Contains(got, "\033[31m") {
		t.Errorf("colorFor(FAIL) should colorize, got: %q", got)
	}

	fSkip := colorFor("SKIP")
	if got := fSkip("text"); !strings.Contains(got, "\033[33m") {
		t.Errorf("colorFor(SKIP) should colorize yellow, got: %q", got)
	}

	fDefault := colorFor("UNKNOWN")
	if got := fDefault("text"); got != "text" {
		t.Errorf("colorFor(UNKNOWN) = %q, want text", got)
	}

	noColor = saved
}

func TestAnsiBoldDim(t *testing.T) {
	saved := noColor
	noColor = true
	if got := bold("text"); got != "text" {
		t.Errorf("bold under noColor = %q, want text", got)
	}
	if got := dim("text"); got != "text" {
		t.Errorf("dim under noColor = %q, want text", got)
	}

	noColor = false
	if got := bold("text"); !strings.Contains(got, "\033[1m") {
		t.Errorf("bold = %q, should contain ANSI escape", got)
	}
	if got := dim("text"); !strings.Contains(got, "\033[2m") {
		t.Errorf("dim = %q, should contain ANSI escape", got)
	}
	noColor = saved
}

func captureStdout(fn func()) string {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	old := os.Stdout
	os.Stdout = w
	defer func() {
		os.Stdout = old
	}()

	fn()
	_ = w.Close()
	var buf strings.Builder
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func TestOutputHelpers(t *testing.T) {
	saved := noColor
	noColor = true
	defer func() { noColor = saved }()

	hdr := captureStdout(func() {
		printHeader()
	})
	if !strings.Contains(hdr, "Review Gates") {
		t.Errorf("printHeader output should contain title, got: %q", hdr)
	}

	res := captureStdout(func() {
		printResult(gates.Result{
			Gate:    "format",
			Status:  "PASS",
			Message: "status: OK - everything looks neat",
		})
	})
	if !strings.Contains(res, "format") || !strings.Contains(res, "status: OK") {
		t.Errorf("printResult output invalid, got: %q", res)
	}

	sumPass := captureStdout(func() {
		printSummary(gates.Report{
			Overall: "PASS",
			Gates: []gates.Result{
				{Gate: "format", Status: "PASS", Message: "ok"},
				{Gate: "coverage", Status: "SKIP", Message: "skipped"},
			},
		})
	})
	if !strings.Contains(sumPass, "All gates passed") {
		t.Errorf("printSummary(PASS) output invalid, got: %q", sumPass)
	}

	sumFail := captureStdout(func() {
		printSummary(gates.Report{
			Overall: "FAIL",
			Gates: []gates.Result{
				{Gate: "format", Status: "FAIL", Message: "failed"},
				{Gate: "coverage", Status: "SKIP", Message: "skipped"},
			},
		})
	})
	if !strings.Contains(sumFail, "Some gates failed") {
		t.Errorf("printSummary(FAIL) output invalid, got: %q", sumFail)
	}
}
