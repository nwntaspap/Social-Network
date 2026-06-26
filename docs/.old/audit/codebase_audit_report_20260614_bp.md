# Social Network Codebase Audit Report

## Executive Summary

- **Overall health**: ❌ Poor
  - Architecture: ⚠️ Fair (clean structure but major spec deviations)
  - Allowed Packages: ✅ Good
  - Go Idiom: ⚠️ Fair
  - Security: ⚠️ Fair
  - Performance: ⚠️ Fair
  - **Spec Compliance**: ❌ Poor — core SN features missing

- **Scope**: 280+ Go files, ~25K LOC backend, frontend SPA (vanilla JS + HTML/CSS)
- **Tool findings**: golangci-lint not run (tool not installed), govulncheck not installed, `go vet` — 1 build error in tests, `go mod graph` — all deps allowed

- **Spec compliance**: REQUIRED items passing 5/18, failing 13/18. BONUS items: 4 present, 1 mixed.

---

## Severity Legend

- CRITICAL: Exploitable vulnerability or guaranteed misbehaviour
- HIGH: Likely bug or grading spec violation
- MEDIUM: Best-practice gap or latent risk
- LOW: Style / maintainability suggestion
- BONUS: Optional capability (tagged PRESENT or MISSING)

---

## Spec Compliance Matrix

| Requirement                                 | Status | Location                                                                                   | Spec Ref                   |
| ------------------------------------------- | ------ | ------------------------------------------------------------------------------------------ | -------------------------- |
| Allowed packages only                       | ✅     | `go.mod`                                                                                   | readme.md Allowed Packages |
| WAL mode                                    | ✅     | `config.go:142` default `_journal_mode=WAL`                                                | readme.md Sqlite           |
| Busy timeout                                | ❌     | `config.go:142` — `_busy_timeout` absent                                                   | readme.md Sqlite           |
| Connection pooling (SQLite)                 | ❌\*\* | `sqlite/init.go:59-61` — only `SetMaxOpenConns`, no `SetMaxIdleConns`/`SetConnMaxLifetime` | readme.md Sqlite           |
| Follower request flow                       | ❌     | Entirely missing — no `followers` table, no handlers                                       | readme.md Followers        |
| Profile privacy (public/private)            | ❌     | Entirely missing — no privacy column, no toggle, no enforcement                            | readme.md Profile          |
| Auto-follow on public                       | ❌     | Entirely missing                                                                           | readme.md Followers        |
| Post privacy scopes (3 levels)              | ❌     | Entirely missing — `topics` table has no privacy column                                    | readme.md Posts            |
| Group lifecycle (browse/create/join/invite) | ❌     | Entirely missing — no `groups` table                                                       | readme.md Groups           |
| Group events (title/desc/time/RSVP)         | ❌     | Entirely missing                                                                           | readme.md Groups           |
| Group chat rooms                            | ❌     | Only direct chats exist (`direct_chats` table)                                             | readme.md Chat             |
| WebSocket auth handshake                    | ✅     | `server.go:384-389` — wrapped in `Authorization.Required`, handler checks user in context  | readme.md Chat             |
| Chat follow-based auth                      | ❌     | `chat/commands/initChat.go:28-33` — no follow check before creating chat                   | readme.md Chat             |
| Notification triggers (4 required)          | ❌     | Only 2 types (reply/like). No follow-request/group-invite/group-join/event-creation        | readme.md Notifications    |
| Startup migrations                          | ✅     | `sqlite/init.go:36-41` — `MigrateOnStart` default true                                     | readme.md Migrate          |
| Registration form fields                    | ❌     | Missing Date of Birth (uses Age int), Avatar/Image, About Me                               | readme.md Authentication   |
| Docker (2 containers)                       | ⚠️     | Single `Dockerfile` builds both binaries; `docker-compose.yml` runs both as 1 container    | readme.md docker           |
| Session cookies hardened                    | ✅     | `config.go:147-153` — HttpOnly default true, Secure configurable, SameSite Lax             | readme.md Authentication   |

\*\* \*Connection pooling partially implemented — only `SetMaxOpenConns` set, missing idle conns and max lifetime.

---

## Critical & High Findings

### AUDIT-001 — Spec mismatch: Project is a forum, not a social network

- **Severity**: CRITICAL
- **Location**: Entire codebase
- **Spec ref**: `readme.md` Objectives (Followers, Profile, Groups, Chat, Notifications)
- **Observation**: Project implements a forum/discussion-board (topics, categories, comments) with direct-messaging chat. The following SN features are completely absent:
  - Follow/unfollow system (no `followers` table)
  - Profile privacy (public/private)
  - Privacy-based post visibility (3 levels)
  - Groups (creation, join requests, invitations)
  - Group events (creation, RSVP)
  - Group chat rooms
  - Follow-based chat authorization
- **Evidence**: Schema has `topics`, `categories`, `comments`, `votes` (forum model). No `followers`, `groups`, `group_members`, `events`, `event_responses` tables. Zero grep matches for "follow" in Go code, zero for "group" (feature), zero for "profile privacy".
- **Risk**: Fails all grading criteria for Followers, Profile, Groups, Post privacy, and most of Chat/Notifications.
- **Remediation**: Full rewrite of ~70% of the backend to implement SN features per spec.
- **Judge Verdict**: CONFIRMED
- **Confidence**: HIGH

### AUDIT-002 — SQL migration parser splits on colon not semicolon

- **Severity**: CRITICAL
- **Location**: `internal/infra/storage/sqlite/init.go:118`
- **Spec ref**: readme.md Migrate
- **Observation**: `execSQLFile` uses `strings.SplitSeq(string(content), ":")` instead of `";"`. SQL statements are terminated by semicolons. Colons appear in SQL (e.g., `DATETIME DEFAULT CURRENT_TIMESTAMP`, `CHECK(...)`, `REFERENCES`). This will produce malformed SQL fragments.
- **Evidence**: Line 118: `statements := strings.SplitSeq(string(content), ":")`
- **Risk**: Migration execution may fail or produce incorrect schema. Some SQL may still work by accident if semicolons happen at line boundaries or the sqlite driver ignores trailing content. Under high concurrency or certain SQL constructs, this can corrupt the database.
- **Remediation**: Replace `":"` with `";"`:
  ```go
  statements := strings.SplitSeq(string(content), ";")
  ```
- **Judge Verdict**: CONFIRMED
- **Confidence**: HIGH

### AUDIT-003 — Missing 4 required notification triggers

- **Severity**: CRITICAL
- **Location**: `internal/domain/notification/notification.go:7-12`
- **Spec ref**: readme.md Notifications
- **Observation**: Only `reply`, `mention`, `like`, `dislike` notification types defined. Spec requires notifications for: (1) follow request received, (2) group invitation received, (3) group join request received, (4) group event created. These cannot be implemented without the underlying features, but even the type constants are missing.
- **Evidence**: `domain/notification/notification.go` lines 7-12 — only 4 types, none SN-related.
- **Risk**: Fails all 4 notification grading requirements. Zero bonus for extra notifications attempted.
- **Remediation**: Requires full follower/group/event features first, then add notification types: `following_request`, `group_invitation`, `group_join_request`, `event_created`.
- **Judge Verdict**: CONFIRMED
- **Confidence**: HIGH

### AUDIT-004 — No follow-based authorization for chat creation

- **Severity**: HIGH
- **Location**: `internal/app/chat/commands/initChat.go:28-33`, `internal/infra/http/chat/initChat/initChatHandler.go:28-65`
- **Spec ref**: readme.md Chat ("at least one of the users must be following the other")
- **Observation**: `InitChat` handler and command never check if `UserID` follows `MeID` or vice versa. Any authenticated user can initiate a chat with any other user.
- **Evidence**: `initChat.go:28-33` directly calls `chatRepo.GetOrCreateChat` without any follow check. `initChatHandler.go:51-58` only checks `UserID != MeID`.
- **Risk**: Spec violation. Users can message anyone regardless of follow relationship.
- **Remediation**: Add follow check in `initChatHandler` or `initChatCommand`:
  ```go
  // Before GetOrCreateChat
  if !h.followRepo.IsFollowing(ctx, req.MeID, req.UserID) &&
     !h.followRepo.IsFollowing(ctx, req.UserID, req.MeID) {
      return nil, errors.New("chat requires at least one user to follow the other")
  }
  ```
  Requires `followRepository` to exist.
- **Judge Verdict**: CONFIRMED
- **Confidence**: HIGH

### AUDIT-005 — Login error messages leak user existence

- **Severity**: HIGH
- **Location**: `internal/infra/http/user/login/LoginEmailHandler.go:65-68`
- **Spec ref**: readme.md Authentication
- **Observation**: Login handler returns `err.Error()` verbatim on failed login. The error contains the identifier (email/username), telling attackers whether a specific email is registered. The `userRepo` errors like `"user with email X not found: user not found"` are returned directly.
- **Evidence**: Line 67: `err.Error()` passed directly to response.
- **Risk**: User enumeration vulnerability. Attacker can probe which emails have accounts.
- **Remediation**: Return generic "invalid email or password" without differentiating:
  ```go
  helpers.RespondWithError(w, http.StatusUnauthorized, "Invalid email or password")
  ```
- **Judge Verdict**: CONFIRMED
- **Confidence**: HIGH

### AUDIT-006 — Go vet: tests fail to compile

- **Severity**: HIGH
- **Location**:
  - `internal/app/topics/commands/createTopic_test.go:87`
  - `internal/app/topics/commands/createTopic_test.go:100`
  - `internal/app/topics/commands/updateTopic_test.go:89`
  - `internal/app/topics/commands/updateTopic_test.go:102`
- **Spec ref**: General correctness
- **Observation**: `NewCreateTopicHandler` and `NewUpdateTopicHandler` now require 2 arguments (repo + fileStorage), but tests pass only 1 argument. Tests fail to build with `go vet`.
- **Evidence**: `go vet ./...` output: "not enough arguments in call to NewCreateTopicHandler"
- **Risk**: CI pipeline broken. Tests cannot run. Future regressions undetected.
- **Remediation**: Update test calls to pass `fileStorage` mock:
  ```go
  handler := NewCreateTopicHandler(repo, &mockFileStorage{})
  ```
- **Judge Verdict**: CONFIRMED
- **Confidence**: HIGH

### AUDIT-007 — Missing SQLite connection pool limits

- **Severity**: HIGH
- **Location**: `internal/infra/storage/sqlite/init.go:53-63`
- **Spec ref**: readme.md Sqlite
- **Observation**: Only `SetMaxOpenConns` configured (default 1). `SetMaxIdleConns` (default 2) and `SetConnMaxLifetime` (default unlimited) are not set. Unlimited idle connections and unlimited connection lifetime are problematic for SQLite — idle connections can hold locks and stale connections accumulate.
- **Evidence**: Line 59-61 only sets `SetMaxOpenConns`. No calls to `SetMaxIdleConns` or `SetConnMaxLifetime`.
- **Risk**: Under load, idle connections accumulate. Stale connections may cause `database is locked` errors or memory leaks.
- **Remediation**:
  ```go
  db.SetMaxOpenConns(cfg.Database.OpenConn)   // already done
  db.SetMaxIdleConns(cfg.Database.OpenConn)   // match max open
  db.SetConnMaxLifetime(5 * time.Minute)       // recycle connections
  ```
- **Judge Verdict**: CONFIRMED
- **Confidence**: HIGH

### AUDIT-008 — Missing `_busy_timeout` in SQLite DSN

- **Severity**: HIGH
- **Location**: `internal/config/config.go:142`
- **Spec ref**: readme.md Sqlite
- **Observation**: Default `DB_PRAGMA` is `"_foreign_keys=on&_journal_mode=WAL"`. No `_busy_timeout` parameter. Without busy timeout, SQLite immediately returns `database is locked` error when a write conflict occurs, rather than waiting for the lock to release.
- **Evidence**: Line 142: `Pragma: helpers.GetEnv("DB_PRAGMA", envMap, "_foreign_keys=on&_journal_mode=WAL")`
- **Risk**: Under concurrent writes (chat messages, notifications), users may see `database is locked` errors. This is especially problematic with `SetMaxOpenConns > 1`.
- **Remediation**: Change default to include busy timeout:
  ```go
  Pragma: helpers.GetEnv("DB_PRAGMA", envMap, "_foreign_keys=on&_journal_mode=WAL&_busy_timeout=5000")
  ```
- **Judge Verdict**: CONFIRMED
- **Confidence**: HIGH

### AUDIT-009 — Registration form incomplete per spec

- **Severity**: HIGH
- **Location**:
  - `internal/infra/http/user/register/registerhandler.go:17-25` (backend model)
  - `frontend/static/js/pages/register.js:53-116` (frontend form)
  - `internal/pkg/validator/validationCases.go:19-51` (validation rules)
- **Spec ref**: readme.md Authentication
- **Observation**: Spec requires: Email, Password, First Name, Last Name, Date of Birth, Avatar/Image (Optional), Nickname (Optional), About Me (Optional). Implementation has: Email, Password, Nickname (required), First Name, Last Name, Age (int, not DOB), Gender. Missing: Date of Birth (field should be date, not integer age), Avatar/Image upload on registration, About Me text field. Nickname is required but spec says optional.
- **Evidence**: Registration model has `Age int` not `DateOfBirth`. No `AboutMe` field. No `AvatarURL` in request model. Frontend form has no avatar upload or about me inputs.
- **Risk**: Spec violation. Testing will check form elements.
- **Remediation**: Add DateOfBirth (date string), AboutMe (optional text), Avatar/Image upload to registration form and handler. Make Nickname optional per spec.
- **Judge Verdict**: CONFIRMED
- **Confidence**: HIGH

### AUDIT-010 — Registration does not auto-login

- **Severity**: HIGH
- **Location**: `frontend/static/js/pages/register.js:267` — navigates to `/login` after registration
- **Spec ref**: readme.md Authentication
- **Observation**: After successful registration, user is redirected to login page instead of being automatically logged in (session created). The backend creates no session after registration.
- **Evidence**: Frontend line 267: `navigate('/login')`. Backend `registerhandler.go:110-120` only returns user ID and message, no session creation.
- **Risk**: Poor UX. Spec says "logged in state persists until explicit logout" — registration should start a session.
- **Remediation**: Backend: create session after `UserRegister` succeeds. Frontend: store session and redirect to home.
- **Judge Verdict**: CONFIRMED
- **Confidence**: HIGH

---

## Medium & Low Findings

### AUDIT-011 — Duplicate registration returns 500 not 409

- **Severity**: MEDIUM
- **Location**: `internal/infra/http/user/register/registerhandler.go:98-108`
- **Observation**: When registering with existing email/username, the handler returns HTTP 500 Internal Server Error. Should return 409 Conflict.
- **Risk**: Poor API semantics. Client cannot distinguish server errors from constraint violations.
- **Remediation**: Map SQLite UNIQUE constraint violation to 409:
  ```go
  if errors.Is(err, users.ErrUserAlreadyExists) || strings.Contains(err.Error(), "UNIQUE constraint") {
      helpers.RespondWithError(w, http.StatusConflict, "Email or username already registered")
      return
  }
  ```
- **Judge Verdict**: CONFIRMED
- **Confidence**: HIGH

### AUDIT-012 — Session store leaks password hash in API response

- **Severity**: MEDIUM
- **Location**: `internal/infra/storage/sessionstore/sessionManager.go:183-226` (`GetUserFromSession`)
- **Observation**: `GetUserFromSession` returns a `user.User` struct with `Password` field set to the bcrypt hash (from `password_hash` column). This user object is set into the request context by middleware and accessed by handlers. If any handler serializes the user to JSON, the hash is exposed.
- **Evidence**: Line 216: `&User.Password` is scanned from `u.password_hash` column.
- **Risk**: If any endpoint (e.g., `/me`) serializes the full user object, bcrypt hashes are leaked. Could enable offline brute-force of passwords.
- **Remediation**: Zero out the password field before returning or use a separate DTO:
  ```go
  User.Password = ""
  ```
- **Judge Verdict**: CONFIRMED
- **Confidence**: MEDIUM

### AUDIT-013 — gorilla/websocket mislabeled as indirect dependency

- **Severity**: LOW
- **Location**: `go.mod:6`
- **Observation**: `github.com/gorilla/websocket v1.5.3` is marked `// indirect` but is directly imported in multiple Go files (`internal/infra/ws/client.go`, `internal/infra/http/ws/handler.go`).
- **Risk**: `go mod tidy` may remove it if it thinks it's unused. Causes confusion.
- **Remediation**: Move to direct dependencies block:
  ```
  require (
      github.com/google/uuid v1.6.0
      github.com/gorilla/websocket v1.5.3
      github.com/mattn/go-sqlite3 v1.14.28
      golang.org/x/crypto v0.40.0
  )
  ```
- **Judge Verdict**: CONFIRMED
- **Confidence**: HIGH

### AUDIT-014 — Typo: "Middlware" instead of "Middleware"

- **Severity**: LOW
- **Location**: `internal/bootstrap/bootstrap.go:32`
- **Observation**: Field name `Middlware` is missing the 'e'. Used consistently throughout, so not a functional bug, but signals lack of attention.
- **Remediation**: Rename to `Middleware`.
- **Judge Verdict**: CONFIRMED
- **Confidence**: HIGH

### AUDIT-015 — OpenDB returns misleading double \*sql.DB

- **Severity**: LOW
- **Location**: `internal/infra/storage/sqlite/init.go:53`
- **Observation**: `OpenDB` signature is `func OpenDB(cfg config.ServerConfig) (*sql.DB, *sql.DB, error)`. Second return value is always `nil, nil`. The caller at `InitializeDB` ignores it correctly, but it's confusing.
- **Remediation**: Change signature to return single `*sql.DB`:
  ```go
  func OpenDB(cfg config.ServerConfig) (*sql.DB, error)
  ```
- **Judge Verdict**: CONFIRMED
- **Confidence**: HIGH

### AUDIT-016 — Unconditional session deletion on new session creation

- **Severity**: MEDIUM
- **Location**: `internal/infra/storage/sessionstore/sessionManager.go:244-255`
- **Observation**: `DeleteSessionWhenNewCreated` deletes ALL other sessions for the user unconditionally, ignoring `MaxSessionsPerUser` config (default 5). The config field exists but is never checked.
- **Risk**: Users can only have 1 active session. Logging in on a new device logs out all other devices.
- **Remediation**: Count sessions first, delete only if exceeding max:
  ```go
  // Before deletion, count sessions
  // Only delete oldest sessions exceeding MaxSessionsPerUser
  ```
- **Judge Verdict**: CONFIRMED
- **Confidence**: HIGH

### AUDIT-017 — execSQLFile rollback pattern is unsafe

- **Severity**: MEDIUM
- **Location**: `internal/infra/storage/sqlite/init.go:96-136`
- **Observation**: The `defer transaction.Rollback()` always runs, including after successful `transaction.Commit()`. While the code checks `sql.ErrTxDone` to avoid errors, the pattern is fragile. If `Commit` fails, the rollback in defer will attempt to rollback an already-rolled-back transaction.
- **Evidence**: Lines 102-116: defer rollback with ErrTxDone check. Line 131: commit. If commit fails, defer rollback also fails silently.
- **Risk**: Silent transaction failures could leave partial migrations applied.
- **Remediation**:
  ```go
  committed := false
  err = transaction.Commit()
  if err == nil {
      committed = true
  }
  // In defer:
  if !committed {
      transaction.Rollback()
  }
  ```
- **Judge Verdict**: CONFIRMED
- **Confidence**: MEDIUM

### AUDIT-018 — Migration files don't follow spec directory structure

- **Severity**: MEDIUM
- **Location**: `db/migrations/schema.sql`, `db/migrations/indexes.sql`
- **Spec ref**: readme.md Migrate (expects `000001_create_users_table.up.sql` / `.down.sql` format)
- **Observation**: Spec expects numbered migration files with up/down pairs. Implementation uses single `schema.sql` + `indexes.sql` without up/down separation or version numbering. No rollback support.
- **Risk**: Cannot rollback migrations. Hard to track which version is applied.
- **Remediation**: Use `golang-migrate/migrate` compatible file naming.
- **Judge Verdict**: WEAKENED — custom migration system is permitted by "or other package that better suits your project", but grading checklist specifically checks for the up/down pattern.
- **Confidence**: MEDIUM

### AUDIT-019 — No goroutine recovery in WebSocket goroutines

- **Severity**: MEDIUM
- **Location**: `internal/infra/ws/client.go:40,67`
- **Observation**: `ReadPump` and `WritePump` run in separate goroutines (`handler.go:53-54`) but have no `defer recover()`. A panic in any goroutine crashes the entire server.
- **Risk**: A malformed WebSocket message or send to closed channel causes server-wide crash.
- **Remediation**: Add recover:
  ```go
  func (c *Client) ReadPump(onMessage func(client *Client, msg []byte)) {
      defer func() {
          if r := recover(); r != nil {
              log.Printf("ws read pump panic: %v", r)
          }
          c.hub.Unregister(c)
          c.conn.Close()
      }()
      // ...
  }
  ```
- **Judge Verdict**: CONFIRMED
- **Confidence**: MEDIUM

### AUDIT-020 — Docker runs both services in 1 container

- **Severity**: MEDIUM (spec deviation)
- **Location**: `Dockerfile`, `docker-compose.yml`
- **Spec ref**: readme.md docker ("two Docker images, one for backend and another for frontend")
- **Observation**: Single `Dockerfile` builds both binaries. `docker-compose.yml` defines 1 service (`forum`) running both backend and frontend via `entrypoint.sh`. Spec expects 2 separate containers.
- **Risk**: Grading violation. Single container couples frontend/backend lifecycle. Cannot scale independently.
- **Remediation**: Split into `Dockerfile.backend` and `Dockerfile.frontend`, define 2 services in docker-compose.
- **Judge Verdict**: CONFIRMED
- **Confidence**: HIGH

---

## Bonus Feature Inventory

| Feature                              | Status     | Location                                                                                                                                                                      |
| ------------------------------------ | ---------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| OAuth (GitHub + Google)              | ✅ PRESENT | `internal/pkg/oAuth/` — GitHub and Google OAuth providers fully implemented with state management                                                                             |
| DB seeding                           | ✅ PRESENT | `db/seeds/dev_data.sql`, executed on `InitializeDB` in dev/staging                                                                                                            |
| Confirmation popups (unfollow)       | ❌ MISSING | No unfollow feature exists, so no confirmation. Search found zero popup implementations                                                                                       |
| Confirmation popups (privacy toggle) | ❌ MISSING | No privacy toggle exists, so no confirmation                                                                                                                                  |
| Docker build scripts                 | ⚠️ PARTIAL | Single `Dockerfile` + `docker-compose.yml` + `entrypoint.sh` + `DOCKER_SETUP.md`. Builds both services in 1 image                                                             |
| Extra notifications                  | ✅ PRESENT | `reply`, `like`, `dislike` types beyond the 4 required (but implemented as a forum, not SN context)                                                                           |
| Overall project quality              | ⚠️ Fair    | Clean architecture respected. App layer separate. SQL injection safe. But massive spec mismatch — project is a forum, not a social network. Tests sparse, largely out of date |

---

## Judge Phase: Adversarial Validation

### Step 3.1 — Evidence Audit

| Finding                             | Verdict   | Notes                                                                 |
| ----------------------------------- | --------- | --------------------------------------------------------------------- |
| AUDIT-001 (Spec mismatch)           | CONFIRMED | Zero matching code; schema confirms forum model                       |
| AUDIT-002 (Colon split)             | CONFIRMED | `strings.SplitSeq(content, ":")` at init.go:118                       |
| AUDIT-003 (Missing notifications)   | CONFIRMED | Only 4 types in domain model, none SN-related                         |
| AUDIT-004 (No follow check chat)    | CONFIRMED | `initChat.go` has zero follow logic                                   |
| AUDIT-005 (Login info leak)         | CONFIRMED | `err.Error()` returned verbatim at LoginEmailHandler.go:67            |
| AUDIT-006 (Test compilation)        | CONFIRMED | `go vet` confirmed the build error                                    |
| AUDIT-007 (Missing pool config)     | CONFIRMED | Only SetMaxOpenConns called                                           |
| AUDIT-008 (Missing busy_timeout)    | CONFIRMED | Default pragma verified at config.go:142                              |
| AUDIT-009 (Incomplete registration) | CONFIRMED | Missing DOB, Avatar, AboutMe — confirmed in both frontend and backend |
| AUDIT-010 (No auto-login)           | CONFIRMED | Registration returns JSON, no session created                         |
| AUDIT-011 (Wrong HTTP status)       | CONFIRMED | 500 returned, not 409                                                 |
| AUDIT-012 (Password hash leak)      | CONFIRMED | `password_hash` scanned into User.Password at sessionManager.go:216   |
| AUDIT-013 (indirect mislabel)       | CONFIRMED | go.mod line 6 verified                                                |
| AUDIT-014 (Middlware typo)          | CONFIRMED | bootstrap.go:32                                                       |
| AUDIT-015 (Double \*sql.DB)         | CONFIRMED | OpenDB signature at init.go:53                                        |
| AUDIT-016 (Unconditional delete)    | CONFIRMED | sessionManager.go:244-255                                             |
| AUDIT-017 (Rollback pattern)        | CONFIRMED | init.go:102-136 reviewed                                              |
| AUDIT-018 (Migration format)        | WEAKENED  | Custom format works but doesn't match spec example                    |
| AUDIT-019 (No goroutine recover)    | CONFIRMED | client.go:40,67 no recover                                            |
| AUDIT-020 (Single Docker container) | CONFIRMED | docker-compose.yml:1 service                                          |

### Step 3.2 — Gap Analysis (JUDGE-FOUND)

No additional critical findings beyond those in Phase 2. Phase 2 analysis was thorough.

### Step 3.3 — Mitigation Check

- AUDIT-012 (password leak): Partially mitigated — no handler currently serializes the full user object to JSON. The `/me` handler returns only specific fields. Risk is low in current code but latent.
- AUDIT-019 (goroutine panic): No mitigation — unchecked panics will crash server.

### Step 3.4 — Socratic Challenge

- AUDIT-001 (Spec mismatch): Fix would require rewriting ~70% of backend. Is it worth it? If this is a fork of a forum project submitted as an SN, the gap is fundamental. Recommendation: REDESIGN from scratch with proper SN features.
- AUDIT-004 (Chat follow check): Fix complexity is low (add repo + one query). No negative tradeoffs. Must fix.
- AUDIT-005 (Login error): Fix is trivial. Must fix.
- AUDIT-012 (Password hash leak): Fix is 1 line. No tradeoff. Must fix.

### Step 3.5 — Spec Compliance Double-Check

All findings cross-referenced against `docs/requirements/readme.md` and `docs/requirements/audit.md`. No findings were dropped for spec misreading.

---

## Verification Plan

### Commands to run fixes

```bash
# Fix test compilation
cd internal/app/topics/commands && go test ./...

# Full test suite
go test -race -coverprofile=coverage.out -covermode=atomic ./...

# Run vet after fixes
go vet ./...

# Lint (requires golangci-lint installed via make tools)
make lint
```

### Manual testing steps

1. **Registration flow**: Create user, check auto-login, verify form fields (Email, Password, First Name, Last Name, Date of Birth)
2. **Login flow**: Login with correct/wrong credentials — no user enumeration via error messages
3. **Session persistence**: Open two browsers, login to each, refresh — sessions persist correctly
4. **WebSocket chat**: Two browsers, follow-based auth, verify only followed users can chat
5. **Profile privacy**: Toggle public/private, verify non-followers blocked, followers have access
6. **Post privacy**: Create posts with 3 visibility levels, verify enforcement
7. **Group lifecycle**: Create/browse/join/invite groups, events with RSVP
8. **Notifications**: Trigger all 4 notification types, verify delivery and display
9. **Docker verification**: `docker compose up` + `docker ps -a` — verify 2 containers both non-zero
10. **SQLite config**: Connect with `sqlite3 db/data/forum.db`, verify PRAGMA journal_mode=wal, busy_timeout=5000

---

## AGENTS.md Update Summary

### Patterns discovered

- Always check SQLite DSN parameters: `_journal_mode=WAL`, `_busy_timeout=5000`
- Never trust project name — verify actual implementation matches spec (forum != social network)
- Check migration file format — spec expects numbered up/down pairs
- Verify registration form fields match spec exactly (field names, optional/required)
- Session management must check `password_hash` isn't leaked in API responses

### Common false positives to suppress

- None specific to this codebase yet

### Tool invocations

- `go vet ./...` — caught test build failures immediately
- `go mod graph` — verified allowed packages
- `rtk grep "pattern" internal/ --type go` — fast codebase search

### Spec items frequently misread

- Allowed packages: `gorilla/websocket` IS allowed (listed in readme.md) — do NOT flag it
- Migration format: custom systems allowed but up/down pairs are expected by audit
- Docker: "two Docker images" means 2 services in docker-compose, not 2 binaries in 1 image
