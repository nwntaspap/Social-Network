# Verification Gates (`internal/gates/`)

Deterministic Go-based validation gates enforcing architectural rules, security posture, and code conventions. Each gate implements the `Gate` interface (`Name() string`, `Run() Result`) and registers in `cmd/gates/main.go`.

## Goal

Catch violations early — at commit time, pre-push, and CI — before code reaches review. Gates replace subjective review with objective checks: boundary rules, dependency acyclicity, security patterns, branch naming, coverage thresholds, scope drift, formatting, linting, tests, and frontend validation.

## Gate Catalog

| #   | Gate             | File                 | Rule                       | What It Checks                                                                                                               |
| --- | ---------------- | -------------------- | -------------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| 1   | `stack`          | `gate_stack.go`      | Stack Configuration        | `go.mod` (Go version and module name), `.env` (SQLite WAL/busy timeout), Next.js frontend, and `internal/platform` readiness |
| 2   | `d1-layout`      | `gate_layout.go`     | D1 — Vertical Slice Layout | Each feature has `<feature>.go`, `commands/`, `queries/`, `transport/`, `store/`                                             |
| 3   | `d5-boundaries`  | `gate_boundaries.go` | D5 — Import Boundaries     | No feature root/commands/queries imports own store/transport; no store imports commands/queries/transport                    |
| 4   | `d6-dag`         | `gate_dag.go`        | D6 — Dependency DAG        | Feature import graph acyclic (DFS cycle detection); no feature imports `notification`                                        |
| 5   | `tdd`            | `gate_tdd.go`        | Test Coverage per Slice    | Each `commands/` dir with Go files must have matching `*_test.go`                                                            |
| 6   | `migrations`     | `gate_migrations.go` | Migration Integrity        | Every `.up.sql` has matching `.down.sql`; no colon-terminated statements                                                     |
| 7   | `security`       | `gate_security.go`   | Security & Vulns           | `gosec` + `govulncheck` + AST checks: bcrypt cost ≥ 12, no SQL concat, no unconditional `CheckOrigin`                        |
| 8   | `branch`         | `gate_branch.go`     | Branch & Commits           | Branch matches `<user>/<ticket-ID>-<detail>`; commits follow Conventional Commits                                            |
| 9   | `coverage-delta` | `gate_coverage.go`   | Coverage Threshold         | Test coverage drop ≤ 5% vs base branch (git worktree)                                                                        |
| 10  | `scope-drift`    | `gate_scopedrift.go` | Scope Drift                | Advisory: files changed vs base branch count                                                                                 |
| 11  | `format`         | `gate_format.go`     | Code Formatting            | Checks formatting via `gofumpt` and `goimports`                                                                              |
| 12  | `lint`           | `gate_lint.go`       | Code Quality / Lint        | Checks Go code style using `golangci-lint` (or fallback tools `staticcheck` + `go vet`)                                      |
| 13  | `go-test`        | `gate_unittest.go`   | Go Unit Tests              | Runs Go unit tests via `go test -race <new_packages>`                                                                        |
| 14  | `frontend`       | `gate_frontend.go`   | Frontend CI                | Runs lint, format, typecheck, and tests in frontend modules                                                                  |

## File Map

```
internal/gates/
├── README.md                   # This file
├── runner.go                   # Gate interface, Runner, Result, Report, JSON helpers
├── runner_test.go              # Runner unit tests
├── git.go                      # Git helpers: branch, log, diff, merge-base
├── git_test.go                 # Git helper tests
├── helper_test.go              # Mock exec.Command for all gate tests
├── gate_stack.go               # Gate #1: Stack configuration (Go version, module, SQLite WAL, Next.js, platform)
├── gate_layout.go              # Gate #2: D1 — vertical slice directory structure
├── gate_boundaries.go          # Gate #3: D5 — import boundary rules (golangci-lint depguard / AST fallback)
├── gate_dag.go                 # Gate #4: D6 — dependency graph acyclicity (go-arch-lint / DFS)
├── gate_tdd.go                 # Gate #5: TDD — test file presence per commands/ dir
├── gate_migrations.go          # Gate #6: Migration naming + delimiter checks
├── gate_security.go            # Gate #7: Security — gosec + govulncheck + AST checks
├── gate_branch.go              # Gate #8: Branch naming + conventional commits
├── gate_coverage.go            # Gate #9: Coverage delta vs base branch
├── gate_scopedrift.go          # Gate #10: Scope drift advisory
├── gate_format.go              # Gate #11: Formatting verification
├── gate_lint.go                # Gate #12: Lint scoping and validation
├── gate_unittest.go            # Gate #13: Go unit testing
├── gate_frontend.go            # Gate #14: Frontend validation
├── gate_stack_test.go          # Stack gate tests
├── gate_layout_test.go         # Layout gate tests
├── gate_boundaries_test.go     # Boundaries gate tests
├── gate_dag_test.go            # DAG gate tests
├── gate_tdd_test.go            # TDD gate tests
├── gate_migrations_test.go     # Migrations gate tests
├── gate_security_test.go       # Security gate tests
├── gate_branch_test.go         # Branch gate tests
├── gate_coverage_test.go       # Coverage gate tests
├── gate_scopedrift_test.go     # Scope drift gate tests
├── gate_format_test.go         # Format gate tests
├── gate_lint_test.go           # Lint gate tests
├── gate_unittest_test.go       # Unit test gate tests
└── gate_frontend_test.go       # Frontend gate tests
```

## How to Run

```bash
# All gates (human-readable text default)
make gates
# or
go run cmd/gates/main.go --all

# Output structured JSON report
go run cmd/gates/main.go --all --json

# Single gate
go run cmd/gates/main.go --gate=boundaries
```

## Output Format

By default, the runner prints a color-coded summary with emoji indicators:

- ✅ Green bold gate name + dimmed detail message
- ❌ Red bold gate name + error details
- ⏭️ Yellow bold gate name + skip reason

After all gates, a footer shows pass/fail/skip tally and overall result.

ANSI colors auto-disable when `NO_COLOR` env var is set or output piped (emojis replaced with `[PASS]`/`[FAIL]`/`[SKIP]` text badges).

## Architecture

Each gate has a primary tool (e.g. `golangci-lint`, `gosec`, `govulncheck`, `bun`) and a pure Go fallback (AST scan, DFS, `go list`). Tests mock `ExecCommand` via `helper_test.go` to verify both paths without requiring external tools.

Gates are integration-tested as a Go package: `go test ./internal/gates/...` with mock CLI binaries injected through the `ExecCommand` variable.
