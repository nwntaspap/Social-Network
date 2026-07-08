# Sprint 5: Chat & OAuth (Week 11–12)

**Outcome:** 1-on-1 private messaging featuring follow check constraints, emojis, and third-party login delegations (GitHub/Google OAuth) work end-to-end.

> **Old chat commands dropped:** Old code had `markAsRead`, `initChat`, `getChatUsers` WS/RPC methods. These are not mapped to the new slice. `initChat` is replaced by implicit conversation creation on first message (S5-BE-86). `markAsRead` dropped (no per-message read tracking in arch). `getChatUsers` dropped (conversation partner derived from `GET /api/chat/conversations`). Document this in the FE migration notes.

---

### S5-BE-83: Wire Chat & OAuth bootstrap routes

- **Priority:** P0
- **Type:** Cleanup/Integration (Bootstrap slice wiring)
- **Assignee:** BE-A + BE-B
- **Story Points:** 3
- **Dependencies:** S5-BE-89, S5-BE-90, S5-BE-96
- **Description:** Register new slice routes in `bootstrap.go` so endpoints are live immediately after this sprint. This is bootstrap-level dependency wiring — gluing the newly created chat and OAuth vertical slices into the application's startup path so they become reachable at runtime.
- **Detailed Steps:**
  1. In `internal/bootstrap/bootstrap.go`, import chat and oauth transport packages.
  2. Call their route registration functions on the HTTP mux and WS router.
  3. Wire OAuth provider clients (github, google) per S5-BE-97/98.
- **Verification:** `go build ./...` passes, new endpoints respond 200/401/403 (not 404).

---

## BE-A (Backend A) Tickets

### S5-BE-84: Chat: Entity & Repository Interface

- **Priority:** P0
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-A
- **Story Points:** 2
- **Description:** Establish domain structures mapping messaging records. This is a refactoring/migration ticket — the chat domain model already exists in the old layered codebase (`internal/domain/chat/`) and is being restructured into the new vertical-slice layout under `internal/chat/`. This migration moves chat domain and WebSocket handlers from `internal/infra/ws/handlers/` into the `internal/chat/` slice, integrating the new `FollowChecker` interface for cross-slice authorization.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1. Create `internal/chat/chat.go`.
  2. Define `PrivateMessage` (ID, ChatID string, SenderID, ReceiverID, Content, CreatedAt) and `Repository` interface. **Note:** Include `ChatID` (string) as required by the websocket payloads and `messages` table schema in `sds.md`.
  3. Define a local `FollowChecker` interface containing `AreConnected(ctx context.Context, a, b string) (bool, error)`.
- **Verification:** Compile check `go build ./internal/chat/...`.

---

### S5-BE-85: Chat: SQLite Store

- **Priority:** P0
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-A
- **Story Points:** 2
- **Dependencies:** S5-BE-84, S5-BE-91
- **Description:** Create SQLite storage queries mapping messages history. This migration moves chat domain and WebSocket handlers from `internal/infra/ws/handlers/` into the `internal/chat/` slice, integrating the new `FollowChecker` interface for cross-slice authorization.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1. Create `internal/chat/store/sqlite.go`. Implement queries saving messages and loading chats history mapping to the new `chats` and `messages` tables.
- **Verification:** Integration tests checking message writes and retrieval against the newly migrated schema.

---

### S5-BE-86: Chat: Send Private Message Command

- **Priority:** P0
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-A
- **Story Points:** 3
- **Dependencies:** S5-BE-84
- **Description:** Process messaging delivery. Enforce follow relationship checks and route through realtime sockets. This migration moves chat domain and WebSocket handlers from `internal/infra/ws/handlers/` into the `internal/chat/` slice, integrating the new `FollowChecker` interface for cross-slice authorization.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1. Create `internal/chat/commands/send_private_msg.go`.
  2. Call local `FollowChecker` to verify that at least one of the users follows the other. If not, reject with error.
  3. Save message to database and dispatch to WebSocket coordinator.
- **Verification:** Unit tests verifying: reject messages between unconnected profiles, allow messages between followers, and verify socket dispatch calls.

---

### S5-BE-87: Chat: Get Chat History Query

- **Priority:** P1
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-A
- **Story Points:** 2
- **Dependencies:** S5-BE-84
- **Description:** Retrieve historic messages log between active user and partner. This migration moves chat domain and WebSocket handlers from `internal/infra/ws/handlers/` into the `internal/chat/` slice, integrating the new `FollowChecker` interface for cross-slice authorization.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1. Create `internal/chat/queries/get_chat_history.go`. Check credentials.
- **Verification:** Test log query mapping.

---

### S5-BE-88: Chat: List Conversations Query

- **Priority:** P1
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-A
- **Story Points:** 2
- **Dependencies:** S5-BE-84
- **Description:** Retrieve distinct list of active chat partners for sidebar listings. This migration moves chat domain and WebSocket handlers from `internal/infra/ws/handlers/` into the `internal/chat/` slice, integrating the new `FollowChecker` interface for cross-slice authorization.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1. Create `internal/chat/queries/list_conversations.go`.
- **Verification:** Test conversations listing maps correctly.

---

### S5-BE-89: Chat: HTTP Transport Routing

- **Priority:** P1
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-A
- **Story Points:** 2
- **Dependencies:** S5-BE-86..05
- **Description:** Bind HTTP REST handlers. This migration moves chat domain and WebSocket handlers from `internal/infra/ws/handlers/` into the `internal/chat/` slice, integrating the new `FollowChecker` interface for cross-slice authorization.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1. Create `internal/chat/transport/http.go`.
  2. Route `GET /api/chat/conversations`, `GET /api/chat/:userId/history`.
- **Verification:** Mock integration HTTP tests.

---

### S5-BE-90: Chat: WS Transport Routing

- **Priority:** P0
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-A
- **Story Points:** 5
- **Dependencies:** S5-BE-89
- **Description:** Migrate old WebSocket chat handlers from `internal/infra/ws/handlers/` into the new vertical slice and bind to core WebSocket hub. This migration moves chat domain and WebSocket handlers from `internal/infra/ws/handlers/` into the `internal/chat/` slice, integrating the new `FollowChecker` interface for cross-slice authorization.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1.  Create `internal/chat/transport/ws.go`.
  2.  Move WS message handlers from `internal/infra/ws/handlers/` (chat_send.go, chat_history.go, etc.) into the new slice.
  3.  Connect routing events to `internal/core/realtime/` hub.
- **Verification:** Test messaging over connections.

---

## BE-B (Backend B) Tickets

### S5-BE-92: OAuth: Entity & Repository Interface

- **Priority:** P0
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-B
- **Story Points:** 1
- **Description:** Establish domain structures to validate third-party tokens state hashes. This migration refactors OAuth logic into `pkg/oauth/` and the new vertical slice structure for seamless third-party authentication.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1. Create `internal/oauth/oauth.go`. Define `OAuthState` (StateString, Provider: Github/Google, CreatedAt) and `Repository` interface.
- **Verification:** Compile checks.

---

### S5-BE-93: OAuth: SQLite Store

- **Priority:** P0
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-B
- **Story Points:** 1
- **Dependencies:** S5-BE-92
- **Description:** Create SQLite query maps for state validations. This migration refactors OAuth logic into `pkg/oauth/` and the new vertical slice structure for seamless third-party authentication.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1. Create `internal/oauth/store/sqlite.go`.
- **Verification:** Store integration tests checks.

---

### S5-BE-94: OAuth: Initiate Login Command

- **Priority:** P0
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-B
- **Story Points:** 2
- **Dependencies:** S5-BE-92
- **Description:** Generate state hashes and return third-party login redirection URLs. This migration refactors OAuth logic into `pkg/oauth/` and the new vertical slice structure for seamless third-party authentication.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1. Create `internal/oauth/commands/initiate.go`.
  2. Create state string token, save to database, format URL parameters, and redirect client browser.
- **Verification:** Unit tests asserting that unique states write successfully to DB.

---

### S5-BE-95: OAuth: Callback Processor Command

- **Priority:** P0
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-B
- **Story Points:** 3
- **Dependencies:** S5-BE-92
- **Description:** Consume callback queries. Validate states, request token swaps, login/register user, and return cookies. This migration refactors OAuth logic into `pkg/oauth/` and the new vertical slice structure for seamless third-party authentication.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1. Create `internal/oauth/commands/callback.go`.
  2. Validate request state against database.
  3. Call provider clients to swap code for access tokens and load user profiles (email, first/last names).
  4. If email is not in db -> register user record. Else -> load user.
  5. Setup session cookie mappings.
- **Verification:** Unit tests mocking provider services verifying register/login logic.

---

### S5-BE-96: OAuth: HTTP Transport Routing

- **Priority:** P1
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-B
- **Story Points:** 2
- **Dependencies:** S5-BE-94, S5-BE-95
- **Description:** Bind OAuth REST routes. This migration refactors OAuth logic into `pkg/oauth/` and the new vertical slice structure for seamless third-party authentication.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1. Create `internal/oauth/transport/http.go`. Route `GET /api/auth/oauth/:provider/init`, `GET /api/auth/oauth/:provider/callback`.
- **Verification:** Integration path calls test.

---

### S5-BE-97: OAuth Client: GitHub Implementation

- **Priority:** P1
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-B
- **Story Points:** 2
- **Dependencies:** S5-BE-99
- **Description:** Move and adapt the existing GitHub profile client from `internal/pkg/oAuth/githubclient/` to `pkg/oauth/github/client.go` per the target architecture. This migration refactors OAuth logic into `pkg/oauth/` and the new vertical slice structure for seamless third-party authentication.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1. Move the existing file and update imports and structure.
  2. Implement any necessary adjustments to fit the new slice interface.
- **Verification:** Unit tests using mock web server responses.

---

### S5-BE-98: OAuth Client: Google Implementation

- **Priority:** P1
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-B
- **Story Points:** 2
- **Dependencies:** S5-BE-99
- **Description:** Move and adapt the existing Google profile client from `internal/pkg/oAuth/googleclient/` to `pkg/oauth/google/client.go` per the target architecture. This migration refactors OAuth logic into `pkg/oauth/` and the new vertical slice structure for seamless third-party authentication.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1. Move the existing file and update imports and structure.
  2. Implement any necessary adjustments to fit the new slice interface.
- **Verification:** Unit tests using mock web server responses.

---

### S5-BE-99: Shared: Refactor OAuth Packages

- **Priority:** P0 (Prerequisite for S5-BE-97/98)
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-B
- **Story Points:** 1
- **Dependencies:** Sprint 0
- **Description:** Move and restructure OAuth packages to `pkg/oauth/` per the target architecture, doing the rename in Sprint 5 to prevent breaking the old bootstrap compilation earlier. This migration refactors OAuth logic into `pkg/oauth/` and the new vertical slice structure for seamless third-party authentication.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1.  Move `internal/pkg/` to `pkg/` (repo root, per target architecture — not `internal/pkg/`), which includes moving hashing, uuid, validator, helpers, and imgutil packages to root `pkg/` and renaming `oAuth/` to `oauth/`.
  2.  Flatten subdirectories: `internal/pkg/oAuth/githubclient/` → `pkg/oauth/github/client.go`, `internal/pkg/oAuth/googleclient/` → `pkg/oauth/google/client.go`.
  3.  Move raw HTTP OAuth token exchange clients from `internal/pkg/oAuth/httpclient/` into `pkg/oauth/client.go`.
- **Verification:** Ensure old auth compilation paths are updated (using alias imports if necessary) and all projects compile. `go build ./...` passes.

---

### S5-BE-91: Platform: Chat Migrations (Gap Fix)

- **Priority:** P0
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-A
- **Story Points:** 3
- **Dependencies:** S1-BE-06
- **Description:** Create the database migration files for the Chat vertical slice to transition legacy chats storage to the architecture's standard schemas. This migration moves chat domain and WebSocket handlers from `internal/infra/ws/handlers/` into the `internal/chat/` slice, integrating the new `FollowChecker` interface for cross-slice authorization.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1. Create `db/migrations/000008_migrate_chats.up.sql` to create `chats` and `messages` tables (with clean columns and UUID message IDs) and migrate existing data from the legacy `direct_chats` and `chat_messages` tables.
  2. Create `db/migrations/000008_migrate_chats.down.sql` to reverse this migration.
- **Verification:** Run up/down migration tests and verify message logs and active conversation entries are preserved.

---

## FE-A (Frontend A) Tickets

### S5-FE-28: Chat Feed View

- **Priority:** P0
- **Type:** Greenfield (New Frontend UI)
- **Assignee:** FE-A
- **Story Points:** 5
- **Description:** Implement direct messaging workspace `/chat` displaying conversational partner threads list and chats panel view. As a greenfield frontend task, this implements new Next.js UI components in `frontend/src/` utilizing shadcn/ui and Vanilla CSS, wiring them to the Next.js App Router.
- **Detailed Steps:**
  - _Greenfield Note:_ Use ESLint + Prettier for linting/formatting and ensure session cookies are handled securely without localStorage leakage.
  1. Retrieve conversations list. Render chat cards. Selecting card loads chat pane.
- **Verification:** Visual validation and states checking tests.

---

### S5-FE-29: Realtime Live Sockets Hook

- **Priority:** P0
- **Type:** Greenfield (New Frontend UI)
- **Assignee:** FE-A
- **Story Points:** 5
- **Dependencies:** S5-FE-28
- **Description:** Connect real-time WebSocket messaging handling typing indicators, online presence indicators, and incoming message dispatches. As a greenfield frontend task, this implements new Next.js UI components in `frontend/src/` utilizing shadcn/ui and Vanilla CSS, wiring them to the Next.js App Router.
- **Detailed Steps:**
  - _Greenfield Note:_ Use ESLint + Prettier for linting/formatting and ensure session cookies are handled securely without localStorage leakage.
  1.  Connect to websocket. Handle incoming payload types (`chat.message`, `chat.typing`, `chat.presence`).
  2.  **Typing indicators:** On keystroke (debounced 500ms), send `chat.typing` WS message with recipientID. On receiving `chat.typing`, show "typing..." bubble for 2s after last event.
  3.  **Online presence:** On WS connect/disconnect, broadcast `chat.presence` with status (online/offline). Track via hub client registry. Display green dot on online conversation partners.
  4.  Dispatch incoming `chat.message` payloads to chat store and update badge count.
- **Verification:** Playwright message delivery checks. Verify typing bubble appears and disappears. Verify online indicator shows for active WS connections.

---

### S5-FE-30: Chat Message Bubble Component

- **Priority:** P1
- **Type:** Greenfield (New Frontend UI)
- **Assignee:** FE-A
- **Story Points:** 2
- **Dependencies:** S5-FE-28
- **Description:** Render chat text bubble matching timestamps and emoji characters. As a greenfield frontend task, this implements new Next.js UI components in `frontend/src/` utilizing shadcn/ui and Vanilla CSS, wiring them to the Next.js App Router.
- **Detailed Steps:**
  - _Greenfield Note:_ Use ESLint + Prettier for linting/formatting and ensure session cookies are handled securely without localStorage leakage.
  1. Render styled message cells. Support emojis (Unicode formatting).
- **Verification:** Verify HTML characters render correctly.

---

## FE-B (Frontend B) Tickets

### S5-FE-31: GitHub OAuth Button Integration

- **Priority:** P1
- **Type:** Greenfield (New Frontend UI)
- **Assignee:** FE-B
- **Story Points:** 3
- **Description:** Implement GitHub button mapping clicks to initiation pathways. As a greenfield frontend task, this implements new Next.js UI components in `frontend/src/` utilizing shadcn/ui and Vanilla CSS, wiring them to the Next.js App Router.
- **Detailed Steps:**
  - _Greenfield Note:_ Use ESLint + Prettier for linting/formatting and ensure session cookies are handled securely without localStorage leakage.
  1. Add login option. Click routes to `/api/auth/oauth/github/init`.
- **Verification:** Test clicking routes to correct URL.

---

### S5-FE-32: Google OAuth Button Integration

- **Priority:** P1
- **Type:** Greenfield (New Frontend UI)
- **Assignee:** FE-B
- **Story Points:** 3
- **Dependencies:** S5-FE-31
- **Description:** Implement Google button mapping clicks to initiation pathways. As a greenfield frontend task, this implements new Next.js UI components in `frontend/src/` utilizing shadcn/ui and Vanilla CSS, wiring them to the Next.js App Router.
- **Detailed Steps:**
  - _Greenfield Note:_ Use ESLint + Prettier for linting/formatting and ensure session cookies are handled securely without localStorage leakage.
  1. Add login option. Click routes to `/api/auth/oauth/google/init`.
- **Verification:** Test clicking routes to correct URL.

---

## SD-QA (System Design/QA) Tickets

### S5-SD-18: Chat Slice: Contract Tests

- **Priority:** P1
- **Type:** Testing/Verification
- **Assignee:** SD-QA
- **Story Points:** 2
- **Dependencies:** S5-BE-90
- **Description:** Verify chat vertical slice compatibility with old domain.
- **Detailed Steps:**
  1. Create `internal/chat/store/sqlite_migration_test.go`.
- **Verification:** Assert equality of returned structures.

---

### S5-SD-19: OAuth Slice: Contract Tests

- **Priority:** P1
- **Type:** Testing/Verification
- **Assignee:** SD-QA
- **Story Points:** 2
- **Dependencies:** S5-BE-96
- **Description:** Ensure OAuth vertical slice compatibility with old domain.
- **Detailed Steps:**
  1. Create `internal/oauth/store/sqlite_migration_test.go`.
- **Verification:** Assert equality of returned structures.

---

### S5-SD-20: E2E: Messaging Real-Time Delivery Journey

- **Priority:** P0
- **Type:** Testing/Verification
- **Assignee:** SD-QA
- **Story Points:** 3
- **Dependencies:** S5-FE-29
- **Description:** Full E2E Playwright test validating messaging loops.
- **Detailed Steps:**
  1. User A follows User B -> A messages B -> B receives in real-time.
- **Verification:** Runs successfully in CI.

---

### S5-SD-21: E2E: GitHub OAuth Sign In

- **Priority:** P1
- **Type:** Testing/Verification
- **Assignee:** SD-QA
- **Story Points:** 3
- **Dependencies:** S5-FE-31
- **Description:** E2E Playwright testing OAuth mock flows.
- **Detailed Steps:**
  1. Launch Playwright browser -> Click GitHub login -> redirect callback success -> logged in.
- **Verification:** Test validates successfully.
