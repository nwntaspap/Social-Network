# Sprint 2: User & Topic Features (Week 5–6)

**Outcome:** User account features (register, login, profile, privacy toggle) and Topic features (posts with public/almost_private/private visibility, post creation, voting) work end-to-end. Both frontend and backend implementations are completed.

> **FollowChecker stubs:** S2-BE-22 and S2-BE-30 define local `FollowChecker` interfaces. The Follow slice does not exist until Sprint 3. Until then, inject a stub that always returns `true` (public profiles bypass) or `false` (private profiles blocked, no follow-gating until Sprint 3). Mark with `// TODO: replace with real FollowChecker in Sprint 3`.
>
> **Migration dependencies:** S1-BE-06 must have created `000002_user_profile_fields` before S2-BE-15 User repo works, and `000003_topic_privacy` before S2-BE-26 Topic repo works. Verify migration order — these are implicit dependencies on S1-BE-06.

---

### S2-BE-14: Wire User & Topic bootstrap routes

- **Priority:** P0
- **Type:** Cleanup/Integration (Bootstrap slice wiring)
- **Assignee:** BE-A + BE-B
- **Story Points:** 3
- **Dependencies:** S2-BE-25, S2-BE-34
- **Description:** Bootstrap wiring ticket. Register new slice HTTP routes in `bootstrap.go` so endpoints are live immediately after this sprint. This is final dependency wiring, not new logic — it makes the previously-built User and Topic slices reachable via HTTP routing after their transport layers are complete.
- **Detailed Steps:**
  1. In `internal/bootstrap/bootstrap.go`, import `internal/user/transport` and `internal/topic/transport`. This wires the already-built transport layers into the running server — no new handlers or route logic is created here.
  2. Call their route registration functions on the HTTP mux (e.g. `userTransport.RegisterRoutes(mux)`).
  3. Verify with `curl http://localhost:8080/api/register` and `curl http://localhost:8080/api/feed`.
- **Verification:** `go build ./...` passes, endpoints respond 200/401 (not 404).

---

## BE-A (Backend A) Tickets

### S2-BE-15: User: Entity & Repository Interface

- **Priority:** P0
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-A
- **Story Points:** 2
- **Description:** Define the domain entity model for User and the repository interface mapping SQLite operations. This is a refactoring/migration ticket: the User entity and repository interface already exist in the old layered codebase at `internal/domain/user/` (entity definitions) and `internal/infra/storage/sqlite/users/` (repository queries). This ticket restructures that existing logic into the new vertical-slice layout under `internal/user/`, replacing the old domain/infra split with a single cohesive slice. The User slice also absorbs old `internal/domain/activity/` — user's activity (post counts, follower counts) becomes a query on user data. This migration extracts user and activity logic from the legacy layered architecture (`internal/domain/user/`, `internal/domain/activity/`, `internal/infra/...`) into the new `internal/user/` vertical slice, adhering to the Strangler Fig pattern.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1.  Create `internal/user/user.go`. Define the `User` struct (ID, Email, PasswordHash, FirstName, LastName, DateOfBirth, Nickname, AboutMe, AvatarPath, IsPrivate, CreatedAt). **Explicitly drop `Age` field** — replaced by `DateOfBirth` for age calculation at runtime.
  2.  **DB Schema Sync:** The Go struct's `Nickname` field must map to the SQLite `username` column, which is retained to prevent breaking legacy code running side-by-side.
  3.  Define the `Repository` interface specifying required CRUD queries (e.g. `Create`, `GetByID`, `GetByEmail`, `Update`, `TogglePrivacy`, `ListAll`).
- **Verification:** Compile check `go build ./internal/user/...`.

---

### S2-BE-16: User: SQLite Store

- **Priority:** P0
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-A
- **Story Points:** 3
- **Dependencies:** S2-BE-15
- **Description:** Implement the User `Repository` interface using SQLite database operations. This migration extracts user and activity logic from the legacy layered architecture (`internal/domain/user/`, `internal/domain/activity/`, `internal/infra/...`) into the new `internal/user/` vertical slice, adhering to the Strangler Fig pattern.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1. Create `internal/user/store/sqlite.go`. Implement the repository using the `platform/database.DB` interface.
  2. Map the `Nickname` field on the Go `User` struct to the `username` column in SQL queries.
  3. Implement scan functions translating SQLite rows into `User` domain structures.
- **Verification:** Write store integration tests in `sqlite_test.go` utilizing an in-memory SQLite database connection. Run `go test -v ./internal/user/store/...`.

---

### S2-BE-17: User: Register Command

- **Priority:** P0
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-A
- **Story Points:** 3
- **Dependencies:** S2-BE-15
- **Description:** Create the write-use-case handler for user registration with input validation. This migration extracts user and activity logic from the legacy layered architecture (`internal/domain/user/`, `internal/domain/activity/`, `internal/infra/...`) into the new `internal/user/` vertical slice, adhering to the Strangler Fig pattern.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1. Create `internal/user/commands/register.go`.
  2. Implement age validation (must be at least 13 years old), password strength rules, and duplicate email checking.
  3. Hash password using bcrypt. Store optional avatar image using `pkg/imgutil` validation.
- **Verification:** Write table-driven unit tests checking valid registration, underage rejection, duplicate email blocking, and invalid fields.

---

### S2-BE-18: User: Login Command

- **Priority:** P0
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-A
- **Story Points:** 2
- **Dependencies:** S2-BE-15
- **Description:** Implement credential validation and user session mapping. This migration extracts user and activity logic from the legacy layered architecture (`internal/domain/user/`, `internal/domain/activity/`, `internal/infra/...`) into the new `internal/user/` vertical slice, adhering to the Strangler Fig pattern.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1. Create `internal/user/commands/login.go`.
  2. Query user by email or username (maps to `Nickname` in code). Check password matching with bcrypt.
  3. Call session manager to generate a session cookie token.
- **Verification:** Unit tests validating correct credentials login, wrong email rejection, and wrong password lockouts.

---

### S2-BE-19: User: Logout Command

- **Priority:** P1
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-A
- **Story Points:** 1
- **Dependencies:** S2-BE-15
- **Description:** Handle session termination. This migration extracts user and activity logic from the legacy layered architecture (`internal/domain/user/`, `internal/domain/activity/`, `internal/infra/...`) into the new `internal/user/` vertical slice, adhering to the Strangler Fig pattern.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1. Create `internal/user/commands/logout.go`.
  2. Revoke active session token from the session store.
- **Verification:** Unit test asserting lookups of revoked sessions fail.

---

### S2-BE-20: User: Update Profile Command

- **Priority:** P1
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-A
- **Story Points:** 2
- **Dependencies:** S2-BE-15
- **Description:** Handle user profile edits (First Name, Last Name, Nickname, About Me, Avatar). This migration extracts user and activity logic from the legacy layered architecture (`internal/domain/user/`, `internal/domain/activity/`, `internal/infra/...`) into the new `internal/user/` vertical slice, adhering to the Strangler Fig pattern.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1. Create `internal/user/commands/update_profile.go`.
  2. Implement input sanitation, optional fields validation, and update db records (mapping Nickname to the username column).
- **Verification:** Test updates to verify information modifies correctly in database.

---

### S2-BE-21: User: Toggle Privacy Command

- **Priority:** P1
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-A
- **Story Points:** 2
- **Dependencies:** S2-BE-15
- **Description:** Implement profile visibility toggle (public/private profiles). This migration extracts user and activity logic from the legacy layered architecture (`internal/domain/user/`, `internal/domain/activity/`, `internal/infra/...`) into the new `internal/user/` vertical slice, adhering to the Strangler Fig pattern.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1. Create `internal/user/commands/toggle_privacy.go`.
  2. Flip `is_private` boolean field in database.
- **Verification:** Unit tests asserting that toggle updates the flag successfully.

---

### S2-BE-22: User: Get Profile Query

- **Priority:** P1
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-A
- **Story Points:** 3
- **Dependencies:** S2-BE-15
- **Description:** Implement profile retrieval. Block contents for non-followers when profile is private. This migration extracts user and activity logic from the legacy layered architecture (`internal/domain/user/`, `internal/domain/activity/`, `internal/infra/...`) into the new `internal/user/` vertical slice, adhering to the Strangler Fig pattern.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1. Create `internal/user/queries/get_profile.go`.
  2. Accept target UserID and requester UserID.
  3. Define a local `FollowChecker` interface to check if requester follows the profile.
  4. **Stub for Sprint 2:** Inject a `FollowChecker` stub — use `true` (no follow-gating, show all) or `false` (block all private profiles). Annotate with `// TODO: replace with real FollowChecker in Sprint 3`.
  5. If private and not following, return limited details (nickname/avatar only) and a privacy error.
- **Verification:** Unit tests checking: public profile gets full info, private profile followed gets full info, private profile non-followed gets blocked.

---

### S2-BE-23: User: Get Activity Query

- **Priority:** P1
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-A
- **Story Points:** 2
- **Dependencies:** S2-BE-15
- **Description:** Retrieve the user's activity log. This query absorbs the old `internal/domain/activity/` and `internal/app/activities/` logic. It retrieves the user's own posts, comments, votes, as well as lists of followers and following. This migration extracts user and activity logic from the legacy layered architecture (`internal/domain/user/`, `internal/domain/activity/`, `internal/infra/...`) into the new `internal/user/` vertical slice, adhering to the Strangler Fig pattern.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1. Create `internal/user/queries/get_activity.go`.
  2. Query the counts of posts, comments, and votes created by the user.
  3. Retrieve lists of the user's followers and users they are following (in Sprint 2 this uses the `FollowChecker` stub; it will be wired to the real follow slice in Sprint 3).
- **Verification:** Unit tests asserting correctness of count retrievals and follower/following list responses.

---

### S2-BE-24: User: List Users Query

- **Priority:** P1
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-A
- **Story Points:** 2
- **Dependencies:** S2-BE-15
- **Description:** List/browse all registered users for exploration. This migration extracts user and activity logic from the legacy layered architecture (`internal/domain/user/`, `internal/domain/activity/`, `internal/infra/...`) into the new `internal/user/` vertical slice, adhering to the Strangler Fig pattern.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1. Create `internal/user/queries/list_users.go`.
  2. Retrieve list of users (excluding sensitive details like passwords).
- **Verification:** Unit tests checking pagination and correct output mapping.

---

### S2-BE-25: User: HTTP Transport Routing

- **Priority:** P1
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-A
- **Story Points:** 3
- **Dependencies:** S2-BE-17..10
- **Description:** Bind all user commands and queries to HTTP handler endpoints. This migration extracts user and activity logic from the legacy layered architecture (`internal/domain/user/`, `internal/domain/activity/`, `internal/infra/...`) into the new `internal/user/` vertical slice, adhering to the Strangler Fig pattern.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1.  Create `internal/user/transport/http.go`.
  2.  Wire up `POST /api/register`, `POST /api/login`, `POST /api/logout`, `GET /api/users/:id/profile`, `GET /api/users/:id/activity` (maps to S2-BE-23 get_activity), `GET /api/users` (maps to S2-BE-24 list_users), `PUT /api/profile`, `POST /api/profile/privacy`.
- **Verification:** Integration tests verifying status codes and JSON response outputs over mock HTTP requests. Every command and query must have at least one route.

---

## BE-B (Backend B) Tickets

### S2-BE-26: Topic: Entity & Repository Interface

- **Priority:** P0
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-B
- **Story Points:** 2
- **Description:** Define domain entity model for posts/topics, visibility scopes, and votes. Topic absorbs the old `internal/domain/category/` and `internal/domain/vote/` domains. Category entities from the old code are merged directly into Topic as a visibility aspect or replaced by the `Visibility` enum; votes are modeled directly inside the Topic slice. This migration extracts topic, category, and vote logic from legacy layered domains into the new `internal/topic/` vertical slice, introducing the new visibility enum (public/almost_private/private) without breaking existing routes.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1. Create `internal/topic/topic.go`.
  2. Define `Topic` entity containing: ID (int), UserID (string) [maps to user_id in DB], Title, Content, ImagePath, Visibility (public, almost_private, private), and CreatedAt. **Note:** Use `int` for IDs to match legacy database types and keep the side-by-side running code functional.
  3. Define `AllowedUser` entity to map which specific users can view private posts: TopicID (int), UserID (string).
  4. Define `Vote` entity: UserID (string), TopicID (int), CommentID (int), ReactionType (int: +1/-1) to absorb the legacy post voting schema. **Defer TargetType: comment voting to Sprint 3** (comment slice) to avoid a circular `topic → comment` dependency at compilation time.
  5. Define the `Repository` interface.
- **Verification:** Compile check `go build ./internal/topic/...`.

---

### S2-BE-27: Topic: SQLite Store

- **Priority:** P0
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-B
- **Story Points:** 3
- **Dependencies:** S2-BE-26
- **Description:** Implement Topic `Repository` queries in SQLite. Topic absorbs the old `internal/infra/storage/sqlite/topics/` and `internal/infra/storage/sqlite/categories/` storage queries. This migration extracts topic, category, and vote logic from legacy layered domains into the new `internal/topic/` vertical slice, introducing the new visibility enum (public/almost_private/private) without breaking existing routes.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1. Create `internal/topic/store/sqlite.go`.
  2. Implement storage queries using `platform/database.DB`. Map Go entity columns directly to legacy DB columns: `user_id` instead of `author_id`, and `reaction_type` instead of `value` for votes.
  3. Write complex visibility queries checking permissions, followers, and allowed lists.
- **Verification:** Store integration tests using in-memory SQLite checking correct write/read of visibility permissions.

---

### S2-BE-28: Topic: Create Topic Command

- **Priority:** P0
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-B
- **Story Points:** 3
- **Dependencies:** S2-BE-26
- **Description:** Build write-use-case for creating posts with visibility restrictions and file attachments. This migration extracts topic, category, and vote logic from legacy layered domains into the new `internal/topic/` vertical slice, introducing the new visibility enum (public/almost_private/private) without breaking existing routes.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1. Create `internal/topic/commands/create_topic.go`.
  2. Implement input checking. Extract and validate attached images (JPG, PNG, GIF) using magic bytes verification.
  3. Save attached image to path if present. Save visibility permission records.
- **Verification:** Unit tests verifying public post creation, and private post creation with specific allowed users.

---

### S2-BE-29: Topic: Cast Vote Command

- **Priority:** P1
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-B
- **Story Points:** 2
- **Dependencies:** S2-BE-26
- **Description:** Cast upvotes and downvotes on posts. This migration extracts topic, category, and vote logic from legacy layered domains into the new `internal/topic/` vertical slice, introducing the new visibility enum (public/almost_private/private) without breaking existing routes.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1. Create `internal/topic/commands/cast_vote.go`.
  2. Check if vote already exists. If direction is identical, remove vote. If different, update direction.
- **Verification:** Unit tests validating: voting up, changing to down, and canceling votes.

---

### S2-BE-30: Topic: Get Feed Query

- **Priority:** P0
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-B
- **Story Points:** 5
- **Dependencies:** S2-BE-26
- **Description:** Get home feed posts filtered by privacy scopes. This migration extracts topic, category, and vote logic from legacy layered domains into the new `internal/topic/` vertical slice, introducing the new visibility enum (public/almost_private/private) without breaking existing routes.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1. Create `internal/topic/queries/get_feed.go`.
  2. Retrieve list of topics. Accept requester UserID.
  3. Define a local `FollowChecker` interface.
  4. **Stub for Sprint 2:** Inject a `FollowChecker` stub — use `true` (show all visible posts) or annotate with `// TODO: replace with real FollowChecker in Sprint 3`.
  5. Build visibility filter:
  - Public posts are visible to everyone.
  - `almost_private` posts are visible to author and followers.
  - `private` posts are visible to author and explicitly allowed users.
- **Verification:** Unit tests checking that posts with various privacy scopes are correctly filtered out depending on follow status.

---

### S2-BE-31: Topic: Get User Topics Query

- **Priority:** P1
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-B
- **Story Points:** 2
- **Dependencies:** S2-BE-26
- **Description:** Retrieve posts created by a specific user, ensuring visibility checks are enforced. This migration extracts user and activity logic from the legacy layered architecture (`internal/domain/user/`, `internal/domain/activity/`, `internal/infra/...`) into the new `internal/user/` vertical slice, adhering to the Strangler Fig pattern.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1. Create `internal/topic/queries/get_user_topics.go`.
- **Verification:** Unit tests validating privacy check boundaries.

---

### S2-BE-32: Topic: Get Topic Query

- **Priority:** P1
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-B
- **Story Points:** 2
- **Dependencies:** S2-BE-26
- **Description:** Retrieve details for a single post with visibility checks. This migration extracts topic, category, and vote logic from legacy layered domains into the new `internal/topic/` vertical slice, introducing the new visibility enum (public/almost_private/private) without breaking existing routes.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1. Create `internal/topic/queries/get_topic.go`.
- **Verification:** Tests checking access controls.

---

### S2-BE-33: Topic: Get Votes Query

- **Priority:** P1
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-B
- **Story Points:** 2
- **Dependencies:** S2-BE-26
- **Description:** Retrieve total count of upvotes and downvotes for a target post. This migration extracts topic, category, and vote logic from legacy layered domains into the new `internal/topic/` vertical slice, introducing the new visibility enum (public/almost_private/private) without breaking existing routes.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1. Create `internal/topic/queries/get_votes.go`.
- **Verification:** Tests verifying correct sum outputs.

---

### S2-BE-34: Topic: HTTP Transport Routing

- **Priority:** P1
- **Type:** Refactoring/Migration (Existing Codebase)
- **Assignee:** BE-B
- **Story Points:** 3
- **Dependencies:** S2-BE-28..20
- **Description:** Setup HTTP routing handlers for posts and votes. This migration extracts topic, category, and vote logic from legacy layered domains into the new `internal/topic/` vertical slice, introducing the new visibility enum (public/almost_private/private) without breaking existing routes.
- **Detailed Steps:**
  - _Migration Note:_ Follow the Strangler Fig pattern (R1). Write contract tests against the old API first, build the new CQRS slice, and swap routing only when tests match.
  1.  Create `internal/topic/transport/http.go`.
  2.  Wire up `POST /api/posts`, `GET /api/feed`, `GET /api/posts/:id`, `GET /api/users/:id/posts` (maps to S2-BE-31 get_user_topics), `GET /api/posts/:id/votes` (maps to S2-BE-33 get_votes), `POST /api/posts/:id/vote`.
- **Verification:** HTTP mock integration tests verifying correct endpoint codes. Every command and query must have at least one route.

---

## FE-A (Frontend A) Tickets

### S2-FE-06: Registration Form

- **Priority:** P0
- **Type:** Greenfield (New Frontend UI)
- **Assignee:** FE-A
- **Story Points:** 5
- **Description:** Implement registration form details, image file selection, and payload submissions. As a greenfield frontend task, this implements new Next.js UI components in `frontend/src/` utilizing shadcn/ui and Vanilla CSS, wiring them to the Next.js App Router.
- **Detailed Steps:**
  - _Greenfield Note:_ Use ESLint + Prettier for linting/formatting and ensure session cookies are handled securely without localStorage leakage.
  1. Bind form inputs to state hooks.
  2. Convert upload image file to multipart form data structure.
  3. Validate forms for local client bounds before posting.
- **Verification:** Playwright interactive flow testing correct submissions.

---

### S2-FE-07: Login Page

- **Priority:** P0
- **Type:** Greenfield (New Frontend UI)
- **Assignee:** FE-A
- **Story Points:** 3
- **Description:** Complete login page layout and credentials request mapping. As a greenfield frontend task, this implements new Next.js UI components in `frontend/src/` utilizing shadcn/ui and Vanilla CSS, wiring them to the Next.js App Router.
- **Detailed Steps:**
  - _Greenfield Note:_ Use ESLint + Prettier for linting/formatting and ensure session cookies are handled securely without localStorage leakage.
  1. Implement login view. Post username/email and password parameters to `/api/login`.
- **Verification:** Check visual styling and test input submits.

---

### S2-FE-08: Profile Page

- **Priority:** P1
- **Type:** Greenfield (New Frontend UI)
- **Assignee:** FE-A
- **Story Points:** 5
- **Description:** Implement `/profile/[id]` layout with follower tallies, user activity posts feed, and private lock screens. As a greenfield frontend task, this implements new Next.js UI components in `frontend/src/` utilizing shadcn/ui and Vanilla CSS, wiring them to the Next.js App Router.
- **Detailed Steps:**
  - _Greenfield Note:_ Use ESLint + Prettier for linting/formatting and ensure session cookies are handled securely without localStorage leakage.
  1. Render user details, follow counts, and followers lists.
  2. Fetch profile data: if locked, show overlay screen with locked lock icon.
  3. Render feed list of posts created by the user.
- **Verification:** Mock api states (locked vs open) and test page output.

---

### S2-FE-09: Privacy Toggle with Confirmation Popup (Bonus)

- **Priority:** P1
- **Type:** Greenfield (New Frontend UI)
- **Assignee:** FE-A
- **Story Points:** 2
- **Description:** Build privacy toggle switch with confirmation dialog box before updating state. As a greenfield frontend task, this implements new Next.js UI components in `frontend/src/` utilizing shadcn/ui and Vanilla CSS, wiring them to the Next.js App Router.
- **Detailed Steps:**
  - _Greenfield Note:_ Use ESLint + Prettier for linting/formatting and ensure session cookies are handled securely without localStorage leakage.
  1. Render switch component. When toggled, pop up confirmation Dialog.
  2. Send request to `/api/profile/privacy` only if confirmed.
- **Verification:** Interactive test confirming cancel ignores, accept updates database.

---

## FE-B (Frontend B) Tickets

### S2-FE-10: Home Feed Page

- **Priority:** P0
- **Type:** Greenfield (New Frontend UI)
- **Assignee:** FE-B
- **Story Points:** 5
- **Description:** Build the main feed page with infinite scroll loading or pagination. As a greenfield frontend task, this implements new Next.js UI components in `frontend/src/` utilizing shadcn/ui and Vanilla CSS, wiring them to the Next.js App Router.
- **Detailed Steps:**
  - _Greenfield Note:_ Use ESLint + Prettier for linting/formatting and ensure session cookies are handled securely without localStorage leakage.
  1. Fetch items from `/api/feed`. Handle loading skeleton states.
- **Verification:** Render page with mock data arrays.

---

### S2-FE-11: Post Creation Form

- **Priority:** P0
- **Type:** Greenfield (New Frontend UI)
- **Assignee:** FE-B
- **Story Points:** 5
- **Description:** Build post creation UI container with image file attachment selector and visibility settings. As a greenfield frontend task, this implements new Next.js UI components in `frontend/src/` utilizing shadcn/ui and Vanilla CSS, wiring them to the Next.js App Router.
- **Detailed Steps:**
  - _Greenfield Note:_ Use ESLint + Prettier for linting/formatting and ensure session cookies are handled securely without localStorage leakage.
  1. Build creation box. Add Select dropdown (public, almost private, private).
  2. If private is chosen, render user selector to pick permitted users.
- **Verification:** Interactive submission check testing parameters sent.

---

### S2-FE-12: Post Card Component

- **Priority:** P1
- **Type:** Greenfield (New Frontend UI)
- **Assignee:** FE-B
- **Story Points:** 3
- **Description:** Render feed post item. As a greenfield frontend task, this implements new Next.js UI components in `frontend/src/` utilizing shadcn/ui and Vanilla CSS, wiring them to the Next.js App Router.
- **Detailed Steps:**
  - _Greenfield Note:_ Use ESLint + Prettier for linting/formatting and ensure session cookies are handled securely without localStorage leakage.
  1. Render header (author name, timestamp), Content, Image if present, and vote counts buttons.
- **Verification:** Test component renders correctly with varying input datasets.

---

## SD-QA (System Design/QA) Tickets

### S2-SD-07: User Slice: Migration Verification Contract Tests

- **Priority:** P1
- **Type:** Testing/Verification
- **Assignee:** SD-QA
- **Story Points:** 3
- **Dependencies:** S0-BE-01 (old repo exists), S2-BE-25 (new slice for verification)
- **Description:** Per Strangler Fig (Step 1 then Step 3), first write contract tests against the OLD repo, then verify the NEW slice passes the same tests.
- **Detailed Steps:**
  1.  **(Step 1)** Create `internal/user/store/sqlite_migration_test.go`. Write tests against the old repository (`internal/infra/storage/sqlite/...` queries) to capture current behavior.
  2.  **(Step 2)** New slice is built (S2-BE-16).
  3.  **(Step 3)** Run same contract tests against the new `internal/user/store/sqlite.go` — assert identical data mapping.
- **Verification:** Contract tests pass with 100% data compatibility against old repo first, then new slice.

---

### S2-SD-08: Topic Slice: Migration Verification Contract Tests

- **Priority:** P1
- **Type:** Testing/Verification
- **Assignee:** SD-QA
- **Story Points:** 3
- **Dependencies:** S0-BE-01 (old repo exists), S2-BE-34 (new slice for verification)
- **Description:** Per Strangler Fig (Step 1 then Step 3), write contract tests against OLD repo first, then verify NEW slice passes the same tests.
- **Detailed Steps:**
  1.  **(Step 1)** Create `internal/topic/store/sqlite_migration_test.go`. Write tests against old topic repository.
  2.  **(Step 3)** Run same tests against new vertical slice store.
- **Verification:** Assert equality of returned structures. Old-then-new ordering verified.

---

### S2-SD-10: E2E: User Signup to Feed Journey

- **Priority:** P0
- **Type:** Testing/Verification
- **Assignee:** SD-QA
- **Story Points:** 3
- **Description:** Core integration Playwright test validating registration, login, and posting to feed.
- **Detailed Steps:**
  1. Playwright scripts: Signup user -> Log in -> Write Post -> Inspect Feed for new post.
- **Verification:** Script executes successfully in headless CI runner.

---

### S2-SD-09: Platform: User & Topic Migrations (000002 & 000003)

- **Priority:** P0
- **Type:** Greenfield (New Module/Feature - DB Migrations)
- **Assignee:** SD-QA
- **Story Points:** 2
- **Dependencies:** S1-BE-06
- **Description:** Create the database migration files for the User and Topic vertical slices (Phases 2.4 / 5). As a greenfield task, this builds new platform or feature abstractions from scratch that do not exist in the old legacy codebase.
- **Detailed Steps:**
  - _Greenfield Note:_ Follow TDD (Red-Green-Refactor). Ensure the slice adheres strictly to boundary rules (D5) without importing other slices' stores/transports.
  1. Create `db/migrations/000002_user_profile_fields.up.sql` to add `date_of_birth`, `about_me`, and `is_private` to the `users` table, and drop the `age` field (retaining `username` column).
  2. Create `db/migrations/000002_user_profile_fields.down.sql` to reverse these changes.
  3. Create `db/migrations/000003_topic_privacy.up.sql` to add `visibility` and `image_url` to the `topics` table and create the `topic_allowed_users` table (retaining topics auto-incrementing integer IDs).
  4. Create `db/migrations/000003_topic_privacy.down.sql` to reverse these changes.
- **Verification:** Run `make db-reset` or execute the migration runner and verify that these migrations apply and rollback cleanly, creating the expected schema changes.
