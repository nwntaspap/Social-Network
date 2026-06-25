# Context Engineering & Agent Architecture Optimization — Go Validation Gates

This implementation plan details the target state for local quality gates and agent execution boundaries using compiled Go gates under `internal/gates/` and `cmd/gates/` rather than legacy bash scripts.

---

## Architecture Overview

We use a Go-based gate runner to perform validation checks at the PR and pre-push stage.

```
┌─────────────────────────────────────────────────┐
│ L0: Local Pre-commit/Pre-push (Lefthook)        │ ← instant, file-specific
│ L1: Deterministic Go Gates (cmd/gates/)         │ ← blocking, PR-time
│ L2: Linter Ecosystem (golangci-lint + go-arch)  │ ← blocking, make ci
│ L3: LLM Semantic Review (agent subagents)       │ ← advisory, deep reasoning
└─────────────────────────────────────────────────┘
```

## Go Gate Runner Component Details

All quality gates are implemented as standard Go code within `internal/gates/` and registered in `cmd/gates/main.go`.

### Gates Taxonomy

1. **`stack`**: Verifies Go version is 1.25 and module path is `social-network`.
2. **`d1-layout`**: Verifies D1 vertical slice folder layout.
3. **`d5-boundaries`**: Invokes `golangci-lint run --enable-only=depguard` to enforce D5 boundaries, with an AST fallback that also checks feature root files.
4. **`d6-dag`**: Checks for circular imports between features via `go-arch-lint check` or a custom DFS fallback. Blocks `notification` package imports.
5. **`tdd`**: Enforces that if `commands/` contains Go files, matching `*_test.go` must exist.
6. **`migrations`**: Validates db migration files (sequential names, up/down pairs, semicolon delimiters).
7. **`security`**: Runs `gosec ./...`, detects unconditional WebSocket `CheckOrigin` returning `true`, checks bcrypt costs (>= 12), and rejects `bcrypt.DefaultCost`.
8. **`branch`**: Enforces ticket branch naming schema `<username>/<ticket-ID>-<detail>` and conventional commit formatting.
9. **`scope-drift`**: Advisory gate reporting number of modified files from base branch.
10. **`coverage-delta`**: Computes test coverage using `git worktree add --detach` and checks that coverage doesn't drop below the allowed threshold.

## Makefile Targets

- `make gates`: Runs `go run cmd/gates/main.go --all`
- `make ci`: Runs full backend CI checks (format, lint, test)
- `make fe-ci`: Runs frontend ESLint + Prettier/Vitest/build checks
- `make setup-hooks`: Installs and configures Lefthook hooks
