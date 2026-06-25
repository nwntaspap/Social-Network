# Database Evolution Plan: Dynamic Dual-Driver Support (SQLite & PostgreSQL)

This document consolidates findings, critiques, and architectural designs from the reviews of the `refactor/PostgresMigration` branch. It proposes a unified, clean, and database-agnostic strategy to support both SQLite and PostgreSQL database engines dynamically based on configuration.

---

## 1. Consolidated Review & Critique

A deep audit of the PostgreSQL migration branch shows several critical issues that must be addressed to ensure code quality, build stability, and performance.

### A. Code Rot & Dead Code

- **SQLite Repository Orphaned**: The SQLite repositories in `internal/infra/storage/sqlite/` are fully present and compile, but are never instantiated.
- **Dead Code Bugs**: Rotted dead code introduces bugs. For example, `sqlite/oauth/oauthRepo.go:183` attempts to scan query results directly into `ctx` (`rows.Scan(ctx, ...)`), which is a runtime bug.
- **`go-sqlite3` Dependency**: The `go-sqlite3` dependency is still in `go.mod`, but the entry point ignores it entirely.

### B. Configuration Ignored & Zombie Config Fields

- **Hardcoded Driver**: The database configuration is loaded correctly, but `main.go` and `bootstrap.go` hardcode PostgreSQL initialization (`postgres.InitializeDB` / `postgres.NewRepositories`), completely ignoring `DB_DRIVER` in `.env` and `config.go`.
- **Zombie Parameters**: The SQLite-specific fields in `DatabaseConfig` (such as `Pragma_Foreign_Keys` and `Pragma_Journal_Mode`) are present but unused for PostgreSQL, and Postgres-specific parameters (such as `PG_URL`) are ignored if someone attempts to run SQLite.

### C. Connection Pool & Performance Bottlenecks

- **PostgreSQL Serialized Pool**: The default `DB_OPEN_CONN` configuration is set to `1` (which is appropriate to prevent locks in SQLite). Applying this limit to PostgreSQL throttles the connection pool size to `1`, forcing all concurrent queries to execute serially.
- **No Keep-Alives**: The Go connection pool lacks configuration for `MaxIdleConns` and `ConnMaxLifetime`, which default Go's database driver to only 2 idle connections, leading to connection thrashing under load in production.

### D. Convoluted Initialization Signatures

- **Bogus OpenDB Return**: In `postgres/init.go`, the initialization function has the signature:
  ```go
  func OpenDB(cfg config.ServerConfig) (*sql.DB, *sql.DB, error)
  ```
  The second `*sql.DB` return is always `nil`. This signature was copy-pasted from the SQLite code and must be refactored to a standard `(*sql.DB, error)`.

### E. Broken Compilation

- **Unmaintained infra package**: Running `go test ./...` fails because [services.go](../../internal/infra/services.go) is left with outdated code that references `sqlite.Repositories` but has incorrect imports.

### F. Redundant Schema Constraints & Indexes

- **Votes Table Redundancy**: In PostgreSQL migration `000008_create_votes.up.sql`, table-level `UNIQUE(user_id, topic_id)` and `UNIQUE(user_id, comment_id)` constraints are defined alongside partial unique indexes (`idx_topic_votes` and `idx_comment_votes`). These table-level constraints are redundant and prevent conflict target matching during `ON CONFLICT` updates on partial criteria.

### G. Inconsistent & Un-abstracted Error Handling

- **Inconsistent Mappers**: The `users` repository package implements `MapPQError()`, but the other 8 repositories do not.
- **Direct Driver Leak**: Packages like `categories` check `pq` errors directly via `isUniqueViolation(err)`, leaking driver-specific packages (`github.com/lib/pq`) into the repository code.
- **No Common Error Interface**: Callers receive driver-specific error codes instead of mapped domain/storage errors (e.g., `UNIQUE_VIOLATION`, `NOT_FOUND`).

### H. Repository Query Anti-Patterns

- **Single-Use `PrepareContext`**: Repositories use `PrepareContext` for almost every single-use query, causing an unnecessary network round-trip to the DB server for statement preparation. Direct `ExecContext` or `QueryContext` should be used instead.
- **Fragile Param-Position Counting**: Manual counting of Postgres positional placeholders using `paramPos++` format formatting is prone to bugs.
- **Transaction Boilerplate**: Clean deferred rollback + error wrapping is copy-pasted in multiple methods.
- **Unoptimized Category Syncing**: The `syncTopicCategories` method performs N deletes and N inserts instead of using a single `DELETE WHERE NOT IN` and batch updates.
- **Presentation Leak**: Date formatting like `Format("02/01/2006")` is performed inside the repositories. These are serialization concerns and must reside in the HTTP handler/serializer layer.

---

## 2. Proposed Architecture: Unified Registry Pattern

**The Problem**:

- The database initialization is hardcoded to PostgreSQL in `cmd/server/main.go` and `internal/bootstrap/bootstrap.go`.
- SQLite code is orphaned, and `services.go` fails to compile because it references `sqlite.Repositories`.
- There is no database-agnostic interface configuration at the storage level, requiring complete codebase modifications to change drivers.

**Best Practice Solution**:
To enable seamless driver switching at runtime, the database storage layer will be redesigned using a **Registry/Provider Factory Pattern** that dynamically selects the repository backend at boot based on `DB_DRIVER`.

```
internal/
  domain/                    # Unchanged domain interfaces
  infra/storage/
    storage.go               # Shared types & Config
    registry.go              # Storage registry & initialization factory
    common/
      errors.go              # Shared error types & mapper interfaces
      transactions.go        # Boilerplate-reduction transaction helpers
    postgres/
      postgres.go            # Postgres driver implementation
      ...                    # PostgreSQL repos (users, topics, etc.)
    sqlite/
      sqlite.go              # SQLite driver implementation
      ...                    # SQLite repos (users, topics, etc.)
```

### A. Shared Repositories Struct

Rather than duplicating the registry struct inside both driver packages, a single shared `Repositories` struct containing domain interfaces is defined:

```go
package storage

import (
	"github.com/arnald/forum/internal/domain/activity"
	"github.com/arnald/forum/internal/domain/category"
	"github.com/arnald/forum/internal/domain/chat"
	"github.com/arnald/forum/internal/domain/comment"
	"github.com/arnald/forum/internal/domain/notification"
	"github.com/arnald/forum/internal/domain/oauth"
	"github.com/arnald/forum/internal/domain/session"
	"github.com/arnald/forum/internal/domain/topic"
	"github.com/arnald/forum/internal/domain/user"
	"github.com/arnald/forum/internal/domain/vote"
)

type Repositories struct {
	UserRepo         user.Repository
	CategoryRepo     category.Repository
	TopicRepo        topic.Repository
	CommentRepo      comment.Repository
	VoteRepo         vote.Repository
	NotificationRepo notification.Repository
	OauthRepo        oauth.Repository
	ActivityRepo     activity.Repository
	ChatRepo         chat.Repository
}
```

### B. Storage Provider Interface

Each database backend implements a standard `Provider` interface:

```go
package storage

import (
	"context"
	"database/sql"
	"github.com/arnald/forum/internal/config"
	"github.com/arnald/forum/internal/domain/session"
)

type Provider interface {
	Open(ctx context.Context, cfg config.DatabaseConfig) (*sql.DB, error)
	Migrate(ctx context.Context, db *sql.DB, cfg config.DatabaseConfig) error
	Seed(ctx context.Context, db *sql.DB, env string) error
	NewRepositories(db *sql.DB) *Repositories
	NewSessionManager(db *sql.DB, sessionCfg config.SessionManagerConfig) session.Manager
}
```

### C. Driver Registry

A central registry coordinates initialization:

```go
package storage

import (
	"context"
	"database/sql"
	"fmt"
)

var providers = make(map[string]Provider)

func Register(name string, provider Provider) {
	providers[name] = provider
}

func Open(ctx context.Context, driver string, cfg config.DatabaseConfig) (*sql.DB, Provider, error) {
	p, ok := providers[driver]
	if !ok {
		return nil, nil, fmt.Errorf("unsupported database driver: %s", driver)
	}
	db, err := p.Open(ctx, cfg)
	return db, p, err
}
```

### D. Dynamic Connection Pool Tuning

Each driver tunes the SQL pool specifically in its `Open` method:

- **SQLite**: Forces `MaxOpenConns = 1` internally to prevent file-locking conflicts, applying pragmas `_foreign_keys=on&_journal_mode=WAL`.
- **Postgres**: Sets configurable `MaxOpenConns` (default 25), `MaxIdleConns` (default 5), and `ConnMaxLifetime` (default 5 minutes).

---

## 3. Migration Integration

**The Problem**:

- SQLite runs migrations inline using a custom schema-exec function, whereas PostgreSQL requires installing and running the external `golang-migrate` CLI tool manually (via `make migrate-up`).
- There is no automated PostgreSQL database migration on server startup, causing runtime errors if migrations aren't executed first via the CLI.
- This fragmented, manual CLI-based workflow creates deployment friction.

**Best Practice Solution**:

- **Eliminate the Migration CLI**: Completely avoid the requirement for external CLI migrations. Both databases will run migrations programmatically inside the Go application code.
- **Embedded Migrations**: Embed all migration files using Go's standard `//go:embed` directive. This packages SQL files with the binary, allowing zero-dependency deployments.
- **Standardized Migration Library**: Use `golang-migrate/migrate/v4` inside the Go application. Both the Postgres and SQLite providers will run schema migrations dynamically on startup if `DB_MIGRATE_ON_START` is true, ensuring consistent state across all backends.

---

## 4. Error Mapping Strategy

**The Problem**:

- Only the `users` repository maps PostgreSQL errors (`MapPQError`), while others like `categories` directly check driver-specific libraries (`github.com/lib/pq`).
- Native database error codes (like Postgres `23505` unique violation) leak into repository and HTTP layers, creating tight coupling with specific database engines.

**Best Practice Solution**:
Introduce a common error mapper interface and standardized repository errors to keep the use cases completely isolated from driver-specific types:

```go
// internal/infra/storage/common/errors.go
package common

import "errors"

var (
	ErrUniqueViolation   = errors.New("record already exists")
	ErrNotFound           = errors.New("record not found")
	ErrForeignKeyViolation = errors.New("foreign key constraint failed")
)
```

Each driver package maps its native error codes (`pq.Error` for Postgres, `sqlite3.Error` for SQLite) to these common errors before returning them to calling domains.

---

## 5. Refactoring & Code Quality Improvements

**The Problem**:

- **Boilerplate Duplication**: Transaction management (defer rollback, commit, check) is copy-pasted across multiple repository files (e.g., `CreateTopic`, `UpdateTopic`, `SendMessage`).
- **Performance Overhead**: Simple, single-use queries utilize `PrepareContext` statements unnecessarily, forcing redundant round-trips to the DB.
- **Leaking Concerns**: Date formatting like `.Format("02/01/2006")` is performed directly in read queries, leaking presentation/HTTP serialization details to the storage layer.
- **Fragile Queries**: Manual placeholder positional tracking `paramPos++` makes query building highly error-prone.
- **Sub-optimal sync operations**: Topic categories are updated using loop-based serial DELETE and INSERT statements rather than batching.

**Best Practice Solution**:

- **Transaction Helper**: Create a common `WithTransaction` utility to avoid copy-pasting defer rollback and commit blocks:
  ```go
  func WithTransaction(ctx context.Context, db *sql.DB, fn func(tx *sql.Tx) error) error {
      tx, err := db.BeginTx(ctx, nil)
      if err != nil {
          return err
      }
      defer func() {
          if p := recover(); p != nil {
              tx.Rollback()
              panic(p)
          }
      }()
      if err := fn(tx); err != nil {
          tx.Rollback()
          return err
      }
      return tx.Commit()
  }
  ```
- **Eliminate Single-Use Prepared Statements**: Replace `PrepareContext` calls with direct query/exec functions like `db.ExecContext(ctx, query, args...)` or `db.QueryRowContext(ctx, query, args...)` to reduce SQL server roundtrips.
- **Remove Date Formatting in Storage**: Repositories must parse and return standard `time.Time` values. Formatting constraints like `Format("02/01/2006")` must be delegated to JSON serializers or HTTP handler layers.

---

## 6. Detailed Implementation Checklist

> [!NOTE]
> **Branch Strategy**: The following steps are performed as an incremental refactoring directly on top of the current `refactor/PostgresMigration` branch. We do not throw away the current Postgres code; instead, we reorganize it alongside the SQLite implementation to implement the dynamic registry.

### Step 1: Compilation Fixes & Dependency Check

- [ ] Fix compiler errors in [services.go](../../internal/infra/services.go) by removing dead references and correcting imports.
- [ ] Call `scan_dependencies` to validate the safety of importing `github.com/golang-migrate/migrate/v4` and `github.com/jackc/pgx/v5` (recommended over legacy `lib/pq`).
- [ ] Update `go.mod` with the approved versions of those packages.

### Step 2: Establish the Registry Pattern

- [ ] Create `internal/infra/storage/storage.go` containing the shared `Repositories` struct and `Provider` interface.
- [ ] Create `internal/infra/storage/registry.go` implementing registration and registry-based database opening.
- [ ] Create transaction helpers in `internal/infra/storage/common/transactions.go`.
- [ ] Create shared error sentinels in `internal/infra/storage/common/errors.go`.

### Step 3: Implement SQLite Provider

- [ ] Implement the `storage.Provider` interface in `internal/infra/storage/sqlite/sqlite.go`.
- [ ] Update SQLite repos to map SQLite driver errors to the shared error types.
- [ ] Fix the OAuth repo Scan bug (`rows.Scan(ctx, ...)`).
- [ ] Clean up SQLite `init.go` and remove dead code.

### Step 4: Implement PostgreSQL Provider

- [ ] Implement the `storage.Provider` interface in `internal/infra/storage/postgres/postgres.go`.
- [ ] Refactor PostgreSQL connection pooling parameters inside `postgres.go` (`MaxOpenConns = 25`, etc.).
- [ ] Update Postgres repos to map `pq` (or `pgx`) driver errors to the shared error types.
- [ ] Fix PostgreSQL schema redundant constraints in `000008_create_votes.up.sql`.

### Step 5: Clean Up Queries & Code Quality

- [ ] Remove `PrepareContext` calls for single-use operations across both implementations.
- [ ] Move date formatting from all repositories to the HTTP serializer/handler layer.
- [ ] Optimize the `syncTopicCategories` implementation to avoid serial N DELETE/INSERT queries.

### Step 6: Bootstrap Integration

- [ ] Update `cmd/server/main.go` to initialize the database through `storage.Open` based on `cfg.Database.Driver`.
- [ ] Update `internal/bootstrap/bootstrap.go` to wire repositories and session managers dynamically using the selected provider.
- [ ] Verify local execution and migrations run seamlessly on startup.
- [ ] Run test suite to verify no regressions.

---

# Database Unification: Dynamic Dual-Driver SQLite & PostgreSQL Support

This plan implements dynamic support for both SQLite3 and PostgreSQL database backends at runtime using a Registry/Provider Factory pattern, based on the loaded configuration. It resolves compiling errors, dead code, connection pool limitations, and database query inefficiencies identified in the review.

## User Review Required

> [!IMPORTANT]
>
> - **Refactoring in Place**: We will execute this plan by refactoring the current branch in place (building on top of the existing PostgreSQL repo implementations) rather than starting from scratch from `main`.
> - **Dynamic Runtime Switch**: The database engine is determined at startup via the `DB_DRIVER` configuration variable (set to either `sqlite3` or `postgres`).
> - **New Dependency Approval**: Implementing embedded migrations on startup requires adding `github.com/golang-migrate/migrate/v4`. Additionally, we plan to swap the legacy PostgreSQL driver `github.com/lib/pq` for `github.com/jackc/pgx/v5` for improved performance and maintenance. These will be scanned for safety using `scan_dependencies` first.
> - **Redundant Constraints Removals**: In PostgreSQL migration `db/postgres/migrations/000008_create_votes.up.sql`, the redundant table-level `UNIQUE (user_id, topic_id)` and `UNIQUE (user_id, comment_id)` constraints will be removed, leaving only the partial unique indexes to manage uniqueness and enable clean upsert operations.

## Open Questions

> [!NOTE]
> There are no open questions currently. The design decisions have been aligned to standard Go clean-architecture practices.

## Proposed Changes

### 1. Storage Core Abstraction

**The Problem**:

- There is no storage abstraction layer, causing hardcoded DB references to break when swapping drivers.
- Database errors are unmapped and leak driver specifics (like `pq.Error`).
- Transaction boilerplate is copied everywhere.

**Best Practice Solution**:
Introduce a registry provider system and common package.

- #### [NEW] [storage.go](internal/infra/storage/storage.go)
  Defines the shared `Repositories` struct and the backend `Provider` interface.
- #### [NEW] [registry.go](internal/infra/storage/registry.go)
  Coordinates database provider registration and opening.
- #### [NEW] [errors.go](internal/infra/storage/common/errors.go)
  Defines shared database-agnostic error sentinels (`ErrUniqueViolation`, `ErrNotFound`, `ErrForeignKeyViolation`).
- #### [NEW] [transactions.go](internal/infra/storage/common/transactions.go)
  Provides a generic transaction helper `WithTransaction` to eliminate duplicate rollback/commit boilerplates.

---

### 2. Infrastructure Code Update

**The Problem**:

- Initializing repositories and session manager in `main.go` and `bootstrap.go` is hardcoded to Postgres.
- The file `services.go` is outdated and references SQLite, causing compilation errors.

**Best Practice Solution**:
Dynamically bind database backends during bootstrap.

- #### [MODIFY] [services.go](internal/infra/services.go)
  Fix compilation issues by referencing the new driver-agnostic `storage.Repositories` instead of `sqlite.Repositories`.
- #### [MODIFY] [bootstrap.go](internal/bootstrap/bootstrap.go)
  Update to wire repositories and session stores dynamically based on the configured database provider.
- #### [MODIFY] [main.go](cmd/server/main.go)
  Change database initialization to call the unified registry `storage.Open(...)` rather than hardcoding the PostgreSQL initialization function.

---

### 3. Database Driver Implementations

**The Problem**:

- SQLite code is rotted and contains an OAuth scan syntax bug (`rows.Scan(ctx, ...)`).
- PostgreSQL lacks connection pooling configurations (`MaxOpenConns` defaults to `1` which is slow).
- PostgreSQL migrations require manual execution of an external CLI tool.

**Best Practice Solution**:

- #### [NEW] [sqlite.go](internal/infra/storage/sqlite/sqlite.go)
  Implements `storage.Provider` for SQLite3. Sets `MaxOpenConns = 1` and runs embedded migrations.
- #### [MODIFY] [init.go](internal/infra/storage/sqlite/init.go)
  Remove dead/unused initialization code, delegating standard setup tasks to `sqlite.go`.
- #### [MODIFY] [oauthRepo.go](internal/infra/storage/sqlite/oauth/oauthRepo.go)
  Fix the syntax error where `ctx` was passed into `rows.Scan(...)`.
- #### [NEW] [postgres.go](internal/infra/storage/postgres/postgres.go)
  Implements `storage.Provider` for PostgreSQL. Configures connection pooling limits dynamically (`MaxOpenConns = 25`, etc.) and runs embedded PostgreSQL migrations, eliminating the CLI tool dependency.
- #### [MODIFY] [init.go](internal/infra/storage/postgres/init.go)
  Remove dead/unused initialization code, clean up signature of `OpenDB` to return `(*sql.DB, error)`.
- #### [MODIFY] [000008_create_votes.up.sql](db/postgres/migrations/000008_create_votes.up.sql)
  Remove redundant table-level unique constraints, leaving only the partial unique indexes.

---

### 4. Code Quality & Repo Optimizations

**The Problem**:

- `PrepareContext` is abused for single-use operations.
- Date formatting logic is leaked into storage read queries.
- Query parameter formatting uses manual tracking.
- Category syncing performs N serial DELETE/INSERT queries.

**Best Practice Solution**:

- #### [MODIFY] [topicRepo.go](internal/infra/storage/postgres/topics/topicRepo.go) (and similar files in `postgres` and `sqlite` repos)
  - Replace single-use `PrepareContext` calls with direct `QueryContext`/`ExecContext`.
  - Move presentation/date-formatting logic to HTTP handler/serializer layer.
  - Simplify the `syncTopicCategories` delete and insert queries to avoid serial loops.
  - Update repos to map native driver errors to shared error types in `storage/common/errors.go`.

---

## Verification Plan

### Automated Tests

- Run `go test ./...` to verify that code compiles cleanly and unit tests pass.
- Implement storage integration tests that execute identical query test cases against both SQLite3 (in-memory or temporary file) and PostgreSQL backends.

### Manual Verification

1. Spin up the Postgres container using `make postgres-up`.
2. Configure `.env` with `DB_DRIVER=postgres` and launch the app via `go run cmd/server/main.go`. Verify embedded migrations execute on startup and the server functions correctly.
3. Configure `.env` with `DB_DRIVER=sqlite3` and launch the app. Verify SQLite file is created, migrations apply, and the server runs successfully.
