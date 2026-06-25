package gates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLintGate_Run(t *testing.T) {
	oldExec := ExecCommand
	oldLook := lookPath
	defer func() {
		ExecCommand = oldExec
		lookPath = oldLook
	}()
	ExecCommand = mockExecCommand

	g := &LintGate{MaxLines: 10000, Dirs: []string{"."}}

	// 1. Tool available (golangci-lint), PASS
	lookPath = func(name string) (string, error) {
		if name == "golangci-lint" {
			return name, nil
		}
		return "", os.ErrNotExist
	}
	t.Setenv("MOCK_FAIL", "0")
	res := g.Run()
	if res.Status != "PASS" {
		t.Errorf("expected lint PASS (golangci-lint), got: %s (%s)", res.Status, res.Message)
	}

	// 2. Tool available (golangci-lint), FAIL
	t.Setenv("MOCK_FAIL", "1")
	res = g.Run()
	if res.Status != "FAIL" {
		t.Errorf("expected lint FAIL (golangci-lint), got: %s (%s)", res.Status, res.Message)
	}

	// 3. Fallbacks (staticcheck + go vet), PASS
	lookPath = func(name string) (string, error) {
		if name == "staticcheck" || name == "go" {
			return name, nil
		}
		return "", os.ErrNotExist
	}
	t.Setenv("MOCK_FAIL", "0")
	res = g.Run()
	if res.Status != "PASS" {
		t.Errorf("expected lint PASS (fallbacks), got: %s (%s)", res.Status, res.Message)
	}

	// 4. Fallbacks, FAIL
	t.Setenv("MOCK_FAIL", "1")
	res = g.Run()
	if res.Status != "FAIL" {
		t.Errorf("expected lint FAIL (fallbacks), got: %s (%s)", res.Status, res.Message)
	}
}

func TestLintGate_LineLimit(t *testing.T) {
	// Use a small limit so that stack gate (59 lines) triggers failure
	g := &LintGate{MaxLines: 10, Dirs: []string{"."}}

	// Mock lookPath to do nothing
	oldLook := lookPath
	defer func() { lookPath = oldLook }()
	lookPath = func(name string) (string, error) {
		return "", os.ErrNotExist
	}

	res := g.Run()
	if res.Status != "FAIL" {
		t.Errorf("expected FAIL due to line limit, got: %s (%s)", res.Status, res.Message)
	}
	if !strings.Contains(res.Message, "exceeding maximum line limit") {
		t.Errorf("expected error message to mention line limit, got: %s", res.Message)
	}
}

func TestLintGate_CheckLineLimitsVariants(t *testing.T) {
	dir := t.TempDir()

	// File at warning threshold (340 lines for limit 300: 300*1.1=330, 300*1.2=360)
	warnContent := strings.Repeat("line\n", 340)
	if err := os.WriteFile(filepath.Join(dir, "warn.go"), []byte(warnContent), 0o600); err != nil {
		t.Fatal(err)
	}

	// File at fail threshold (400 lines)
	failContent := strings.Repeat("line\n", 400)
	if err := os.WriteFile(filepath.Join(dir, "fail.go"), []byte(failContent), 0o600); err != nil {
		t.Fatal(err)
	}

	// File that passes (100 lines)
	passContent := strings.Repeat("line\n", 100)
	if err := os.WriteFile(filepath.Join(dir, "pass.go"), []byte(passContent), 0o600); err != nil {
		t.Fatal(err)
	}

	oldLook := lookPath
	defer func() { lookPath = oldLook }()
	lookPath = func(name string) (string, error) {
		return "", os.ErrNotExist
	}

	g := &LintGate{MaxLines: 300, Dirs: []string{dir}}
	res := g.Run()
	// Should FAIL due to the 400-line file exceeding 360 (300*1.2)
	if res.Status != "FAIL" {
		t.Errorf("expected FAIL due to line limits, got: %s (%s)", res.Status, res.Message)
	}
	if !strings.Contains(res.Message, "exceeding maximum line limit") {
		t.Errorf("expected message about line limit, got: %s", res.Message)
	}
}

func TestLintGate_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	oldLook := lookPath
	defer func() { lookPath = oldLook }()
	lookPath = func(name string) (string, error) {
		return "", os.ErrNotExist
	}

	g := &LintGate{MaxLines: 300, Dirs: []string{dir}}
	res := g.Run()
	if res.Status != "PASS" {
		t.Errorf("expected PASS for empty dir, got: %s (%s)", res.Status, res.Message)
	}
}

func TestCountLines_Error(t *testing.T) {
	_, err := countLines("/nonexistent-file-path")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
