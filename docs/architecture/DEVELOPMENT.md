# 🛠️ Development Environment & Contributor Guide

Welcome to the **Social Network** project! This document outlines the instructions to set up your local development environment, build the project, run tests, and adhere to our team's contribution guidelines.

---

## 📋 Table of Contents

1. [Prerequisite Tools](#-prerequisite-tools)
2. [Step-by-Step Setup](#%EF%B8%8F-step-by-step-setup)
3. [Docker Compose Commands](#-docker-compose-commands)
4. [Local Development Without Docker](#-local-development-without-docker)
5. [Linting & Formatting Checks](#-linting--formatting-checks)
6. [Branch Naming & PR Conventions](#-branch-naming--pr-conventions)
7. [Refactoring & TDD Methodology](#-refactoring--tdd-methodology)

---

## 🛠️ Prerequisite Tools

Before starting, ensure you have the following tools installed on your local machine:

| Tool        | Recommended Version | Purpose                             | Installation Guide                                                |
| ----------- | ------------------- | ----------------------------------- | ----------------------------------------------------------------- |
| **Go**      | `≥ 1.25`            | Backend service and tools execution | [Go Installation](https://go.dev/doc/install)                     |
| **Bun**     | `1.1`+              | Frontend package manager & runtime  | [Bun Installation](https://bun.sh)                                |
| **Docker**  | Latest              | Containerization for local services | [Docker Desktop](https://www.docker.com/products/docker-desktop/) |
| **openssl** | Latest              | Local HTTPS certificate generation  | Pre-installed on most Unix systems                                |

---

## 🚀 Step-by-Step Setup

Follow these steps to configure your local development environment:

### 1. Install All Dependencies

Install every project dependency (Go modules, root JS tooling, `.env` config, SSL certs, Go dev tools, git hooks, and frontend packages) with a single command:

```bash
make install
```

This runs:

- `go mod download` — Go library dependencies
- `npm ci` — Root JS tooling (from `package-lock.json`)
- `cp -n .env.example .env` — Environment config (safe, won't overwrite)
- `bash scripts/makecerts.sh` — SSL/TLS certificates for HTTPS/WSS (idempotent: skips if certs exist)
- `make tools` — Go development tools (`goimports`, `staticcheck`, `golangci-lint`, `govulncheck`, `gofumpt`, `gosec`, `go-arch-lint`)
- `make setup-hooks` — Lefthook git hooks (pre-commit/pre-push)
- `bun install` — Frontend dependencies

> [!TIP]
> If you already have the environment set up and just need to reinstall Go tools, use `make setup` (runs tools + hooks only).

### 2. Build and Start the Environment

Launch the local Docker containers (running Go backend, Next.js frontend, and persistent SQLite database):

```bash
make dev
```

> [!NOTE]
> `make dev` is an alias to `make docker-dev` which launches the dev compose configuration.

### 3. Verify Setup (Run CI Checks)

Ensure everything compiles, formats, lints, and tests successfully (both BE and FE):

```bash
make ci
```

---

## 🐳 Docker Compose Commands

We support Docker Compose configurations for both hot-reloading development and production-like builds.

### Start Dev Services (With Hot-Reload)

Starts the backend on port `8080` and frontend on port `3000` with volume mounts to reload on code changes:

```bash
make docker-dev
```

If you need to rebuild the containers during startup:

```bash
make docker-dev-build
```

### Stop Dev Services

Tear down running containers:

```bash
make docker-down
```

### Reset Database & Clean Resources

Remove containers, volumes, local SQLite databases, and local images:

```bash
make docker-clean
```

Or to clean just the SQLite database file:

```bash
make db-clean
```

### Run SQL Queries Inside Container

If you need to inspect database state inside the running container, run:

```bash
make docker-db
```

_(This opens the database `db/data/forum.db` with SQLite header output)._

---

## 💻 Local Development Without Docker

If you prefer to run services natively on your host OS:

> [!TIP]
> Run `make install` first — it handles Go module download, `.env` setup, certs, and frontend deps (`bun install`) in one step. Then use the commands below.

### Backend Development

1. Start the Go server:
   ```bash
   go run cmd/server/main.go
   ```

### Frontend Development

1. Start the Next.js development server:
   ```bash
   cd frontend && bun run dev
   ```
   _(By default, the server runs on `http://localhost:3000`)_

> If you skipped `make install`, run `bun install` inside `frontend/` first.

---

## 🔍 Linting, Formatting & Verification Gates

Before pushing code or opening a pull request, ensure it passes all verification gates.

### Full CI Pipeline & Gates

```bash
make gates
```

Decoupled from legacy blanket check. Performs:

1. `go build ./...` — Compiles all code (legacy + new) for build safety.
2. `go run cmd/gates/main.go --all` — Runs the custom Go verification gates.
3. `make be-ci-new` — Runs new-code scoped check.
4. `make fe-ci` — Runs frontend CI check (scoped to `frontend-next/` if it exists).

_Note: You can run legacy blanket checks via `make ci` (runs `make be-ci` + `make fe-ci`), which checks all legacy files as well._

### Go Verification Gates

Deterministic Go-based gates under `internal/gates/` (see [README](../../internal/gates/README.md)) enforce architectural and convention rules:

```bash
make gates
# Or directly: go run cmd/gates/main.go --all
```

| Gate       | Tool/Fallback                | What It Checks                                                                     |
| ---------- | ---------------------------- | ---------------------------------------------------------------------------------- |
| Stack      | go version / go.mod          | Go ≥ 1.25, module path `social-network`                                            |
| Layout     | os.Stat                      | Target directory structure exists                                                  |
| Boundaries | golangci-lint depguard / AST | D5 — forbidden cross-slice imports                                                 |
| DAG        | go-arch-lint / DFS           | D6 — dependency graph acyclicity (supports go-arch-lint)                           |
| TDD        | os.Stat                      | Test files exist per command/query                                                 |
| Migrations | glob / grep                  | Migration naming (`NNNNNN_name.up.sql`), delimiter (`";"`)                         |
| Security   | gosec / custom AST           | SQL concat, WebSocket CheckOrigin, bcrypt cost (scoped to new code)                |
| Branch     | regex                        | Branch naming convention `<user>/<ticket>-<detail>` (includes `dev` scope support) |
| Coverage   | git worktree + go test       | Test coverage threshold (>90%)                                                     |
| ScopeDrift | git diff                     | Unplanned file changes                                                             |

Output is JSON with exit codes (0 = pass). Run individual gates: `--gate=<name>`.

### Pre-commit & Pre-push Hooks (Lefthook)

Install hooks:

```bash
make setup-hooks
```

Pre-commit auto-formats staged Go/frontend files. Pre-push runs `go vet`, `go test -short`, `go build`, `go-arch-lint`, `tsc --noEmit`, `eslint`. Bypass: `--no-verify`.

### Backend Validation

Run scoped linter (`golangci-lint`, `staticcheck`, and `govulncheck`), formatting, and test checks:

```bash
make be-ci-new
```

To run tests separately with code coverage:

```bash
make test-new
```

### Frontend Validation

Run the composite frontend gate (scopes to `frontend-next/` if it exists, falling back to legacy `frontend/`):

```bash
make fe-ci
```

Individual commands (run from `frontend-next/`):

```bash
bun run lint        # ESLint lint
bun run format:check  # Prettier format check
tsc --noEmit          # TypeScript type check
bun run test          # Vitest
```

---

## 🌿 Branch Naming & PR Conventions

We enforce a strict branching and commit strategy to maintain a clean git history.

### Branch Naming Convention

Branches must be named in the following format:

```
<username>/<ticket/issue-ID>-<detail>
```

- **username**: Your own Gitea username — resolve via `tea whoami` or `cat ~/.config/tea/config.yml | grep 'user:' | head -1 | awk '{print $2}'`. Known devs: `epapamic`, `ekaramet`, `dkotsi`, `geoikonomou`, `smichail`
- **ticket/issue-ID**: Ticket ID from `docs/sprints/ticket-tracker.md` (e.g. `S3-BE-01`) or GitHub/Gitea issue number (e.g. `42`). **Required** — maps branch to work item.
- **detail**: kebab-case description (e.g. `db-factory`, `fix-sqlite-busy-timeout`).

**Examples:**

- `epapamic/S1-BE-05-db-factory`
- `ekaramet/fix-websocket-panic`
- `geoikonomou/docs-dev-env`

### Pull Request & Commit Strategy

1. **Trunk-Based Development**: Keep feature branches short-lived (≤ 3 days).
2. **Squash Merge**: All branches are squashed into `main` as a single commit.
3. **Conventional Commits with Ticket ID**: The squashed commit must follow the [Conventional Commits](https://www.conventionalcommits.org/) format with ticket ID:
   ```
   <type>(<scope>)[<ID>]: <description>
   ```
   _Examples:_
   - `feat(user)[S2-BE-17]: add register command with age validation`
   - `fix(core)[42]: recover from WebSocket goroutine panic`
   - `docs(dev): add onboarding setup instructions`
4. **PR Description**: Copy `.github/PULL_REQUEST_TEMPLATE.md` into `.git/PR_DESCRIPTION.md` and fill in the details (source: [general-instructions.md](../sprints/general-instructions.md)).

---

## 🗺️ Refactoring & TDD Methodology

### Test-Driven Development (TDD)

Implement every command, query, and store method using the **Red-Green-Refactor** pattern:

1. **RED**: Write a failing test first.
2. **GREEN**: Write the minimal code to make the test pass.
3. **REFACTOR**: Clean up code and verify everything remains green.

### Strangler Fig Pattern

For migrating legacy endpoints, follow the Strangler Fig approach:

1. Write contract tests against the old API (verify current behavior).
2. Build the new slice alongside old code (no routing changes).
3. Verify contract tests pass on the new slice (identical behavior).
4. Swap routing in `bootstrap.go` (one-line change).
5. Monitor via tests + manual smoke (confidence window).
6. Delete old directories (`domain/`, `app/`, `infra/`).
