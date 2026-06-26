# Software Design Specification (SDS)

> **Target architecture.** The current codebase uses a layered structure under `internal/domain/`, `internal/app/`, `internal/infra/`. This document describes the vertical-slice target state after all refactoring phases. See [target-architecture-with-phases.md](target-architecture-with-phases.md) for the migration plan.

This Software Design Specification (SDS) defines the technical architecture, data model, component communication patterns, and detailed interface structures for the Social Network application.

---

## 1. Data Model & Database Specification

The system uses SQLite (with Write-Ahead Logging `WAL` and busy timeout configured) as its primary storage. The database schema is managed via sequential, numbered up/down migration scripts stored in the `db/migrations/` directory.

### 1.1 Core Database Configurations

- **SQLite DSN Parameters**: All database connections must open with:
  `file:social.db?_journal_mode=WAL&_busy_timeout=5000`
- **SQL Execution**: Standard parameterized queries using `?` placeholders. String concatenation or formatting for dynamic variables is strictly prohibited to prevent SQL injection.
- **Migration Delimiter**: The custom migration runner splits commands by the literal semicolon string `";"` (never `":"`, which clashes with timestamps and strings).

### 1.2 Table Definitions (SQL Schema)

```sql
-- 000001_initial_schema.up.sql
-- Baseline tables: users (with old fields), sessions, topics (old fields), comments, votes, chats, messages, notifications, oauth_states, schema_migrations.
-- Note: schema_migrations is used by custom migration system.

-- Altered in 000002_user_profile_fields.up.sql: added date_of_birth, about_me, is_private; dropped age.
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    date_of_birth DATE NOT NULL,
    avatar_url TEXT,
    username TEXT UNIQUE,
    about_me TEXT,
    is_private BOOLEAN NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sessions (
    token TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Altered in 000003_topic_privacy.up.sql: added visibility, image_url.
CREATE TABLE IF NOT EXISTS topics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    user_id TEXT NOT NULL,
    visibility TEXT NOT NULL CHECK(visibility IN ('public', 'almost_private', 'private')),
    image_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Created in 000003_topic_privacy.up.sql
CREATE TABLE IF NOT EXISTS topic_allowed_users (
    topic_id INTEGER NOT NULL,
    user_id TEXT NOT NULL,
    PRIMARY KEY (topic_id, user_id),
    FOREIGN KEY(topic_id) REFERENCES topics(id) ON DELETE CASCADE,
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 000001_initial_schema.up.sql (comments: image_url for image/GIF attachments)
CREATE TABLE IF NOT EXISTS comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    topic_id INTEGER NOT NULL,
    user_id TEXT NOT NULL,
    content TEXT NOT NULL,
    image_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(topic_id) REFERENCES topics(id) ON DELETE CASCADE,
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 000001_initial_schema.up.sql (votes)
CREATE TABLE IF NOT EXISTS votes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL,
    topic_id INTEGER,
    comment_id INTEGER,
    reaction_type INTEGER NOT NULL CHECK(reaction_type IN (-1, 1)),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, topic_id),
    UNIQUE (user_id, comment_id),
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY(topic_id) REFERENCES topics(id) ON DELETE CASCADE,
    FOREIGN KEY(comment_id) REFERENCES comments(id) ON DELETE CASCADE
);


-- 000004_follow_system.up.sql
CREATE TABLE IF NOT EXISTS follows (
    follower_id TEXT NOT NULL,
    followee_id TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(follower_id, followee_id),
    FOREIGN KEY(follower_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY(followee_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS follow_requests (
    follower_id TEXT NOT NULL,
    followee_id TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(follower_id, followee_id),
    FOREIGN KEY(follower_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY(followee_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 000006_groups.up.sql
CREATE TABLE IF NOT EXISTS groups (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    creator_id TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(creator_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS group_members (
    group_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    role TEXT NOT NULL CHECK(role IN ('creator', 'admin', 'member')),
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(group_id, user_id),
    FOREIGN KEY(group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS group_invitations (
    group_id TEXT NOT NULL,
    inviter_id TEXT NOT NULL,
    invitee_id TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(group_id, invitee_id),
    FOREIGN KEY(group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY(inviter_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY(invitee_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS group_join_requests (
    group_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(group_id, user_id),
    FOREIGN KEY(group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS group_chat_messages (
    id TEXT PRIMARY KEY,
    group_id TEXT NOT NULL,
    sender_id TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY(sender_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS group_posts (
    id TEXT PRIMARY KEY,
    group_id TEXT NOT NULL,
    author_id TEXT NOT NULL,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    image_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY(author_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS group_post_comments (
    id TEXT PRIMARY KEY,
    post_id TEXT NOT NULL,
    author_id TEXT NOT NULL,
    content TEXT NOT NULL,
    image_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(post_id) REFERENCES group_posts(id) ON DELETE CASCADE,
    FOREIGN KEY(author_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 000007_events.up.sql
CREATE TABLE IF NOT EXISTS events (
    id TEXT PRIMARY KEY,
    group_id TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    event_time TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(group_id) REFERENCES groups(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS event_options (
    id TEXT PRIMARY KEY,
    event_id TEXT NOT NULL,
    name TEXT NOT NULL,
    FOREIGN KEY(event_id) REFERENCES events(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS event_rsvps (
    event_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    option_id TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(event_id, user_id),
    FOREIGN KEY(event_id) REFERENCES events(id) ON DELETE CASCADE,
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY(option_id) REFERENCES event_options(id) ON DELETE CASCADE
);

-- 000001_initial_schema.up.sql (infrastructure: chats, messages, notifications, oauth_states)
CREATE TABLE IF NOT EXISTS chats (
    id TEXT PRIMARY KEY,
    user_one_id TEXT NOT NULL,
    user_two_id TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_one_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY(user_two_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS messages (
    id TEXT PRIMARY KEY,
    chat_id TEXT NOT NULL,
    sender_id TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(chat_id) REFERENCES chats(id) ON DELETE CASCADE,
    FOREIGN KEY(sender_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS notifications (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    type TEXT NOT NULL, -- 'follow.requested', 'follow.accepted', 'group.invited', 'group.join_requested', 'event.created'
    source_id TEXT NOT NULL, -- references the triggering entity (user, group, event, chat)
    content TEXT NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS oauth_states (
    state TEXT PRIMARY KEY,
    provider TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

```

### 1.3 Migration Numbering (Consolidated)

| Migration | File                    | Description                                                                                                                 |
| --------- | ----------------------- | --------------------------------------------------------------------------------------------------------------------------- |
| 000001    | `initial_schema`        | Baseline: users, sessions, topics, comments, votes, chats, messages, notifications, oauth_states, schema_migrations         |
| 000002    | `user_profile_fields`   | Add date_of_birth, about_me, is_private; drop age                                                                           |
| 000003    | `topic_privacy`         | Add visibility, image_url to topics; create topic_allowed_users                                                             |
| 000004    | `follow_system`         | Create follows, follow_requests                                                                                             |
| 000005    | `migrate_notifications` | Convert old notification rows to new schema                                                                                 |
| 000006    | `groups`                | Create groups, group_members, group_invitations, group_join_requests, group_chat_messages, group_posts, group_post_comments |
| 000007    | `events`                | Create events, event_options, event_rsvps                                                                                   |
| 000008    | `migrate_chats`         | Create chats, messages; migrate legacy chat data                                                                            |
| 000009    | `seed_data`             | Optional: demo users, posts, groups, follows                                                                                |

**Gap note:** Actual `db/migrations/` currently contains only `schema.sql` and `indexes.sql`. Numbered migration files (000001–000009) are not yet created. Sprint 5 (S5-BE-91) creates 000008. See `target-architecture-with-phases.md` for the full migration plan.

**Note on SDS SQL comments above:** The inline migration-number comments in the SQL schema (e.g. `-- 000005_groups.up.sql`) are approximate references. The canonical numbering is this table.

---

## 2. Component Layout & Clean Boundaries

Each feature is standard Go code built inside `internal/<feature>/`. Features must respect the clean architecture dependency boundaries as detailed in the guidelines.

### 2.1 File Organization & Content Rules

#### 2.1.1 Entities and Store Definition (`<feature>.go`)

Defines the struct representations of the business models, the `Repository` interface that commands/queries use to talk to the database, and domain-specific errors.

```go
package follow

import "context"

type Follow struct {
    FollowerID string
    FolloweeID string
}

type FollowRequest struct {
    FollowerID string
    FolloweeID string
}

// Repository is implemented by store/sqlite.go.
// Each command/query file in commands/ and queries/ accepts this interface.
type Repository interface {
    CreateFollow(ctx context.Context, f *Follow) error
    DeleteFollow(ctx context.Context, followerID, followeeID string) error
    GetFollowers(ctx context.Context, userID string) ([]Follow, error)
    GetFollowing(ctx context.Context, userID string) ([]Follow, error)
    CreateFollowRequest(ctx context.Context, req *FollowRequest) error
    DeleteFollowRequest(ctx context.Context, followerID, followeeID string) error
    AreConnected(ctx context.Context, a, b string) (bool, error)
}
```

#### 2.1.2 Commands Layer (`commands/<use_case>.go`)

Each write use case lives in its own file inside `commands/`. Each file contains the command struct, validation, business logic, event publishing, and a handler.

```go
// follow/commands/follow_user.go
package commands

import (
    "context"
    "errors"

    "social-network/internal/follow"
)

// Cross-slice interface: defined locally, satisfied by user store
type UserPrivacyChecker interface {
    IsPrivate(ctx context.Context, userID string) (bool, error)
}

type EventBus interface {
    Publish(ctx context.Context, eventType string, payload any) error
}

type FollowUserCommand struct {
    FollowerID string
    TargetID   string
}

type FollowUserHandler struct {
    repo    follow.Repository
    privacy UserPrivacyChecker
    bus     EventBus
}

func NewFollowUserHandler(repo follow.Repository, p UserPrivacyChecker, bus EventBus) *FollowUserHandler {
    return &FollowUserHandler{repo: repo, privacy: p, bus: bus}
}

func (h *FollowUserHandler) Execute(ctx context.Context, cmd *FollowUserCommand) error {
    isPrivate, err := h.privacy.IsPrivate(ctx, cmd.TargetID)
    if err != nil {
        return err
    }

    if isPrivate {
        req := &follow.FollowRequest{FollowerID: cmd.FollowerID, FolloweeID: cmd.TargetID}
        if err := h.repo.CreateFollowRequest(ctx, req); err != nil {
            return err
        }
        return h.bus.Publish(ctx, "follow.requested", req)
    }

    f := &follow.Follow{FollowerID: cmd.FollowerID, FolloweeID: cmd.TargetID}
    if err := h.repo.CreateFollow(ctx, f); err != nil {
        return err
    }
    return h.bus.Publish(ctx, "follow.accepted", f)
}
```

#### 2.1.3 Queries Layer (`queries/<use_case>.go`)

Each read use case lives in its own file inside `queries/`. Read-only queries that extract data projections and perform access checks.

```go
// follow/queries/get_followers.go
package queries

import (
    "context"

    "social-network/internal/follow"
)

type GetFollowersQuery struct {
    UserID string
}

type GetFollowersResolver struct {
    repo follow.Repository
}

func NewGetFollowersResolver(repo follow.Repository) *GetFollowersResolver {
    return &GetFollowersResolver{repo: repo}
}

func (r *GetFollowersResolver) Execute(ctx context.Context, q *GetFollowersQuery) ([]follow.Follow, error) {
    return r.repo.GetFollowers(ctx, q.UserID)
}
```

#### 2.1.4 Transport Layer (`transport/http.go` + `transport/ws.go`)

Defines HTTP REST handlers and WebSocket handlers that delegate to commands/queries. One file per feature avoids handler fragmentation.

```go
// follow/transport/http.go
package transport

import (
    "net/http"
    "social-network/internal/follow/commands"
    "social-network/internal/follow/queries"
)

type Handler struct {
    followUser     *commands.FollowUserHandler
    unfollowUser   *commands.UnfollowUserHandler
    acceptRequest  *commands.AcceptRequestHandler
    declineRequest *commands.DeclineRequestHandler
    getFollowers   *queries.GetFollowersResolver
    getFollowing   *queries.GetFollowingResolver
}

func NewHandler(
    fu *commands.FollowUserHandler,
    uu *commands.UnfollowUserHandler,
    ar *commands.AcceptRequestHandler,
    dr *commands.DeclineRequestHandler,
    gf *queries.GetFollowersResolver,
    gf2 *queries.GetFollowingResolver,
) *Handler {
    return &Handler{followUser: fu, unfollowUser: uu, acceptRequest: ar, declineRequest: dr, getFollowers: gf, getFollowing: gf2}
}
```

`ws.go` exists only for features with WebSocket traffic (chat, group chat).

#### 2.1.5 SQLite Adapter (`store/sqlite.go`)

Implements the full `Repository` interface using standard SQL. All SQL for a feature domain lives in one file. Translates platform DB connection methods into feature actions.

```go
package store

import (
    "context"
    "social-network/internal/follow"
    "social-network/internal/platform/database"
)

type SQLiteStore struct {
    db database.DB
}

func NewSQLiteStore(db database.DB) *SQLiteStore {
    return &SQLiteStore{db: db}
}

// Used by: commands/follow_user.go
func (s *SQLiteStore) CreateFollow(ctx context.Context, f *follow.Follow) error {
    _, err := s.db.ExecContext(ctx,
        `INSERT INTO follows (follower_id, followee_id) VALUES (?, ?)`,
        f.FollowerID, f.FolloweeID)
    return err
}

// Used by: queries/are_connected.go (cross-slice FollowChecker)
func (s *SQLiteStore) AreConnected(ctx context.Context, a, b string) (bool, error) {
    query := `SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = ? AND followee_id = ?)`
    var exists bool
    err := s.db.QueryRowContext(ctx, query, a, b).Scan(&exists)
    return exists, err
}
```

Comments like `// Used by:` make the mapping from store method → command/query slice explicit without splitting files.

---

## 3. Platform Abstractions

Infrastructure frameworks are abstract concepts in the core code. They are wired to real engines inside `internal/bootstrap/bootstrap.go`.

### 3.1 Database Adapter (`internal/platform/database/`)

The database factory decouples standard features from database-specific libraries (e.g. `pgx` vs native SQLite).

```go
package database

import (
    "context"
    "database/sql"
)

type DB interface {
    QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
    QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
    ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}
```

### 3.2 In-Process / Remote Event Bus (`internal/platform/eventbus/`)

Cross-feature events are delivered via standard event buses.

```go
package eventbus

import "context"

type EventHandler func(ctx context.Context, payload []byte) error

type EventBus interface {
    Publish(ctx context.Context, eventType string, payload any) error
    Subscribe(eventType string, handler EventHandler)
}
```

- **In-Memory Channel**: Implemented using buffered Go channels and concurrent dispatch pools.
- **RabbitMQ Portability**: Plugs in with `amqp091-go`, matching the same interface methods. Feature slices remain untouched.

### 3.3 Cache Subsystem (`internal/platform/cache/`)

```go
package cache

import (
    "context"
    "time"
)

type Cache interface {
    Get(ctx context.Context, key string, dst any) error
    Set(ctx context.Context, key string, val any, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
}
```

---

## 4. Real-time Communication Specification

Real-time interactions are controlled by a WebSocket Hub located in `internal/core/realtime/`.

### 4.1 Connection Flow

1. **Authentication Handshake**: Next.js client initiates a WebSocket connection request. The gateway (via `internal/core/realtime/`) delegates token verification to `internal/core/session/` to validate the browser's auth token cookie _during the initial HTTP Upgrade request_. Unauthenticated sockets are rejected with a `401 Unauthorized` status (no anonymous sockets permitted).
2. **Origin Verification**: The `CheckOrigin` configuration on WebSocket Upgrader **must not** return `true` unconditionally. It must check the request header against configured allowed CORS origins (e.g., matching the server's `.env` origins) to mitigate Cross-Site WebSocket Hijacking.
3. **Panic Recovery**: Spawning user-driven read/write pumps introduces panic risk. All client loops must catch internal failures:
   ```go
   func (c *Client) ReadPump() {
       defer func() {
           if r := recover(); r != nil {
               log.Printf("Recovered from read pump panic: %v", r)
           }
           c.Hub.Unregister <- c
           c.Conn.Close()
       }()
       // Read loop ...
   }
   ```
4. **Deadlines & Rate Limits**:
   - `writeWait`: Time allowed to write a message to the client (e.g., 10 seconds).
   - `pongWait`: Time allowed to read the next pong message from the peer (e.g., 60 seconds).
   - `pingPeriod`: Send pings to peer with this period (e.g., 54 seconds). Must be less than `pongWait`.
   - Max message size: Limit incoming payload size (e.g., 512 KB) to prevent OOM exploits.

### 4.2 Messaging Contract

Messages are framed in JSON format:

```json
{
  "type": "chat_message",
  "payload": {
    "chat_id": "uuid-string",
    "content": "Hello world 😊"
  }
}
```

Messages support Unicode text, including emojis, transmitted as UTF-8 encoded JSON strings. No special encoding or stripping is applied to message content.

---

## 5. Security & Core Middleware

All middleware reside in `internal/core/middleware/` and compose standard HTTP routing filters.

### 5.1 Password Protection

All user passwords must be hashed using `bcrypt` (with a cost factor of at least `12`) via the `golang.org/x/crypto/bcrypt` library. Plaintext password variables must be wiped immediately from memory when no longer needed.

### 5.2 Image Upload & MIME-Type Validation

Do not rely on user-supplied HTTP `Content-Type` headers for uploads. Implement magic-byte validation using `http.DetectContentType` on the initial 512 bytes of the file content in the upload controller:

```go
package imgutil

import (
    "errors"
    "net/http"
)

var AllowedMimeTypes = map[string]bool{
    "image/jpeg": true,
    "image/png":  true,
    "image/gif":  true,
}

func ValidateImageHeader(data []byte) error {
    mime := http.DetectContentType(data)
    if !AllowedMimeTypes[mime] {
        return errors.New("invalid image format")
    }
    return nil
}
```

### 5.3 Concurrency & Memory Safety

- **Goroutine Leak Prevention**: Rate limiting modules or clean-up workers utilizing `time.Ticker` must include a `stop chan struct{}` closure channel to release system threads:

  ```go
  type RateLimiter struct {
      cleanupTicker *time.Ticker
      stop          chan struct{}
  }

  func (rl *RateLimiter) Start() {
      go func() {
          for {
              select {
              case <-rl.cleanupTicker.C:
                  rl.cleanup()
              case <-rl.stop:
                  rl.cleanupTicker.Stop()
                  return
              }
          }
      }()
  }
  ```

---

## 6. Frontend Specifications (Next.js, Tailwind CSS & shadcn/ui)

The user-facing client application is built with the Next.js App Router inside `frontend/` using HTML5, Tailwind CSS, and the `shadcn/ui` component library.

### 6.1 Component Library & Setup

We use `shadcn/ui` to implement accessible, premium, styled interactive components.

- **Configuration (`components.json`)**:
  ```json
  {
    "$schema": "https://ui.shadcn.com/schema.json",
    "style": "default",
    "rsc": true,
    "tsx": true,
    "tailwind": {
      "config": "tailwind.config.js",
      "css": "src/styles/globals.css",
      "baseColor": "slate",
      "cssVariables": true
    },
    "aliases": {
      "components": "@/components",
      "utils": "@/lib/utils"
    }
  }
  ```
- **Component File Structure**:
  - Reusable primitive components (e.g., `button.tsx`, `dialog.tsx`, `input.tsx`, `dropdown-menu.tsx`) reside in `src/components/ui/`.
  - Composite feature components (e.g., `chat-window.tsx`, `post-creator.tsx`, `follow-request-card.tsx`) reside in `src/components/features/`.

### 6.2 Unified CSS & Tailwind Configuration

Tailwind utility classes form the basis of the visual design system, augmented by semantic CSS variables defined in `src/styles/globals.css`:

```css
@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
  :root {
    --background: 222.2 84% 4.9%;
    --foreground: 210 40% 98%;
    --card: 222.2 84% 4.9%;
    --card-foreground: 210 40% 98%;
    --popover: 222.2 84% 4.9%;
    --popover-foreground: 210 40% 98%;
    --primary: 210 40% 98%;
    --primary-foreground: 222.2 47.4% 11.2%;
    --secondary: 217.2 32.6% 17.5%;
    --secondary-foreground: 210 40% 98%;
    --muted: 217.2 32.6% 17.5%;
    --muted-foreground: 215 20.2% 65.1%;
    --accent: 217.2 32.6% 17.5%;
    --accent-foreground: 210 40% 98%;
    --border: 217.2 32.6% 17.5%;
    --input: 217.2 32.6% 17.5%;
    --ring: 224.3 76.3% 48%;

    --glass-effect: backdrop-blur-md bg-slate-900/70 border border-white/10;
  }
}
```

- **Aesthetic Styling Rules**: All custom layouts leverage modern fonts like Inter or Outfit. Use smooth hover and focus transitions on all custom Tailwind designs (`transition-all duration-200 ease-in-out`). Glassmorphism cards must use the `--glass-effect` style pattern.

### 6.3 Forms and Validations

- **Registration Form Fields**:
  - `Email`: Text input with built-in regex pattern verification (`^[^\s@]+@[^\s@]+\.[^\s@]+$`).
  - `Password`: Enforced strength rules (minimum 8 characters, combination of uppercase, numbers, and symbols).
  - `First Name` & `Last Name`: Required string validations.
  - `Date of Birth`: Date picker enforcing a minimum age constraint of 13.
  - `Avatar` (optional): Upload handler incorporating magic-byte checking on client/server side to prevent non-image extensions.
  - `Nickname` & `About Me`: Optional string inputs.
- **Confirmation dialogs**: All destructive or critical user operations (e.g., deleting a post, leaving a group, unfollowing, declining a request, toggling profile privacy) must prompt the user with a `shadcn/ui` Dialog overlay for confirmation before executing.
- **Notification vs message display**: Notifications are displayed in a dedicated UI panel (bell icon with unread count), visually and structurally distinct from the Chat panel. This ensures users can differentiate new notifications (follow requests, group invites, event alerts) from new private messages at a glance.

---

## 7. Testing, Validation & Linting Framework

A multi-tiered testing and validation pipeline ensures security, correctness, and style compliance across the entire codebase.

### 7.1 Go Backend Validation Pipeline

#### 7.1.1 Static Analysis & Scoped Linting

- **Linter**: `golangci-lint` acts as the primary quality gate.
  - Configuration: Defined in `.golangci.yml` (formatting linters disabled in golangci-lint to prevent conflicts; checked by standalone tools instead).
  - Scoped execution command: `golangci-lint run --timeout=5m $(addsuffix /..., $(NEW_DIRS))` (runs as part of `make be-ci-new`).
- **Native compiler warnings**: `go vet $(NEW_PKGS)` runs in scoped CI and hooks.
- **CVE Scan**: `govulncheck $(NEW_PKGS) || true` is run during scoped CI and local builds to identify vulnerabilities without blocking PRs on stdlib issues.
- **Architecture verification**: Deterministic Go gates in `internal/gates/` (see [README](../../internal/gates/README.md)) enforce boundary rules (D5), dependency DAG (D6), branch naming (including `dev` scope support), security checks (gosec + custom AST, scoped to new code), test coverage threshold, and scope drift detection. Run via `make gates` or `go run cmd/gates/main.go --all`.
- **Pre-commit hooks**: Lefthook auto-formats staged files (gofumpt/goimports for BE, eslint + prettier for FE) on commit. Pre-push runs `go vet`, `go test -short`, `go build`, `go-arch-lint`, `tsc --noEmit`, `eslint`, `vitest`. Install: `make setup-hooks`.

#### 7.1.2 Automated Testing

- **Test suite runner**: `go test -race -v -coverprofile=coverage.out $(NEW_PKGS)` (runs via `make test-new`).
  - Runs all unit and integration tests of the new codebase under the Go race detector.
- **Database Test Assertions**: Unit tests for SQL repository layers verify SQLite WAL settings and query isolation by ensuring the application runs migrations on a fresh memory/test DB file.
- **WebSocket Verification**: Integration tests mock connection lifecycle events, WebSocket upgrades, token validations, and hub routing to prevent crashes.

---

### 7.2 Next.js Frontend Validation Pipeline

#### 7.2.1 Static Analysis & Code Style

- **Linter & Formatter**: **ESLint + Prettier** are used for linting, formatting, and import sorting.
  - Configuration: Defined in `eslint.config.mjs` and `.prettierrc` in the frontend root.
  - Linting command: `npx eslint src/`
  - Formatting command: `npx prettier --write src/`
- **Type Checking**: TypeScript compiler running without emitting output files to ensure complete type safety across API calls, state properties, and socket payloads.
  - Execution command: `tsc --noEmit`

#### 7.2.2 Automated Testing

- **Unit & Component Testing**: **Vitest** paired with **React Testing Library** handles state, hook execution, and custom DOM element checks.
  - Execution command: `bun run test`
- **End-to-End (E2E) Testing**: **Playwright** performs full-browser flow validation for cross-feature features:
  - User registration flow, cookie generation, and login.
  - Real-time chat (spawning parallel browser contexts to test bidirectional messaging).
  - Notification receipt on events (group invitations, follow requests).
  - Profile locks and access permissions.
  - Execution command: `bunx playwright test`

---

## 8. Docker & Build Automation

### 8.1 Docker Compose

Two Docker images are built via `docker-compose.yml`:

- **backend**: Go server on port `8080`, SQLite volume mounted at `./data:/app/data`
- **frontend**: Next.js app on port `3000`

### 8.2 Build Script (Bonus Feature)

A convenience script `scripts/docker-build.sh` automates image building and container startup:

```bash
#!/bin/sh
docker compose build --parallel
docker compose up -d
```
