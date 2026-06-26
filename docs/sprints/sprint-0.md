# Sprint 0: Foundation (Week 1)

**Outcome:** Setup project tooling, backend scaffold, frontend Next.js base, docker development environment, and fix critical bugs. All team members can run a clean build and development setup.

> **Note:** Several files referenced in Sprint 0 already exist in the repo (Makefile, `.golangci.yml`, `docker-compose.yml`, `docker-compose.dev.yml`, `frontend/`, etc.). Tickets describing creation from scratch **must audit and update** existing files rather than recreating them. Verify existing content first, then extend or fix as needed.

---

## BE-A (Backend A) Tickets

### S0-BE-01: Go Project Scaffold

- **Priority:** P0 (Blocks everything)
- **Type:** Scaffolding/Setup
- **Assignee:** BE-A
- **Story Points:** 3
- **Description:** Set up new project infrastructure from scratch — adjust the existing codebase layout and establish the directory structure according to the vertical-slice target architecture, reusing existing files. This is scaffolding work that creates or restructures project layout, not bug fixing.
- **Detailed Steps:**
  1. Initialize the module at the project root: `go mod init social-network` (or reuse the **existing** `go.mod`, updating imports as needed).
  2. Adjust the existing structure to align with target structure (these directories are NEW structural additions; existing code under the old layered layout stays in place and is not rewritten — only moved/reorganized):
     - `cmd/server/main.go` (main entry point)
     - `internal/core/` (sessions, middlewares, websocket base)
     - `internal/platform/` (database factory, cache, eventbus)
     - `internal/pkg/` (shared packages like hashing, uuid)
     - `db/migrations/` (SQL migration scripts)
  3. Ensure the Go version in `go.mod` is set to `1.24` (audit the **existing** `go.mod` and update the version if stale).
- **Verification:** Run `go vet ./...` and `go build ./cmd/server/`. It should compile successfully (even with empty main function).

---

### S0-BE-02: Bug Fixes (B1.1, B1.2, B1.5)

- **Priority:** P0
- **Type:** Bug Fix (Existing Codebase)
- **Assignee:** BE-A
- **Story Points:** 3
- **Dependencies:** S0-BE-01
- **Description:** Fix migration path delimiter, SQLiteWAL/timeout settings, and SQL injections.
- **Detailed Steps:**
  1. **B1.1 (Migration Delimiter):** In `internal/infra/storage/sqlite/init.go`, change the SQL parsing delimiter from `":"` to `";"` so multi-statement SQL runs correctly.
  2. **B1.2 (SQLite DSN WAL/Timeout):** Modify SQLite connection setup in `internal/infra/storage/sqlite/init.go` and `.env` to pass query parameters: `?_journal_mode=WAL&_busy_timeout=5000` to prevent database locks.
  3. **B1.5 (SQL Injection):** In `internal/infra/storage/sqlite/topics/topicRepo.go` and `internal/infra/storage/sqlite/categories/categoryRepo.go`, sanitize inputs or use an allowed whitelist for dynamic `ORDER BY` directions (`ASC` or `DESC` only). Do NOT interpolate raw input directly into the query template.
- **Verification:** Write unit/integration tests reproducing the issues, ensure they fail before, and pass after implementing the fixes.

---

## BE-B (Backend B) Tickets

### S0-BE-03: Makefile + CI Pipeline

- **Priority:** P0 (Blocks everything)
- **Type:** Scaffolding/Setup
- **Assignee:** BE-B
- **Story Points:** 2
- **Dependencies:** S0-BE-01
- **Description:** Create a robust `Makefile` that handles project setup, linting, formatting, tests, and cleanups.
- **Detailed Steps:**
  1.  Add standard commands:
      - `make install` -> Install ALL project dependencies: Go modules, root JS tooling, `.env` config, SSL certs, Go dev tools, git hooks, and frontend deps.
      - `make setup` -> Install Go development tools and git hooks only (subset of `make install`).
      - `make format` -> Format backend using `gofumpt -s` and `goimports`.
  - `make check-format` -> Fail if code format changes.
  - `make lint` -> Run static checks and `golangci-lint`.
  - `make test` -> Run Go unit tests with race detector: `go test -race -cover ./...`.
  - `make be-ci` -> Backend: `go mod tidy`, format checking, linting (staticcheck + golangci-lint + govulncheck), running tests.
  - `make fe-ci` -> Frontend: ESLint, Prettier format check, `tsc --noEmit`, Vitest.
  - `make ci` -> Full gate (runs both `be-ci` and `fe-ci`).
  - `make db-reset` -> Helper to wipe local SQLite db files for fresh runs.
  - `make seed` -> seed database with test data
  - `make run-backend` -> Run backend application locally.
  - `make run-frontend` -> Run frontend application locally.
  - `make run-all` -> Run both backend and frontend concurrently locally.
  - `make dev` / `make docker-dev` -> Start development environment in Docker.
  - `make build-backend` -> Build backend application locally.
  - `make build-frontend` -> Build frontend application locally.
  - `make build` -> Build both backend and frontend locally.
  - `make docker-build` -> Build Docker images for both backend and frontend.
- **Verification:** Running `make ci` on the command line should execute both BE and FE checks and finish with exit code 0.

---

### S0-BE-04: Bug Fixes (B1.3, B1.4, B1.6, B1.7, B1.8)

- **Priority:** P0
- **Type:** Bug Fix (Existing Codebase)
- **Assignee:** BE-B
- **Story Points:** 3
- **Dependencies:** S0-BE-01
- **Description:** Fix OAuth Scanner, WebSocket origin policy, Prepared statements, WS panics, and RateLimiter leaks.
- **Detailed Steps:**
  1. **B1.3 (OAuth Scan):** In `internal/infra/storage/sqlite/oauth/oauthRepo.go`, adjust `Scan()` arguments to exclude `ctx` since the driver does not take context directly inside `Scan()`.
  2. **B1.4 (WS CheckOrigin):** In `internal/infra/http/ws/handler.go`, restrict origins. Do not return unconditional `true` for `CheckOrigin`. Read origin configuration environment variables.
  3. **B1.6 (Prepared Stmt db.Exec):** In `internal/infra/storage/sqlite/users/userRepo.go`, call `stmt.ExecContext(...)` instead of `db.Exec(...)` to ensure prepared statements are actually executed on the prepared query plan.
  4. **B1.7 (WS Panic Recovery):** In `internal/infra/ws/client.go` inside `ReadPump` and `WritePump` goroutines, add `defer func() { if r := recover(); r != nil { ... } }()` to prevent a single connection crash from bringing down the entire server.
  5. **B1.8 (RateLimiter Leak - core GCRA):** In `internal/infra/middleware/ratelimiter/rateLimiter.go` (not `internal/infra/middleware/rateLimiter.go` which is the HTTP wrapper), add a `stop` channel to close the cleanup ticker when shutting down rate limiting instances to prevent thread/memory leaks.
- **Verification:** Write specific unit tests and run `go test -race ./...` to check concurrent behaviors.

---

## FE-A (Frontend A) Tickets

### S0-FE-01: Next.js Scaffold + Tooling (Frontend Setup)

- **Priority:** P0 (Blocks everything)
- **Type:** Scaffolding/Setup
- **Assignee:** FE-A
- **Story Points:** 5
- **Dependencies:** S0-SD-03 (CI pipeline for frontend gates)
- **Description:** Set up the frontend workspace using Next.js App Router, TypeScript, Tailwind CSS, ESLint + Prettier, testing frameworks, and Bun runtime. Includes all dependency management, lint/format/test gates, and CI integration.
- **Detailed Steps:**
  1. **Install Bun runtime** (prerequisite):
     ```bash
     curl -fsSL https://bun.sh/install | bash
     ```
     Verify: `bun --version` (requires ≥ 1.1).
  2. Create a `frontend/` subdirectory in the project root.
  3. Bootstrap Next.js using `npx -y create-next-app@latest ./` (App router, TypeScript, Tailwind CSS, strict settings, Bun runtime compatibility).
  4. Install frontend dependencies:
     ```bash
     cd frontend && bun install
     ```
  5. Configure ESLint + Prettier formatter and linter in `eslint.config.mjs + .prettierrc`. Set indentation, formatting, and strict rules. Add scripts in `package.json`:
     ```json
     {
       "scripts": {
         "lint": "eslint src/",
         "format": "prettier --write src/",
         "format:check": "prettier --check src/",
         "typecheck": "tsc --noEmit",
         "test": "vitest run",
         "test:watch": "vitest"
       }
     }
     ```
  6. Install `shadcn/ui` tooling and Tailwind custom presets.
  7. Install `Vitest` and `Playwright` testing setups (`vitest.config.ts`, `playwright.config.ts`).
  8. Set up `frontend/Dockerfile` for multi-stage Next.js build (port 3000).
  9. Verify all frontend gates pass:
     - `bun run lint` (ESLint)
     - `bun run format:check` (Prettier format check)
     - `tsc --noEmit` (TypeScript type check)
     - `bun run test` (Vitest)
     - `bun run build` (Next.js production build)
- **Verification:**
  - Run `bun --version` — version ≥ 1.1.
  - Run `bun install` from `frontend/` — no errors.
  - Run `bun run lint`, `bun run format:check`, `tsc --noEmit`, `bun run test`, `bun run build` — all green.
  - Run `make fe-ci` from root — frontend CI pipeline passes.
  - Verify `frontend/Dockerfile` builds: `docker build -f frontend/Dockerfile -t social-fe .`

---

## FE-B (Frontend B) Tickets

### S0-FE-02: shadcn/ui Components + Layout

- **Priority:** P1
- **Type:** Scaffolding/Setup
- **Assignee:** FE-B
- **Story Points:** 3
- **Dependencies:** S0-FE-01
- **Description:** Set up core UI shell structure and shadcn primitives.
- **Detailed Steps:**
  1. Add shadcn components: Button, Input, Card, Dialog, DropdownMenu, Avatar, Badge, Switch, Toast, Tooltip.
  2. Create a global layout file `src/app/layout.tsx` containing:
     - Navigation sidebar (responsive)
     - Header area (user profile dropdown, theme switcher, notification icon placeholder)
     - Main content container (glassmorphic styling, responsive layout)
  3. Create routing page files for `/login`, `/register`, `/feed`, and `/profile/[id]`.
- **Verification:** Start the dev server (`bun run dev`) and visually check that the layout and components render correctly on port 3000.

---

## SD-QA (System Design/QA) Tickets

### S0-SD-01: golangci-lint Config

- **Priority:** P0
- **Type:** Scaffolding/Setup
- **Assignee:** SD-QA
- **Story Points:** 1
- **Dependencies:** S0-BE-03
- **Description:** Configure `.golangci.yml` at the project root to enforce code cleanliness and strict slice boundaries.
- **Detailed Steps:**
  1. Enable linters: `gofumpt`, `goimports`, `gci`, `staticcheck`, `errcheck`, `govet`, `revive`.
  2. Configure `revive` rules to enforce boundary checks (e.g. preventing direct imports between transport/store directories of different slices).
  3. Set execution timeout to 5 minutes.
- **Verification:** Execute `golangci-lint run --timeout=5m` and confirm it succeeds.

---

### S0-SD-02: Docker Compose Development Environment

- **Priority:** P1
- **Type:** Scaffolding/Setup
- **Assignee:** SD-QA
- **Story Points:** 3
- **Dependencies:** S0-BE-01, S0-FE-01
- **Description:** Set up local docker compose environments for both development (hot reload) and production.
- **Detailed Steps:**
  1. Create a `Dockerfile` for the Go backend (multi-stage build: builder -> final minimal alpine image).
  2. Create a `frontend-next/Dockerfile` for the Next.js frontend.
  3. Create `docker-compose.dev.yml` to mount volumes (e.g. backend code, SQLite local file, node_modules) for hot-reloading development.
  4. Create a production `docker-compose.yml` that configures environment variables, ports (8080 for backend, 3000 for frontend), and volumes for persistence.
  5. Create `scripts/docker-build.sh` to automate building and launching.
  6. Create `scripts/makecerts.sh` using openssl to issue local SSL certificates for localhost HTTPS.
- **Verification:** Run `sh scripts/docker-build.sh` and make sure both services start up and communicate properly.

---

### S0-SD-03: CI Pipeline, Go Gates & Pre-commit Hooks

- **Priority:** P1
- **Type:** Scaffolding/Setup
- **Assignee:** SD-QA
- **Story Points:** 5
- **Dependencies:** S0-BE-01, S0-BE-03
- **Description:** Establish the full quality enforcement pipeline: local git hooks via lefthook, automated Go-based verification gates (`internal/gates/`), and CI integration. Ensures bad code is caught at commit time, pre-push, and in CI.
- **Detailed Steps:**
  1. **Lefthook setup:**
     - Install lefthook: `go install github.com/evilmartians/lefthook/v2@latest && lefthook install`.
     - Configure `lefthook.yml` with pre-commit hooks (staged files only):
       - Backend: `gofumpt -l {staged_files} | xargs -r gofumpt -w` + `goimports -w -local social-network {staged_files}` with `stage_fixed: true`.
       - Frontend: `prettier --write` and `eslint` on staged files.
     - Configure pre-push hooks:
       - Backend: `go vet ./...`, `go test -short ./...`, `go build ./...`, `go-arch-lint check`.
       - Frontend: `tsc --noEmit`, `bun run lint`, `bun run test`.
  2. **Go-based verification gates (`internal/gates/`):**
     - Implement gate runner in `cmd/gates/main.go` registering all 10 gates.
     - `runner.go`: exported `var ExecCommand = exec.Command` for testability.
     - `gate_stack.go`: verify Go version ≥ 1.24, module path `social-network`.
     - `gate_layout.go`: verify target directory structure exists (`cmd/server/`, `internal/core/`, `internal/platform/`, `internal/pkg/`, `db/migrations/`).
     - `gate_boundaries.go`: run `golangci-lint run --enable-only=depguard --timeout=5m`, AST fallback when tool missing (checks feature subdirectories for forbidden imports per D5 rules).
     - `gate_dag.go`: run `go-arch-lint check`, DFS cycle-detection fallback with path output (A → B → C → A). Block notification imports per D6.
     - `gate_tdd.go`: verify test files exist for each command/query package.
     - `gate_migrations.go`: verify migration file naming convention (`NNNNNN_name.up.sql` / `.down.sql`), delimiter (`";"`).
     - `gate_security.go`: run `gosec ./...` + 3 custom AST checks:
       - SQL concatenation detection.
       - WebSocket `CheckOrigin` returning unconditional `true` — detect and reject.
       - bcrypt cost: resolve constant-based definitions (not just literals), reject `bcrypt.DefaultCost`.
     - `gate_branch.go`: regex `^[a-z]+/[A-Za-z0-9-]+-[A-Za-z0-9-]+$`, skip merge commits.
     - `gate_coverage.go`: `git worktree add --detach` for safe coverage computation.
     - `gate_scopedrift.go`: detect scope creep (unplanned file changes).
     - Flags: `--all`, `--gate=<name>`. JSON output with exit codes.
  3. **CI pipeline integration:**
     - Add `Makefile` targets:
       - `make gates`: `go run cmd/gates/main.go --all`.
       - `make setup-hooks`: install + configure lefthook.
       - `ci-mod`, `format`, `check-format`, `lint`, `test`, `be-ci`, `fe-ci`, `ci`.
     - `make ci` runs full pipeline: `ci-mod → format → check-format → lint (staticcheck + golangci-lint + govulncheck) → test`.
  4. **Test coverage (target >90%):**
     - Table-driven tests with `t.Run()` subtests for all gates.
     - Mock `ExecCommand` to test tool success/failure/missing-binary paths.
     - Test `Run()` on every gate (Boundaries, DAG, Security, Branch, Coverage, ScopeDrift).
     - Test `toolAvailable()`, `WriteJSON()`, `FindBaseBranch()`, `GitLog()`, `GitBranch()`, `GitDiffFiles()`, `getFeatureDeps()`, `runFallback()`, `checkNotificationImports()`.
     - Assert coverage exceeds 90%.
- **Verification:**
  - Run `lefthook run pre-commit` and `lefthook run pre-push` — all hooks pass.
  - Try to commit an unformatted file — commit gets blocked. Format it — succeeds.
  - Run `go run cmd/gates/main.go --all` — all gates pass with JSON output.
  - Run `go test -race -coverprofile=coverage_gates.out ./internal/gates/...` — coverage >90%.
  - Run `make gates` — full pipeline green.

---

### S0-SD-04: Dev Environment Docs

- **Priority:** P2
- **Type:** Scaffolding/Setup
- **Assignee:** SD-QA
- **Story Points:** 2
- **Dependencies:** All Sprint 0 tickets
- **Description:** Create onboarding documentation for the development environment.
- **Detailed Steps:**
  1. Create `DEVELOPMENT.md` in the project root.
  2. Document:
     - Prerequisite tools (Go 1.24, Bun, Docker, openssl)
     - Step-by-step setup (`make install`, `make dev`, `make ci` (full gate))
     - Docker compose command sequences
     - Branch naming conventions and pull request rules
- **Verification:** A fresh developer should be able to set up their machine by following `DEVELOPMENT.md` alone.
