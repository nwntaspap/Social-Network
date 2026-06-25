# Consolidated Code Issues (Excluding Spec Compliance)

_Deduplicated from 3 audit reports: bp.md, 022136.md, 022747_fs.md_

---

## CRITICAL

### 1. Migration Statement Delimiter Uses `:` Instead of `;`

- **Severity**: CRITICAL
- **Location**: `internal/infra/storage/sqlite/init.go:118`
- **Issue**: `strings.SplitSeq(string(content), ":")` splits SQL on colons, not semicolons. Future migrations with colons (timestamps, constraints) will silently break.
- **Fix**: Change `":"` to `";"`.

### 2. `ctx` Passed as Scan Destination in oauthRepo

- **Severity**: CRITICAL
- **Location**: `internal/infra/storage/sqlite/oauth/oauthRepo.go:182-183`
- **Issue**: `Scan(ctx, &oauthUser.ProviderID, ...)` — first argument is `context.Context`, not a scan target. Will panic/corrupt data on OAuth code path.
- **Fix**: Remove `ctx` from `Scan()` arguments.

---

## HIGH

### 3. Login Error Leaks User Email/Username

- **Severity**: HIGH
- **Location**: `internal/infra/http/user/login/LoginEmailHandler.go:65-68`, `internal/infra/storage/sqlite/users/userRepo.go:138`
- **Issue**: Repository errors contain the email (`"user with email X not found"`), returned verbatim to client via `err.Error()`. Enables user enumeration.
- **Fix**: Return generic `"Invalid email or password"` for all login failures.

### 4. WebSocket `CheckOrigin` Allows All Origins (CSWSH)

- **Severity**: HIGH
- **Location**: `internal/infra/http/ws/handler.go:16-20`
- **Issue**: `CheckOrigin: func(r *http.Request) bool { return true }` — any website can hijack authenticated WebSocket sessions.
- **Fix**: Validate `Origin` header against allowed origins.

### 5. Missing `_busy_timeout` in SQLite DSN

- **Severity**: HIGH
- **Location**: `internal/config/config.go:142`
- **Issue**: Default pragma is `"_foreign_keys=on&_journal_mode=WAL"` — no `_busy_timeout`. Concurrent writes return `database is locked` instead of waiting.
- **Fix**: Append `&_busy_timeout=5000` to default pragma.

### 6. No Follow-Based Auth for Chat Creation

- **Severity**: HIGH
- **Location**: `internal/app/chat/commands/initChat.go:28-33`, `internal/infra/http/chat/initChat/initChatHandler.go:58`
- **Issue**: Any user can chat with any other user; no follow-relationship check.
- **Fix**: Add follow check before creating/accessing chat — at least one direction must exist.

### 7. StateManager/RateLimiter Goroutine Leaks on Shutdown

- **Severity**: HIGH
- **Location**: `internal/pkg/oAuth/stateManager.go:43,85-98`, `internal/infra/middleware/ratelimiter/rateLimiter.go:34`
- **Issue**: `go sm.cleanup()` and `go rl.cleanup()` run `for range ticker.C` forever with no stop signal. Leaks on every server restart.
- **Fix**: Add `stop chan struct{}` and use `select` to break on close.

### 8. Missing SQLite Connection Pool Limits

- **Severity**: HIGH
- **Location**: `internal/infra/storage/sqlite/init.go:59-61`
- **Issue**: Only `SetMaxOpenConns` set. Missing `SetMaxIdleConns` and `SetConnMaxLifetime`. Idle connections accumulate; stale connections never recycle.
- **Fix**: Add `db.SetMaxIdleConns(cfg.Database.OpenConn)` and `db.SetConnMaxLifetime(5 * time.Minute)`.

### 9. Go Vet: Tests Fail to Compile

- **Severity**: HIGH
- **Location**: `internal/app/topics/commands/createTopic_test.go:87,100`, `internal/app/topics/commands/updateTopic_test.go:89,102`
- **Issue**: `NewCreateTopicHandler`/`NewUpdateTopicHandler` now require 2 args (repo + fileStorage), tests pass only 1. CI broken.
- **Fix**: Update test calls to pass `&mockFileStorage{}`.

---

## MEDIUM

### 10. ORDER BY String Concatenation Without Validation

- **Severity**: MEDIUM
- **Location**: `internal/infra/storage/sqlite/topics/topicRepo.go:414-420`, `internal/infra/storage/sqlite/categories/categoryRepo.go:68`
- **Issue**: `order` (ASC/DESC) concatenated directly into SQL without whitelist validation.
- **Fix**: Whitelist `order` to only `"ASC"` or `"DESC"`.

### 11. Prepared Statement Created But Never Used

- **Severity**: MEDIUM
- **Location**: `internal/infra/storage/sqlite/users/userRepo.go:70-76`
- **Issue**: `stmt` prepared via `PrepareContext` but `r.DB.ExecContext` used instead of `stmt.ExecContext`.
- **Fix**: Use `stmt.ExecContext(...)` or remove the `Prepare` call.

### 12. Duplicate Registration Returns 500 Not 409

- **Severity**: MEDIUM
- **Location**: `internal/infra/http/user/register/registerhandler.go:98-108`
- **Issue**: UNIQUE constraint violation returns HTTP 500 instead of 409 Conflict.
- **Fix**: Map constraint violations to 409.

### 13. Session Store Leaks Password Hash in API Response

- **Severity**: MEDIUM
- **Location**: `internal/infra/storage/sessionstore/sessionManager.go:183-226`
- **Issue**: `GetUserFromSession` returns `User` struct with `Password` field set to bcrypt hash. If any handler serializes full user, hash leaks.
- **Fix**: Zero out `User.Password = ""` before returning.

### 14. Unconditional Session Deletion on New Session

- **Severity**: MEDIUM
- **Location**: `internal/infra/storage/sessionstore/sessionManager.go:244-255`
- **Issue**: `DeleteSessionWhenNewCreated` deletes ALL other sessions unconditionally, ignoring `MaxSessionsPerUser` config.
- **Fix**: Count sessions first; delete oldest only if exceeding max.

### 15. No Goroutine Recovery in WebSocket Goroutines

- **Severity**: MEDIUM
- **Location**: `internal/infra/ws/client.go:40,67`
- **Issue**: `ReadPump`/`WritePump` goroutines have no `defer recover()`. A panic crashes the entire server.
- **Fix**: Add `recover()` with logging in both pump functions.

### 16. execSQLFile Rollback Pattern Unsafe

- **Severity**: MEDIUM
- **Location**: `internal/infra/storage/sqlite/init.go:96-136`
- **Issue**: Deferred `Rollback()` always runs after `Commit()`. If commit fails, rollback error shadows the commit error.
- **Fix**: Use `committed` boolean flag to conditionally rollback in defer.

### 17. Layer Violation: Handler Imports Storage Error

- **Severity**: MEDIUM
- **Location**: `internal/infra/http/topic/getTopic/getTopicHandler.go:14,102`
- **Issue**: Handler imports `sqlite/topics` to compare `ErrTopicNotFound`. Breaks clean architecture boundary.
- **Fix**: Export domain-level errors from `app/` layer.

### 18. Registration Handler Leaks Raw Error

- **Severity**: MEDIUM
- **Location**: `internal/infra/http/user/register/registerhandler.go:99-102`
- **Issue**: `err.Error()` returned verbatim on registration failure. May reveal SQL constraint details.
- **Fix**: Return generic error message.

### 19. Migration Uses `context.TODO()` — No Cancellable Context

- **Severity**: MEDIUM
- **Location**: `internal/infra/storage/sqlite/init.go:83`
- **Issue**: `execSQLFile` uses `context.TODO()`. Migration cannot be cancelled or timed out.
- **Fix**: Accept `context.Context` from caller.

---

## LOW

### 20. gorilla/websocket Mislabeled as Indirect Dependency

- **Severity**: LOW
- **Location**: `go.mod:6`
- **Issue**: `github.com/gorilla/websocket v1.5.3` marked `// indirect` but directly imported in multiple files.
- **Fix**: Run `go mod tidy`.

### 21. Typo: "Middlware" Instead of "Middleware"

- **Severity**: LOW
- **Location**: `internal/bootstrap/bootstrap.go:32`
- **Issue**: Field name `Middlware` is missing the 'e'.
- **Fix**: Rename to `Middleware`.

### 22. OpenDB Returns Misleading Double `*sql.DB`

- **Severity**: LOW
- **Location**: `internal/infra/storage/sqlite/init.go:53`
- **Issue**: Signature `func OpenDB(...) (*sql.DB, *sql.DB, error)`. Second return value is always `nil, nil`.
- **Fix**: Change to single `*sql.DB` return.

### 23. Image MIME Validation Uses Client Content-Type Only

- **Severity**: LOW
- **Location**: `internal/pkg/validator/validator.go:152`
- **Issue**: Relies on client-sent `Content-Type` header without server-side magic-byte inspection.
- **Fix**: Use `http.DetectContentType` on first 512 bytes.

### 24. Google OAuth Route Uses Misleading Service Name

- **Severity**: LOW
- **Location**: `internal/infra/http/server.go:197`
- **Issue**: Google OAuth passes `UserLoginGithub` service — works because `Provider` interface is generic, but name is misleading.
- **Fix**: Rename to generic `UserLogin` or add Google-specific service.
