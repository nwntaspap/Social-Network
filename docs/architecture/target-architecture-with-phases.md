# Optimized Architecture Plan — Vertical Slices with Plug-in Infrastructure Services

## Guiding Principle

**One pattern, everywhere.** When there are two ways to do something, pick one and use it consistently. Simplicity and readability beat optimization.

---

## Current Codebase: Quantified Pain Points

| Metric                         | Actual                          | Target (Vertical Slices)                         |
| ------------------------------ | ------------------------------- | ------------------------------------------------ |
| Handler packages               | **32 directories**              | **10 (one per feature)**                         |
| Aliased imports in `server.go` | **38 aliases**                  | **0-4 (one per route group)**                    |
| CQRS command/query packages    | **27 directories**              | **20 (10 commands + 10 queries)**                |
| Total Go files in `internal/`  | **182**                         | **~140** (fewer files due to no handler subdirs) |
| Domain packages                | **10** (each in its own subdir) | **0** (merged into feature roots)                |
| Storage repo packages          | **9 subdirs**                   | **0** (merged into `store/sqlite.go`)            |
| Overhead/Logic Ratio           | **~3:1**                        | **~1.5:1**                                       |

> **Core Problem:** The current codebase has **32 handler packages** and **27 CQRS directories** for ~10 entities. The ratio of packaging overhead to business logic is roughly 3:1. This results in excessive file nesting, deeply indented import paths (e.g., `internal/infra/http/user/profile/`), and cognitive overload when navigating the codebase.

---

## System Overview

```
                       +-------------------+
                       |      Browser      |
                       |   (Next.js App)   |
                       |    Port 3000      |
                       +---------+---------+
                                 | (HTTP / WebSocket)
                                 v
                       +-------------------+
                       |    Go Backend     |
                       |    Port 8080      |
                       +---------+---------+
                          /      |      \
                         v       v       v
              +----------+  +---------+  +----------+
              |  SQLite  |  |  Redis  |  | RabbitMQ |
              |  (store) |  | (cache) |  | (events) |
              +----------+  +---------+  +----------+
```

### Components

**Frontend (Next.js)** — Serves the client-side UI on port 3000. Uses the App Router for server-side and client-side rendering. Communicates with the backend via HTTP REST for CRUD operations and WebSocket for real-time chat and notifications. Built with **shadcn/ui components + Tailwind CSS styling + ESLint + Prettier for linting/formatting**.

**Backend (Go)** — HTTP server on port 8080. Entry point for all API requests. Organized as **vertical feature slices** under `internal/<feature>/`, each encapsulating domain entities, CQRS commands/queries, HTTP transport handlers, and a SQLite store implementation. Cross-cutting concerns (auth, sessions, WebSocket hub, middleware) live in `internal/core/`. Platform abstractions (database factory, event bus, cache) live in `internal/platform/`.

**Infrastructure Services** — Pluggable behind interfaces in `internal/platform/`:

- **SQLite** (required): Primary storage with Write-Ahead Logging (`_journal_mode=WAL`) and busy timeout (`_busy_timeout=5000`).
- **PostgreSQL** (optional): Demonstrates database portability via the factory pattern — swap implementations without touching feature code.
- **Redis** (optional): In-memory cache, session store, rate limiter backend, and real-time pub/sub.
- **RabbitMQ** (optional): Async event bus for cross-feature notifications, replacing the in-process channel implementation.

The backend starts with in-memory infrastructure (channels, maps) and swaps to Redis/RabbitMQ later by reimplementing the same `platform` interfaces — zero changes to feature code.

### Feature Overview

| Feature      | Description                                                                                     | New/Migrated            |
| ------------ | ----------------------------------------------------------------------------------------------- | ----------------------- |
| User         | Registration, login, profile, privacy toggle, avatar                                            | Migrate from old layers |
| Topic        | Posts with visibility (public/almost_private/private), images                                   | Migrate from old layers |
| Comment      | Comments on posts with optional images                                                          | Migrate from old layers |
| Vote         | Upvote/downvote on posts and comments                                                           | Migrate from old layers |
| Follow       | Public follow, private follow request/accept/decline                                            | Greenfield              |
| Group        | Create, invite, request join, group chat, group posts                                           | Greenfield              |
| Event        | Create event with title/description/day-time, RSVP (going/not going)                            | Greenfield              |
| Chat         | 1-on-1 direct messaging via WebSocket, follow-gated                                             | Migrate from old layers |
| Notification | Event bus subscriber: follow-request, follow-accepted, group-invite, group-join, event-creation | Migrate from old layers |
| OAuth        | GitHub and Google third-party authentication                                                    | Migrate from old layers |

---

## Technology & Tooling (SDLC Overview)

Quick-reference for all tools across the software development lifecycle.

### Languages & Runtimes

| Layer          | Technology                                                |
| -------------- | --------------------------------------------------------- |
| Backend        | Go 1.25                                                   |
| Frontend       | TypeScript / JavaScript (Next.js App Router, Bun runtime) |
| Database       | SQL (SQLite / PostgreSQL)                                 |
| Infrastructure | Docker Compose v5.1.1                                     |

### Backend (Go)

| Phase                     | Tool                                 | Where                                |
| ------------------------- | ------------------------------------ | ------------------------------------ |
| Build                     | `go build`                           | `Dockerfile` (multi-stage)           |
| Unit/Integration tests    | `go test -race -coverprofile`        | `Makefile` `test`                    |
| Linting (aggregator)      | `golangci-lint` v2.2.1 (30+ linters) | `.golangci.yml`                      |
| Linting (static analysis) | `staticcheck`                        | `Makefile` `lint`                    |
| Linting (official)        | `go vet`                             | `.golangci.yml` + CLI                |
| Formatting                | `gofmt -s`, `gofumpt`                | `Makefile` `format`, `.golangci.yml` |
| Imports                   | `goimports`, `gci`                   | `Makefile` `format`, `.golangci.yml` |
| Module hygiene            | `go mod tidy`                        | `Makefile` `ci-mod`                  |
| Benchmarking              | `benchstat`                          | `Makefile`                           |
| Profiling                 | `go tool pprof` (+ Graphviz)         | `Makefile`                           |
| Vulnerability scan        | `govulncheck`                        | Manual / CI                          |

### Frontend (Next.js, Bun)

| Phase                | Tool              | Where                               |
| -------------------- | ----------------- | ----------------------------------- |
| Runtime              | Bun               | `package.json` (scripts)            |
| Package manager      | Bun               | `bun.lock`                          |
| Linting + formatting | ESLint + Prettier | `eslint.config.mjs` + `.prettierrc` |
| Type checking        | `tsc --noEmit`    | `package.json`                      |
| Unit/component tests | Vitest            | `vitest.config.ts`                  |
| E2E tests            | Playwright        | `playwright.config.ts`              |

### Infrastructure

| Phase                     | Tool                  | Where                  |
| ------------------------- | --------------------- | ---------------------- |
| Container build           | Docker (multi-stage)  | `Dockerfile`           |
| Orchestration             | Docker Compose v5.1.1 | `docker-compose.yml`   |
| Dev TLS certs             | `openssl`             | `scripts/makecerts.sh` |
| Task runner / CI pipeline | Makefile              | `Makefile` `ci` target |

### CI Pipeline (`make gates` / `make be-ci-new`)

Local development and PR verification run the decoupled quality gate pipeline:

```bash
make gates
```

This performs:

1. `go build ./...` — Compiles all code (legacy + new) for build safety.
2. `go run cmd/gates/main.go --all` — Runs the custom Go verification gates.
3. `make be-ci-new` — Runs new-code scoped check:
   `ci-mod → check-format-new → lint-new (staticcheck-new + golangci-lint-new + vet-new + vulncheck-new + gosec-new) → test-new`
4. `make fe-ci` — Runs frontend CI check (scoped to `frontend-next/` if it exists).

_Note: Legacy blanket checks can be run via `make ci` (runs `make be-ci` + `make fe-ci`), which checks all legacy files and is informational._

---

## Design Decisions (from Architecture Discussions)

These are the settled conclusions. Every phase below follows them.

### D1: Vertical Slices with CQRS

All features live in `internal/<feature>/` with the same internal structure. Each **use case** (one business operation) is its own file inside `commands/` (writes) or `queries/` (reads). The store and transport remain shared per-feature — they are thin layers where splitting adds no value.

```
internal/<feature>/
  <feature>.go           # Entity structs + Repository interface + domain errors
  commands/
    <use_case>.go        # One file per write operation (validation + logic + event publishing)
    <use_case>.go        # ...
  queries/
    <use_case>.go        # One file per read operation (access checks + data projection)
    <use_case>.go        # ...
  transport/
    http.go              # HTTP handlers — delegates to commands/queries
    ws.go                # WebSocket handlers (only for chat, group chat)
  store/
    sqlite.go            # SQLite implementation of Repository (all SQL in one file)
```

**Why this layout:**

- The **logic layer** is where complexity lives (privacy checks, event publishing, MIME validation). Splitting it per use case isolates each operation.
- The **store layer** is thin SQL (5–15 lines per method). One file per feature keeps all queries reviewable in one place.
- The **transport layer** is a thin HTTP adapter. One file per feature avoids handler fragmentation.

No exceptions. Every feature, new or migrated, follows this layout.

### D2: Interface Strategy — One Rule

- **Within a feature**: the `Repository` interface lives in `<feature>.go`. Commands and queries in the same package accept the full `Repository`. This is simple and readable — no fragmentation.
- **Across features**: the consumer defines a narrow local interface. The producer implements it implicitly (Go duck typing). Wired at boot in `bootstrap.go`.

```go
// WITHIN feature — full interface, same package
// follow/follow.go
type Repository interface {
    CreateFollow(ctx context.Context, f *Follow) error
    DeleteFollow(ctx context.Context, followerID, followeeID string) error
    GetFollowers(ctx context.Context, userID string) ([]Follow, error)
    GetFollowing(ctx context.Context, userID string) ([]Follow, error)
    AreConnected(ctx context.Context, a, b string) (bool, error)
    // ...
}

// follow/commands/follow_user.go
type FollowUserHandler struct { repo Repository }  // accepts the full interface — fine, same package
```

```go
// ACROSS features — narrow interface, consumer-defined
// chat/commands/send_private_msg.go
type FollowChecker interface {
    AreConnected(ctx context.Context, a, b string) (bool, error)
}

type SendPrivateMsgHandler struct {
    repo    Repository      // own repo
    follows FollowChecker   // narrow — only what chat needs from follow
}
```

```go
// bootstrap/bootstrap.go — wires concrete into narrow interface
sendMsgH := chat.NewSendPrivateMsgHandler(chatRepo, followSvc)  // followSvc satisfies chat.FollowChecker
```

### D3: Cross-Slice Communication — Three Strategies, Consistently Applied

| When                      | Strategy                   | How                                                              |
| ------------------------- | -------------------------- | ---------------------------------------------------------------- |
| **Data references**       | ID-only                    | `Comment` has `AuthorID string`, never `Author user.User`        |
| **Sync behavior checks**  | Consumer-defined interface | `chat` defines `FollowChecker`, `follow` implements it           |
| **Mutation side-effects** | Event bus publish          | `follow` publishes `follow.requested`, `notification` subscribes |

The event bus starts as an **in-process Go implementation** (channels). Later, we swap it for RabbitMQ by implementing the same interface. The feature code never changes.

### D4: Database Access — Factory Pattern from Day 1

One `DB` interface. One factory function. Start with SQLite. Add PostgreSQL later without changing any feature code.

```go
// platform/database/database.go
type DB interface {
    QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
    QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
    ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func NewDB(cfg Config) (DB, error) {
    switch cfg.Driver {
    case "sqlite":
        return newSQLite(cfg.DSN)  // _journal_mode=WAL, _busy_timeout=5000
    default:
        return nil, fmt.Errorf("unsupported driver: %s", cfg.Driver)
    }
}
```

Feature stores accept `DB`, not `*sql.DB`. When PostgreSQL support arrives, we add `case "postgres"` — one line in the factory, zero changes in features.

### D5: Boundary Rules

```
feature root (<feature>.go) + commands/ + queries/
  ├── MUST NOT import own transport/ or store/
  ├── MAY import platform/eventbus (interface only)
  └── MUST NOT import another feature's transport/ or store/

commands/<use_case>.go
  ├── Imports own feature root (entities, Repository interface)
  ├── Defines cross-slice interfaces locally (e.g. FollowChecker)
  └── MUST NOT import store/ or transport/

queries/<use_case>.go
  ├── Imports own feature root
  ├── Defines cross-slice interfaces locally
  └── MUST NOT import store/ or transport/

transport/http.go
  ├── Imports own feature root + commands/ + queries/
  ├── Imports internal/core/session/ for auth context
  └── MUST NOT import store/

store/sqlite.go
  ├── Imports own feature root (entities + repository interface)
  ├── Imports platform/database (DB interface)
  └── MUST NOT import transport/ or commands/ or queries/

bootstrap/bootstrap.go
  └── Imports everything, wires concrete implementations
```

### D6: Dependency Graph

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

`notification` is never imported by other features. It subscribes to events at boot time. This prevents circular dependencies and keeps notification as a pure side-effect consumer.

---

## Directory Tree — Final Target

```
cmd/
  server/
    main.go                              # Application entry point
  gates/
    main.go                              # CLI runner for verification gates
  client/
    main.go                              # CLI client for testing/ops

internal/
  # ─── Feature Slices (all follow D1 layout) ───

  user/                                  # Absorbs activity. Entities: User
    user.go                              # Entity: User (DateOfBirth, AboutMe, IsPrivate), Repository iface
    commands/
      register.go                        # Validate 8 fields, bcrypt hash, create user
      login.go                           # Password verify, create session
      logout.go                          # Revoke session
      update_profile.go                  # Update user fields (name, about me, avatar)
      toggle_privacy.go                  # Flip is_private flag
    queries/
      get_profile.go                     # Privacy lock check (FollowChecker), return user info
      get_activity.go                    # User posts + follower/following lists
      list_users.go                      # Browse/search users
    transport/
      http.go                            # All user HTTP handlers
    store/
      sqlite.go                          # All user SQL

  follow/                                # NEW. Entities: Follow, FollowRequest
    follow.go                            # Entity: Follow, FollowRequest, Repository iface
    commands/
      follow_user.go                     # Auto-follow (public) or create request (private) + publish event
      unfollow_user.go                   # Remove follow relationship
      accept_request.go                  # Accept follow request → insert follow + publish event
      decline_request.go                 # Decline follow request
    queries/
      get_followers.go                   # List followers of a user
      get_following.go                   # List users followed by a user
      get_pending_requests.go            # List pending follow requests
      are_connected.go                   # Shared helper — implements FollowChecker for cross-slice use
    transport/
      http.go                            # All follow HTTP handlers
    store/
      sqlite.go                          # All follow SQL

  topic/                                 # Absorbs category. Entities: Topic, Category, AllowedUser, Vote
    topic.go                             # Entity: Topic, Visibility enum, AllowedUser, Vote, Repository iface
    commands/
      create_topic.go                    # MIME validation (magic bytes), set visibility + allowed users
      cast_vote.go                       # Upvote/downvote a topic
    queries/
      get_feed.go                        # Home feed with visibility filtering (public/almost_private/private)
      get_user_topics.go                 # Topics by user (privacy-aware)
      get_topic.go                       # Single topic detail + privacy check
      get_votes.go                       # Vote counts for a topic
    transport/
      http.go                            # All topic HTTP handlers
    store/
      sqlite.go                          # All topic + vote SQL

  comment/                               # Entities: Comment
    comment.go                           # Entity: Comment, Repository iface
    commands/
      create_comment.go                  # MIME validation for image/GIF attachment
    queries/
      get_comments.go                    # Comments for a topic
    transport/
      http.go                            # All comment HTTP handlers
    store/
      sqlite.go                          # All comment SQL

  group/                                 # NEW. Entities: Group, Member, Invitation, JoinRequest, GroupPost
    group.go                             # Entity: Group, GroupMember, Invitation, JoinRequest, Repository iface
    commands/
      create_group.go                    # Create group + auto-add creator as owner
      invite_member.go                   # Invite follower + publish group.invited event
      respond_invite.go                  # Accept/decline group invitation
      request_join.go                    # Request to join group + publish group.join_requested event
      respond_join.go                    # Owner accepts/declines join request
      create_group_post.go               # Post inside group (membership check)
      send_group_message.go              # Group chat message (membership check)
    queries/
      list_groups.go                     # Browse all groups
      get_group.go                       # Group detail + membership check
      get_group_feed.go                  # Group posts (membership check)
      get_group_chat.go                  # Group chat history
    transport/
      http.go                            # Group REST handlers
      ws.go                              # Group chat WebSocket handlers
    store/
      sqlite.go                          # All group SQL

  event/                                 # NEW. Entities: Event, EventRSVP
    event.go                             # Entity: Event, RSVP, Option, Repository iface
    commands/
      create_event.go                    # Validate ≥2 options + membership check + publish event.created
      rsvp.go                            # Vote going/not going
    queries/
      list_group_events.go               # Events for a group + vote counts
    transport/
      http.go                            # All event HTTP handlers
    store/
      sqlite.go                          # All event SQL

  chat/                                  # Entities: Chat, Message
    chat.go                              # Entity: PrivateMessage, Repository iface
    commands/
      send_private_msg.go                # Follow relationship check (FollowChecker) + WS dispatch
    queries/
      get_chat_history.go                # Private message history
      list_conversations.go              # User's conversation list
    transport/
      http.go                            # Chat REST handlers
      ws.go                              # Chat WebSocket handlers
    store/
      sqlite.go                          # All chat SQL

  notification/                          # Entities: Notification
    notification.go                      # Entity: Notification, Type enum, Repository iface
    commands/
      consume_events.go                  # EventBus subscriber — handles all event types, writes notifications
      mark_read.go                       # Mark notification as read
    queries/
      list_notifications.go              # List user notifications
    transport/
      http.go                            # All notification HTTP handlers
    store/
      sqlite.go                          # All notification SQL

  oauth/                                 # Entities: OAuthState
    oauth.go                             # Entity: OAuthState, Provider enum, Repository iface
    commands/
      initiate.go                        # Generate state + redirect URL
      callback.go                        # Exchange code, upsert user, create session
    transport/
      http.go                            # OAuth HTTP handlers
    store/
      sqlite.go                          # OAuth SQL

  # ─── Cross-cutting Core ───
  core/
    session/        # Session entity + manager + store
    realtime/       # WebSocket hub, client, router
    middleware/     # Auth, CORS, rate limiter, logging
    server/         # HTTP server, route registration, graceful shutdown

  # ─── Platform (behind interfaces) ───
  platform/
    database/       # DB interface + factory (SQLite now, PostgreSQL later)
    eventbus/       # EventBus interface + in-process impl (RabbitMQ later)
    cache/          # Cache interface + in-memory map (Redis later)

  # ─── Quality Gates ───
  gates/                                # Deterministic verification gates (boundaries, DAG, security, coverage, TDD)

  # ─── Shared Utilities ───
  pkg/
    bcrypt/         # Password hashing
    uuid/           # ID generation
    validator/      # Request validation
    helpers/        # Generic utilities
    oauth/          # OAuth provider clients (github/, google/, client.go)
    imgutil/        # Magic-byte MIME validation

  config/           # App config loader
  bootstrap/        # Composition root — wires everything
```

---

## Phase 1: Critical Bug Fixes

_Fix bugs that cause crashes or security holes. No structural changes._

| #   | Bug                                    | File (current path)                                                                                             | Fix                                         |
| --- | -------------------------------------- | --------------------------------------------------------------------------------------------------------------- | ------------------------------------------- |
| 1.1 | Migration delimiter                    | `internal/infra/storage/sqlite/init.go`                                                                         | `":"` → `";"`                               |
| 1.2 | SQLite DSN missing WAL/timeout         | `internal/infra/storage/sqlite/init.go`, `.env`                                                                 | Add `_journal_mode=WAL&_busy_timeout=5000`  |
| 1.3 | OAuth `Scan()` with `ctx` arg          | `internal/infra/storage/sqlite/oauth/oauthRepo.go`                                                              | Remove `ctx` from `Scan()` params           |
| 1.4 | WebSocket `CheckOrigin` returns `true` | `internal/infra/http/ws/handler.go`                                                                             | Validate against configured origin          |
| 1.5 | SQL injection in ORDER BY              | `internal/infra/storage/sqlite/topics/topicRepo.go`, `internal/infra/storage/sqlite/categories/categoryRepo.go` | Whitelist `["ASC", "DESC"]`                 |
| 1.6 | Prepared stmt uses `db.Exec`           | `internal/infra/storage/sqlite/users/userRepo.go`                                                               | Use `stmt.ExecContext`                      |
| 1.7 | WS goroutine panic recovery            | `internal/infra/ws/client.go`                                                                                   | Add `defer recover()` to ReadPump/WritePump |
| 1.8 | RateLimiter ticker leak                | `internal/infra/middleware/ratelimiter/rateLimiter.go`                                                          | Add `stop chan struct{}`                    |

**Verify**: `go vet ./...` + `go test -race ./...`

---

## Phase 2: Platform Foundation + Migration System

_Set up the platform abstractions that all features will use. Still SQLite only, but behind interfaces._

### 2.1 Database Factory (`internal/platform/database/`)

- `database.go` — `DB` interface wrapping `*sql.DB` methods, `NewDB(cfg)` factory
- `sqlite.go` — SQLite init with WAL + busy timeout
- `migrations.go` — Sequential migration runner, `schema_migrations` tracking table

The factory returns `DB`. All feature stores will accept `DB`, not `*sql.DB`.

### 2.2 In-Process Event Bus (`internal/platform/eventbus/`)

- `eventbus.go` — `EventBus` interface: `Publish(ctx, eventType, payload)`, `Subscribe(eventType, handler)`
- `memory.go` — In-process implementation using Go channels and goroutines

~40 lines of code. No external dependencies. Features publish events; subscribers receive them asynchronously in-process.

### 2.3 Simple In-Memory Cache (`internal/platform/cache/`)

- `cache.go` — `Cache` interface: `Get(ctx, key, val)`, `Set(ctx, key, val, ttl)`, `Delete(ctx, key)`
- `memory.go` — In-process implementation using a thread-safe map (`sync.RWMutex`) and cleanup goroutine.

### 2.4 Migration Scripts (`db/migrations/`)

Migration files follow the numbered up/down format per spec requirements. The custom migration runner splits commands on `";"` and tracks applied versions in `schema_migrations`.

Create numbered migration scripts:

- `000001_initial_schema.up.sql` — Current tables (users, topics, comments, categories, votes, sessions, chats, notifications, oauth_states)
- `000001_initial_schema.down.sql` — Drop all
- `000002_user_profile_fields.up.sql` — Add `date_of_birth`, `about_me`, `is_private` to users; drop `age`
- `000002_user_profile_fields.down.sql` — Reverse
- `000003_topic_privacy.up.sql` — Add `visibility` to topics; create `topic_allowed_users`
- `000003_topic_privacy.down.sql` — Reverse
- `000004_follow_system.up.sql` — Create `follows`, `follow_requests`
- `000005_migrate_notifications.up.sql` — Convert old notifications data to new schema
- `000006_groups.up.sql` — Create `groups`, `group_members`, `group_invitations`, `group_join_requests`, `group_chat_messages`
- `000007_events.up.sql` — Create `events`, `event_rsvps`
- `000008_migrate_chats.up.sql` — Create `chats`, `messages` and migrate legacy chats
- `000009_seed_data.up.sql` — Optional seed demo data (users, posts, groups, follows) — bonus feature
- `000009_seed_data.down.sql` — Remove seed data

**Verify**: Run migrations on fresh DB. `go vet ./...`.

---

## Phase 3: Cross-Cutting Core

_Move existing cross-cutting concerns into their target locations._

### 3.1 Session (`internal/core/session/`)

- `session.go` — Session entity, `Manager` interface
- `store/sqlite.go` — SQLite session store (moved from `internal/infra/storage/sessionstore/`)

### 3.2 Realtime (`internal/core/realtime/`)

- `hub.go` — WebSocket hub (moved from `internal/infra/ws/`)
- `client.go` — Client lifecycle with `defer recover()` (bug 1.7 applied here)
- `router.go` — WS message routing by type

### 3.3 Middleware (`internal/core/middleware/`)

- `auth.go` — Session auth middleware
- `cors.go` — CORS with proper origin validation
- `ratelimiter.go` — Rate limiter (flattened from sub-package, bug 1.8 applied)
- `logging.go` — Request logging

### 3.4 Server (`internal/core/server/`)

- `server.go` — HTTP server, `ListenAndServe()`, graceful shutdown
- `routes.go` — Route registration

### 3.5 Shared Utilities

- Rename `internal/pkg/oAuth/` → `pkg/oauth/`
- Rename `internal/pkg/oAuth/githubclient/` → `pkg/oauth/github/`
- Rename `internal/pkg/oAuth/googleclient/` → `pkg/oauth/google/`
- Flatten `internal/pkg/oAuth/httpclient/` → `pkg/oauth/client.go`
- Create `pkg/imgutil/detect.go` — `http.DetectContentType` wrapper

**Verify**: `go vet ./...` + `go test -race ./...` — everything still compiles and passes.

---

## Phase 4: Greenfield Feature Slices

_Build new features that don't exist yet. Vertical slices from scratch. No migration needed._

### 4.1 Follow (`internal/follow/`)

| File                              | Contents                                                               |
| --------------------------------- | ---------------------------------------------------------------------- |
| `follow.go`                       | `Follow`, `FollowRequest` entities, `Repository` interface             |
| `commands/follow_user.go`         | Auto-follow (public) or create request (private), publish event        |
| `commands/unfollow_user.go`       | Remove follow relationship                                             |
| `commands/accept_request.go`      | Accept request → insert follow + publish `follow.accepted`             |
| `commands/decline_request.go`     | Decline request                                                        |
| `queries/get_followers.go`        | List followers of a user                                               |
| `queries/get_following.go`        | List users followed by a user                                          |
| `queries/get_pending_requests.go` | List pending follow requests for current user                          |
| `queries/are_connected.go`        | Check follow relationship (implements `FollowChecker` for cross-slice) |
| `transport/http.go`               | HTTP handlers                                                          |
| `store/sqlite.go`                 | SQLite implementation                                                  |

**Key behavior**:

- Public profile → instant follow, publish `follow.accepted` to eventbus
- Private profile → creates follow request, publish `follow.requested` to eventbus
- `notification` subscribes to both events and creates appropriate notifications

### 4.2 Group (`internal/group/`)

| File                             | Contents                                                                             |
| -------------------------------- | ------------------------------------------------------------------------------------ |
| `group.go`                       | `Group`, `GroupMember`, `Invitation`, `JoinRequest` entities, `Repository` interface |
| `commands/create_group.go`       | Create group + auto-add creator as owner                                             |
| `commands/invite_member.go`      | Any member invites a follower, publish `group.invited`                               |
| `commands/respond_invite.go`     | Invitee accepts/declines invitation                                                  |
| `commands/request_join.go`       | User requests to join group, publish `group.join_requested`                          |
| `commands/respond_join.go`       | Owner accepts/declines join request                                                  |
| `commands/create_group_post.go`  | Post inside group (membership check)                                                 |
| `commands/send_group_message.go` | Group chat message (membership check)                                                |
| `queries/list_groups.go`         | Browse all groups                                                                    |
| `queries/get_group.go`           | Group detail + membership info                                                       |
| `queries/get_group_feed.go`      | Group posts (membership check)                                                       |
| `queries/get_group_chat.go`      | Group chat history                                                                   |
| `transport/http.go`              | REST handlers                                                                        |
| `transport/ws.go`                | Group chat WebSocket handlers                                                        |
| `store/sqlite.go`                | SQLite implementation                                                                |

**Key behavior**:

- Invite/join request → publish `group.invited` / `group.join_requested` to eventbus
- Group chat uses WebSocket hub for real-time message delivery
- Any group member can invite others, only creator can accept/decline join requests

### 4.3 Event (`internal/event/`)

| File                           | Contents                                                                                      |
| ------------------------------ | --------------------------------------------------------------------------------------------- |
| `event.go`                     | `Event`, `EventRSVP`, `Option` entities, `Repository` interface                               |
| `commands/create_event.go`     | Validate ≥2 options (going/not going at minimum) + membership check + publish `event.created` |
| `commands/rsvp.go`             | User votes going/not going on an event                                                        |
| `queries/list_group_events.go` | Events for a group + options + vote counts                                                    |
| `transport/http.go`            | HTTP handlers                                                                                 |
| `store/sqlite.go`              | SQLite implementation                                                                         |

**Key behavior**:

- Event creation → publish `event.created` to eventbus (fans out to all group members)
- Requires `GroupMemberChecker` interface (defined locally, satisfied by group store)
- Event fields: title, description, day/time, at least 2 options (going, not going)

**Verify**: All new features compile, handlers respond, events publish and trigger notifications. `go test -race ./...`.

---

## Phase 5: Migrate Existing Features to Vertical Slices

_Move existing code from `internal/domain/` → `internal/app/` → `internal/infra/` into vertical slices. One feature at a time._

### Per-Feature Migration Steps

For each feature (user, topic, comment, chat, notification, oauth):

1. Create `internal/<feature>/` with D1 layout (`<feature>.go`, `commands/`, `queries/`, `transport/`, `store/`)
2. Copy entity from `internal/domain/<feature>/` → `internal/<feature>/<feature>.go`
3. Split CQRS from `internal/app/<feature>/commands/` into individual files under `commands/` (one per use case)
4. Split CQRS from `internal/app/<feature>/queries/` into individual files under `queries/` (one per use case)
5. Merge handlers from `internal/infra/http/<feature>/` → `transport/http.go`
6. Copy store from `internal/infra/storage/sqlite/<feature>/` → `store/sqlite.go`
7. Update imports
8. `go vet ./...` + `go test -race ./...`
9. Delete old directories

### Special Merge Notes

**`user/` — absorbs `activity/`**

- `internal/domain/user/user.go` + `internal/domain/activity/` → `internal/user/user.go`
- Add `DateOfBirth`, `AboutMe`, `IsPrivate` fields (drop `Age`)
- Split into: `commands/register.go`, `commands/login.go`, `commands/logout.go`, `commands/update_profile.go`, `commands/toggle_privacy.go`
- Split into: `queries/get_profile.go`, `queries/get_activity.go`, `queries/list_users.go`
- All handlers → `user/transport/http.go`

**`topic/` — absorbs `category/` and `vote/`**

- `internal/domain/topic/` + `internal/domain/category/` + `internal/domain/vote/` → `internal/topic/topic.go`
- Add `Visibility` enum (`public`, `almost_private`, `private`), `AllowedUser` entity, `Vote` entity
- Split into: `commands/create_topic.go`, `commands/cast_vote.go`
- Split into: `queries/get_feed.go`, `queries/get_user_topics.go`, `queries/get_topic.go`, `queries/get_votes.go`
- Add `http.DetectContentType` for upload MIME validation (uses `pkg/imgutil/`)

**`comment/`**

- Split into: `commands/create_comment.go` (with MIME validation for image/GIF)
- Split into: `queries/get_comments.go`

**`chat/` — gets `transport/ws.go`**

- Move WS message handlers from `internal/infra/ws/handlers/` → `chat/transport/ws.go`
- Split into: `commands/send_private_msg.go` (requires follow relationship via `FollowChecker` interface)
- Split into: `queries/get_chat_history.go`, `queries/list_conversations.go`

**`notification/` — becomes event consumer**

- Wire eventbus subscriptions in `bootstrap.go`
- Split into: `commands/consume_events.go` (subscribes to all events), `commands/mark_read.go`
- Split into: `queries/list_notifications.go`
- Subscribes to: `follow.requested`, `follow.accepted`, `group.invited`, `group.join_requested`, `event.created`
- Notifications displayed via dedicated UI panel, visually distinct from chat messages (separate icon, sidebar panel, unread count)

**`oauth/`**

- Split into: `commands/initiate.go`, `commands/callback.go`

### Cleanup

After all features migrated:

- Delete `internal/domain/`
- Delete `internal/app/`
- Delete `internal/infra/`

**Verify**: Full `go vet ./...` + `go test -race ./...` + manual smoke test of all API endpoints.

---

## Phase 6: Next.js Frontend

### 6.1 Scaffold

- Scaffold Next.js app in `frontend/` (App Router)
- Component Library: Integrate **shadcn/ui** for UI components.
- Styling: **Tailwind CSS** with custom HSL values (dark mode, glassmorphism, micro-animations).
- Structure: `src/app/` (routes), `src/components/ui/` (primitives), `src/components/features/` (composite elements), `src/styles/`.
- Code Quality: **ESLint + Prettier** for linting, formatting, and import sorting (configured via `eslint.config.mjs` and `.prettierrc`).
- Typography: Google Fonts (Inter or Outfit).

### 6.2 Core Pages

- Registration form: email, password, first name, last name, date of birth, avatar (optional), nickname (optional), about me (optional)
- Login (email/username + password, GitHub OAuth, Google OAuth)
- Home feed with posts (filtered by visibility)
- Profile page (`/profile/[id]`) with privacy lock screen for non-followers
- Post creation with visibility selector (public / almost_private / private + user picker)

### 6.3 Social Features

- Follow/unfollow buttons with confirmation popups
- Privacy toggle with confirmation popup
- Follow request notification with accept/decline
- Groups directory, group page with members, posts, events, chat
- Event RSVP (going / not going)
- Notifications panel with live SSE/WebSocket updates

### 6.4 Real-time

- WebSocket connection for chat (direct + group)
- SSE or WebSocket for notification streaming
- Emoji support in chat messages (Unicode text)
- Typing indicators, online presence

---

## Phase 7: Docker Compose (2 Services)

```yaml
services:
  backend:
    build: .
    ports: ['8080:8080']
    volumes: ['./data:/app/data'] # SQLite persistence
    environment:
      DB_DRIVER: sqlite3
      DB_PATH: /app/data/forum.db
      DB_PRAGMA: '_foreign_keys=on&_journal_mode=WAL'

  frontend:
    build: ./frontend
    ports: ['3000:3000']
    environment:
      NEXT_PUBLIC_API_URL: http://backend:8080
```

- Optional convenience script: `scripts/docker-build.sh` to build images and start containers

**This completes all spec requirements.** Everything below is optional learning.

---

## Phase 8: PostgreSQL Support (Optional — Learning)

_Goal: understand database portability and the factory pattern in practice._

### 8.1 Add PostgreSQL Driver

- Add `pgx` to `go.mod`
- Create `platform/database/postgres.go` — connection pool init
- Add `case "postgres"` to `NewDB()` factory

### 8.2 Per-Feature `store/postgres.go`

For each feature, create `store/postgres.go` implementing the same `Repository` interface with PostgreSQL-specific SQL syntax.

Since every store accepts the `DB` interface (not `*sql.DB`), and every store implements a `Repository` interface defined in the feature root, this is a parallel implementation — no feature code changes.

### 8.3 Config Switch

```env
DATABASE_DRIVER=postgres
DATABASE_DSN=postgres://user:pass@localhost:5432/social?sslmode=disable
```

### 8.4 Docker Compose Update

Add `postgres` service. Backend switches to PostgreSQL by changing `DATABASE_DRIVER`.

---

## Phase 9: Redis (Optional — Learning)

_Goal: understand caching, distributed rate limiting, and pub/sub._

### 9.1 Redis Cache (`internal/platform/cache/`)

- `redis.go` — Implements `Cache` interface using Redis connection pool
- `pubsub.go` — Pub/sub wrapper for cross-instance messaging

### 9.2 Session Cache

- Create `internal/core/session/store/redis.go` — Cache implementation utilizing `platform/cache` interface wrapper with TTL. DB remains source of truth.
- Wire in `bootstrap.go` to wrap the SQLite/Postgres session store.

### 9.3 Rate Limiter

- Update `internal/core/middleware/ratelimiter.go` to use the `platform/cache` Redis implementation (e.g. `INCR` + `EXPIRE`) instead of in-memory map.
- Enables consistent rate limiting across multiple backend instances.

### 9.4 Realtime Pub/Sub

- Update `internal/core/realtime/hub.go` to subscribe to Redis channels via the `pubsub` client.
- When a notification is created, publish to Redis. All backend instances receive it and push to their connected WebSocket clients.

### 9.5 Docker Compose Update

Add `redis` service.

---

## Phase 10: RabbitMQ (Optional — Learning)

_Goal: understand message brokers, exchanges, queues, dead-letter handling._

### 10.1 RabbitMQ Client (`internal/platform/rabbitmq/`)

- `rabbitmq.go` — Connection, channel management, auto-reconnect
- `publisher.go` — Implements `eventbus.EventBus` interface using AMQP
- `consumer.go` — Background consumer dispatching to service methods
- `exchanges.go` — Exchange/queue/binding declarations

### 10.2 Swap EventBus Implementation

In `bootstrap.go`, change one line:

```go
// Before (in-process)
bus := eventbus.NewMemoryBus()

// After (RabbitMQ)
bus := rabbitmq.NewPublisher(amqpConn)
```

Feature code (follow, group, event, notification) doesn't change — they all use the `eventbus.EventBus` interface.

### 10.3 Exchange/Queue Topology

| Exchange        | Routing Key            | Queue                  | Consumer             |
| --------------- | ---------------------- | ---------------------- | -------------------- |
| `social.events` | `follow.requested`     | `notifications.follow` | notification service |
| `social.events` | `follow.accepted`      | `notifications.follow` | notification service |
| `social.events` | `group.invited`        | `notifications.group`  | notification service |
| `social.events` | `group.join_requested` | `notifications.group`  | notification service |
| `social.events` | `event.created`        | `notifications.event`  | notification service |

Dead-letter exchange for failed messages with configurable retry.

### 10.4 Docker Compose Update

Add `rabbitmq` service. Final compose has 4-5 services.

---

## Future Scale Expansions: Microservice Promotion & Independent Scaling (Requirement)

The optimized architecture is designed to support the seamless promotion of modules into standalone **Microservices** and the physical scaling of **Commands** and **Queries** independently. By keeping slices logically segregated and minimizing cross-slice references, this transition is treated as a routing and infrastructure concern rather than a code-rewrite task.

### 1. Promotion to Microservices (Requirement)

Any feature slice (e.g., `user/`, `group/`, `notification/`) must be ready for promotion to an independent microservice:

- **No Shared Storage:** Slices access only their own database tables. Cross-slice joins are forbidden.
- **Strict Boundary Import Rules:** Slices interact only through whitelisted domain interfaces or clean APIs.
- **Transition Path:** To extract a slice (e.g., `notification`) into its own microservice:
  1. Move the `internal/notification/` directory to a new Go service repository.
  2. Implement an HTTP/gRPC transport layer for its endpoints.
  3. Replace in-memory service calls in other slices with HTTP/gRPC API client calls.

### 2. Kubernetes Readiness (App Alignment)

To ensure the application deploys reliably on Kubernetes and handles orchestrator lifecycle events:

- **Liveness & Readiness Probes:** The HTTP server exposes `/healthz` (always returns `200 OK`) and `/readyz` (dynamically verifies active SQLite/Postgres connections, Redis connectivity, and RabbitMQ health). Traffic is routed only when `/readyz` is healthy.
- **Graceful Shutdown (`SIGTERM` handling):** The Go binary traps orchestrator `SIGTERM`/`SIGINT` signals, halts the HTTP/WS listener, drains in-flight requests, gracefully closes RabbitMQ consumer channels, and releases database connection pools to prevent active connection dropouts.
- **12-Factor Config:** Configuration is populated strictly via environment variables, aligning with Kubernetes `ConfigMaps` and `Secrets` injection.

### 3. Message Broker Swappability (RabbitMQ to Kafka)

To prevent cloud-provider/vendor lock-in to RabbitMQ and allow dropping in Kafka or NATS in the future:

- **Abstracted Event Bus:** Feature commands publish events strictly via a generic `EventBus` interface defined in `internal/platform/eventbus/`.
- **Decoupled Business Logic:** Slices have zero import dependencies on RabbitMQ (`amqp`) libraries.
- **Interchangeable Implementations:** The concrete RabbitMQ client resides in `platform/rabbitmq`. If you migrate to Kafka, you only need to create `platform/kafka`, implement the same `EventBus` interface, and update the wiring in `bootstrap/bootstrap.go`.

### 4. Path to Independent CQRS Scaling

If read traffic heavily outweighs write traffic, the unified binary can be scaled asymmetrically in Kubernetes (e.g., running 10 replicas of Queries and 2 replicas of Commands) using these steps:

- **Separate Entrypoints:** Create separate binaries under `cmd/commands/main.go` and `cmd/queries/main.go` using the same underlying code but wiring only the necessary controllers.
- **Asymmetric Routing:** Configure Kubernetes Ingress, Nginx, or an API Gateway to forward read requests (`GET`) to query replicas and write requests (`POST`, `PUT`, `DELETE`) to command replicas.
- **Database Replication:** Update the `platform/database` factory to accept write and read connection strings (DSNs), supplying the primary database to commands and read-replicas to queries.

---

## Verification Checklist

### Automated Verification (run after every phase)

#### Backend (Go)

```bash
go vet ./...
go build ./...
go test -race -coverprofile=coverage.out ./...
golangci-lint run
govulncheck ./...
```

#### Frontend (Next.js)

```bash
# Lint and Format checks (ESLint + Prettier)
npx eslint src/
npx prettier --check src/

# Type Checking
tsc --noEmit

# Unit & Component Testing
bun run test # runs Vitest

# E2E Testing
bunx playwright test
```

### Boundary Verification

```bash
# No feature's transport/ or store/ imports another feature's transport/ or store/
grep -rn 'import' internal/*/transport/ internal/*/store/ | grep 'internal/' | grep -v 'platform/' | grep -v 'pkg/' | grep -v 'infra/'
```

### Manual Test Scenarios

**A: Registration & Login**

1. Register under-13 → rejected
2. Register without nickname/about → succeeds
3. Upload non-image as avatar → rejected (magic bytes)

**B: Follow & Privacy**

1. Set User B private
2. A follows B → follow request + notification to B
3. B declines → no relationship
4. A views B's profile → "Private" lock screen
5. B accepts → A sees full profile, can start chat
6. A unfollows → confirmation popup, relationship severed

**C: Post Privacy**

1. A creates "almost_private" post
   - B (follower of A) → visible
   - C (not a follower) → invisible
2. A creates "private" post selecting only B
   - B (selected user) → visible
   - D (follower of A, but not selected) → invisible

**D: Group & Event**

1. A creates group
2. A invites B → B gets notification, accepts, joins chat
3. C requests join → A gets notification, accepts
4. A creates event → all members notified
5. B RSVPs "going" → count updates
