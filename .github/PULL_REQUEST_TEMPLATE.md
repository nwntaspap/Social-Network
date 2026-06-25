# 🚀 PR: [Ticket ID] — [Brief Title]

> **Ticket** `SX-YY-NN` · **Sprint** `N` · **Branch** `[owner]/<type>-<detail>`
> Resolves [Ticket Details](docs/sprints/sprint-[N].md#SX-YY-NN)

---

## 📝 Description

### 🔄 What Changed
- **Scope**: [Brief description of the work done]
- **Files**: [Key paths changed]

### 🎯 Why
- **Problem**: [What problem this solves]
- **Impact**: [How this affects the rest of the system]

---

## 🧪 Verification & Audit

### ✅ Backend Verification (`make be-ci`)
- [ ] **Backend**: `make be-ci` — runs ci-mod → format → check-format → lint (staticcheck + golangci-lint + govulncheck) → test (race + coverage)

### ✅ Frontend Verification (`make fe-ci`)
- [ ] **Frontend**: `make fe-ci` — runs `bun run lint` (ESLint) → `bun run format:check` (Prettier) → `tsc --noEmit` → `bun run test` (Vitest)

### 📋 Feature Traceability
| Slice | Layer Affected | Status | Evidence |
| :--- | :--- | :--- | :--- |
| `user/` | Entity / Command / Query / Transport / Store | | |
| `follow/` | Entity / Command / Query / Transport / Store | | |
| `topic/` | Entity / Command / Query / Transport / Store | | |
| `comment/` | Entity / Command / Query / Transport / Store | | |
| `group/` | Entity / Command / Query / Transport / Store / WS | | |
| `event/` | Entity / Command / Query / Transport / Store | | |
| `chat/` | Entity / Command / Query / Transport / Store / WS | | |
| `notification/` | Entity / Command / Query / Transport / Store | | |
| `oauth/` | Entity / Command / Query / Transport / Store | | |
| `core/` | Session / Realtime / Middleware / Server | | |
| `platform/` | Database / EventBus / Cache | | |

---

## ✅ PR Gate Checklist

### 📋 Required Checks
- [ ] **Standards**: Reviewed [AGENTS.md](AGENTS.md) and [conventions.md](.agents/rules/conventions.md).
- [ ] **Policy Compliance**: Ran `make ci` locally; all pass.
- [ ] **Sprint Scope**: Work matches declared sprint ticket(s); no scope creep.
- [ ] **Branching**: Branch name follows `<owner>/<type>-<detail>` convention.
- [ ] **Conventional Commit**: Commits use Conventional Commits format (e.g. `feat(user): add login handler`).
- [ ] **No Dead Code**: No unused imports, variables, or functions from your changes.
- [ ] **Integration Test Coverage**: New commands/queries have corresponding unit/integration tests.

### 🏗️ Architecture — D1–D5 Compliance
- [ ] **D1 Layout**: Feature follows `internal/<feature>/` with `<feature>.go`, `commands/`, `queries/`, `transport/`, `store/`.
- [ ] **D2 Interface Strategy**: Within-slice uses full `Repository`; cross-slice uses narrow consumer-defined interfaces.
- [ ] **D3 Cross-Slice**: ID-only refs (no embedded entities); sync via narrow interfaces; side-effects via Event Bus.
- [ ] **D4 DB Factory**: Stores accept `platform/database.DB`, not `*sql.DB`.
- [ ] **D5 Boundary Rules**: No outward imports from `<feature>.go`/`commands/`/`queries/` to `transport/` or `store/`.
- [ ] **D6 Acyclic Graph**: No circular imports; `notification` never imported by other features.
- [ ] **Safe Sinks**: Untrusted content uses `textContent` or explicit attribute APIs.
- [ ] **DB Migrations**: Named `00000X_name.up.sql` / `00000X_name.down.sql` (if applicable).
- [ ] **Dependencies**: Checked `go.mod` / `package.json` and lockfile impact.

### 🔒 Security
- [ ] **Params**: SQL uses `?` placeholders; ORDER BY whitelisted to `ASC`/`DESC`.
- [ ] **MIME**: Image upload validated via `http.DetectContentType` (magic bytes), not Content-Type header.
- [ ] **WebSocket**: `CheckOrigin` validated against configured origins (not unconditional `true`).
- [ ] **bcrypt**: Password hashing with cost factor ≥ 12.
- [ ] **Session**: Cookies use `HttpOnly` + `Secure` + `SameSite=Lax`.
- [ ] **Goroutines**: WS read/write pumps have `defer recover()`; tickers have `stop chan struct{}`.

---

## 🛡️ Security & Architecture Notes
- **Security**: [Notes on trust boundaries or potential sinks]
- **Architecture**: [Notes on system interactions, Strangler Fig migration status, or dependency changes]
- **Risks**: [Potential regressions or performance considerations]

---

## 🗄️ Database Migrations
- `00000X_name.up.sql`
- `00000X_name.down.sql`

---

## 🏁 Definition of Done
- [ ] D1–D5 boundary rules followed
- [ ] Concurrency & SQLite rules followed (`SetMaxOpenConns(1)`, WAL, busy timeout)
- [ ] Tests passing (`make ci`)
- [ ] Type checking clean (`make ci`)
- [ ] Lint clean (`make ci`)
- [ ] Branch name & commit convention correct
- [ ] Strangler Fig: contract tests pass for both old and new slices before swapping routes (if migrating)

---

<details>
<summary>📖 <b>Local Command Reference</b> (Click to expand)</summary>

| Command | Purpose |
| :--- | :--- |
| **`make ci`** | **Full CI gate (BE + FE)** |
| **`make be-ci`** | **Backend gate (ci-mod → format → check-format → lint → test)** |
| **`make fe-ci`** | **Frontend gate (lint → format:check → tsc → test)** |
| `make test` | BE: `go test -race -coverprofile=coverage.out ./...` |
| `make lint` | BE: staticcheck + golangci-lint + govulncheck |
| `go vet ./...` | BE: Go static analysis |
| `go build ./...` | BE: Go build check |
| `npx playwright test` | E2E: Playwright tests |

</details>
