package gates

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// mockExecCommand is executed when ExecCommand calls our mocked binary.
// It acts as a fake CLI tool depending on environment variables and arguments.
//
//nolint:gocognit,nestif,gosec
func mockExecCommand(command string, args ...string) *exec.Cmd {
	var script string

	switch command {
	case "golangci-lint":
		if os.Getenv("MOCK_FAIL") == "1" {
			script = "echo 'depguard violation: transport importing store'; exit 1"
		} else {
			script = "echo 'no violations'; exit 0"
		}
	case "go-arch-lint":
		if os.Getenv("MOCK_FAIL") == "1" {
			script = "echo 'cycle detected'; exit 1"
		} else {
			script = "echo 'clean'; exit 0"
		}
	case "gosec":
		if os.Getenv("MOCK_FAIL") == "1" {
			script = "echo 'gosec violation: hardcoded password'; exit 1"
		} else {
			script = "echo 'clean'; exit 0"
		}
	case "staticcheck":
		if os.Getenv("MOCK_FAIL") == "1" {
			script = "exit 1"
		} else {
			script = "exit 0"
		}
	case "gofumpt":
		if os.Getenv("MOCK_FAIL") == "1" {
			script = "echo 'main.go'; exit 1"
		} else {
			script = "exit 0"
		}
	case "goimports":
		if os.Getenv("MOCK_FAIL") == "1" {
			script = "echo 'main.go'; exit 1"
		} else {
			script = "exit 0"
		}
	case "govulncheck":
		if os.Getenv("MOCK_FAIL") == "1" {
			script = "echo 'vulnerability found'; exit 1"
		} else {
			script = "exit 0"
		}
	case "bun":
		if os.Getenv("MOCK_FAIL") == "1" {
			script = "exit 1"
		} else {
			script = "exit 0"
		}
	case "go":
		if len(args) > 0 {
			switch args[0] {
			case "list":
				if os.Getenv("MOCK_FAIL") == "1" {
					script = "exit 1"
				} else if os.Getenv("MOCK_CYCLE") == "1" {
					pkg := args[len(args)-1]
					if strings.Contains(pkg, "/user/...") {
						script = `echo '{"Imports": ["social-network/internal/follow"]}'`
					} else if strings.Contains(pkg, "/follow/...") {
						script = `echo '{"Imports": ["social-network/internal/user"]}'`
					}
				} else if os.Getenv("MOCK_NOTIF") == "1" {
					pkg := args[len(args)-1]
					if strings.Contains(pkg, "/user/...") {
						script = `echo '{"Imports": ["social-network/internal/notification"]}'`
					}
				} else {
					script = `echo '{"Imports": []}'`
				}
			case "test":
				if os.Getenv("MOCK_FAIL") == "1" {
					script = "exit 1"
				} else {
					// Locate coverprofile arg and touch it
					covPath := ""
					for _, arg := range args {
						if val, ok := strings.CutPrefix(arg, "-coverprofile="); ok {
							covPath = val
						}
					}
					if covPath != "" {
						script = fmt.Sprintf("touch %s; exit 0", covPath)
					} else {
						script = "exit 0"
					}
				}
			case "vet":
				if os.Getenv("MOCK_FAIL") == "1" {
					script = "exit 1"
				} else {
					script = "exit 0"
				}
			case "tool":
				if len(args) > 1 && args[1] == "cover" {
					if os.Getenv("MOCK_FAIL") == "1" {
						script = "exit 1"
					} else if os.Getenv("MOCK_COVER_MALFORMED") == "1" {
						script = "echo 'total'"
					} else {
						script = "echo 'total: (statements) 92.5%'"
					}
				}
			}
		}
	case "git":
		if os.Getenv("MOCK_FAIL") == "1" {
			script = "exit 1"
		} else if len(args) > 0 {
			switch args[0] {
			case "merge-base":
				script = "echo 'abcdef123456'"
			case "log":
				if os.Getenv("MOCK_GIT_EMPTY") == "1" {
					script = "echo ''"
				} else if os.Getenv("MOCK_COMMIT_FAIL") == "1" {
					script = "echo 'Fixing code'"
				} else {
					script = "echo 'feat(user): add auth handler'"
				}
			case "rev-parse":
				if os.Getenv("MOCK_REV_FAIL") == "1" {
					script = "exit 1"
				} else if len(args) > 1 && args[1] == "--show-toplevel" {
					script = "echo '/mock/root'"
				} else if os.Getenv("MOCK_REV_MAIN") == "1" {
					script = "echo 'main'"
				} else if os.Getenv("MOCK_REV_FAIL") == "1" {
					script = "exit 1"
				} else if os.Getenv("MOCK_REV_UNAPPROVED") == "1" {
					script = "echo 'unapproved/S1-BE-01-add-auth'"
				} else {
					script = "echo 'ekaramet/S1-BE-01-add-auth'"
				}
			case "diff":
				if os.Getenv("MOCK_GIT_EMPTY") == "1" {
					script = "echo ''"
				} else {
					script = "echo 'internal/user/user.go'"
				}
			case "worktree":
				if len(args) > 3 && args[1] == "add" {
					script = fmt.Sprintf("mkdir -p %s; exit 0", args[3])
				} else {
					script = "exit 0"
				}
			}
		}
	default:
		script = "exit 0"
	}

	return exec.CommandContext(context.Background(), "sh", "-c", script)
}

func createTempGoMod(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "go.mod")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
