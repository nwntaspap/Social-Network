# Social Network Codebase Audit Report

**Date**: 2026-06-14
**Codebase**: `github.com/arnald/forum` (social-network)
**Audit scope**: Full codebase audit against grading specification (`docs/requirements/readme.md`, `docs/requirements/audit.md`)
**Audit tooling**: golangci-lint v1.64.8, govulncheck v1.3.0, go vet, go mod graph, manual code review

---

## Executive Summary

### Overall Health

| Layer                       | Verdict | Description                                                                                                                                                                                                            |
| --------------------------- | ------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Architecture & Design**   | ⚠️ Fair | Hexagonal/DDD 3-layer architecture is generally respected, but layer skipping (handler calling storage error definitions) and lack of formal JS framework are compliance gaps.                                         |
| **Allowed Packages**        | ✅ Pass | All external direct and indirect dependencies are fully compliant with the specification's allowed list.                                                                                                               |
| **Go Idioms & Correctness** | ❌ Poor | Multiple critical/high correctness defects exist: a `context.Context` passed to `Scan()`, unused prepared statements, and two indefinite goroutine leaks.                                                              |
| **Security**                | ❌ Poor | Highly vulnerable state: cross-origin websocket hijacking (CSWSH) allowed, credentials validation leaks, SQL injection vectors in query ordering, and complete absence of profile privacy/follower/post authorization. |
| **Performance**             | ⚠️ Fair | Uses WAL mode and pooling max conns, but suffers from missing `_busy_timeout` write contention limits and multiple background leaks.                                                                                   |
| **Spec Compliance**         | ❌ Poor | Major functional sections (Followers, Groups, Events, Post/Profile Privacy, 4 Notification types, 2-Container Docker Setup, and JS Framework usage) are entirely missing or violating spec mandates.                   |

---

## Severity Legend

- **CRITICAL**: Exploitable vulnerability, runtime panic, or guaranteed misbehavior under normal operation.
- **HIGH**: Grading spec violation or highly likely bug causing security bypass or resource leak.
- **MEDIUM**: Code structure, maintenance, or latent security risk.
- **LOW**: Minor code style, dead code, or documentation suggestion.
- **BONUS**: Optional capabilities checking.

---

## Spec Compliance Matrix

| Requirement                                           | Status     | Location                                           | Spec Ref                   |
| ----------------------------------------------------- | ---------- | -------------------------------------------------- | -------------------------- |
| **Allowed packages only**                             | ✅ PASS    | `go.mod`                                           | readme.md Allowed Packages |
| **WAL mode**                                          | ✅ PASS    | `config.go:142`                                    | readme.md Sqlite           |
| **Busy timeout (`_busy_timeout=5000`)**               | ❌ FAIL    | `config.go:142` (missing)                          | readme.md Sqlite           |
| **Connection pooling**                                | ⚠️ PARTIAL | `init.go:59-61` (only SetMaxOpenConns is set)      | readme.md Sqlite           |
| **Follower request flow (accept/decline)**            | ❌ FAIL    | No code exists                                     | readme.md Followers        |
| **Auto-follow on public profiles**                    | ❌ FAIL    | No code exists                                     | readme.md Followers        |
| **Unfollow**                                          | ❌ FAIL    | No code exists                                     | readme.md Followers        |
| **Profile privacy toggle (public/private)**           | ❌ FAIL    | No `is_private` column or toggle code              | readme.md Profile          |
| **Profile display (all registration fields)**         | ❌ FAIL    | `getMe/handler.go:43-50` (incomplete fields)       | readme.md Profile          |
| **Post creation (with auth)**                         | ✅ PASS    | `createTopicHandler.go`                            | readme.md Posts            |
| **Post privacy scopes (3 levels)**                    | ❌ FAIL    | No privacy scope columns on `topics`               | readme.md Posts            |
| **Media MIME validation (JPG/PNG/GIF)**               | ✅ PASS    | `validator.go:113-163`                             | readme.md Posts            |
| **Group creation**                                    | ❌ FAIL    | No code exists                                     | readme.md Groups           |
| **Group discovery / browse**                          | ❌ FAIL    | No code exists                                     | readme.md Groups           |
| **Group join requests (accept/refuse)**               | ❌ FAIL    | No code exists                                     | readme.md Groups           |
| **Group invitations (accept/decline)**                | ❌ FAIL    | No code exists                                     | readme.md Groups           |
| **Group posts and comments member visibility**        | ❌ FAIL    | No code exists                                     | readme.md Groups           |
| **Group chat rooms**                                  | ❌ FAIL    | No code exists                                     | readme.md Groups           |
| **Group events (title, desc, time, Going/Not Going)** | ❌ FAIL    | No code exists                                     | readme.md Groups           |
| **Event voting (Going/Not Going)**                    | ❌ FAIL    | No code exists                                     | readme.md Groups           |
| **WebSocket handshake auth**                          | ✅ PASS    | `ws/handler.go:37-43`, `server.go:384-389`         | readme.md Chat             |
| **Chat follow-based auth**                            | ❌ FAIL    | `initChat.go:28-33` (does not check follow)        | readme.md Chat             |
| **Chat read limits (conn.SetReadLimit)**              | ✅ PASS    | `client.go:14,46` (4096 bytes limit)               | readme.md Chat             |
| **Chat deadlines (SetRead/WriteDeadline)**            | ✅ PASS    | `client.go:47-49,77,88`                            | readme.md Chat             |
| **Emoji support**                                     | ✅ PASS    | Go UTF-8 support for Unicode                       | readme.md Chat             |
| **Private message targeting (recipient-only)**        | ✅ PASS    | `chatSend.go:44-55`, `hub.go:176-199`              | readme.md Chat             |
| **Notification: follow request received**             | ❌ FAIL    | No code exists                                     | readme.md Notifications    |
| **Notification: group invitation received**           | ❌ FAIL    | No code exists                                     | readme.md Notifications    |
| **Notification: group join request received**         | ❌ FAIL    | No code exists                                     | readme.md Notifications    |
| **Notification: new event created**                   | ❌ FAIL    | No code exists                                     | readme.md Notifications    |
| **Global notification access (on every page)**        | ✅ PASS    | SSE streaming at `/api/v1/notifications/stream`    | readme.md Notifications    |
| **Notification vs message separation**                | ✅ PASS    | Distinct models, tables, and streams               | readme.md Notifications    |
| **Startup migrations**                                | ✅ PASS    | `InitializeDB` automatically called in `main.go`   | readme.md Migrate          |
| **Migration format (numbered up/down)**               | ❌ FAIL    | Custom SQL files `schema.sql` and `indexes.sql`    | readme.md Migrate          |
| **Registration: Email, Password, First/Last Name**    | ✅ PASS    | `registerhandler.go`                               | readme.md Authentication   |
| **Registration: Date of Birth**                       | ❌ FAIL    | Uses `Age` integer instead of `Date of Birth` date | readme.md Authentication   |
| **Registration: Avatar (optional)**                   | ❌ FAIL    | Absent from form and user creation                 | readme.md Authentication   |
| **Registration: Nickname (optional)**                 | ✅ PASS    | `registerhandler.go:18`                            | readme.md Authentication   |
| **Registration: About Me (optional)**                 | ❌ FAIL    | Absent from model and schema                       | readme.md Authentication   |
| **bcrypt password hashing (cost >= 10)**              | ✅ PASS    | `encryption.go:9` uses cost=12                     | readme.md Authentication   |
| **Session cookies (HttpOnly, Secure, SameSite)**      | ✅ PASS    | `manager.go:102-123`                               | readme.md Authentication   |
| **Session persistence**                               | ✅ PASS    | Survives page reload, refresh-token rotation       | readme.md Authentication   |
| **Login error (no email leak)**                       | ❌ FAIL    | `LoginEmailHandler.go` leaks credentials existence | audit.md Authentication    |
| **Duplicate registration rejection**                  | ✅ PASS    | Checked via SQL UNIQUE constraints                 | audit.md Authentication    |
| **Docker (2 containers)**                             | ❌ FAIL    | Single container running both binaries             | readme.md docker           |
| **JS Framework used**                                 | ❌ FAIL    | Code uses custom Vanilla JS SPA, no JS framework   | readme.md Frontend         |

---

## Critical & High Findings

### AUDIT-001: Follower System Completely Absent

- **ID**: AUDIT-001
- **Severity**: CRITICAL
- **Location**: Entire codebase
- **Spec ref**: `readme.md` Followers
- **Observation**: There is no follow or unfollow system implemented. The database schema has no tables for follows or follow requests, and no domain, application, or infrastructure code references this logic.
- **Evidence**:
  `grep_search` for `follow` or `follower` in `internal/` yields zero results.
- **Risk**: Complete failure to satisfy grading specifications for followers, which also breaks related privacy permissions in chats, profiles, and post visibilities.
- **Remediation**:
  Create the database tables `follows` and `follow_requests`.
  Implement domain model interfaces, use cases, HTTP routing, and handlers to support request flows, accept/decline, unfollowing, and auto-follow rules for public profiles.
- **Judge Verdict**: CONFIRMED
- **Confidence**: HIGH

---

### AUDIT-002: Groups and Events Completely Absent

- **ID**: AUDIT-002
- **Severity**: CRITICAL
- **Location**: Entire codebase
- **Spec ref**: `readme.md` Groups
- **Observation**: No groups, group memberships, group posts, group events, event RSVP, or group chat rooms are present in the codebase.
- **Evidence**:
  `grep_search` for `group` or `event` returns zero domain-specific files or tables. `schema.sql` has no group tables.
- **Risk**: Fails to meet the groups and events specification requirements entirely.
- **Remediation**:
  Create schema tables: `groups`, `group_members`, `group_invitations`, `group_posts`, `group_events`, `event_rsvps`.
  Implement full greenfield functionality in all layers (`internal/domain/group`, `internal/app/group`, and handlers).
- **Judge Verdict**: CONFIRMED
- **Confidence**: HIGH

---

### AUDIT-003: `ctx` Passed as Scan Destination in oauthRepo

- **ID**: AUDIT-003
- **Severity**: CRITICAL
- **Location**: [oauthRepo.go:182-188](file:///home/ertval/code/zone-modules/social-network/internal/infra/storage/sqlite/oauth/oauthRepo.go#L182-L188)
- **Spec ref**: N/A (Go Correctness)
- **Observation**: In `GetOAuthProvider()`, the `Scan()` destination list incorrectly starts with the `ctx` variable (of type `context.Context`) as a scanning target. This causes a type mismatch and runtime error since the SQL statement only queries 4 columns but receives 5 destination variables, with the first being a non-pointer interface.
- **Evidence**:
  ```go
  err = stmt.QueryRowContext(ctx, userID, string(provider)).Scan(
  	ctx,
  	&oauthUser.ProviderID,
  	&oauthUser.Email,
  	&oauthUser.Username,
  	&oauthUser.AvatarURL,
  )
  ```
- **Risk**: Guaranteed runtime panic or query failure whenever users attempt to link or sign in via OAuth.
- **Remediation**:
  Remove `ctx` from the `.Scan()` parameters list:
  ```diff
  -	err = stmt.QueryRowContext(ctx, userID, string(provider)).Scan(
  -		ctx,
  -		&oauthUser.ProviderID,
  +	err = stmt.QueryRowContext(ctx, userID, string(provider)).Scan(
  +		&oauthUser.ProviderID,
  ```
- **Judge Verdict**: CONFIRMED
- **Confidence**: HIGH

---

### AUDIT-004: Migration Statement Delimiter Uses `:` Instead of `;`

- **ID**: AUDIT-004
- **Severity**: HIGH
- **Location**: [init.go:118](file:///home/ertval/code/zone-modules/social-network/internal/infra/storage/sqlite/init.go#L118)
- **Spec ref**: `readme.md` Migrate
- **Observation**: The custom SQL execution function `execSQLFile` splits migration scripts on a colon (`":"`) rather than a semicolon (`";"`). This makes it impossible to include colons in migration files (such as database constraints, string literals, or times) without splitting the commands incorrectly.
- **Evidence**:
  ```go
  statements := strings.SplitSeq(string(content), ":")
  ```
- **Risk**: Silent compilation or query failures for any future migrations containing colon characters.
- **Remediation**:
  Change the delimiter from `":"` to `";"`:
  ```diff
  -	statements := strings.SplitSeq(string(content), ":")
  +	statements := strings.SplitSeq(string(content), ";")
  ```
- **Judge Verdict**: CONFIRMED
- **Confidence**: HIGH

---

### AUDIT-005: WebSocket `CheckOrigin` Allows All Origins (CSRF Vector)

- **ID**: AUDIT-005
- **Severity**: HIGH
- **Location**: [handler.go:16-20](file:///home/ertval/code/zone-modules/social-network/internal/infra/http/ws/handler.go#L16-L20)
- **Spec ref**: `readme.md` Chat (WebSocket)
- **Observation**: The WebSocket upgrader implements a `CheckOrigin` function that always returns `true`, allowing cross-origin WebSocket requests from any origin.
- **Evidence**:
  ```go
  var upgrader = websocket.Upgrader{
  	ReadBufferSize:  1024,
  	WriteBufferSize: 1024,
  	CheckOrigin: func(r *http.Request) bool {
  		// TODO: actual origin in production
  		return true
  	},
  }
  ```
- **Risk**: Cross-Site WebSocket Hijacking (CSWSH). A malicious external page can connect to the logged-in user's WebSocket stream, reading and sending messages on their behalf.
- **Remediation**:
  Inspect the `Origin` header and restrict connection upgrades to matched trusted origins (e.g. comparing with configured host/port).
- **Judge Verdict**: JUDGE-FOUND
- **Confidence**: HIGH

---

### AUDIT-006: StateManager and RateLimiter Goroutines Leak on Shutdown

- **ID**: AUDIT-006
- **Severity**: HIGH
- **Location**: [stateManager.go:43](file:///home/ertval/code/zone-modules/social-network/internal/pkg/oAuth/stateManager.go#L43), [rateLimiter.go:34](file:///home/ertval/code/zone-modules/social-network/internal/infra/middleware/ratelimiter/rateLimiter.go#L34)
- **Spec ref**: N/A (Go Correctness)
- **Observation**: Both `StateManager` and `RateLimiter` spawn background cleanup loops in goroutines using infinite `for range ticker.C` structures, but neither receives a shutdown signal (like a `stop` channel or `context.Context`).
- **Evidence**:
  In `stateManager.go`:
  ```go
  go sm.cleanup()
  ```
  In `rateLimiter.go`:
  ```go
  go rl.cleanup()
  ```
- **Risk**: Goroutine leak on application reload or test execution. The leaked goroutines continue holding structural references, causing memory bloat.
- **Remediation**:
  Add a `stop chan struct{}` field to the structs. Check for channel closure in select blocks along with ticker events, and close the channels on shutdown.
- **Judge Verdict**: CONFIRMED (expanded to include RateLimiter)
- **Confidence**: HIGH

---

### AUDIT-007: Login Error Leaks User Email/Username Existence

- **ID**: AUDIT-007
- **Severity**: HIGH
- **Location**: [LoginEmailHandler.go:64-73](file:///home/ertval/code/zone-modules/social-network/internal/infra/http/user/login/LoginEmailHandler.go#L64-L73) and [userRepo.go:138](file:///home/ertval/code/zone-modules/social-network/internal/infra/storage/sqlite/users/userRepo.go#L138)
- **Spec ref**: `audit.md` Authentication
- **Observation**: When user lookup fails during login, the repository returns detailed errors containing the email (e.g. `"user with email ... not found"`). The HTTP handler then responds directly with `err.Error()`, exposing user details to the client.
- **Evidence**:
  ```go
  if err != nil {
  	helpers.RespondWithError(w,
  		http.StatusInternalServerError,
  		err.Error(),
  	)
  ```
- **Risk**: Credentials enumeration. Attackers can test email lists to discover who is registered on the platform.
- **Remediation**:
  Do not leak repository errors. Always return a generic message (e.g. `"Invalid email or password"`) for any authentication failure.
- **Judge Verdict**: CONFIRMED
- **Confidence**: HIGH

---

### AUDIT-008: Missing `_busy_timeout` in SQLite Connection String

- **ID**: AUDIT-008
- **Severity**: HIGH
- **Location**: [config.go:142](file:///home/ertval/code/zone-modules/social-network/internal/config/config.go#L142)
- **Spec ref**: `readme.md` Sqlite
- **Observation**: The DSN configuration is missing `_busy_timeout=5000`. Without this setting, SQLite does not block and wait for writes under contention, but instantly throws lock exceptions.
- **Evidence**:
  ```go
  Pragma: helpers.GetEnv("DB_PRAGMA", envMap, "_foreign_keys=on&_journal_mode=WAL")
  ```
- **Risk**: Highly probable `database is locked` error events when concurrent writes happen.
- **Remediation**:
  Append `&_busy_timeout=5000` to the default database pragma query string.
- **Judge Verdict**: CONFIRMED
- **Confidence**: HIGH

---

### AUDIT-009: JS Framework Requirement Violations

- **ID**: AUDIT-009
- **Severity**: HIGH
- **Location**: [frontend/static/js/](file:///home/ertval/code/zone-modules/social-network/frontend/static/js/)
- **Spec ref**: `readme.md` Frontend (Framework)
- **Observation**: The subject explicitly requires the use of a JS framework (e.g. Next.js, Vue, Svelte, or Mithril). The project instead uses custom Vanilla JS scripts with hand-rolled routing.
- **Evidence**:
  Directory `frontend/static/js` contains only custom vanilla `.js` files. No `package.json` config exists inside frontend for framework compilation.
- **Risk**: Direct grading spec violation.
- **Remediation**:
  Migrate the frontend code to an approved framework (e.g., Svelte or Next.js).
- **Judge Verdict**: CONFIRMED
- **Confidence**: HIGH

---

### AUDIT-010: Docker Setup Violates Multi-Container Requirement

- **ID**: AUDIT-010
- **Severity**: HIGH
- **Location**: [docker-compose.yml](file:///home/ertval/code/zone-modules/social-network/docker-compose.yml) and [Dockerfile](file:///home/ertval/code/zone-modules/social-network/Dockerfile)
- **Spec ref**: `readme.md` docker
- **Observation**: The spec requires two distinct Docker images (one for frontend, one for backend) configured as separate services in docker-compose. The codebase instead builds one multi-binary Docker image and runs them concurrently in a single container via a shell wrapper.
- **Evidence**:
  `docker-compose.yml` only has a single service (`forum`) mapping both ports. `Dockerfile` compiles both binaries and runs `entrypoint.sh` which forks both.
- **Risk**: Grading non-compliance. Under concurrent loads, failure to monitor independent containers could result in silent crashes.
- **Remediation**:
  Split `Dockerfile` into a `Dockerfile.backend` and `Dockerfile.frontend`. Define two separate services under `docker-compose.yml` (`backend` and `frontend`).
- **Judge Verdict**: CONFIRMED
- **Confidence**: HIGH

---

### AUDIT-011: Post Privacy Scopes and Profile Privacy Toggles Missing

- **ID**: AUDIT-011
- **Severity**: HIGH
- **Location**: `db/migrations/schema.sql`
- **Spec ref**: `readme.md` Posts, `readme.md` Profile
- **Observation**: There is no database or controller support for public/private profile states or post privacy scopes (public/almost private/private).
- **Evidence**:
  `schema.sql` has no `is_private` column on `users`, nor does `topics` have a `visibility` column.
- **Risk**: Fails to implement privacy permissions as required.
- **Remediation**:
  Add `is_private` boolean to users table and `visibility` constraint enum to topics table. Implement server-side check queries.
- **Judge Verdict**: CONFIRMED
- **Confidence**: HIGH

---

## Medium & Low Findings

### AUDIT-012: Registration Form Field Non-compliance

- **Severity**: MEDIUM
- **Location**: [registerhandler.go:23](file:///home/ertval/code/zone-modules/social-network/internal/infra/http/user/register/registerhandler.go#L23)
- **Spec ref**: `readme.md` Authentication
- **Observation**: Registration uses `Age` (integer) instead of Date of Birth (date format), and does not support optional "Avatar" or "About Me" registration fields.
- **Remediation**: Update models and register schema to use Date of Birth, and add the optional fields.
- **Judge Verdict**: CONFIRMED

---

### AUDIT-013: Lack of Panic Recovery in WebSocket Goroutines

- **Severity**: MEDIUM
- **Location**: [client.go:40](file:///home/ertval/code/zone-modules/social-network/internal/infra/ws/client.go#L40) and [client.go:67](file:///home/ertval/code/zone-modules/social-network/internal/infra/ws/client.go#L67)
- **Spec ref**: N/A (Resiliency)
- **Observation**: The `ReadPump` and `WritePump` goroutines lack `recover()` wrappers. Any runtime panic inside them will immediately terminate the entire server process.
- **Remediation**: Add standard `defer recover()` statements to client pump routines.
- **Judge Verdict**: CONFIRMED

---

### AUDIT-014: SQLite Connection Pool Incomplete

- **Severity**: MEDIUM
- **Location**: [init.go:59-61](file:///home/ertval/code/zone-modules/social-network/internal/infra/storage/sqlite/init.go#L59-L61)
- **Spec ref**: N/A (Performance)
- **Observation**: The connection setup calls `SetMaxOpenConns(1)` but does not configure `SetMaxIdleConns()` or `SetConnMaxLifetime()`, which leaves connection recycling unoptimized.
- **Remediation**: Set max idle connections to 1 and set a lifetime limit (e.g. 15 minutes).
- **Judge Verdict**: CONFIRMED

---

### AUDIT-015: SQL Injection Vector in ORDER BY Sorting

- **Severity**: MEDIUM
- **Location**: [categoryRepo.go:68](file:///home/ertval/code/zone-modules/social-network/internal/infra/storage/sqlite/categories/categoryRepo.go#L68) and [topicRepo.go:414-420](file:///home/ertval/code/zone-modules/social-network/internal/infra/storage/sqlite/topics/topicRepo.go#L414-L420)
- **Spec ref**: N/A (Security)
- **Observation**: The `order` sorting parameter (ASC/DESC) is directly concatenated into queries. Although the sorting fields are whitelisted, the sort direction isn't, allowing query string modifications.
- **Remediation**: Whitelist the sort direction to restrict it to only `"ASC"` and `"DESC"`.
- **Judge Verdict**: CONFIRMED

---

### AUDIT-016: Prepared Statement Prepared But Never Executed

- **Severity**: MEDIUM
- **Location**: [userRepo.go:70-76](file:///home/ertval/code/zone-modules/social-network/internal/infra/storage/sqlite/users/userRepo.go#L70-L76)
- **Spec ref**: N/A (Performance)
- **Observation**: In `UserRegister()`, `stmt` is prepared but never used. Instead, the query executes using raw `r.DB.ExecContext` directly.
- **Remediation**: Change execution to `stmt.ExecContext(...)`.
- **Judge Verdict**: CONFIRMED

---

### AUDIT-017: Layer Skipping (Handler calling Storage Package Definitions)

- **Severity**: MEDIUM
- **Location**: [getTopicHandler.go:14,102](file:///home/ertval/code/zone-modules/social-network/internal/infra/http/topic/getTopic/getTopicHandler.go#L14)
- **Spec ref**: `readme.md` Backend
- **Observation**: The HTTP handler imports the storage package definition `sqlite/topics` directly to compare `topics.ErrTopicNotFound`, skipping the application domain layer.
- **Remediation**: Return domain errors from the app layer.
- **Judge Verdict**: CONFIRMED

---

### AUDIT-018: Image MIME Check relies solely on Client Content-Type

- **Severity**: LOW
- **Location**: `internal/pkg/validator/validator.go:152`
- **Spec ref**: N/A (Security)
- **Observation**: The image attachment validation relies solely on the client-sent `Content-Type` header without performing server-side magic-byte inspection on the upload files.
- **Remediation**: Read the first 512 bytes of files and check via `http.DetectContentType`.
- **Judge Verdict**: JUDGE-FOUND

---

## Bonus Feature Inventory

| Feature                                  | Status     | Location                                          |
| ---------------------------------------- | ---------- | ------------------------------------------------- |
| **OAuth (GitHub / Google)**              | ✅ PRESENT | `internal/pkg/oAuth/githubclient/githubClient.go` |
| **DB seeding**                           | ✅ PRESENT | `db/seeds/dev_data.sql`                           |
| **Confirmation popups (unfollow)**       | ❌ MISSING | (Feature does not exist)                          |
| **Confirmation popups (privacy toggle)** | ❌ MISSING | (Feature does not exist)                          |
| **Docker build scripts**                 | ⚠️ PARTIAL | Builds single multi-binary image                  |
| **Extra notifications**                  | ❌ MISSING | Only implements default triggers                  |

---

## Verification Plan

### Automated Verification

Run verification checks regularly during development:

```bash
# Verify standard packages build correctly
go build ./...

# Verify standard patterns
go vet ./...

# Run linters
~/go/bin/golangci-lint run --config /home/ertval/.gemini/antigravity-ide/brain/514d0cbb-3c8e-4c74-ba64-0da156bcaee1/scratch/golangci.yml

# Check vulnerability occurrences
~/go/bin/govulncheck ./...
```

### Manual Verification Steps

Once followers, privacy scopes, and groups are implemented:

1. **Docker separation**: Build using `docker compose up -d` and run `docker ps` to verify two containers (frontend and backend) run independently.
2. **Profile & Follower flows**: Register private and public accounts. Send a follow request to a private account, accept/decline, verify notifications, and check that content visibility is restricted appropriately.
3. **Post privacy scopes**: Create public, follower-only, and private posts with attachments. Confirm visibility filters are enforced in database query listings.
4. **Group Event RSVP**: Create a group event and test voting to ensure options update in real-time.
5. **WebSocket & origin CSWSH checks**: Run cross-origin requests using tools to verify that connection upgrades are correctly restricted to verified host origins.
