package gates

import (
	"testing"
)

func TestGitHelpers(t *testing.T) {
	oldExec := ExecCommand
	defer func() { ExecCommand = oldExec }()
	ExecCommand = mockExecCommand

	// FindBaseBranch
	base := FindBaseBranch()
	if base != "main" {
		t.Errorf("expected base main, got: %s", base)
	}

	// GitLog
	log, err := GitLog("main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(log) != 1 || log[0] != "feat(user): add auth handler" {
		t.Errorf("unexpected git log: %v", log)
	}

	// GitBranch
	branch := GitBranch()
	if branch != "ekaramet/S1-BE-01-add-auth" {
		t.Errorf("expected ekaramet/S1-BE-01-add-auth, got: %s", branch)
	}

	// GitDiffFiles
	files, err := GitDiffFiles("main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 || files[0] != "internal/user/user.go" {
		t.Errorf("unexpected files: %v", files)
	}
}

func TestGitHelpers_Empty(t *testing.T) {
	oldExec := ExecCommand
	defer func() { ExecCommand = oldExec }()
	ExecCommand = mockExecCommand

	t.Setenv("MOCK_GIT_EMPTY", "1")

	log, err := GitLog("main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if log != nil {
		t.Errorf("expected nil log for empty output, got: %v", log)
	}

	files, err := GitDiffFiles("main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if files != nil {
		t.Errorf("expected nil files for empty output, got: %v", files)
	}
}

func TestGitHelpers_RepoRoot(t *testing.T) {
	oldExec := ExecCommand
	defer func() { ExecCommand = oldExec }()
	ExecCommand = mockExecCommand

	root := GitRepoRoot()
	if root != "/mock/root" {
		t.Errorf("expected /mock/root, got: %s", root)
	}
}

func TestGitHelpers_RepoRootFail(t *testing.T) {
	oldExec := ExecCommand
	defer func() { ExecCommand = oldExec }()
	ExecCommand = mockExecCommand

	t.Setenv("MOCK_REV_FAIL", "1")
	root := GitRepoRoot()
	if root != "" {
		t.Errorf("expected empty string on failure, got: %s", root)
	}
}
