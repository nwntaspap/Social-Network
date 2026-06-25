# Refactor: Consolidate Action-Per-Package HTTP Handlers → Entity-Per-Package

## Goal

Collapse ~30 sub-packages (`topic/createTopic`, `topic/deleteTopic`, ...) into **one package per domain entity** (`topic/`, `comment/`, `category/`, `vote/`, `notification/`, `chat/`, `user/`, `activity/`).

## Why

- **Before**: 35+ imports in `server.go`, all with aliases to disambiguate
- **After**: 8 entity imports, zero aliases
- Removes ~25 duplicate `Handler` struct / `NewHandler` definitions
- Fixes typo: `updatecommentHanlder.go` → `update.go`

## Migration Strategy

Each entity follows the same template:

```
infra/http/<entity>/
├── handler.go      # one Handler struct + NewHandler (shared)
├── create.go       # Create handler method
├── delete.go       # Delete handler method
├── get.go          # Get-by-ID handler method
├── get_all.go      # List-all handler method
├── update.go       # Update handler method
├── ...             # entity-specific files (cast.go, counts.go, stream.go, etc.)
├── models.go       # (optional) shared request/response types
└── routes.go       # RegisterRoutes(*http.ServeMux, deps...) — wires everything
```

Every action file in the new package:

- Has `package topic` (not `package createtopic`)
- Defines a method on the shared Handler: `func (h *Handler) CreateTopic(...)`
- Has no `Handler struct` or `NewHandler` — those live in `handler.go`

`routes.go` exports `RegisterRoutes(mux, appServices, config, logger, ...)` that wires all routes for the entity, applying middleware where needed.

## Step-by-Step Execution

Do **one entity at a time**, run `go vet ./internal/infra/http/...` after each.

### 1. `topic/`

**Delete directories** (5):

```
internal/infra/http/topic/createTopic/
internal/infra/http/topic/deleteTopic/
internal/infra/http/topic/getAllTopics/
internal/infra/http/topic/getTopic/
internal/infra/http/topic/updateTopic/
```

**Create files** in `internal/infra/http/topic/`:

#### `handler.go`

```go
package topic

import (
    "github.com/arnald/forum/internal/app"
    "github.com/arnald/forum/internal/config"
    "github.com/arnald/forum/internal/infra/logger"
)

type Handler struct {
    UserServices app.Services
    Config       *config.ServerConfig
    Logger       logger.Logger
}

func NewHandler(userServices app.Services, config *config.ServerConfig, logger logger.Logger) *Handler {
    return &Handler{
        UserServices: userServices,
        Config:       config,
        Logger:       logger,
    }
}
```

#### `create.go`

Copy content from `createTopicHandler.go`. Changes:

- `package createtopic` → `package topic`
- Remove `Handler struct` and `NewHandler`
- Method receiver: `(h *Handler) CreateTopic(...)`
- Import adjustments: remove redundant imports already in `handler.go` (they'll be in the file scope, Go handles this per-file)

**All topic action files** (`create.go`, `delete.go`, `get.go`, `get_all.go`, `update.go`) follow the same pattern — just paste the handler function body into `package topic` as a method on `*Handler`.

`get.go` also moves content from `topic/getTopic/errors.go`:

```go
package topic

import "errors"

var ErrTopicIDRequired = errors.New("topic id required")
```

#### `routes.go`

```go
package topic

import (
    "net/http"
    "github.com/arnald/forum/internal/infra/middleware"
)

const apiContext = "/api/v1"

func RegisterRoutes(mux *http.ServeMux, h *Handler, mw *middleware.Middleware) {
    auth := mw.Authorization.Required

    mux.HandleFunc(apiContext+"/topics/create", middlewareChain(auth, h.CreateTopic))
    mux.HandleFunc(apiContext+"/topics/update", middlewareChain(auth, h.UpdateTopic))
    mux.HandleFunc(apiContext+"/topics/delete", middlewareChain(auth, h.DeleteTopic))
    mux.HandleFunc(apiContext+"/topic",       middlewareChain(auth, h.GetTopic))
    mux.HandleFunc(apiContext+"/topics/all",   middlewareChain(auth, h.GetAllTopics))
}

func middlewareChain(next func(http.ResponseWriter, *http.Request), mws ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
    // same as current server.go middlewareChain
}
```

### 2. `comment/`

**Delete directories** (5):

```
internal/infra/http/comment/createComment/
internal/infra/http/comment/deleteComment/
internal/infra/http/comment/getComment/
internal/infra/http/comment/getCommentsByTopic/
internal/infra/http/comment/updateComment/
```

**Create files** in `internal/infra/http/comment/`:

#### `handler.go`

Two Handler struct patterns coexist for comment — use the common one:

```go
package comment

import (
    "github.com/arnald/forum/internal/app"
    "github.com/arnald/forum/internal/config"
    "github.com/arnald/forum/internal/infra/logger"
)

type Handler struct {
    UserServices app.Services
    Config       *config.ServerConfig
    Logger       logger.Logger
}

func NewHandler(userServices app.Services, config *config.ServerConfig, logger logger.Logger) *Handler {
    return &Handler{
        UserServices: userServices,
        Config:       config,
        Logger:       logger,
    }
}
```

Note: `createComment` currently uses a specific interface `commentCommands.CreateCommentRequestHandler` instead of `app.Services`. When you move it, change the method to access `h.UserServices.Commands.CreateComment` instead, OR keep the specific interface in the Handler struct. **Prefer the specific interface pattern** (it's cleaner — no `app.Services` leak). If so, handler.go would need:

```go
type Handler struct {
    CreateComment commentCommands.CreateCommentRequestHandler
    UserServices  app.Services   // for delete, get, update, getByTopic
    Config        *config.ServerConfig
    Logger        logger.Logger
}
```

**Better approach**: Keep the specific interface pattern that `createComment` already uses. The handler.go should include fields for ALL needed services:

```go
type Handler struct {
    CreateComment commentCommands.CreateCommentRequestHandler
    UserServices  app.Services
    Config        *config.ServerConfig
    Logger        logger.Logger
}
```

But this creates a split: `createComment` accesses `h.CreateComment`, others access `h.UserServices.Commands.*`. Acceptable.

**Files**:

- `create.go` — `CreateComment` method (uses `h.CreateComment`)
- `delete.go` — `DeleteComment` method (uses `h.UserServices`)
- `get.go` — `GetComment` method
- `get_by_topic.go` — `GetCommentsByTopic` method
- `update.go` — `UpdateComment` method

#### `routes.go`

```go
func RegisterRoutes(mux *http.ServeMux, h *Handler, mw *middleware.Middleware) {
    auth := mw.Authorization.Required

    mux.HandleFunc(apiContext+"/comments/create", middlewareChain(auth, h.CreateComment))
    mux.HandleFunc(apiContext+"/comments/update", middlewareChain(auth, h.UpdateComment))
    mux.HandleFunc(apiContext+"/comments/delete", middlewareChain(auth, h.DeleteComment))
    mux.HandleFunc(apiContext+"/comments/get",    h.GetComment)           // NO auth
    mux.HandleFunc(apiContext+"/comments/topic",  h.GetCommentsByTopic)   // NO auth
}
```

### 3. `category/`

**Delete directories** (5):

```
internal/infra/http/category/createCategory/
internal/infra/http/category/deleteCategory/
internal/infra/http/category/getAllCategories/
internal/infra/http/category/getCategoryByID/
internal/infra/http/category/updateCategory/
```

**Create files** in `internal/infra/http/category/`:

`handler.go` — same as topic Handler (app.Services, config, logger)

**Files**: `create.go`, `delete.go`, `get.go`, `get_all.go`, `update.go`

#### `routes.go`

```go
func RegisterRoutes(mux *http.ServeMux, h *Handler, mw *middleware.Middleware) {
    auth := mw.Authorization.Required

    mux.HandleFunc(apiContext+"/category/create",  middlewareChain(auth, h.CreateCategory))
    mux.HandleFunc(apiContext+"/category/delete",  middlewareChain(auth, h.DeleteCategory))
    mux.HandleFunc(apiContext+"/category/update",  middlewareChain(auth, h.UpdateCategory))
    mux.HandleFunc(apiContext+"/category",          middlewareChain(auth, h.GetCategoryByID))
    mux.HandleFunc(apiContext+"/categories/all",   h.GetAllCategories)    // NO auth
}
```

### 4. `vote/`

**Delete directories** (3):

```
internal/infra/http/vote/castVote/
internal/infra/http/vote/deleteVote/
internal/infra/http/vote/getVoteCounts/
```

**Create files** in `internal/infra/http/vote/`:

`handler.go`:

```go
package vote

import (
    "github.com/arnald/forum/internal/app"
    "github.com/arnald/forum/internal/config"
    "github.com/arnald/forum/internal/infra/logger"
    votecommands "github.com/arnald/forum/internal/app/votes/commands"
)

type Handler struct {
    CastVote votecommands.CastVoteRequestHandler
    Services app.Services
    Config   *config.ServerConfig
    Logger   logger.Logger
}

func NewHandler(
    castvoteHandler votecommands.CastVoteRequestHandler,
    services app.Services,
    config *config.ServerConfig,
    logger logger.Logger,
) *Handler {
    return &Handler{
        CastVote: castvoteHandler,
        Services: services,
        Config:   config,
        Logger:   logger,
    }
}
```

**Files**: `cast.go` (CastVote), `delete.go` (DeleteVote), `counts.go` (GetCounts — note: method is named `GetCounts`, not `GetVoteCounts`)

#### `routes.go`

```go
func RegisterRoutes(mux *http.ServeMux, h *Handler, mw *middleware.Middleware) {
    auth := mw.Authorization.Required

    mux.HandleFunc(apiContext+"/vote/cast",   middlewareChain(auth, h.CastVote))
    mux.HandleFunc(apiContext+"/vote/delete", middlewareChain(auth, h.DeleteVote))
    mux.HandleFunc(apiContext+"/vote/counts", middlewareChain(auth, h.GetCounts))
}
```

### 5. `notification/`

**Delete directories** (5):

```
internal/infra/http/notification/getNotifications/
internal/infra/http/notification/getUnreadCount/
internal/infra/http/notification/markAllAsRead/
internal/infra/http/notification/markAsRead/
internal/infra/http/notification/streamNotification/
```

**Create files** in `internal/infra/http/notification/`:

This entity uses the "single service interface" pattern exclusively — each handler takes one specific interface. That's fine, keep it.

`handler.go`:

```go
package notification

import (
    notificationcommands "github.com/arnald/forum/internal/app/notifications/commands"
    notificationqueries "github.com/arnald/forum/internal/app/notifications/queries"
)

type Handler struct {
    GetNotifications notificationqueries.GetNotificationsHandler
    GetUnreadCount   notificationqueries.GetUnreadCountHandler
    MarkAllAsRead    notificationcommands.MarkAllAsReadHandler
    MarkAsRead       notificationcommands.MarkAsReadHandler
    OpenStream       notificationcommands.OpenStreamHandler
}

func NewHandler(
    getNotifications notificationqueries.GetNotificationsHandler,
    getUnreadCount notificationqueries.GetUnreadCountHandler,
    markAllAsRead notificationcommands.MarkAllAsReadHandler,
    markAsRead notificationcommands.MarkAsReadHandler,
    openStream notificationcommands.OpenStreamHandler,
) *Handler {
    return &Handler{
        GetNotifications: getNotifications,
        GetUnreadCount:   getUnreadCount,
        MarkAllAsRead:    markAllAsRead,
        MarkAsRead:       markAsRead,
        OpenStream:       openStream,
    }
}
```

**Files**:

- `get.go` — `GetNotifications`
- `unread_count.go` — `GetUnread`
- `mark_all_read.go` — `MarkAllAsRead`
- `mark_read.go` — `MarkAsRead`
- `stream.go` — `StreamNotifications`

#### `routes.go`

```go
func RegisterRoutes(mux *http.ServeMux, h *Handler, mw *middleware.Middleware) {
    auth := mw.Authorization.Required

    mux.HandleFunc(apiContext+"/notifications/stream",       middlewareChain(auth, h.StreamNotifications))
    mux.HandleFunc(apiContext+"/notifications/unread-count", middlewareChain(auth, h.GetUnread))
    mux.HandleFunc(apiContext+"/notifications",              middlewareChain(auth, h.GetNotifications))
    mux.HandleFunc(apiContext+"/notifications/mark-read",    middlewareChain(auth, h.MarkAsRead))
    mux.HandleFunc(apiContext+"/notifications/mark-all-read", middlewareChain(auth, h.MarkAllAsRead))
}
```

### 6. `chat/`

**Delete directories** (2):

```
internal/infra/http/chat/getChatUsers/
internal/infra/http/chat/initChat/
```

**Create files** in `internal/infra/http/chat/`:

`handler.go`:

```go
package chat

import (
    chatcommands "github.com/arnald/forum/internal/app/chat/commands"
    chatqueries "github.com/arnald/forum/internal/app/chat/queries"
    "github.com/arnald/forum/internal/infra/logger"
)

type Handler struct {
    InitChat     chatcommands.InitChatHandler
    GetChatUsers chatqueries.GetChatUsersHandler
    Logger       logger.Logger
}

func NewHandler(
    initChat chatcommands.InitChatHandler,
    getChatUsers chatqueries.GetChatUsersHandler,
    logger logger.Logger,
) *Handler {
    return &Handler{
        InitChat:     initChat,
        GetChatUsers: getChatUsers,
        Logger:       logger,
    }
}
```

**Files**: `init.go`, `users.go`

#### `routes.go`

```go
func RegisterRoutes(mux *http.ServeMux, h *Handler, mw *middleware.Middleware) {
    auth := mw.Authorization.Required

    mux.HandleFunc(apiContext+"/chat/init",  middlewareChain(auth, h.InitChat))
    mux.HandleFunc(apiContext+"/chat/users", middlewareChain(auth, h.GetChatUsers))
}
```

### 7. `user/`

**Delete directories** (4):

```
internal/infra/http/user/getMe/
internal/infra/http/user/login/
internal/infra/http/user/logout/
internal/infra/http/user/register/
```

**Create files** in `internal/infra/http/user/`:

`handler.go`:

```go
package user

import (
    "github.com/arnald/forum/internal/app"
    "github.com/arnald/forum/internal/config"
    "github.com/arnald/forum/internal/domain/session"
    "github.com/arnald/forum/internal/infra/http/authcookies"
    "github.com/arnald/forum/internal/infra/logger"
)

type Handler struct {
    UserServices   app.Services
    SessionManager session.Manager
    CookieManager  *authcookies.Manager
    Config         *config.ServerConfig
    Logger         logger.Logger
}

func NewHandler(
    config *config.ServerConfig,
    appServices app.Services,
    sm session.Manager,
    logger logger.Logger,
    cookieManager *authcookies.Manager,
) *Handler {
    return &Handler{
        UserServices:   appServices,
        SessionManager: sm,
        CookieManager:  cookieManager,
        Config:         config,
        Logger:         logger,
    }
}
```

**Files**:

- `login.go` — both `UserLoginEmail` and `UserLoginUsername` methods (merge from `LoginEmailHandler.go` and `loginUsernameHandler.go`)
- `register.go` — `UserRegister`
- `logout.go` — `Logout`
- `me.go` — `GetMe`

#### `routes.go`

```go
func RegisterRoutes(mux *http.ServeMux, h *Handler, mw *middleware.Middleware) {
    auth := mw.Authorization.Required

    mux.HandleFunc(apiContext+"/login/email",    h.UserLoginEmail)        // NO auth
    mux.HandleFunc(apiContext+"/login/username", h.UserLoginUsername)     // NO auth
    mux.HandleFunc(apiContext+"/register",       h.UserRegister)          // NO auth
    mux.HandleFunc(apiContext+"/logout",         middlewareChain(auth, h.Logout))
    mux.HandleFunc(apiContext+"/me",             middlewareChain(auth, h.GetMe))
}
```

### 8. `activity/`

**Delete directories** (1):

```
internal/infra/http/activity/getUserActivity/
```

**Create files** in `internal/infra/http/activity/`:

`handler.go`:

```go
package activity

import (
    "github.com/arnald/forum/internal/app"
    "github.com/arnald/forum/internal/config"
    "github.com/arnald/forum/internal/infra/logger"
)

type Handler struct {
    Services app.Services
    Config   *config.ServerConfig
    Logger   logger.Logger
}

func NewHandler(services app.Services, config *config.ServerConfig, logger logger.Logger) *Handler {
    return &Handler{
        Services: services,
        Config:   config,
        Logger:   logger,
    }
}
```

**File**: `get.go` — `GetUserActivity`

#### `routes.go`

```go
func RegisterRoutes(mux *http.ServeMux, h *Handler, mw *middleware.Middleware) {
    mux.HandleFunc(apiContext+"/user/activity", middlewareChain(mw.Authorization.Required, h.GetUserActivity))
}
```

### 9. Update `server.go`

**Before** (35+ import lines + 280 lines of route registration):

```go
import (
    createtopic "..."
    deletetopic "..."
    getalltopics "..."
    gettopic "..."
    updatetopic "..."
    createcomment "..."
    // ... 30+ more
)
```

**After**:

```go
import (
    "github.com/arnald/forum/internal/infra/http/activity"
    "github.com/arnald/forum/internal/infra/http/authcookies"
    "github.com/arnald/forum/internal/infra/http/category"
    "github.com/arnald/forum/internal/infra/http/chat"
    "github.com/arnald/forum/internal/infra/http/comment"
    "github.com/arnald/forum/internal/infra/http/health"
    "github.com/arnald/forum/internal/infra/http/notification"
    "github.com/arnald/forum/internal/infra/http/oauth"
    "github.com/arnald/forum/internal/infra/http/topic"
    "github.com/arnald/forum/internal/infra/http/user"
    "github.com/arnald/forum/internal/infra/http/vote"
    wshttp "github.com/arnald/forum/internal/infra/http/ws"
)
```

Keep imports that are NOT moving into entity packages (health, oauth, ws, authcookies).

**`AddHTTPRoutes` replacement** — the 280-line method becomes:

```go
func (server *Server) AddHTTPRoutes() {
    server.router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("frontend/static"))))
    server.router.HandleFunc("/", spaHandler("frontend/static/index.html"))

    server.router.HandleFunc(apiContext+"/health",
        middlewareChain(health.NewHandler(server.logger, server.appServices.Commands.CreateNotification).HealthCheck,
            server.middleware.Authorization.Required))

    // Each entity registers its own routes
    topic.RegisterRoutes(server.router, topic.NewHandler(server.appServices, server.config, server.logger), server.middleware)
    comment.RegisterRoutes(server.router, comment.NewHandler(
        server.appServices.Commands.CreateComment,
        server.appServices,
        server.config,
        server.logger,
    ), server.middleware)
    category.RegisterRoutes(server.router, category.NewHandler(server.appServices, server.config, server.logger), server.middleware)
    vote.RegisterRoutes(server.router, vote.NewHandler(
        server.appServices.Commands.CastVote,
        server.appServices,
        server.config,
        server.logger,
    ), server.middleware)
    notification.RegisterRoutes(server.router, notification.NewHandler(
        server.appServices.Queries.GetNotifications,
        server.appServices.Queries.GetUnreadCount,
        server.appServices.Commands.MarkAllAsRead,
        server.appServices.Commands.MarkAsRead,
        server.appServices.Commands.OpenStream,
    ), server.middleware)
    chat.RegisterRoutes(server.router, chat.NewHandler(
        server.appServices.Commands.InitChat,
        server.appServices.Queries.GetChatUsers,
        server.logger,
    ), server.middleware)
    user.RegisterRoutes(server.router, user.NewHandler(
        server.config,
        server.appServices,
        server.sessionManager,
        server.logger,
        server.cookieManager,
    ), server.middleware)
    activity.RegisterRoutes(server.router, activity.NewHandler(server.appServices, server.config, server.logger), server.middleware)

    // OAuth routes stay inline (complex, reused handler construction)
    // ...
    // WS routes stay inline
    // ...
}
```

OAuth and WS routes remain inline in `server.go` because their handler construction is complex (multiple providers, StateManager, hub wiring, etc.).

**Remove** `middlewareChain` from `server.go` — each entity's `routes.go` defines its own (or define it once in a shared helper package like `infra/http/middlewareutil/`).

Better: Extract `middlewareChain` into `infra/http/middlewareutil/middleware.go` so all entity packages can import it:

```go
package middlewareutil

import "net/http"

func Chain(handler http.HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
    for i := len(middlewares) - 1; i >= 0; i-- {
        handler = middlewares[i](handler)
    }
    return handler
}
```

### 10. Cleanup

After all 8 entities are migrated:

- `go vet ./internal/infra/http/...`
- `go build ./...`
- Remove any empty parent directories (e.g., `activity/` if it only had `getUserActivity/`, the old `topic/` dir is now replaced)
- Run `go mod tidy`
- Run tests: `go test -race -count=1 ./internal/...`

## Summary of Changes

| Entity       | Packages removed | Files created        | Imports saved in server.go |
| ------------ | ---------------- | -------------------- | -------------------------- |
| topic        | 5                | 6+1 handler + routes | 4                          |
| comment      | 5                | 6+1 handler + routes | 4                          |
| category     | 5                | 6+1 handler + routes | 4                          |
| vote         | 3                | 4+1 handler + routes | 2                          |
| notification | 5                | 6+1 handler + routes | 4                          |
| chat         | 2                | 3+1 handler + routes | 1                          |
| user         | 4                | 5+1 handler + routes | 3                          |
| activity     | 1                | 2+1 handler + routes | 0                          |
| **Total**    | **30**           | **~42**              | **-27 import lines**       |

## Error Handling

- If `go vet` fails, check:
  - unused imports after removing `Handler struct` / `NewHandler`
  - package name mismatches
  - method name collisions
  - `handler.go` needs all imports that any action file uses (Go per-file imports means each file must still import its own deps — handler.go only needs its own deps)

Yes, each `.go` file in the same package still needs its own imports. The benefit is NOT eliminating per-file imports — it's eliminating per-package Handler structs, per-package NewHandler, and per-package imports in server.go.

## Verification

```bash
go vet ./internal/infra/http/...
go build ./...
go test -race -count=1 ./internal/...
```
