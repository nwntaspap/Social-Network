# Sprint 1: Platform & Core Infrastructure (Week 3–4)

**Outcome:** All platform abstractions (Database factory, Event bus, Cache, migration engine) and cross-cutting core layers (Sessions, Real-time WebSockets, auth middleware, and HTTP servers) are fully built and verified using TDD. The frontend has complete auth screens and api/mock wrappers.

---

## BE-A (Backend A) Tickets

### S1-BE-05: Platform: DB Factory

- **Priority:** P0 (Prerequisite for migrations and features)
- **Type:** Greenfield (New Module/Feature - Platform Abstraction)
- **Assignee:** BE-A
- **Story Points:** 5
- **Dependencies:** Sprint 0
- **Description:** Create the pluggable Database connection provider interface and SQLite initializer with specific pooling rules. This is an entirely NEW module with NO existing implementation in the legacy codebase -- the old `internal/infra/storage/sqlite/` layer uses direct `sql.Open` calls without a factory abstraction or interface. This platform package must be designed and built from scratch. As a greenfield task, this builds new platform or feature abstractions from scratch that do not exist in the old legacy codebase.
- **Detailed Steps:**
  - _Greenfield Note:_ Follow TDD (Red-Green-Refactor). Ensure the slice adheres strictly to boundary rules (D5) without importing other slices' stores/transports.
  1. Create `internal/platform/database/database.go` and define the `DB` interface (containing standard methods `QueryContext`, `QueryRowContext`, `ExecContext`). There is no old code to reference for this abstraction; the implementation is new.
  2. Implement `newSQLite(dsn string) (DB, error)` in `internal/platform/database/sqlite.go`.
  3. Ensure WAL mode is active: execute `PRAGMA journal_mode=WAL;`.
  4. Ensure busy timeout is set: execute `PRAGMA busy_timeout=5000;`.
  5. **SQLite Pooling Limit (Gap Fix):** Configure SQLite connection limit explicitly to prevent concurrency locking: `db.SetMaxOpenConns(1)` (since SQLite does not support concurrent write operations across multiple connections).
  6. Expose `NewDB(cfg Config) (DB, error)` as a factory.
- **Verification:** Write unit tests in `internal/platform/database/sqlite_test.go` verifying WAL is active, timeout is set, and maximum open connections is restricted to 1.

---

### S1-BE-06: Custom Migration System

- **Priority:** P0
- **Type:** Greenfield (New Module/Feature - Platform Abstraction)
- **Assignee:** BE-A
- **Story Points:** 8
- **Dependencies:** S1-BE-05
- **Description:** Build the backend SQL migrations runner that applies `.up.sql` and `.down.sql` scripts dynamically. This is an entirely NEW module with NO existing implementation in the legacy codebase -- the old `internal/infra/storage/sqlite/init.go` does direct schema creation but has no migration runner, version tracking, or up/down logic. This platform package must be designed and built from scratch. Convert the existing schema (`db/migrations/schema.sql` + `indexes.sql`) into the initial numbered migration file (`000001_initial_schema.up.sql`). Migration files for subsequent features will be created in their respective feature sprints (Sprint 2 for 000002-000003, Sprint 3 for 000004, Sprint 4 for 000005-000006). Seed migration (000009) is handled by S1-SD-05. As a greenfield task, this builds new platform or feature abstractions from scratch that do not exist in the old legacy codebase.
- **Detailed Steps:**
  - _Greenfield Note:_ Follow TDD (Red-Green-Refactor). Ensure the slice adheres strictly to boundary rules (D5) without importing other slices' stores/transports.
  1.  Implement a migration runner in `internal/platform/database/migrations.go`.
  2.  Create a metadata database table named `schema_migrations` to track applied version IDs.
  3.  Implement parser that reads SQL scripts, breaks statements on `";"`, and executes them sequentially.
  4.  Write migration files under `db/migrations/`:
      - `000001_initial_schema.up.sql` (Convert existing tables from `schema.sql` + `indexes.sql`: users, posts, categories, comments, votes, sessions, chats, notifications, oauth_states)
      - `000001_initial_schema.down.sql` (Drop all tables)
- **Verification:** Write integration tests verifying that executing the migrations runner creates the correct database tables, running up twice does not error, and executing down rolls back the database cleanly. Run migrations up to 000001 and verify all tables exist.

---

### S1-BE-07: Core: Session Management

- **Priority:** P1
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-A
- **Story Points:** 3
- **Dependencies:** S1-BE-05
- **Description:** Setup backend cookie-based session management. This migrates EXISTING session logic from the old layered codebase into the new vertical-slice layout. The current implementation lives in `internal/domain/session/session.go` (domain model and `SessionManager` interface) and `internal/infra/http/authcookies/manager.go` (cookie handling). The new layout consolidates session structs, the manager interface, and SQLite-backed storage under `internal/core/session/` with a dedicated store sub-package, replacing the old domain/infra split. This migration moves session logic from `internal/infra/http/authcookies/` to `internal/core/session/`, separating it from the HTTP transport layer.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1. Create `internal/core/session/session.go`. Define `Session` structs (Token, UserID, ExpiresAt) and a `SessionManager` interface. Restructure from the old `internal/domain/session/session.go` and `internal/domain/session/sessionManager.go` which contain the current domain model and manager. The existing code has the logic; this ticket moves and restructures it into the new layout.
  2. Implement SQLite-backed session storage in `internal/core/session/store/sqlite.go` matching the migration schema.
  3. Provide creation (`Create`), lookup (`Get`), and invalidation (`Revoke`) methods.
- **Verification:** Write tests using TDD to assert sessions write to SQLite, lookups find correct IDs, and expired session lookups return errors.

---

### S1-BE-08: Core: Middlewares

- **Priority:** P1
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-A
- **Story Points:** 3
- **Dependencies:** S1-BE-07, S1-BE-11
- **Description:** Build the generic middleware pipeline (Auth check, CORS origin validator, RateLimiter). This migration extracts core logic from the legacy layered codebase into its own isolated vertical slice following the Strangler Fig pattern.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1.  **Auth Middleware:** Verifies request session cookie, queries session store, and attaches UserID into the Go request Context.
  2.  **CORS Middleware:** Performs strict header checks against configured origin domains.
  3.  **Rate Limiting Middleware:** Uses a sliding-window token bucket algorithm (utilizing S1-BE-11 Cache) to limit requests per IP. Ensure ticker cleanup doesn't leak threads.
  4.  **(Phase 3.3)** Implement `internal/core/middleware/logging.go` — request logging middleware that records method, path, status, duration, and request ID.
- **Note:** Rate limiter depends on Cache (S1-BE-11, P1). Since both are P1 in same sprint, implement Cache first or mark rate limiter as soft-blocked on Cache completion.
- **Verification:** Write tests hitting mock HTTP endpoints through the middleware chain, confirming correct status codes (401 Unauthorized for bad sessions, 429 Too Many Requests for rate limits, etc.).

---

### S1-BE-09: Shared: Image Type Verification Utility

- **Priority:** P2
- **Type:** Greenfield (New Module/Feature - Utility)
- **Assignee:** BE-A
- **Story Points:** 1
- **Description:** Add helper to verify image mime types using Magic Bytes headers. As a greenfield task, this builds new platform or feature abstractions from scratch that do not exist in the old legacy codebase.
- **Detailed Steps:**
  - _Greenfield Note:_ Follow TDD (Red-Green-Refactor). Ensure the slice adheres strictly to boundary rules (D5) without importing other slices' stores/transports.
  1. Create `pkg/imgutil/detect.go`.
  2. Write a helper function that inspects the first 512 bytes of a file reader using `http.DetectContentType` to enforce only `image/jpeg`, `image/png`, and `image/gif` are accepted.
- **Verification:** Unit tests asserting that valid image files (JPG, PNG, GIF) return true, and PDFs or executables are rejected.

---

## BE-B (Backend B) Tickets

### S1-BE-10: Platform: Event Bus

- **Priority:** P0
- **Type:** Greenfield (New Module/Feature - Platform Abstraction)
- **Assignee:** BE-B
- **Story Points:** 3
- **Description:** Create the Event Bus interface and an in-process, channel-based implementation for async cross-slice notifications. As a greenfield platform package, this implements the central messaging bus under `internal/platform/eventbus/` to support async decoupled communication between vertical slices.
- **Detailed Steps:**
  - _Greenfield Note:_ Follow TDD (Red-Green-Refactor). Ensure the slice adheres strictly to boundary rules (D5) without importing other slices' stores/transports.
  1. Create `internal/platform/eventbus/eventbus.go`. Define the `EventBus` interface:
     - `Publish(ctx context.Context, eventType string, payload any) error`
     - `Subscribe(eventType string, handler HandlerFunc)`
  2. Implement `memory.go` using standard Go channels, dynamic worker pools (goroutines), and a mapping registry under a `sync.RWMutex`.
  3. Ensure that panic/errors in subscriber callbacks are caught using `recover()` so they do not crash the publisher context.
- **Verification:** Write unit tests asserting that publishing an event async routes correctly to all registered handlers, and verifying handler panics are handled gracefully.

---

### S1-BE-11: Platform: Cache

- **Priority:** P1
- **Type:** Greenfield (New Module/Feature - Platform Abstraction)
- **Assignee:** BE-B
- **Story Points:** 2
- **Description:** Build the caching system interface and a concurrent-safe memory map implementation. As a greenfield task, this builds new platform or feature abstractions from scratch that do not exist in the old legacy codebase.
- **Detailed Steps:**
  - _Greenfield Note:_ Follow TDD (Red-Green-Refactor). Ensure the slice adheres strictly to boundary rules (D5) without importing other slices' stores/transports.
  1. Create `internal/platform/cache/cache.go`. Define the `Cache` interface (`Get`, `Set`, `Delete`).
  2. Implement `memory.go` using a map protected by a `sync.RWMutex`.
  3. Implement TTL (Time-To-Live) expiration checks using a background ticker goroutine that periodically sweeps expired keys.
- **Verification:** Write unit tests checking cache write/read, key expiration after TTL, and concurrent access safety under `go test -race`.

---

### S1-BE-12: Core: Realtime WebSocket Hub

- **Priority:** P1
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-B
- **Story Points:** 5
- **Description:** Setup the real-time WebSocket connection manager (Hub + Clients) for messaging and notification pushes. This migration extracts core logic from the legacy layered codebase into its own isolated vertical slice following the Strangler Fig pattern.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1.  Create `internal/core/realtime/hub.go` containing the central coordinator tracking active clients.
  2.  Implement client registration, deregistration, and broadcast loops.
  3.  Implement `internal/core/realtime/client.go` mapping WebSocket connections. Ensure panic recovery is set up on read/write loops.
  4.  Enforce read connection limit constraints and connection timeouts/deadlines.
  5.  **(Phase 3.2)** Implement `internal/core/realtime/router.go` for WebSocket message routing. Map incoming message types (chat.send, typing, etc.) to handler functions per vertical slice. Register routes for each slice's WS handlers.
- **Verification:** Write mock WS clients and test broadcast delivery speeds, client disconnects, and thread safety.

---

### S1-BE-13: Core: HTTP Server Bootstrap

- **Priority:** P1
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-B
- **Story Points:** 3
- **Dependencies:** S1-BE-12, S1-BE-08
- **Description:** Create the core HTTP server wrapper featuring graceful shutdown. This migration extracts core logic from the legacy layered codebase into its own isolated vertical slice following the Strangler Fig pattern.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1.  Implement `internal/core/server/server.go`.
  2.  Set up standard `http.Server` timeouts (ReadHeaderTimeout, ReadTimeout, WriteTimeout, IdleTimeout).
  3.  Handle OS signals (`SIGINT`, `SIGTERM`) to trigger graceful shutdown, allowing up to 10 seconds for draining ongoing connections.
  4.  **(Phase 3.4)** Implement `internal/core/server/routes.go` for centralized route registration. New vertical slices register their routes here rather than modifying `server.go` directly.
- **Verification:** Send mock kill signals to a running server process and ensure it logs shutdown progress and terminates cleanly.

---

## FE-A (Frontend A) Tickets

### S1-FE-03: Auth Pages (Login & Registration UI)

- **Priority:** P0
- **Type:** Greenfield (New Frontend UI - Next.js)
- **Assignee:** FE-A
- **Story Points:** 5
- **Description:** Implement complete interactive screens for Login and Registration according to strict spec fields. As a greenfield frontend task, this implements new Next.js UI components in `frontend/src/` utilizing shadcn/ui and Vanilla CSS, wiring them to the Next.js App Router.
- **Detailed Steps:**
  - _Greenfield Note:_ Use ESLint + Prettier for linting/formatting and ensure session cookies are handled securely without localStorage leakage.
  1. Create registration form page under `src/app/register/page.tsx` containing all 8 fields: Email, Password, First Name, Last Name, Date of Birth, Avatar (file upload input), Nickname, and About Me.
  2. Add client-side validation (e.g. valid email syntax, minimum password length, minimum age of 13 years old).
  3. Create login screen page under `src/app/login/page.tsx` with email/nickname and password inputs.
- **Verification:** Write Vitest rendering tests to verify that invalid inputs display form validation errors.

---

### S1-FE-04: API Client Wrapper

- **Priority:** P0
- **Type:** Greenfield (New Frontend Module)
- **Assignee:** FE-A
- **Story Points:** 2
- **Description:** Set up custom fetch API wrapper with automatic session cookies and global error parsing. As a greenfield frontend task, this implements new Next.js UI components in `frontend/src/` utilizing shadcn/ui and Vanilla CSS, wiring them to the Next.js App Router.
- **Detailed Steps:**
  - _Greenfield Note:_ Use ESLint + Prettier for linting/formatting and ensure session cookies are handled securely without localStorage leakage.
  1. Create `src/lib/api-client.ts`.
  2. Implement global default headers (including credential modes: `credentials: 'include'`).
  3. Handle generic API error structures and trigger routing to `/login` upon receiving `401 Unauthorized` responses.
- **Verification:** Write unit tests with mocked network responses to check HTTP code error conversions.

---

## FE-B (Frontend B) Tickets

### S1-FE-05: Nav Layout Shell

- **Priority:** P1
- **Type:** Greenfield (New Frontend UI - Next.js)
- **Assignee:** FE-B
- **Story Points:** 3
- **Description:** Finalize layout navigation, responsive states, and design theme provider. As a greenfield frontend task, this implements new Next.js UI components in `frontend/src/` utilizing shadcn/ui and Vanilla CSS, wiring them to the Next.js App Router.
- **Detailed Steps:**
  - _Greenfield Note:_ Use ESLint + Prettier for linting/formatting and ensure session cookies are handled securely without localStorage leakage.
  1. Build navigation sidebar listing links to Feed, Groups, Profile, and Chat.
  2. Build theme system provider (supporting Dark/Light HSL toggling).
  3. Implement responsive drawer menu for mobile breakpoints.
- **Verification:** Check visual rendering across mobile, tablet, and desktop breakpoints.

---

## SD-QA (System Design/QA) Tickets

### S1-SD-05: Platform: Database Seeding (Gap Fix)

- **Priority:** P2
- **Type:** Greenfield (New Module/Feature - Platform Seeding)
- **Assignee:** SD-QA
- **Story Points:** 2
- **Dependencies:** S1-BE-06
- **Description:** Define the structure for database seed migrations to inject realistic test data into the database in development mode. Note that the seed data script itself is numbered `000009_seed_data.up.sql` to execute _after_ all schemas (Sprints 2-4) are fully established. As a greenfield task, this builds new platform or feature abstractions from scratch that do not exist in the old legacy codebase.
- **Detailed Steps:**
  - _Greenfield Note:_ Follow TDD (Red-Green-Refactor). Ensure the slice adheres strictly to boundary rules (D5) without importing other slices' stores/transports.
  1. Create the template migration file `db/migrations/000009_seed_data.up.sql` containing placeholder SQL statements. Specify that the seeding script will include:
     - 4 test users with secure hashed passwords.
     - 5 posts with various privacy scopes (public, almost private, private).
     - 2 groups with membership details.
     - Basic follow relationships.
  2. Create migration file `db/migrations/000009_seed_data.down.sql` to clean up the seeded data.
  3. Update `internal/platform/database/migrations.go` in Sprint 1 to optionally apply seed migrations only when a development environment flag (e.g. `ENV=development`) is active, laying the platform runner logic.
- **Verification:** Confirm the migration runner compiled during S1-BE-06 correctly processes the `000009` file stub when dev flag is on.

---

### S1-SD-06: API Mocking Service

- **Priority:** P1
- **Type:** Greenfield (New Frontend Module - Testing)
- **Assignee:** SD-QA
- **Story Points:** 3
- **Dependencies:** S1-FE-04
- **Description:** Configure Mock Service Worker (msw) to allow mock testing frontend authentication flows prior to backend completions. As a greenfield frontend task, this implements new Next.js UI components in `frontend/src/` utilizing shadcn/ui and Vanilla CSS, wiring them to the Next.js App Router.
- **Detailed Steps:**
  - _Greenfield Note:_ Use ESLint + Prettier for linting/formatting and ensure session cookies are handled securely without localStorage leakage.
  1. Install `msw`. Configure service worker script in `src/mocks/browser.ts`.
  2. Implement mock route handlers mimicking backend response JSON payloads for login, registration, and logout commands.
  3. Conditionalize browser boot to initialize MSW when running in mock environments.
- **Verification:** Mock login actions in playwright tests and confirm route redirects are working.
