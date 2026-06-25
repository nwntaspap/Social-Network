package gates

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDAGGate_Run(t *testing.T) {
	oldExec := ExecCommand
	oldLook := lookPath
	defer func() {
		ExecCommand = oldExec
		lookPath = oldLook
	}()
	ExecCommand = mockExecCommand

	g := &DAGGate{}

	// 1. Tool available, PASS
	lookPath = func(name string) (string, error) { return name, nil }
	t.Setenv("MOCK_FAIL", "0")
	res := g.Run()
	if res.Status != "PASS" {
		t.Errorf("expected DAG tool PASS, got: %s (%s)", res.Status, res.Message)
	}

	// 2. Tool available, FAIL
	t.Setenv("MOCK_FAIL", "1")
	res = g.Run()
	if res.Status != "FAIL" {
		t.Errorf("expected DAG tool FAIL, got: %s", res.Status)
	}

	// 3. Tool missing, fallback cycle detection FAIL
	lookPath = func(name string) (string, error) { return "", errors.New("missing") }
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "user"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "follow"), 0o700); err != nil {
		t.Fatal(err)
	}
	g.InternalDir = dir
	t.Setenv("MOCK_FAIL", "0")
	t.Setenv("MOCK_CYCLE", "1")
	res = g.Run()
	if res.Status != "FAIL" {
		t.Errorf("expected cycle detection FAIL, got: %s (%s)", res.Status, res.Message)
	}
	if !strings.Contains(res.Message, "CIRCULAR") {
		t.Errorf("expected circular path in message, got: %s", res.Message)
	}

	// 4. Fallback notification import FAIL
	t.Setenv("MOCK_CYCLE", "0")
	t.Setenv("MOCK_NOTIF", "1")
	res = g.Run()
	if res.Status != "FAIL" {
		t.Errorf("expected notification import FAIL, got: %s (%s)", res.Status, res.Message)
	}
}

func TestDAGGate_NotificationImportsNonexistent(t *testing.T) {
	// checkNotificationImports with nonexistent directory
	dg := &DAGGate{InternalDir: "/nonexistent"}
	res := dg.checkNotificationImports()
	if res != nil {
		t.Errorf("expected nil for nonexistent directory, got: %v", res)
	}
}
