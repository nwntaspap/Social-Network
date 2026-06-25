# High-Level Architecture & Code Organization

> **Target architecture.** The current codebase uses a layered structure under `internal/domain/`, `internal/app/`, `internal/infra/`. This document describes the vertical-slice target state after all refactoring phases. See [target-architecture-with-phases.md](target-architecture-with-phases.md) for the migration plan.

This document provides a concise, high-level overview of the architectural patterns, code organization, and structural boundaries governing the Social Network application.

---

## 1. Guiding Principles

1. **One Pattern, Everywhere**: Maintain strict consistency across features. If multiple approaches exist, we select one standard pattern and enforce it uniformly.
2. **Feature-Based Vertical Slices with CQRS**: Features are modular, self-contained packages. Each use case (one business operation) lives in its own file inside `commands/` (writes) or `queries/` (reads) subfolders. Store and transport remain shared per-feature.
3. **Decoupled Infrastructure**: External services (database, event bus, cache) are behind abstract interfaces in `internal/platform/`. Changing the concrete implementation (e.g., SQLite to PostgreSQL, or memory channel to RabbitMQ) requires zero modification to feature slices.

---

## 2. Directory Structure & Code Organization

The project is structured around self-contained vertical slices inside `internal/` and decoupled cross-cutting layers:

```
.
├── cmd/
│   ├── server/
│   │   └── main.go         # Application entry point & service bootstrap configuration
│   ├── gates/
│   │   └── main.go         # CLI runner for verification gates
│   └── client/
│       └── main.go         # CLI client for testing/ops
├── db/
│   └── migrations/         # Numbered database up/down migration SQL scripts (includes optional seed data via 000009_seed_data)
├── internal/
│   # ─── Feature Slices (Vertical Slices with CQRS) ───
│   ├── user/               # User profiles, auth, privacy toggle, activity tracking
│   ├── follow/             # Follow relationships & request workflow
│   ├── topic/              # Posts with visibility, categories, voting
│   ├── comment/            # Comment creation and querying
│   ├── group/              # Group definitions, memberships, invites, group chat, group posts
│   ├── event/              # Event management & RSVPs
│   ├── chat/               # Direct messages & chat presence
│   ├── notification/       # Notification delivery and event subscription
│   ├── oauth/              # OAuth state & third-party auth delegation
│   #
│   # ─── Cross-Cutting Core ───
│   ├── core/
│   │   ├── middleware/     # Auth checks, CORS, logging, and rate limiting
│   │   ├── realtime/       # WebSocket hub, connection lifecycle, and routing
│   │   ├── server/         # HTTP server configuration & graceful shutdown
│   │   └── session/        # Session tokens & state manager
│   #
│   # ─── Decoupled Platform Abstractions ───
│   ├── platform/
│   │   ├── cache/          # In-memory and Redis cache interfaces
│   │   ├── database/       # DB factory (SQLite/PostgreSQL interface)
│   │   └── eventbus/       # Async event publishing & subscription
│   #
│   #
│   # ─── Quality Gates ───
│   ├── gates/              # Deterministic verification gates (boundaries, DAG, security, coverage, TDD, etc.)
│   #
│   # ─── Bootstrap & Config ───
│   ├── bootstrap/          # Composition root (wiring slices and platform services)
│   ├── config/             # Config loaders
│   └── pkg/                # Reusable helper packages (bcrypt, uuid, validator, helpers, oauth, imgutil)
```

---

## 3. Vertical Slice Layout

Each feature slice within `internal/<feature>/` adheres to a strict internal structure. Each **use case** is its own file inside `commands/` or `queries/`:

```
internal/<feature>/
  ├── <feature>.go           # Domain entity structs, Repository interface, domain errors
  ├── commands/
  │     ├── <use_case>.go    # One file per write operation (validation + logic + event publishing)
  │     └── ...              # Additional command use cases
  ├── queries/
  │     ├── <use_case>.go    # One file per read operation (access checks + data projection)
  │     └── ...              # Additional query use cases
  ├── transport/
  │     ├── http.go          # REST HTTP handlers — delegates to commands/queries
  │     └── ws.go            # WebSocket handlers (chat & group chat only)
  └── store/
        └── sqlite.go        # Concrete SQLite repository implementation (all SQL in one file)
```

**Why this layout:**

- The **commands/ and queries/** layer is where complexity lives (privacy checks, event publishing, MIME validation). Splitting it per use case isolates each operation.
- The **store** layer is thin SQL (5–15 lines per method). One file per feature keeps all queries reviewable in one place.
- The **transport** layer is a thin HTTP adapter. One file per feature avoids handler fragmentation.

### Boundary & Dependency Rules

To keep vertical slices clean and decouple business logic from infrastructure details, we enforce strict boundary checks:

1. **No Outward Store/Transport Imports**: The feature root (`<feature>.go`) and `commands/`/`queries/` files **must not** import their own `transport/` or `store/` packages, nor may they import another feature's `transport/` or `store/` packages.
2. **Platform Interfaces Only**: Feature logic interacts with storage only through its own `Repository` interface and with database connections via the `platform/database.DB` interface.
3. **No Direct Feature-to-Feature Cross-Imports**: Slices communicate with each other through **narrow consumer-defined interfaces** (defined locally in the command/query file) or asynchronously via the **Event Bus**.
4. **Store Isolation**: `store/sqlite.go` must not import `commands/`, `queries/`, or `transport/`.

---

## 4. Cross-Slice Communication

To prevent circular dependencies and tight coupling, features interact via three defined patterns:

| Integration Type         | Strategy                    | Implementation Example                                                                                                                                                  |
| ------------------------ | --------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Data References**      | ID-only mapping             | A `Comment` struct contains an `AuthorID string` rather than embedding a `User` struct.                                                                                 |
| **Synchronous Queries**  | Consumer-defined interfaces | `internal/chat/commands/send_private_msg.go` defines a narrow local `FollowChecker` interface, which is satisfied by `internal/follow` during bootstrapping.            |
| **Asynchronous Effects** | Event Bus pub/sub           | `internal/follow/commands/follow_user.go` publishes a `follow.requested` event. `internal/notification/commands/consume_events.go` subscribes to it to dispatch alerts. |

---

## 5. Dependency Graph

Features import only what they need, keeping the graph acyclic:

```
user           → (nothing)
session        → user
follow         → user, eventbus
topic          → user
comment        → user, topic
group          → user, eventbus
event          → group, eventbus
chat           → user, FollowChecker (interface, not follow import)
notification   → user (subscribes to eventbus, no feature imports)
oauth          → user
```

Vote logic is absorbed into `topic/` and `comment/` — there is no standalone `vote` slice.

`notification` is never imported by other features. It subscribes to events at boot time, preventing circular dependencies.

---

## 6. Technology Stack & Runtime Infrastructure

### Backend (Go)

- **Database Engine**: Handled via `platform/database.DB`. Defaults to SQLite with Write-Ahead Logging (`WAL`) enabled and busy timeout (`_busy_timeout=5000`) configured to prevent locking. Portability for PostgreSQL is built-in. Seed data migration (`000009_seed_data`) available as a bonus feature.
- **WebSocket Protocol**: Built-in HTTP upgrade routing to `internal/core/realtime/` with token verification on handshake. Chat messages support Unicode/emoji via standard UTF-8 JSON encoding.
- **Asynchronous Processing**: In-process channel-based Event Bus for localized operations. Portability for RabbitMQ is built-in.

### Frontend (Next.js)

- **Architecture**: Next.js App Router providing server and client-side rendering.
- **Component Library**: **shadcn/ui** is used for core reusable elements (buttons, inputs, dialogs, cards, dropdowns, etc.), providing accessible and customizable components.
- **Styling**: **Tailwind CSS** coupled with Vanilla CSS overrides for the design system (glassmorphism, dark/light themes, customized HSL color palettes, and interactive transitions).
- **Communication**: REST APIs for basic CRUD operations, WebSocket channels for real-time chat (emoji support via UTF-8), and SSE for live notifications.
- **UI Conventions**: Destructive operations (unfollow, privacy toggle, decline requests) use `shadcn/ui` Dialog overlays for confirmation. Notifications are displayed in a dedicated panel (bell icon, unread count), visually distinct from the Chat panel.
- **Full spec**: See [SDS §6](sds.md#6-frontend-specifications-nextjs-tailwind-css--shadcnui) for detailed frontend specifications.

### Docker

- **Two containers**: Backend (Go, port 8080) and Frontend (Next.js, port 3000), orchestrated via `docker-compose.yml`.
- **Build script**: Optional `scripts/docker-build.sh` convenience script for automated image building and container startup.

---

## 7. Development Tooling & SDLC

Quick-reference for all tools used across the software development lifecycle.

### Backend (Go)

| Phase                             | Tool                          | Where                                |
| --------------------------------- | ----------------------------- | ------------------------------------ |
| Build                             | `go build`                    | `Dockerfile` (multi-stage)           |
| Testing                           | `go test -race -coverprofile` | `Makefile` `test`                    |
| Linting (aggregator, 30+ linters) | `golangci-lint` v2.2.1        | `.golangci.yml`                      |
| Linting (static analysis)         | `staticcheck`                 | `Makefile` `lint`                    |
| Linting (official)                | `go vet`                      | `.golangci.yml`, CLI                 |
| Formatting                        | `gofmt -s`, `gofumpt`         | `Makefile` `format`, `.golangci.yml` |
| Imports                           | `goimports`, `gci`            | `Makefile` `format`, `.golangci.yml` |
| Modules                           | `go mod tidy`                 | `Makefile` `ci-mod`                  |
| Benchmarking                      | `benchstat`                   | `Makefile`                           |
| Profiling                         | `go tool pprof`               | `Makefile`                           |
| Vuln scanning                     | `govulncheck`                 | Manual / CI                          |

### Frontend (Bun)

| Phase                | Tool                 | Where                               |
| -------------------- | -------------------- | ----------------------------------- |
| Runtime              | Bun                  | `package.json` (scripts)            |
| Package manager      | Bun                  | `bun.lock`                          |
| Linting + formatting | ESLint + Prettier    | `eslint.config.mjs` + `.prettierrc` |
| Type checking        | `tsc --noEmit`       | `package.json`                      |
| Unit/component tests | Vitest (planned)     | `vitest.config.ts`                  |
| E2E tests            | Playwright (planned) | `playwright.config.ts`              |

### Infrastructure & CI

| Phase              | Tool                  | Where                  |
| ------------------ | --------------------- | ---------------------- |
| Containers         | Docker (multi-stage)  | `Dockerfile`           |
| Orchestration      | Docker Compose v5.1.1 | `docker-compose.yml`   |
| Dev TLS            | `openssl`             | `scripts/makecerts.sh` |
| CI pipeline        | Makefile `ci` target  | `Makefile`             |
| Pre-commit hooks   | Lefthook              | `lefthook.yml`         |
| Verification gates | `cmd/gates/main.go`   | `make gates`           |
| Go arch lint       | `go-arch-lint`        | `.go-arch-lint.yml`    |
| Security scan      | `gosec`               | `Makefile` / gates     |

### CI Pipeline (`make ci` / `make be-ci-new`)

During local development and PR verification, the CI pipeline is split into legacy blanket targets and scoped new-code targets:

**New-Code Scoped CI** (`make be-ci-new`):

```
ci-mod → check-format-new → lint-new (staticcheck-new + golangci-lint-new + vet-new + vulncheck-new + gosec-new) → test-new
```

> This pipeline is scoped exclusively to `NEW_DIRS` (the vertical slices under `internal/`, bootstrap, core, config, platform, gates, etc.), avoiding failures on legacy code.

**Legacy Blanket CI** (`make be-ci`):

```
ci-mod → check-format → lint (staticcheck + golangci-lint + govulncheck) → test
```

> Runs checks across the entire codebase (informational, not blocking PRs).

**Frontend** (`make fe-ci`):

```
bun run lint → bun run format:check → tsc --noEmit → bun run test
```

> Automatically targets the new `frontend-next/` directory if it exists, falling back to legacy `frontend/`, or skips if not scaffolded.

### Verification Gates (`make gates`)

Go-based deterministic gates under `internal/gates/` enforce architectural and convention rules. `make gates` is decoupled from the legacy CI and executes:

1. `go build ./...` — Compiles all code (legacy + new) for build safety.
2. `go run cmd/gates/main.go --all` — Runs the Go verification gates.
3. `make be-ci-new` — Performs the scoped CI pipeline on the new codebase.
4. `make fe-ci` — Performs the scoped CI pipeline on the frontend.

| Gate       | Tool/Fallback                | What It Checks                                                                     |
| ---------- | ---------------------------- | ---------------------------------------------------------------------------------- |
| Stack      | go version / go.mod          | Go ≥ 1.25, module path `social-network`                                            |
| Layout     | os.Stat                      | Target directory structure exists                                                  |
| Boundaries | golangci-lint depguard / AST | D5 — forbidden cross-slice imports                                                 |
| DAG        | go-arch-lint / DFS           | D6 — dependency graph acyclicity                                                   |
| TDD        | os.Stat                      | Test files exist per command/query                                                 |
| Migrations | glob / grep                  | Migration naming (`NNNNNN_name.up.sql`), delimiter (`";"`)                         |
| Security   | gosec / custom AST           | SQL concat, WebSocket CheckOrigin, bcrypt cost                                     |
| Branch     | regex                        | Branch naming convention `<user>/<ticket>-<detail>` (includes `dev` scope support) |
| Coverage   | git worktree + go test       | Test coverage threshold (>90%)                                                     |
| ScopeDrift | git diff                     | Unplanned file changes                                                             |

Output: JSON with exit codes. Run via `make gates` or `go run cmd/gates/main.go --all`.

### Pre-commit / Pre-push Hooks (Lefthook)

Configured in `lefthook.yml`:

- **Pre-commit** (staged files, parallel): `gofumpt` + `goimports` on `*.go`, ESLint + Prettier format+lint on frontend. Auto-fixes via `stage_fixed: true`.
- **Pre-push** (parallel): `go vet`, `go test -short`, `go build`, `go-arch-lint check` (backend); `tsc --noEmit`, `eslint`, `vitest` (frontend).

Install: `make setup-hooks`. Bypass: `--no-verify`.

---

## 8. Requirements Verification

The project's requirements checklist is maintained in [`docs/requirements/audit.md`](../requirements/audit.md). All architectural decisions and feature implementations must satisfy the conditions defined there. Sprint 6 automation tickets (S6-SD-25, S6-SD-34) implement test suites mapped directly to audit.md questions. See also the [progressive disclosure chain](../sprints/general-instructions.md#linear-progressive-disclosure-navigation-chain) (Stage 7).
