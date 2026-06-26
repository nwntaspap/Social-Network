# Vertical Slice Architecture with CQRS for Social Network

> [!IMPORTANT]
> **Superseded.** This document was an early exploration. The final design is captured in:
>
> - [arch_optimized_v2.md](file:///home/ertval/code/zone-modules/social-network/docs/architecture/arch_optimized_v2.md) — full plan with D1 layout, directory tree, and phased execution
> - [architecture.md](file:///home/ertval/code/zone-modules/social-network/docs/architecture/architecture.md) — high-level architecture
> - [sds.md](file:///home/ertval/code/zone-modules/social-network/docs/architecture/sds.md) — detailed software design specification
>
> Key difference from this doc: the final design uses `commands/` and `queries/` **subfolders** (one file per use case) while keeping `store/sqlite.go` as a single shared file per feature. This doc proposed flat per-use-case store files which was rejected as unnecessary granularity.

This document details the transition from a module-centric package structure to a **True Vertical Slice Architecture** combined with **CQRS (Command Query Responsibility Segregation)**. It addresses the requirements from [audit.md](file:///home/ertval/code/zone-modules/social-network/docs/requirements/audit.md) and [readme.md](file:///home/ertval/code/zone-modules/social-network/docs/requirements/readme.md).

---

## 1. Modules vs. True Vertical Slices

In `arch_optimized_v2.md`, features were grouped into directories (e.g., `internal/follow/`) containing large files representing layers: `commands.go`, `queries.go`, and `store/sqlite.go`. While localized, these are still **technical layers (modules)** rather than independent vertical slices. If the follow feature grows, `commands.go` becomes a dump for all write logic, violating the Single Responsibility Principle.

A **True Vertical Slice** encapsulates everything needed to fulfill a single business request (use case)—from the HTTP/WebSocket transport layer, down through validation and business rules, to database operations.

```
                               VERTICAL SLICES (Use Cases)
                         ┌───────────────┬───────────────┐
                         │ Register User │  Get Profile  │
                         ├───────────────┼───────────────┤
  Transport (HTTP/WS)    │  HTTP Route   │  HTTP Route   │
                         ├───────────────┼───────────────┤
  Logic (Command/Query)  │ bcrypt, Val.  │ Privacy check │
                         ├───────────────┼───────────────┤
  Persistence (SQL)      │  INSERT User  │  SELECT User  │
                         └───────────────┴───────────────┘
```

To prevent package dependency hell in Go, we group slices into **feature domains** (`internal/<domain>/`) as flat packages, but slice the code into **dedicated files per use case**.

---

## 2. Standard Vertical Slice File Layout

Within each domain package (e.g., `internal/post`), each use case is defined by a pair of files: one for the **Request/Handler/Logic** and one for the **Repository/SQL Store implementation**.

### For Commands (Write Side)

- `<use_case>_command.go`: Defines the command struct, request DTO, validation, event publishing, and handler.
- `<use_case>_store.go`: Contains the database write implementation.

### For Queries (Read Side)

- `<use_case>_query.go`: Defines the query parameters, DTO, response mapping, and resolver.
- `<use_case>_store.go`: Contains database query execution (optimized for reading, bypassing domain constraints if necessary).

---

## 3. Directory Layout (The Architecture)

```
internal/
  auth/
    # Shared entity/session validation
    auth.go                      # Domain Entity: Session, User (subset)

    # Slice: Register User
    register_command.go          # RegisterCommand, Validation, Handler, HTTP Endpoint
    register_store.go            # SQLite: INSERT user, verify uniqueness

    # Slice: Log In
    login_command.go             # LoginCommand, Password verify, Generate Session Cookie
    login_store.go               # SQLite: SELECT user by email/username, INSERT session

    # Slice: Log Out
    logout_command.go            # LogoutCommand, Revoke Session Cookie
    logout_store.go              # SQLite: DELETE session

  user/
    user.go                      # Domain Entity: User Profile, PrivacyState

    # Slice: Get Profile
    get_profile_query.go         # GetProfileQuery, Privacy Lock checking, Handler
    get_profile_store.go         # SQLite: SELECT user details, follow counts

    # Slice: Update Profile Privacy
    update_privacy_command.go    # UpdatePrivacyCommand, Handler (triggers confirmation)
    update_privacy_store.go      # SQLite: UPDATE users SET is_private = ?

  follow/
    follow.go                    # Domain Entity: Follow, FollowRequest

    # Slice: Follow User (Auto-follow / Request flow)
    follow_command.go            # FollowUserCommand (decides auto-follow vs request, publishes events)
    follow_store.go              # SQLite: INSERT follows OR INSERT follow_requests

    # Slice: Unfollow User
    unfollow_command.go          # UnfollowUserCommand (triggers confirmation popup)
    unfollow_store.go            # SQLite: DELETE follows

    # Slice: Accept Request
    accept_request_command.go    # AcceptFollowCommand (updates status, publishes event)
    accept_request_store.go      # SQLite: DELETE follow_requests, INSERT follows

    # Slice: Decline Request
    decline_request_command.go    # DeclineFollowCommand
    decline_request_store.go      # SQLite: DELETE follow_requests

    # Slice: Get Relationship Lists
    get_followers_query.go       # GetFollowersQuery, Handler
    get_followers_store.go       # SQLite: SELECT followers
    get_following_query.go       # GetFollowingQuery, Handler
    get_following_store.go       # SQLite: SELECT following
    get_pending_requests_query.go # GetPendingRequestsQuery, Handler
    get_pending_requests_store.go # SQLite: SELECT pending follow requests

  post/
    post.go                      # Domain Entity: Post, PrivacyEnum, AllowedUser

    # Slice: Create Post
    create_post_command.go       # CreatePostCommand, Image/GIF magic byte verification, Handler
    create_post_store.go         # SQLite: INSERT posts, INSERT topic_allowed_users

    # Slice: Get Home Feed
    get_feed_query.go            # GetFeedQuery (applies visibility filters)
    get_feed_store.go            # SQLite: SELECT feed posts with complex permissions check

    # Slice: Get User Posts
    get_user_posts_query.go      # GetUserPostsQuery (checks if viewer is allowed based on privacy)
    get_user_posts_store.go      # SQLite: SELECT user posts

  comment/
    comment.go                   # Domain Entity: Comment

    # Slice: Create Comment
    create_comment_command.go    # CreateCommentCommand, MIME validation, Handler
    create_comment_store.go      # SQLite: INSERT comments

    # Slice: Get Post Comments
    get_comments_query.go        # GetCommentsQuery, Handler
    get_comments_store.go        # SQLite: SELECT comments for post

  group/
    group.go                     # Domain Entity: Group, Member, Invitation, JoinRequest

    # Slice: Create Group
    create_group_command.go      # CreateGroupCommand, Handler
    create_group_store.go        # SQLite: INSERT groups, INSERT group_members (role=owner)

    # Slice: Browse Groups
    list_groups_query.go         # ListGroupsQuery, Handler
    list_groups_store.go         # SQLite: SELECT all groups

    # Slice: Get Group Detail
    get_group_query.go           # GetGroupQuery, Member checking, Handler
    get_group_store.go           # SQLite: SELECT group details, members, invites

    # Slice: Request Join
    request_join_command.go      # RequestJoinCommand, Publishes group.join_requested
    request_join_store.go        # SQLite: INSERT group_join_requests

    # Slice: Respond to Join Request
    respond_join_command.go      # RespondToJoinCommand (accept/decline by owner)
    respond_join_store.go        # SQLite: UPDATE/DELETE join requests, INSERT group_members

    # Slice: Invite Member
    invite_member_command.go     # InviteMemberCommand, Publishes group.invited
    invite_member_store.go       # SQLite: INSERT group_invitations

    # Slice: Respond to Invitation
    respond_invite_command.go    # RespondToInviteCommand (accept/decline by invitee)
    respond_invite_store.go      # SQLite: UPDATE/DELETE group_invitations, INSERT group_members

    # Slice: Post inside Group
    create_group_post_command.go # CreateGroupPostCommand (requires membership check)
    create_group_post_store.go   # SQLite: INSERT group posts
    get_group_feed_query.go      # GetGroupFeedQuery (requires membership check)
    get_group_feed_store.go      # SQLite: SELECT posts in group

  event/
    event.go                     # Domain Entity: Event, RSVP

    # Slice: Create Event
    create_event_command.go      # CreateEventCommand (validates 2+ options, publishes event.created)
    create_event_store.go        # SQLite: INSERT events, INSERT options

    # Slice: RSVP Vote
    rsvp_event_command.go        # RsvpEventCommand (updates vote selection)
    rsvp_event_store.go          # SQLite: INSERT/UPDATE event_rsvps

    # Slice: Get Events
    list_group_events_query.go   # ListGroupEventsQuery, Handler
    list_group_events_store.go   # SQLite: SELECT events + options + vote counts

  chat/
    chat.go                      # Domain Entity: PrivateMessage, GroupMessage

    # Slice: Get Chat History
    get_chat_history_query.go    # GetChatHistoryQuery (private/group history checks)
    get_chat_history_store.go    # SQLite: SELECT private or group messages (ordered)

    # Slice: Send Private Message
    send_private_msg_command.go  # SendMessageCommand, WS connection check, follow relationship check
    send_private_msg_store.go    # SQLite: INSERT private_messages

    # Slice: Send Group Message
    send_group_msg_command.go    # SendGroupMessageCommand, Group membership check
    send_group_msg_store.go      # SQLite: INSERT group_messages

  notification/
    notification.go              # Domain Entity: Notification

    # Slice: Consume Event & Notify
    consume_events.go            # Listens to eventbus, writes notification, broadcasts via WS/SSE
    consume_events_store.go      # SQLite: INSERT notifications

    # Slice: List Notifications
    list_notifications_query.go  # ListNotificationsQuery, Handler
    list_notifications_store.go  # SQLite: SELECT user notifications

    # Slice: Mark Read
    mark_read_command.go         # MarkAsReadCommand, Handler
    mark_read_store.go           # SQLite: UPDATE notifications SET is_read = 1
```

---

## 4. CQRS Execution Flow: A Concrete Code Blueprint

Let's illustrate how a single slice—**Create Post**—is structured to show how the logic, validation, HTTP transport, and database operations are encapsulated in a single pair of files.

### 4.1 Command Logic & HTTP Transport: `create_post_command.go`

```go
package post

import (
	"context"
	"net/http"

	"social-network/internal/pkg/imgutil"
	"social-network/internal/platform/database"
	"social-network/internal/platform/eventbus"
)

// DTO representing the incoming HTTP payload
type CreatePostRequest struct {
	Content         string   `json:"content"`
	ImageBase64     string   `json:"image_base64,omitempty"` // optional image
	Visibility      string   `json:"visibility"`             // public, almost_private, private
	AllowedFollowers []string `json:"allowed_followers,omitempty"` // selected followers if private
}

// Command object passed internally to the Handler
type CreatePostCommand struct {
	AuthorID         string
	Content          string
	ImageData        []byte
	MimeType         string
	Visibility       string // "public", "almost_private", "private"
	AllowedFollowers []string
}

// Interface for dependencies needed solely by this command
type CreatePostStore interface {
	InsertPost(ctx context.Context, cmd *CreatePostCommand) (string, error)
}

type FollowChecker interface {
	AreFollowers(ctx context.Context, userID string, checkIDs []string) (bool, error)
}

type CommandHandler struct {
	store   CreatePostStore
	follows FollowChecker
	bus     eventbus.EventBus
}

func NewCommandHandler(s CreatePostStore, f FollowChecker, b eventbus.EventBus) *CommandHandler {
	return &CommandHandler{store: s, follows: f, bus: b}
}

// Execute performs validation, business logic checks, database mutation, and event publishing
func (h *CommandHandler) Execute(ctx context.Context, cmd *CreatePostCommand) (string, error) {
	// 1. Business Validation: Ensure image is correct type if present
	if len(cmd.ImageData) > 0 {
		mime, err := imgutil.DetectMimeType(cmd.ImageData)
		if err != nil || !imgutil.IsAllowedMime(mime) {
			return "", ErrInvalidImageMime
		}
		cmd.MimeType = mime
	}

	// 2. Business Validation: Verify that selected private viewers are actually followers
	if cmd.Visibility == "private" {
		ok, err := h.follows.AreFollowers(ctx, cmd.AuthorID, cmd.AllowedFollowers)
		if err != nil || !ok {
			return "", ErrInvalidPrivateViewers
		}
	}

	// 3. Database operation
	postID, err := h.store.InsertPost(ctx, cmd)
	if err != nil {
		return "", err
	}

	// 4. Publish event for side effects (e.g. indexing or group notifications if applicable)
	_ = h.bus.Publish(ctx, "post.created", map[string]any{
		"post_id":   postID,
		"author_id": cmd.AuthorID,
	})

	return postID, nil
}

// HTTP Handler wired at the transport layer
func MakeHTTPHandler(h *CommandHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get requester user ID from session context
		userID := r.Context().Value("user_id").(string)

		var req CreatePostRequest
		if err := decodeJSON(r, &req); err != nil {
			respondError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Decode image if present
		var imgBytes []byte
		if req.ImageBase64 != "" {
			var err error
			imgBytes, err = decodeBase64(req.ImageBase64)
			if err != nil {
				respondError(w, http.StatusBadRequest, "Malformed image data")
				return
			}
		}

		cmd := &CreatePostCommand{
			AuthorID:         userID,
			Content:          req.Content,
			ImageData:        imgBytes,
			Visibility:       req.Visibility,
			AllowedFollowers: req.AllowedFollowers,
		}

		postID, err := h.Execute(r.Context(), cmd)
		if err != nil {
			respondError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}

		respondJSON(w, http.StatusCreated, map[string]string{"post_id": postID})
	}
}
```

### 4.2 SQLite Store Implementation: `create_post_store.go`

```go
package post

import (
	"context"
	"database/sql"
	"fmt"

	"social-network/internal/pkg/uuid"
	"social-network/internal/platform/database"
)

type SQLiteStore struct {
	db database.DB
}

func NewSQLiteStore(db database.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

// InsertPost implements the CreatePostStore interface for SQLite
func (s *SQLiteStore) InsertPost(ctx context.Context, cmd *CreatePostCommand) (string, error) {
	postID := uuid.New()

	// Transactions are scoped within the store logic of the vertical slice
	err := database.WithTransaction(ctx, s.db, func(tx database.Tx) error {
		// 1. Insert post entity
		query := `
			INSERT INTO posts (id, author_id, content, image_path, mime_type, visibility, created_at)
			VALUES (?, ?, ?, ?, ?, ?, datetime('now'))
		`
		imagePath := ""
		if len(cmd.ImageData) > 0 {
			imagePath = fmt.Sprintf("/uploads/posts/%s", postID)
			// Write file to storage... (omitted for brevity)
		}

		_, err := tx.ExecContext(ctx, query, postID, cmd.AuthorID, cmd.Content, imagePath, cmd.MimeType, cmd.Visibility)
		if err != nil {
			return err
		}

		// 2. If private, insert allowed users
		if cmd.Visibility == "private" && len(cmd.AllowedFollowers) > 0 {
			stmt, err := tx.PrepareContext(ctx, "INSERT INTO post_allowed_users (post_id, user_id) VALUES (?, ?)")
			if err != nil {
				return err
			}
			defer stmt.Close()

			for _, userID := range cmd.AllowedFollowers {
				_, err = stmt.ExecContext(ctx, postID, userID)
				if err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return postID, nil
}
```

---

## 5. Cross-Slice Communication

To ensure vertical slices remain independent and do not import each other's storage/handler concerns (which results in circular dependencies), we utilize Go's **implicit interface satisfaction (duck typing)** and the **in-process Event Bus**.

### Rule 1: No Direct Domain-to-Domain Imports for State Mutation

If domain `follow` needs to notify domain `notification` that a request has been made:

1. `follow` publishes an event on the event bus: `follow.requested`.
2. `notification` registers a subscriber at boot time for `follow.requested`.
3. The notification slice consumes the event and inserts a notification row in the database.

This keeps `follow` completely decoupled from `notification`.

### Rule 2: Local Segregated Interfaces for Queries

If domain `chat` needs to check if two users follow each other before initiating a connection:

1. `chat` defines a local consumer interface inside `internal/chat/initiate_chat_query.go`:
   ```go
   type FollowChecker interface {
       AreConnected(ctx context.Context, userA, userB string) (bool, error)
   }
   ```
2. The `follow` package implements `AreConnected` on its own service/repository.
3. During compilation wiring (`bootstrap.go`), we pass the concrete implementation of the follow service to the chat query resolver.
4. `chat` never imports `internal/follow/store` or the `follow` DB structures directly; it only relies on its locally defined interface.

---

## 6. Composition Root (`internal/bootstrap/bootstrap.go`)

With vertical slices, the bootstrap code acts as the **wiring loom**. It instantiates the DB connection, creates the shared event bus, configures slices, and sets up routing.

```go
package bootstrap

import (
	"net/http"

	"social-network/internal/auth"
	"social-network/internal/follow"
	"social-network/internal/post"
	"social-network/internal/platform/database"
	"social-network/internal/platform/eventbus"
)

func WireApp(db database.DB, bus eventbus.EventBus, mux *http.ServeMux) {
	// 1. Initialize Stores
	authStore := auth.NewSQLiteStore(db)
	followStore := follow.NewSQLiteStore(db)
	postStore := post.NewSQLiteStore(db)

	// 2. Initialize Logic Services (satisfying cross-slice local interfaces)
	followSvc := follow.NewService(followStore, bus) // implements post.FollowChecker

	// 3. Initialize Commands/Queries and Register HTTP routes

	// Slice: Register
	registerHandler := auth.MakeHTTPHandler(auth.NewCommandHandler(authStore, bus))
	mux.HandleFunc("POST /api/auth/register", registerHandler)

	// Slice: Follow User
	followHandler := follow.MakeHTTPHandler(follow.NewFollowCommandHandler(followStore, followSvc, bus))
	mux.HandleFunc("POST /api/follow", followHandler)

	// Slice: Create Post (Wired with cross-slice followSvc checking dependency)
	createPostHandler := post.MakeHTTPHandler(post.NewCommandHandler(postStore, followSvc, bus))
	mux.HandleFunc("POST /api/posts", createPostHandler)

	// 4. Wire Event Bus subscribers
	// notification.SubscribeToEvents(bus, notificationStore)
}
```

---

## 7. Compliance Checklist Mapping

Here is how the vertical slice architecture handles key spec items:

| Spec Feature                   | Vertical Slice Location                | Design / CQRS Handling                                                                                                                      |
| ------------------------------ | -------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| **WAL & Busy Timeout**         | `internal/platform/database/`          | Guaranteed in central SQLite factory; slices only consume `database.DB` wrapper.                                                            |
| **Email/Username Duplicity**   | `internal/auth/register_command.go`    | SQL verification is isolated in `register_store.go`'s INSERT.                                                                               |
| **Profile Privacy Visibility** | `internal/user/get_profile_query.go`   | Privacy lock logic evaluates follower relationship via local interface before returning details.                                            |
| **Post Privacy (3 Types)**     | `internal/post/create_post_command.go` | Visibility state enum ("public", "almost_private", "private") is mapped and written alongside permitted user IDs in `create_post_store.go`. |
| **Group Invites / Joins**      | `internal/group/`                      | Separate command/query files manage requests, status transitions, and invite verification.                                                  |
| **Group Chat & Private Chat**  | `internal/chat/`                       | Chat validation queries follow relationships using local interface; real-time delivery routed by WS dispatcher.                             |
| **Notifications**              | `internal/notification/`               | Driven by decoupled event subscriptions on the EventBus.                                                                                    |
