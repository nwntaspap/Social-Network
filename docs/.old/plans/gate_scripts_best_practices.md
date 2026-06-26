# CI Gate Scripts: Refactor Plan & Tools Best Practices

Following a detailed review of the proposed [deepseek-v4-flash-free-gate-scripts-refactor-plan.md](file:///home/ertval/code/zone-modules/social-network/docs/plans/deepseek-v4-flash-free-gate-scripts-refactor-plan.md) and the master [.opencode/implementation_plan.md](file:///home/ertval/code/zone-modules/social-network/.opencode/implementation_plan.md), this document defines a unified target architecture. It integrates a custom **Go-written gate runner** with industry-standard static-analysis tools (`lefthook`, `go-arch-lint`, `depguard` via `golangci-lint`, and `gosec`).

---

## 1. Analysis of Current Gate Implementation Problems

The current CI gate implementation uses shell scripts (`.sh`) under `scripts/gates/`. These scripts exhibit typical shell scripting anti-patterns that lead to fragility and hard-to-debug crashes:

| Script                    | Primary Issue                                | Root Cause                                                                                                                                                                                                                       | Best Practice Fix                                                                                                            |
| :------------------------ | :------------------------------------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | :--------------------------------------------------------------------------------------------------------------------------- |
| `run-all.sh`              | **SIGPIPE Crash (Exit Code 141)**            | Piping output under `set -euo pipefail` to a truncating command (`head -5`). When `head` terminates, the pipe closes, causing the preceding `echo` to trigger a SIGPIPE crash.                                                   | Programmatic output buffering/truncation in a typed runner.                                                                  |
| `check-coverage-delta.sh` | **Workspace Mutation & Empty Stash Failure** | Running `git stash` on a clean tree (which returns exit 1), checking out `main` in the active workspace (which modifies files and leaves the dev in a detached HEAD if it crashes), and using `git stash pop` on an empty stack. | Use **Git Worktrees** in a temporary directory to perform base branch tests, avoiding any mutations of the active workspace. |
| `check-d6-dag.sh`         | **Empty List Crash (Exit Code 1)**           | Under `set -e`, running a grep search that returns no matches (normal in pre-migration state) causes grep to exit with status 1, crashing the script.                                                                            | Tolerant stream consumption and parsing.                                                                                     |
| `check-branch.sh`         | **Merge-Base History Flood**                 | Missing or incorrect merge-base resolutions (e.g. checking local `main` instead of `origin/main` when branch history is diverged) causes checking hundreds of historical commits.                                                | Robust Git commit log analysis using Git plumbing commands or libraries.                                                     |
| `check-d5-boundaries.sh`  | **False Positives in Import Check**          | Using naive string-matching regexes (`grep -rn 'import'`) which matches comments, documentation, and test code (`*_test.go`).                                                                                                    | **Abstract Syntax Tree (AST)** parsing to analyze actual import declarations in Go code.                                     |
| `check-migrations.sh`     | **Fragile Delimiter Scanning**               | Complex multi-line SQL validation using generic grep and regexes which struggle with statement blocks and SQL comments.                                                                                                          | Programmatic parsing of SQL statements using SQL tokenizers or lexers.                                                       |
| `check-security.sh`       | **Fragile Regex Cost Detection**             | Simple regex search for `cost = N` which misses default constants like `bcrypt.DefaultCost`.                                                                                                                                     | AST parsing to verify code assignments.                                                                                      |

---

## 2. Comparing Alternative Languages: Go vs. TypeScript

To enable type checking, compiler safety, and unit testing, we compare the two primary languages already in the project stack: **Go** (backend) and **TypeScript** (frontend). Introducing a third language (e.g., Python) violates the consistency principle and adds runtime setup overhead.

### Option A: Go (Recommended)

Writing gate checkers in Go is the **most robust** solution for this codebase.

- **Pros:**
  1.  **Zero-Dependency Portability:** Go code compiles into static binaries. It runs instantly on any machine (and in any minimal CI runner) without needing to install `node_modules` or dependencies.
  2.  **First-Class AST Parser:** The Go standard library includes `go/parser`, `go/token`, and `go/ast`. This allows us to parse Go files and programmatically inspect import blocks for **D5 Boundary checks** with 100% accuracy, bypassing fragile regexes.
  3.  **Built-in Unit Testing:** We can write standard `*_test.go` files and execute tests with `go test -race ./...`. We can mock filesystems, Git commands, and package structures.
  4.  **No SIGPIPE Risks:** Using `os/exec` enables programmatic command running, stream piping, and graceful truncation without shell-pipe signals.
  5.  **Clean Git Worktree API:** We can easily write Go logic to spawn a temporary `git worktree` under `/tmp/` to run coverage checks on `origin/main` without touching the developer's working directory.

### Option B: TypeScript (with Node.js / `tsx`)

Writing gate checkers in TypeScript.

- **Pros:**
  1.  Easy for frontend developers to write and maintain.
  2.  Strong type checking and standard testing frameworks (Vitest, Jest).
- **Cons:**
  1.  **Runtime Overhead:** Requires running `npm install` or having `node_modules` loaded. If the backend pipeline is isolated from the frontend, it introduces a Node dependency to Go-only tasks.
  2.  **Weak Go Parsing Support:** Checking Go imports (D5 boundary checks) or dependency graphs (D6) from TypeScript would require either calling out to `go list` (and parsing output) or implementing a custom regex parser in JS, which retains the regex-based fragility.

---

## 3. CI Quality & Validation Tools Setup

We leverage best-of-breed specialized tools for standard checks, allowing the custom Go gates to focus on application-specific rules:

### A. Local Hook Management: Lefthook (`lefthook.yml`)

Replaces manual or Node-based hook runners. Runs checks concurrently on staged files:

- **Pre-commit:** Runs `gofumpt` and `goimports` formatting, and executes `go vet` to catch syntax errors instantly.
- **Pre-push:** Runs short tests (`go test -short ./...`), verifies builds, and checks architecture limits using `go-arch-lint`.

### B. Vertical Slice Boundaries (D5): `depguard` via `golangci-lint`

Instead of parsing Go files with generic scripts, we use `depguard` inside `.golangci.yml` to prevent layer violations (e.g. enforcing that `commands` and `queries` cannot import `store` or `transport` packages).

### C. Dependency DAG Acyclicity (D6): `go-arch-lint` (`.go-arch-lint.yml`)

Ensures clean separation of feature slices and enforces that `notification` is never imported. This completely replaces fragile `grep`-based dependency checking.

### D. Security Pattern Scans: `gosec`

A dedicated AST-based security linter that detects low bcrypt costs, hardcoded secrets, and unsafe SQL string concatenations.

---

## 4. Go Gate Runner Directory Structure (`cmd/gates/` & `internal/gates/`)

For rules requiring custom domain logic (like git-history validation, coverage deltas, and SQL migration checks), we implement a clean, compiled Go CLI runner under `cmd/gates/` and a **single flat library package** `gates` under `internal/gates/`.

We do **not** need a package per gate. A flat package structure avoids over-engineering, reduces import boilerplate, and allows all gates to directly share helpers (such as git and output formatting) without circular imports or exporting private variables.

```
scripts/gates/             ← Deprecated bash scripts (to archive)
cmd/gates/
  └── main.go              # CLI parser, compiles/runs the gates (package main)
internal/gates/
  ├── git.go               # Shared git helpers (merge-base, log, diff) (package gates)
  ├── errors.go            # Structured gate errors (package gates)
  ├── output.go            # JSON formatting schema (no pipefail/SIGPIPE) (package gates)
  ├── gate_stack.go        # Validates Go version & mod config (Gate #1) (package gates)
  ├── gate_layout.go       # Validates D1 vertical slice layout (Gate #2) (package gates)
  ├── gate_boundaries.go   # Runs go-arch-lint & depguard wrapper (Gate #3) (package gates)
  ├── gate_dag.go          # Custom package import acyclicity checks (Gate #4) (package gates)
  ├── gate_tdd.go          # Verifies that test files exist for all code (Gate #6) (package gates)
  ├── gate_migrations.go   # Validates SQLite migration filenames & delimiters (Gate #7) (package gates)
  ├── gate_security.go     # Wraps gosec and custom bcrypt AST rules (Gate #8) (package gates)
  ├── gate_branch.go       # Validates branch names & conventional commits (Gate #9) (package gates)
  ├── gate_scopedrift.go   # Verifies changes are within ticket scope (package gates)
  ├── gate_coverage.go     # Compares branch coverage vs base branch (Gate #13) (package gates)
  └── gates_test.go        # Unified tests for all gates (package gates)
```

---

## 5. Gate-by-Gate Refactor Mapping

| Gate              | Bash Script               | Problem                       | Go Tool / AST Fix                                                             |
| :---------------- | :------------------------ | :---------------------------- | :---------------------------------------------------------------------------- |
| **#1 Stack**      | `check-stack.sh`          | Fragile version parsing       | Read `go.mod` via `go/parser` or JSON editor; inspect `runtime.Version()`.    |
| **#2 Layout**     | `check-d1-layout.sh`      | Rigid filesystem scans        | Programmatic file checks (`os.Stat`) with table-driven tests.                 |
| **#3 Boundaries** | `check-d5-boundaries.sh`  | Naive grep matches comments   | Integrates `depguard` in `.golangci.yml` and Go AST parsing of import blocks. |
| **#4 DAG**        | `check-d6-dag.sh`         | Empty feature list crash      | Programmatic check running `go list -json` and executing topological sort.    |
| **#6 TDD**        | `check-tdd.sh`            | Regex parsing of find results | Programmatic package enumeration via `go/parser` checking file pairs.         |
| **#7 Migrations** | `check-migrations.sh`     | Wrong delimiter regexes       | Reads SQL files and scans statement endings using strict tokenizer regexes.   |
| **#8 Security**   | `check-security.sh`       | Misses constant definitions   | Walks Go AST using `go/ast` to check bcrypt costs and query structures.       |
| **#9 Branch**     | `check-branch.sh`         | Mismatched Git histories      | Resolves base ref dynamically (tries `main`, falls back to `origin/main`).    |
| **#10 Drift**     | `check-scope-drift.sh`    | Duplicate code with branch    | Reuses shared `git.go` package, sets advisory flags.                          |
| **#13 Coverage**  | `check-coverage-delta.sh` | Workspace mutation crash      | Spawns a temporary `git worktree` to compile and test the base branch safely. |
| **Master**        | `run-all.sh`              | SIGPIPE 141 error             | Go orchestrator collects outputs from all packages, formatting to JSON.       |

---

## 6. Architectural Implementation Highlights

### A. Git Worktree Coverage Isolation

To prevent corrupting the local developer workspace with branch switches, the Go implementation uses `git worktree` to execute baseline tests in a temporary system directory:

```go
package gates

import (
	"os"
	"os/exec"
	"path/filepath"
)

func CheckCoverage(baseBranch string) (float64, error) {
	tempDir := filepath.Join(os.TempDir(), "social-network-base-cov")
	defer os.RemoveAll(tempDir)

	// Add temporary worktree for base branch comparison
	worktreeCmd := exec.Command("git", "worktree", "add", tempDir, baseBranch)
	if err := worktreeCmd.Run(); err != nil {
		return 0, err
	}
	defer exec.Command("git", "worktree", "prune").Run()

	// Run tests in the temp worktree
	testCmd := exec.Command("go", "test", "-coverprofile=coverage.out", "./...")
	testCmd.Dir = tempDir
	if err := testCmd.Run(); err != nil {
		return 0, err
	}
	return parseCoverageFile(filepath.Join(tempDir, "coverage.out"))
}
```

### B. Safe Git Commit Resolution

The custom runner resolves merge-bases dynamically without erroring on detached heads or shallow clones:

```go
package gates

import (
	"os/exec"
	"strings"
)

func FindBaseBranch() string {
	// Try local main first
	cmd := exec.Command("git", "merge-base", "main", "HEAD")
	if err := cmd.Run(); err == nil {
		return "main"
	}
	// Fall back to origin/main for Gitea/CI environments
	return "origin/main"
}
```

---

## 7. Testing Strategy

All Go gate files are covered by matching `_test.go` files in the same `gates` package. They use test fixtures (mock directories, temporary git repositories, and malformed files) to assert that:

1. Valid code passes the gate in under 100ms.
2. Deliberately broken layouts, cyclic imports, low-cost bcrypt instances, or bad commit formats fail with explicit, descriptive error messages.
3. All tests run concurrently and easily via `go test ./internal/gates`.

---

## 8. Next Steps

1. **Verify Shell Script Hotfixes:** Apply immediate fixes to the existing bash scripts (temporarily disabling `pipefail` during output truncation, falling back to `origin/main` for unrelated git histories, fixing migration regexes) to make the CI gate runner functional right now.
2. **Implement Go-Based Gates:** Incrementally rewrite each gate script into `scripts/gates/` in Go.
3. **Write Unit Tests:** Add Go tests verifying each gate against intentional violations.
4. **Update Pipeline / Lefthook:** Switch the Makefile and `lefthook.yml` to call `go run scripts/gates/main.go --gate=<name>` or precompile the gates during the bootstrap phase.
