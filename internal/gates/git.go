/*
Git provides shared helper functions for Git repository interaction,
including retrieving the current branch name, finding the merge base,
fetching the commit log, and identifying changed files.
*/

package gates

import (
	"strings"
)

// FindBaseBranch resolves the correct base branch for comparison.
// Tries local main first, falls back to origin/main for CI/Gitea environments.
func FindBaseBranch() string {
	cmd := ExecCommand("git", "merge-base", "main", "HEAD")
	if err := cmd.Run(); err == nil {
		return "main"
	}
	return "origin/main"
}

// GitLog returns commit subjects between base..HEAD.
func GitLog(base string) ([]string, error) {
	// #nosec G204
	cmd := ExecCommand("git", "log", base+"..HEAD", "--format=%s")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return nil, nil
	}
	return strings.Split(raw, "\n"), nil
}

// GitBranch returns the current branch name.
func GitBranch() string {
	cmd := ExecCommand("git", "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

// GitRepoRoot returns the absolute path to the root of the git repository.
func GitRepoRoot() string {
	cmd := ExecCommand("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// GitDiffFiles returns filenames changed between base..HEAD.
func GitDiffFiles(base string) ([]string, error) {
	// #nosec G204
	cmd := ExecCommand("git", "diff", base+"..HEAD", "--name-only")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return nil, nil
	}
	return strings.Split(raw, "\n"), nil
}
