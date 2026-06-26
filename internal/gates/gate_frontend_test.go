package gates

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFrontendGate_Run(t *testing.T) {
	oldExec := ExecCommand
	oldLook := lookPath
	defer func() {
		ExecCommand = oldExec
		lookPath = oldLook
	}()
	ExecCommand = mockExecCommand

	t.Setenv("MOCK_REV_FAIL", "1")

	g := &FrontendGate{}

	// Scenario 1: No frontend exists
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)
	res := g.Run()
	if res.Status != "SKIP" {
		t.Errorf("expected skip when no frontend exists, got: %s", res.Status)
	}

	// Scenario 2: frontend-next exists, PASS
	tmpDir2 := t.TempDir()
	if err := os.Mkdir(filepath.Join(tmpDir2, "frontend-next"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir2, "frontend-next", "package.json"), []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(tmpDir2)

	t.Setenv("MOCK_FAIL", "0")
	res = g.Run()
	if res.Status != "PASS" {
		t.Errorf("expected PASS when frontend exists, got: %s (%s)", res.Status, res.Message)
	}

	// Scenario 3: frontend-next exists, FAIL
	t.Setenv("MOCK_FAIL", "1")
	res = g.Run()
	if res.Status != "FAIL" {
		t.Errorf("expected FAIL when frontend commands fail, got: %s (%s)", res.Status, res.Message)
	}
}
