# 🌐 Social Network — Vertical Slices with CQRS

A premium, high-performance social networking platform built with a **Go 1.25 Backend API** organized around Feature-Based Vertical Slices, and a modern **Next.js App Router Frontend** powered by Bun, Tailwind CSS, and shadcn/ui.

The project features decoupled infrastructure abstractions (SQLite, PostgreSQL, Redis, and RabbitMQ), an asynchronous in-process channel-based Event Bus, strict boundary rules, secure WebSocket/SSE real-time features, and a responsive glassmorphic design.

---

## 🏗️ System Architecture

The application separates client presentation, server business rules, and pluggable infrastructure services:

```mermaid
graph TD
    Client[Browser: Next.js + shadcn/ui] <-->|REST API / WebSockets| Backend[Go Backend: Port 8080]
    Backend <-->|platform/database| SQLite[(SQLite: WAL Mode)]
    Backend -.->|platform/eventbus| RabbitMQ[(RabbitMQ: Async Events)]
    Backend -.->|platform/cache| Redis[(Redis: Session Cache)]
```

### 1. Presentation Layer (Frontend)

- **Next.js App Router (Port 3000)**: Server-side and client-side rendering with Tailwind CSS and shadcn/ui.
- **Real-time Communication**: Persistent WebSockets for chats and Server-Sent Events (SSE) for live notifications.
- **Client Verification**: Native Unicode emoji parsing, magic-byte image validation, and client-side file size and extension checks before transport.

### 2. Business Logic & Feature Layer (Backend)

- **Vertical Slices with CQRS (Port 8080)**: Organized under `internal/<feature>/`. Each slice wraps domain entities, CQRS commands/queries, HTTP/WebSocket transports, and SQLite storage implementations.
- **Cross-Cutting Core**: Reusable utilities, session tracking, WebSocket hubs, and route routers live under `internal/core/`.

### 3. Decoupled Platform Services

- **Infrastructure Adaptability**: Databases, event brokers, and caches sit behind abstract interfaces in `internal/platform/`. Changing the concrete implementation (e.g., swapping SQLite for PostgreSQL or in-memory events for RabbitMQ) requires zero modification to feature slices.

---

## 🌟 Core Features (Finished Product)

### 🔐 Authentication & Session Persistence

- **Rich Registration**: Custom flow requiring Email, Password, First Name, Last Name, and Date of Birth. Optional Avatar, Nickname, and About Me info.
- **Secure Sessions**: Persistent double-cookie auth (`access_token` and `refresh_token` rotation) behind `HttpOnly` and secure flags.
- **OAuth Delegation**: Built-in GitHub and Google authentication.

### 👥 Follows & Profile Privacy

- **Privacy Toggle**: Seamlessly switch profiles between **Public** and **Private** via a confirmation overlay.
- **Auto-follow / Request Flow**: Public profiles accept follows instantly. Private profiles create a follow request, triggering a notification for the user to Accept or Decline.

### 📝 Posts, Comments & Visibility Scopes

- **Image Support**: Attach JPEG, PNG, and GIF files (magic-byte validated).
- **Privacy Scopes**:
  - `public`: Visible to everyone.
  - `almost_private`: Restrained to current followers.
  - `private`: Restricted to selected followers chosen via a user picker.

### 💬 Real-Time Unified Chat

- **Real-time Handshake**: Secure WebSocket connections with session token checks on handshakes.
- **Follower Validation**: Direct messages are follow-gated (at least one user must follow the other).
- **Rich Client UX**: Unicode emoji support, typing indicators, active presence tracking, and message-read status.
- **Group Chats**: Automatic WebSocket rooms created for group members.

### 🏛️ Groups & Events Lifecycle

- **Group Spaces**: Create groups with title/description, invite followers, request to join, and publish group-exclusive posts.
- **Event Planning**: Schedule events with a title, description, date/time, and RSVP options (Going vs. Not Going). RSVPs sync in real-time.

### 🔔 Live Notifications Stream

- Dedicated panel (visually distinct from chat messages) showing instant alerts for:
  - Follow requests received & accepted.
  - Group invitations & join requests.
  - Group event creation.

---

## 🏛️ Core Design Decisions (D1–D6)

Every developer and sub-agent must adhere to these architectural guidelines:

| Decision | Area                   | Summary & Rule                                                                                                                                                                                                          |
| :------- | :--------------------- | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **D1**   | **Vertical Slices**    | All business logic for a feature lives in `internal/<feature>/`. Commands (writes) and queries (reads) reside in separate files inside `commands/` and `queries/`. Stores and transports are thin, unified per-feature. |
| **D2**   | **Interface Strategy** | Within a slice, commands/queries accept the full `Repository` interface from `<feature>.go`. Across slices, consumer features define narrow local interfaces satisfied implicitly via Go duck typing.                   |
| **D3**   | **Communication**      | Slice communication is ID-only for data references, narrow local interfaces for synchronous checks, and the Platform Event Bus for mutation side effects.                                                               |
| **D4**   | **Database Access**    | Feature stores accept `platform/database.DB` interfaces rather than raw `*sql.DB`. The connection factory switches between SQLite (WAL mode + busy timeout) and PostgreSQL dynamically.                                 |
| **D5**   | **Boundary Rules**     | Feature logic and command/query packages **must not** import their own `transport/` or `store/` folders, nor can they import transport/store directories of other features.                                             |
| **D6**   | **Dependency Graph**   | The import tree must remain strictly acyclic (e.g. `user` and `session` have no dependencies on higher-level features, `notification` is a pure subscriber with zero external feature imports).                         |

---

## 📂 Project Structure

```
.
├── cmd/
│   ├── server/
│   │   └── main.go                  # Application entry point & configuration bootstrap
│   ├── gates/
│   │   └── main.go                  # CLI runner for verification gates
│   └── client/
│       └── main.go                  # CLI client for testing/ops
├── db/
│   └── migrations/                  # Sequential up/down SQL database migrations
├── internal/
│   # ─── Vertical Feature Slices ───
│   ├── user/                        # Profiles, registration, privacy toggles
│   ├── follow/                      # Follow relationships & pending requests
│   ├── topic/                       # Posts, feed visibility, voting rules
│   ├── comment/                     # Image-supported comments
│   ├── group/                       # Communities, group posts, group chat
│   ├── event/                       # Group events & RSVP tracking
│   ├── chat/                        # Direct messages, presence, and chat history
│   ├── notification/                # Event bus subscriber notifications
│   ├── oauth/                       # Github/Google authentication pipelines
│   #
│   # ─── Cross-Cutting Core ───
│   ├── core/
│   │   ├── middleware/              # Auth, CORS, logging, rate limiter
│   │   ├── realtime/                # WebSocket hub and connection routing
│   │   ├── server/                  # HTTP engine & graceful shutdown
│   │   └── session/                 # Session lifecycle and token stores
│   #
│   # ─── Platform Abstractions ───
│   ├── platform/
│   │   ├── cache/                   # Memory cache & Redis wrappers
│   │   ├── database/                # DB factory and migrations runner
│   │   └── eventbus/                # Event publishing channels
│   #
│   # ─── Quality Gates ───
│   ├── gates/                       # Deterministic verification gates (boundaries, DAG, security, coverage, TDD, etc.)
│   #
│   # ─── Bootstrap & Helpers ───
│   ├── bootstrap/                   # Composition root & service wiring
│   ├── config/                      # Configuration loader
│   └── pkg/                         # Shared helpers (bcrypt, uuid, validator, helpers, oauth, imgutil) — migrates to repo-root pkg/ in S5-BE-99
│
└── frontend/                        # Next.js Application Root (TypeScript + Bun)
```

---

## 🛠️ Technology & Tooling

### Languages & Runtimes

- **Backend**: Go 1.25
- **Frontend**: TypeScript (Next.js App Router, Bun Runtime)
- **Database**: SQLite3 (Production: WAL mode enabled; Dev/Testing: In-Memory / Local DB)
- **Containers**: Docker & Docker Compose v5.1.1

### Tooling Breakdown

| Layer        | Phase           | Tool                          | Target/File                         |
| :----------- | :-------------- | :---------------------------- | :---------------------------------- |
| **Backend**  | Testing         | `go test -race -coverprofile` | `Makefile` (`make test`)            |
| **Backend**  | Linting         | `golangci-lint` (v2.2.1)      | `.golangci.yml`                     |
| **Backend**  | Static Analysis | `staticcheck`                 | `Makefile` (`make lint`)            |
| **Backend**  | Formatting      | `gofmt -s`, `gofumpt`         | `Makefile` (`make format`)          |
| **Backend**  | Vuln Check      | `govulncheck`                 | Local execution                     |
| **Frontend** | Package/Run     | `Bun`                         | `package.json`                      |
| **Frontend** | Formatting/Lint | `ESLint` + `Prettier`         | `eslint.config.mjs` + `.prettierrc` |
| **Frontend** | Testing         | `Vitest` (planned)            | `vitest.config.ts`                  |
| **Frontend** | E2E Testing     | `Playwright` (planned)        | `playwright.config.ts`              |

---

## 🚀 Getting Started

### 📋 Prerequisites & Installation Guide

To configure your local environment for development and testing, install the following runtimes and tools:

1. **Go 1.25+**: Install the standard library runtime from [go.dev/dl](https://go.dev/dl/).
2. **Bun**: The fast JavaScript package manager and runtime. Install via:
   ```bash
   curl -fsSL https://bun.sh/install | bash
   ```
3. **Docker & Docker Compose**: Essential for orchestrating backend/frontend services in a containerized environment.
4. **Install All Project Dependencies**: Run the single unified install command:

   ```bash
   make install
   ```

   This installs everything: Go modules, root JS tooling, `.env` config, SSL certs, Go development tools (`gofumpt`, `goimports`, `staticcheck`, `golangci-lint`, `govulncheck`, `gosec`, `go-arch-lint`, lefthook), git hooks, and frontend dependencies.

   > If you only need Go tools (without frontend/certs/env), use `make setup`.

---

### 🐳 Running with Docker (Recommended)

1. **Configure Environment Variables**  
   Create a `.env` file at the root:

   ```env
   SERVER_PORT=8080
   CLIENT_PORT=3000
   DATABASE_DRIVER=sqlite
    DATABASE_DSN=db/data/forum.db?_journal_mode=WAL&_busy_timeout=5000
   DB_SEED_ON_START=true
   GITHUB_CLIENT_ID=your_github_client_id
   GITHUB_CLIENT_SECRET=your_github_client_secret
   GOOGLE_CLIENT_ID=your_google_client_id
   GOOGLE_CLIENT_SECRET=your_google_client_secret
   ```

2. **Boot the Platform**  
   Build and start both the backend API and Next.js frontend services:

   ```bash
   make docker-up
   ```

   _To run the development setup with code mounting and hot reload:_

   ```bash
   make docker-dev
   # Or use the alias:
   make dev
   ```

3. **Access points**
   - **Frontend web App**: [http://localhost:3000](http://localhost:3000)
   - **Backend REST API**: [http://localhost:8080/api/v1](http://localhost:8080/api/v1)

4. **Shutdown and Clean**  
   Stop the services:
   ```bash
   make docker-down
   ```
   Purge volumes, cache, and DB files:
   ```bash
   make docker-clean
   ```

---

### 💻 Running Locally

#### 1. Setup Backend

Run the database migrations and boot the Go API Server:

```bash
# Run server (database is initialized automatically via migrations runner)
go run cmd/server/main.go
```

The local SQLite file is generated at `db/data/forum.db` (or as configured per `DB_PATH` in `.env`).

#### 2. Setup Frontend

Install dependencies and run the Next.js development server:

```bash
cd frontend
bun install
bun run dev
```

---

## 🧪 Testing & Code Quality Gates

We enforce strict validation pipelines to ensure codebase stability.

### ⚙️ CI Pipeline

Run the automated check suite locally before pushing:

```bash
make gates
```

This runs the decoupled PR quality gate suite:

1.  **`go build ./...`**: Compiles all Go packages (legacy + new) for basic build safety.
2.  **`go run cmd/gates/main.go --all`**: Runs the custom Go verification gates.

_Note: You can still run the legacy full-system check via `make ci` (runs blanket `make be-ci` + `make fe-ci`), which is informational and does not block PR gates._

To auto-format files in new directories:
`PATH=<GOBIN>:$PATH gofumpt -w <dirs> && PATH=<GOBIN>:$PATH goimports -w -local social-network <dirs>`

### 🧪 Go Verification Gates

Deterministic Go-based gates catch architecture violations, security issues, and branch/convention regressions:

```bash
go run cmd/gates/main.go --all
```

Individual gates:

```bash
go run cmd/gates/main.go --gate=stack           # Go version + module path
go run cmd/gates/main.go --gate=d1-layout       # Directory structure (D1 layout)
go run cmd/gates/main.go --gate=d5-boundaries   # D5 import boundary rules
go run cmd/gates/main.go --gate=d6-dag          # D6 dependency graph acyclicity (supports go-arch-lint)
go run cmd/gates/main.go --gate=tdd             # Test file presence
go run cmd/gates/main.go --gate=migrations      # Migration naming/delimiter
go run cmd/gates/main.go --gate=security        # gosec + govulncheck + custom AST checks (scoped to new code)
go run cmd/gates/main.go --gate=branch          # branch naming convention (includes 'dev' scope)
go run cmd/gates/main.go --gate=coverage-delta  # test coverage threshold
go run cmd/gates/main.go --gate=scope-drift     # scope creep detection
go run cmd/gates/main.go --gate=format          # code formatting (gofumpt / gofmt)
go run cmd/gates/main.go --gate=lint            # code linting (golangci-lint / staticcheck / go vet)
go run cmd/gates/main.go --gate=go-test         # Go unit tests
go run cmd/gates/main.go --gate=frontend        # frontend CI checks (lint, format, tsc, test)
```

Default output is human-readable text; use `--json` for JSON format. Exit code is 0 on success, 1 on failure. Gates are also run via `make gates`.

### 🔗 Pre-commit & Pre-push Hooks (Lefthook)

Quality hooks run automatically on staged files (pre-commit) and before push (pre-push):

```bash
# Install hooks (one-time setup)
make setup-hooks

# Pre-commit (runs on staged files):
#   - Backend: gofumpt + goimports formatting (auto-fixed)
#   - Frontend: Prettier + ESLint

# Pre-push:
#   - Backend: go vet ./..., go test -short ./..., go build ./..., go-arch-lint check
#   - Frontend: tsc --noEmit, eslint, vitest
```

To bypass hooks temporarily: `git commit --no-verify` or `git push --no-verify`.

### ⚛️ Frontend Validation

Run linting, formatting check, and TypeScript compilation gates:

```bash
cd frontend
bun run lint            # ESLint linting
bun run format:check    # Prettier formatting check
bun run typecheck       # tsc --noEmit check
bun run test            # Vitest unit & component tests
```

---

## 📚 Documentation Reading Order

This project uses progressive disclosure to avoid cognitive overload. Read in this order:

1. **Rules & Guidelines**: [AGENTS.md](AGENTS.md) + [conventions.md](.agents/rules/conventions.md)
2. **Methodology**: [general-instructions.md](docs/sprints/general-instructions.md)
3. **Architecture**: [architecture.md](docs/architecture/architecture.md)
4. **Design Specs**: [sds.md](docs/architecture/sds.md)
5. **Roadmap**: [target-architecture-with-phases.md](docs/architecture/target-architecture-with-phases.md)
6. **Sprints**: [sprint-0.md](docs/sprints/sprint-0.md) through [sprint-6.md](docs/sprints/sprint-6.md)

See [conventions.md](.agents/rules/conventions.md#14-progressive-disclosure--doc-reading-order) for the full dependency graph.

---

## 🤖 Agentic Workflows

This repository includes specialized configurations and tools designed to optimize the workflow of agentic AI coding assistants (like `Antigravity` and `Opencode`). Developers using agentic assistants can leverage these files for enhanced context management, token savings, and codebase navigation:

- **⚡ Terminal Token Compression**:
  - [RTK (Rust Token Killer)](.agents/rules/antigravity-rtk-rules.md): Minimizes input/output token consumption by 60–90%. Always prefix terminal commands with `rtk` (e.g. `rtk git status`, `rtk make test`, `rtk go test ./...`).
- **💬 Communication Optimization**:
  - [Caveman Mode (.agents/skills/caveman/SKILL.md)](.agents/skills/caveman/SKILL.md): Ultra-compressed communication skill that cuts conversation token usage by ~75% while maintaining technical accuracy.
- **📊 Codebase Knowledge Graph**:
  - [Graphify Rule (.agents/rules/graphify.md)](.agents/rules/graphify.md) / [Graphify Skill (.agents/skills/graphify/SKILL.md)](.agents/skills/graphify/SKILL.md): Builds and queries an AST-based knowledge graph. Run `graphify query "<question>"` to BFS traverse or `graphify path "<A>" "<B>"` to trace dependencies.
- **🤖 Core Instructions & Conventions**:
  - [AGENTS.md](AGENTS.md): Main code guidelines, including simplicity-first principles and progressive document reading orders.
  - [Conventions (.agents/rules/conventions.md)](.agents/rules/conventions.md): Code boundaries and architectural rules.
  - [Karpathy Guidelines (.agents/rules/karpathy-guidelines.md)](.agents/rules/karpathy-guidelines.md): Red-Green-Refactor, TDD, and codebase change minimization guides.

---

### ⚙️ Critical Skills Installation & Dependency Management

**Agent Skills** for coding assistants (caveman, graphify, diagnose, review, tdd, etc.) are pre-installed under `.agents/skills/` and registered in `skills-lock.json`. Skills auto-activate when the agent detects a matching task. No manual install needed for existing skills.

To install new skills from a remote source (e.g. `mattpocock/skills`):

```bash
# Opencode
opencode skill install <source>/<skill-name>

# Or sync all locked skills from skills-lock.json
opencode skill sync
```

To update all skills to their latest version:

```bash
opencode skill upgrade
```

**Development Dependencies** — one command to install everything:

| Layer                      | Command                                | What it installs                                                                                                  |
| :------------------------- | :------------------------------------- | :---------------------------------------------------------------------------------------------------------------- |
| **All (unified)**          | `make install`                         | Go modules + root JS + `.env` + SSL certs + Go tools + git hooks + frontend deps                                  |
| **Backend (Go)**           | `make setup` (or `make tools`)         | `goimports`, `staticcheck`, `golangci-lint` (v2.2.1), `govulncheck`, `gofumpt`, `gosec`, `go-arch-lint`, lefthook |
| **Backend (Go modules)**   | `go mod download` (via `make install`) | Go library dependencies from `go.sum`                                                                             |
| **Frontend (Bun/Next.js)** | `bun install` (via `make install`)     | npm packages from `frontend/package.json`                                                                         |
| **Docker**                 | `docker compose build`                 | Container images for backend + frontend                                                                           |

To update Go dependencies:

```bash
go get -u ./...        # update all deps
go mod tidy            # clean up
```

To update frontend dependencies:

```bash
cd frontend
bun update             # update all packages per semver ranges
```

---

## 🤝 Contribution & Onboarding Workflow

### 🚀 Developer Branch Strategy

All branches must match the naming schema: `<username>/<ticket/issue-ID>-<detail>`

- _Username_: Your own Gitea username — resolve via `tea whoami` or `cat ~/.config/tea/config.yml | grep 'user:' | head -1 | awk '{print $2}'`. Known devs: `epapamic`, `ekaramet`, `dkotsi`, `geoikonomou`, `smichail`
- _ticket/issue-ID_: Ticket ID from `docs/sprints/ticket-tracker.md` (e.g. `S3-BE-01`) or GitHub/Gitea issue number (e.g. `42`). **Required** — maps branch to work item.
- _Detail_: kebab-case description (e.g. `db-factory`, `fix-sqlite-busy-timeout`).
- _Examples_: `ekaramet/S1-BE-05-db-factory`, `geoikonomou/42-fix-sqlite-busy-timeout`
- Branches must live **$\le$ 3 days** (Trunk-Based Development).

### 📝 Commit Message Convention

Squash-merged commits onto the main branch must follow the **Conventional Commits** standard:

```
<type>(<scope>)[<ID>]: <description>

[Optional Body explaining the why behind changes]
```

- _Examples_:
  - `feat(follow)[S3-BE-37]: add follow requests workflow`
  - `fix(core)[42]: recover from websocket write panic`
  - `refactor(topic)[S2-BE-26]: migrate topic store to vertical slice`

### 🏁 Definition of Done (DoD)

A task is marked completed when:

1. **TDD cycle** is fully executed (write failing test -> make it pass -> refactor).
2. **Boundary checks** (D5) verify no cross-slice http/store imports.
3. Code compiles and linting/formatting passes cleanly (`make ci`).
4. PR follows the standard PR description template (found in `.github/PULL_REQUEST_TEMPLATE.md`).
5. Successfully verified through targeted manual smoke tests (e.g. age locks, privacy gates, follow actions).

### 🏗️ Strangler Fig Migration Pattern

Legacy layered code is migrated to vertical slices using the Strangler Fig pattern. See [conventions.md](.agents/rules/conventions.md#3-strangler-fig-migration) for full steps:

1. Write contract tests against old API
2. Build new slice alongside old code
3. Verify contract tests match
4. Swap routing in bootstrap
5. Delete old code after confidence window
