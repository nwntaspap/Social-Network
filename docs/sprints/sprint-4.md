# Sprint 4: Group & Event Features (Week 9–10)

**Outcome:** Groups with membership, group feed, group chat via WebSocket, and the event RSVP voting system work end-to-end.

> **Missing migration DDL:** Architecture specifies `000006_groups.up.sql` and `000007_events.up.sql`. S1-BE-06 created `000006` and `000007` stubs (empty or minimal). S4-BE-62 and S4-BE-78 must extend these files with actual Group and Event table DDL — or create replacement migration files if stubs were not created.

---

### S4-BE-60: Wire Group & Event bootstrap routes

- **Priority:** P0
- **Type:** Cleanup/Integration (Bootstrap slice wiring)
- **Assignee:** BE-A + BE-B
- **Story Points:** 3
- **Dependencies:** S4-BE-74, S4-BE-75, S4-BE-82
- **Description:** Register new slice routes in `bootstrap.go` so endpoints are live immediately after this sprint. This is bootstrap-level dependency wiring for the new Group and Event vertical slices — connecting their transport packages and event consumers into the existing bootstrap flow.
- **Detailed Steps:**
  1. In `internal/bootstrap/bootstrap.go`, import group and event transport packages.
  2. Call their route registration functions on the HTTP mux and WS router.
  3. Register event bus consumers for group events.
  4. Note: This is wiring-only — no business logic; the new slices are plugged into the existing bootstrap infrastructure.
- **Verification:** `go build ./...` passes, new endpoints respond 200/401/403 (not 404).

---

## BE-A (Backend A) Tickets

### S4-BE-61: Group: Entities & Repository Interface

- **Priority:** P0
- **Type:** Greenfield (New Module/Feature - Group system)
- **Assignee:** BE-A
- **Story Points:** 2
- **Description:** Establish domain model entities mapping group lifecycles. As a greenfield backend feature, this implements the entirely new group ecosystem (creation, invites, join requests, group chat, group posts) under `internal/group/`, isolated from other domains.
- **Detailed Steps:**
  - _Greenfield Note:_ Follow TDD (Red-Green-Refactor). Ensure the slice adheres strictly to boundary rules (D5) without importing other slices' stores/transports.
  1. Create `internal/group/group.go`.
  2. Define `Group` (ID, Title, Description, CreatorID, CreatedAt), `GroupMember` (GroupID, UserID, Role: creator/admin/member, JoinedAt), `Invitation` (GroupID, InviterID, InviteeID, CreatedAt), `JoinRequest` (GroupID, RequesterID, CreatedAt), and `GroupPost` (ID, GroupID, AuthorID, Title, Content, ImagePath, CreatedAt).
  3. **Note:** In alignment with the database schema, `Invitation` and `JoinRequest` do not have a `status` column (row presence denotes a pending status, and accept/decline actions delete the row). Also, `GroupPost` contains a `Title` field to map to the `group_posts.title` NOT NULL database column.
  4. Define the `Repository` interface mapping required storage operations.
- **Verification:** Compile check `go build ./internal/group/...`.

---

### S4-BE-62: Group: SQLite Store

- **Priority:** P0
- **Type:** Greenfield (New Module/Feature - Group system)
- **Assignee:** BE-A
- **Story Points:** 3
- **Dependencies:** S4-BE-61, S4-SD-16
- **Description:** Implement SQLite storage mapping group structure. As a greenfield backend feature, this implements the entirely new group ecosystem (creation, invites, join requests, group chat, group posts) under `internal/group/`, isolated from other domains.
- **Detailed Steps:**
  - _Greenfield Note:_ Follow TDD (Red-Green-Refactor). Ensure the slice adheres strictly to boundary rules (D5) without importing other slices' stores/transports.
  1.  Create `internal/group/store/sqlite.go`. Implement queries checking group membership: `IsMember(ctx context.Context, groupID, userID string) (bool, error)`.
- **Verification:** Integration tests checking memberships writing against the tables created by the `000006` migration.

---

### S4-BE-63: Group: Create Group Command

- **Priority:** P0
- **Type:** Greenfield (New Module/Feature - Group system)
- **Assignee:** BE-A
- **Story Points:** 2
- **Dependencies:** S4-BE-61
- **Description:** Create group record and automatically set the creator as group creator (with admin privileges). As a greenfield backend feature, this implements the entirely new group ecosystem (creation, invites, join requests, group chat, group posts) under `internal/group/`, isolated from other domains.
- **Detailed Steps:**
  - _Greenfield Note:_ Follow TDD (Red-Green-Refactor). Ensure the slice adheres strictly to boundary rules (D5) without importing other slices' stores/transports.
  1. Create `internal/group/commands/create_group.go`.
  2. Validate parameters (title, description bounds). Insert group, insert creator into members.
- **Verification:** Unit tests validating correct creator promotion.

---

### S4-BE-64: Group: Invite Member Command

- **Priority:** P1
- **Type:** Greenfield (New Module/Feature - Group system)
- **Assignee:** BE-A
- **Story Points:** 3
- **Dependencies:** S4-BE-61
- **Description:** Invite follower to group, firing event notifications. Architecture requires invite-gating: only users who follow the inviter can be invited. As a greenfield backend feature, this implements the entirely new group ecosystem (creation, invites, join requests, group chat, group posts) under `internal/group/`, isolated from other domains.
- **Detailed Steps:**
  - _Greenfield Note:_ Follow TDD (Red-Green-Refactor). Ensure the slice adheres strictly to boundary rules (D5) without importing other slices' stores/transports.
  1.  Create `internal/group/commands/invite_member.go`. Ensure requester is group member.
  2.  Define a local `FollowChecker` interface (same pattern as S2-BE-22, S2-BE-30): `AreConnected(ctx context.Context, a, b string) (bool, error)`.
  3.  Before inserting invitation, verify that invitee follows the inviter. Reject with 403 if not connected.
  4.  Insert invitation, publish `group.invited` event.
- **Verification:** Unit tests verifying: successful invite when follower, rejected invite when not connected, event outputs.

---

### S4-BE-65: Group: Respond Invite Command

- **Priority:** P1
- **Type:** Greenfield (New Module/Feature - Group system)
- **Assignee:** BE-A
- **Story Points:** 2
- **Dependencies:** S4-BE-64
- **Description:** Accept/decline group invitations. As a greenfield backend feature, this implements the entirely new group ecosystem (creation, invites, join requests, group chat, group posts) under `internal/group/`, isolated from other domains.
- **Detailed Steps:**
  - _Greenfield Note:_ Follow TDD (Red-Green-Refactor). Ensure the slice adheres strictly to boundary rules (D5) without importing other slices' stores/transports.
  1. Create `internal/group/commands/respond_invite.go`.
  2. Delete invitation row. If accepted -> insert user to member table.
- **Verification:** Check members mapping after accepts.

---

### S4-BE-66: Group: Request Join Command

- **Priority:** P1
- **Type:** Greenfield (New Module/Feature - Group system)
- **Assignee:** BE-A
- **Story Points:** 2
- **Dependencies:** S4-BE-61
- **Description:** Submit request to join group, notifying creator/admin. As a greenfield backend feature, this implements the entirely new group ecosystem (creation, invites, join requests, group chat, group posts) under `internal/group/`, isolated from other domains.
- **Detailed Steps:**
  - _Greenfield Note:_ Follow TDD (Red-Green-Refactor). Ensure the slice adheres strictly to boundary rules (D5) without importing other slices' stores/transports.
  1. Create `internal/group/commands/request_join.go`. Insert record in `group_join_requests`, publish `group.join_requested` event.
- **Verification:** Unit tests checking double request validation bounds.

---

### S4-BE-67: Group: Respond Join Command

- **Priority:** P1
- **Type:** Greenfield (New Module/Feature - Group system)
- **Assignee:** BE-A
- **Story Points:** 2
- **Dependencies:** S4-BE-66
- **Description:** Allow group creator/admin to approve join requests. As a greenfield backend feature, this implements the entirely new group ecosystem (creation, invites, join requests, group chat, group posts) under `internal/group/`, isolated from other domains.
- **Detailed Steps:**
  - _Greenfield Note:_ Follow TDD (Red-Green-Refactor). Ensure the slice adheres strictly to boundary rules (D5) without importing other slices' stores/transports.
  1. Create `internal/group/commands/respond_join.go`. Enforce that only the group creator can approve.
  2. Delete join request row. If accepted -> add user to member list.
- **Verification:** Assert only creator/admin can trigger, and member writes succeed.

---

### S4-BE-68: Group: Create Post Command

- **Priority:** P1
- **Type:** Greenfield (New Module/Feature - Group system)
- **Assignee:** BE-A
- **Story Points:** 3
- **Dependencies:** S4-BE-61
- **Description:** Create post inside a group. As a greenfield task, this builds new platform or feature abstractions from scratch that do not exist in the old legacy codebase.
- **Detailed Steps:**
  - _Greenfield Note:_ Follow TDD (Red-Green-Refactor). Ensure the slice adheres strictly to boundary rules (D5) without importing other slices' stores/transports.
  1. Create `internal/group/commands/create_group_post.go`. Enforce group membership checks. Validate title and content.
- **Verification:** Block creation for non-members, allow for members.

---

### S4-BE-69: Group: Send Group Message Command

- **Priority:** P1
- **Type:** Greenfield (New Module/Feature - Group system)
- **Assignee:** BE-A
- **Story Points:** 3
- **Dependencies:** S4-BE-61
- **Description:** Post a message in the group chat, dispatching to WebSocket connections. As a greenfield task, this builds new platform or feature abstractions from scratch that do not exist in the old legacy codebase.
- **Detailed Steps:**
  - _Greenfield Note:_ Follow TDD (Red-Green-Refactor). Ensure the slice adheres strictly to boundary rules (D5) without importing other slices' stores/transports.
  1. Create `internal/group/commands/send_group_message.go`. Validate membership. Route message through WS coordinator.
- **Verification:** Test socket delivery payload verification.

---

### S4-BE-70: Group: List Groups Query

- **Priority:** P1
- **Type:** Greenfield (New Module/Feature - Group system)
- **Assignee:** BE-A
- **Story Points:** 2
- **Dependencies:** S4-BE-61
- **Description:** Retrieve listing of all existing groups for browsing. As a greenfield backend feature, this implements the entirely new group ecosystem (creation, invites, join requests, group chat, group posts) under `internal/group/`, isolated from other domains.
- **Detailed Steps:**
  - _Greenfield Note:_ Follow TDD (Red-Green-Refactor). Ensure the slice adheres strictly to boundary rules (D5) without importing other slices' stores/transports.
  1. Create `internal/group/queries/list_groups.go`.
- **Verification:** Test list pagination.

---

### S4-BE-71: Group: Get Group Detail Query

- **Priority:** P1
- **Type:** Greenfield (New Module/Feature - Group system)
- **Assignee:** BE-A
- **Story Points:** 2
- **Dependencies:** S4-BE-61
- **Description:** Get specific group profile info. As a greenfield backend feature, this implements the entirely new group ecosystem (creation, invites, join requests, group chat, group posts) under `internal/group/`, isolated from other domains.
- **Detailed Steps:**
  - _Greenfield Note:_ Follow TDD (Red-Green-Refactor). Ensure the slice adheres strictly to boundary rules (D5) without importing other slices' stores/transports.
  1. Create `internal/group/queries/get_group.go`.
- **Verification:** Tests asserting responses structures.

---

### S4-BE-72: Group: Get Group Feed Query

- **Priority:** P1
- **Type:** Greenfield (New Module/Feature - Group system)
- **Assignee:** BE-A
- **Story Points:** 2
- **Dependencies:** S4-BE-61
- **Description:** Retrieve post list inside group. Enforce membership check. As a greenfield backend feature, this implements the entirely new group ecosystem (creation, invites, join requests, group chat, group posts) under `internal/group/`, isolated from other domains.
- **Detailed Steps:**
  - _Greenfield Note:_ Follow TDD (Red-Green-Refactor). Ensure the slice adheres strictly to boundary rules (D5) without importing other slices' stores/transports.
  1. Create `internal/group/queries/get_group_feed.go`.
- **Verification:** Assert non-members cannot read feed details.

---

### S4-BE-73: Group: Get Group Chat History Query

- **Priority:** P1
- **Type:** Greenfield (New Module/Feature - Group system)
- **Assignee:** BE-A
- **Story Points:** 2
- **Dependencies:** S4-BE-61
- **Description:** Get group message history log. As a greenfield task, this builds new platform or feature abstractions from scratch that do not exist in the old legacy codebase.
- **Detailed Steps:**
  - _Greenfield Note:_ Follow TDD (Red-Green-Refactor). Ensure the slice adheres strictly to boundary rules (D5) without importing other slices' stores/transports.
  1. Create `internal/group/queries/get_group_chat.go`.
- **Verification:** Retrieve log chronologically.

---

### S4-BE-74: Group: HTTP Transport Routing

- **Priority:** P1
- **Type:** Greenfield (New Module/Feature - Group system)
- **Assignee:** BE-A
- **Story Points:** 3
- **Dependencies:** S4-BE-63..13
- **Description:** Bind HTTP routes. Every command and query must have at least one route. As a greenfield backend feature, this implements the entirely new group ecosystem (creation, invites, join requests, group chat, group posts) under `internal/group/`, isolated from other domains.
- **Detailed Steps:**
  - _Greenfield Note:_ Follow TDD (Red-Green-Refactor). Ensure the slice adheres strictly to boundary rules (D5) without importing other slices' stores/transports.
  1.  Create `internal/group/transport/http.go`.
  2.  Route:
      - `POST /api/groups` — create_group (S4-BE-63)
      - `GET /api/groups` — list_groups (S4-BE-70)
      - `GET /api/groups/:id` — get_group (S4-BE-71)
      - `GET /api/groups/:id/feed` — get_group_feed (S4-BE-72)
      - `GET /api/groups/:id/chat` — get_group_chat (S4-BE-73)
      - `POST /api/groups/:id/invite` — invite_member (S4-BE-64)
      - `POST /api/groups/:id/invite/respond` — respond_invite (S4-BE-65)
      - `POST /api/groups/:id/join` — request_join (S4-BE-66)
      - `POST /api/groups/:id/join/respond` — respond_join (S4-BE-67)
      - `POST /api/groups/:id/posts` — create_group_post (S4-BE-68)
- **Verification:** Mock requests integration tests. Every command handler has a corresponding route.

---

### S4-BE-75: Group: WS Transport Routing

- **Priority:** P1
- **Type:** Greenfield (New Module/Feature - Group system)
- **Assignee:** BE-A
- **Story Points:** 3
- **Dependencies:** S4-BE-74
- **Description:** Route real-time WS chat events. As a greenfield backend feature, this implements the entirely new group ecosystem (creation, invites, join requests, group chat, group posts) under `internal/group/`, isolated from other domains.
- **Detailed Steps:**
  - _Greenfield Note:_ Follow TDD (Red-Green-Refactor). Ensure the slice adheres strictly to boundary rules (D5) without importing other slices' stores/transports.
  1. Create `internal/group/transport/ws.go`. Connect to core WebSocket.
- **Verification:** Test messaging over connections.

---

### S4-BE-76: Group: Post Comments (Gap Fix)

- **Priority:** P1
- **Type:** Greenfield (New Module/Feature - Group system)
- **Assignee:** BE-A
- **Story Points:** 3
- **Dependencies:** S4-BE-68
- **Description:** Implement commenting capabilities on group posts backend commands/queries. As a greenfield task, this builds new platform or feature abstractions from scratch that do not exist in the old legacy codebase.
- **Detailed Steps:**
  - _Greenfield Note:_ Follow TDD (Red-Green-Refactor). Ensure the slice adheres strictly to boundary rules (D5) without importing other slices' stores/transports.
  1. Define `GroupPostComment` (ID, PostID, AuthorID, Content, ImagePath, CreatedAt) and Repository.
  2. Implement SQLite store matching table `group_post_comments` created by migrations.
  3. Implement `CreateGroupPostCommentCommand` (ensures group membership) and `GetGroupPostCommentsQuery`.
  4. Register routes `POST /api/group-posts/:id/comments` and `GET /api/group-posts/:id/comments` in `transport/http.go`.
- **Verification:** Unit tests verifying commenting on group posts and retrieving comment lists.

---

## BE-B (Backend B) Tickets

### S4-BE-77: Event: Entities & Repository Interface

- **Priority:** P0
- **Type:** Greenfield (New Module/Feature - Event system)
- **Assignee:** BE-B
- **Story Points:** 2
- **Description:** Establish domain structures for events and RSVP votes. Use separate `Option` entity per the architecture spec for extensibility (e.g. adding "maybe" later). As a greenfield backend feature, this implements the new event system for groups (creation, RSVP with multiple options) under `internal/event/`, publishing `event.created` to the event bus.
- **Detailed Steps:**
  - _Greenfield Note:_ Follow TDD (Red-Green-Refactor). Ensure the slice adheres strictly to boundary rules (D5) without importing other slices' stores/transports.
  1.  Create `internal/event/event.go`. Define `Event` (ID, GroupID, CreatorID, Title, Description, ScheduledTime, CreatedAt), `Option` (EventID, Label string, e.g. "going", "not_going"), and `EventRSVP` (EventID, UserID, OptionID, UpdatedAt).
- **Verification:** Build checks.

---

### S4-BE-78: Event: SQLite Store

- **Priority:** P0
- **Type:** Greenfield (New Module/Feature - Event system)
- **Assignee:** BE-B
- **Story Points:** 3
- **Dependencies:** S4-BE-77, S4-SD-16
- **Description:** Store query mappings for events. As a greenfield backend feature, this implements the new event system for groups (creation, RSVP with multiple options) under `internal/event/`, publishing `event.created` to the event bus.
- **Detailed Steps:**
  - _Greenfield Note:_ Follow TDD (Red-Green-Refactor). Ensure the slice adheres strictly to boundary rules (D5) without importing other slices' stores/transports.
  1.  Create `internal/event/store/sqlite.go`.
- **Verification:** Integration tests checking storage against the tables created by the `000007` migration.

---

### S4-BE-79: Event: Create Event Command

- **Priority:** P0
- **Type:** Greenfield (New Module/Feature - Event system)
- **Assignee:** BE-B
- **Story Points:** 3
- **Dependencies:** S4-BE-77
- **Description:** Create event in group, executing group membership constraints, and publishing to notifications. As a greenfield backend feature, this implements the new event system for groups (creation, RSVP with multiple options) under `internal/event/`, publishing `event.created` to the event bus.
- **Detailed Steps:**
  - _Greenfield Note:_ Follow TDD (Red-Green-Refactor). Ensure the slice adheres strictly to boundary rules (D5) without importing other slices' stores/transports.
  1. Create `internal/event/commands/create_event.go`.
  2. Define local `GroupMemberChecker` interface. Enforce creator is group member.
  3. Validate fields: title, description, time in future, minimum of 2 RSVP options.
  4. Publish `event.created` event on EventBus.
- **Verification:** Unit tests verifying validator constraints, member checking, and notifications triggering.

---

### S4-BE-80: Event: RSVP Command

- **Priority:** P1
- **Type:** Greenfield (New Module/Feature - Event system)
- **Assignee:** BE-B
- **Story Points:** 2
- **Dependencies:** S4-BE-77
- **Description:** Record RSVP choices (going/not going). As a greenfield backend feature, this implements the new event system for groups (creation, RSVP with multiple options) under `internal/event/`, publishing `event.created` to the event bus.
- **Detailed Steps:**
  - _Greenfield Note:_ Follow TDD (Red-Green-Refactor). Ensure the slice adheres strictly to boundary rules (D5) without importing other slices' stores/transports.
  1. Create `internal/event/commands/rsvp.go`. Upsert choices.
- **Verification:** Unit tests verifying updates to RSVP states.

---

### S4-BE-81: Event: List Group Events Query

- **Priority:** P1
- **Type:** Greenfield (New Module/Feature - Event system)
- **Assignee:** BE-B
- **Story Points:** 2
- **Dependencies:** S4-BE-77
- **Description:** List events under a group with aggregated vote tallies. As a greenfield backend feature, this implements the entirely new group ecosystem (creation, invites, join requests, group chat, group posts) under `internal/group/`, isolated from other domains.
- **Detailed Steps:**
  - _Greenfield Note:_ Follow TDD (Red-Green-Refactor). Ensure the slice adheres strictly to boundary rules (D5) without importing other slices' stores/transports.
  1. Create `internal/event/queries/list_group_events.go`.
- **Verification:** Test return outputs contain correct count aggregates.

---

### S4-BE-82: Event: HTTP Transport Routing

- **Priority:** P1
- **Type:** Greenfield (New Module/Feature - Event system)
- **Assignee:** BE-B
- **Story Points:** 2
- **Dependencies:** S4-BE-79..81
- **Description:** Bind HTTP handlers. As a greenfield backend feature, this implements the new event system for groups (creation, RSVP with multiple options) under `internal/event/`, publishing `event.created` to the event bus.
- **Detailed Steps:**
  - _Greenfield Note:_ Follow TDD (Red-Green-Refactor). Ensure the slice adheres strictly to boundary rules (D5) without importing other slices' stores/transports.
  1. Create `internal/event/transport/http.go`.
  2. Route `POST /api/groups/:id/events`, `GET /api/groups/:id/events`, `POST /api/events/:id/rsvp`.
- **Verification:** Integration test endpoints mapping.

---

## FE-A (Frontend A) Tickets

### S4-FE-20: Groups Directory Page

- **Priority:** P1
- **Type:** Greenfield (New Frontend UI)
- **Assignee:** FE-A
- **Story Points:** 3
- **Description:** Build list view browsing all existing groups. As a greenfield frontend task, this implements new Next.js UI components in `frontend/src/` utilizing shadcn/ui and Vanilla CSS, wiring them to the Next.js App Router.
- **Detailed Steps:**
  - _Greenfield Note:_ Use ESLint + Prettier for linting/formatting and ensure session cookies are handled securely without localStorage leakage.
  1. Render lists of groups with search triggers and links.
- **Verification:** Check visual outputs.

---

### S4-FE-21: Group Profile Page

- **Priority:** P1
- **Type:** Greenfield (New Frontend UI)
- **Assignee:** FE-A
- **Story Points:** 5
- **Description:** Build view rendering group headers details, admin controls panel, and join toggle buttons. As a greenfield frontend task, this implements new Next.js UI components in `frontend/src/` utilizing shadcn/ui and Vanilla CSS, wiring them to the Next.js App Router.
- **Detailed Steps:**
  - _Greenfield Note:_ Use ESLint + Prettier for linting/formatting and ensure session cookies are handled securely without localStorage leakage.
  1. Render details. If creator/admin -> render invitation search box and join request approvals table.
  2. If non-member -> show join group requests toggles.
- **Verification:** Verify layout modifications per role.

---

### S4-FE-22: Group Posts Feed

- **Priority:** P1
- **Type:** Greenfield (New Frontend UI)
- **Assignee:** FE-A
- **Story Points:** 5
- **Description:** Render posts list cards inside group profile views. As a greenfield frontend task, this implements new Next.js UI components in `frontend/src/` utilizing shadcn/ui and Vanilla CSS, wiring them to the Next.js App Router.
- **Detailed Steps:**
  - _Greenfield Note:_ Use ESLint + Prettier for linting/formatting and ensure session cookies are handled securely without localStorage leakage.
  1. Enforce locked feeds container layout for non-members.
  2. Render post feeds dialog form allowing public attachments posts creation for members.
- **Verification:** Lock triggers check.

---

### S4-FE-23: Group Chat Workspace

- **Priority:** P1
- **Type:** Greenfield (New Frontend UI)
- **Assignee:** FE-A
- **Story Points:** 5
- **Description:** Interactive instant message container workspace inside groups profile page. As a greenfield frontend task, this implements new Next.js UI components in `frontend/src/` utilizing shadcn/ui and Vanilla CSS, wiring them to the Next.js App Router.
- **Detailed Steps:**
  - _Greenfield Note:_ Use ESLint + Prettier for linting/formatting and ensure session cookies are handled securely without localStorage leakage.
  1. Connect to websocket groups chat topic channel. Render streams.
- **Verification:** WS message delivery verification.

---

## FE-B (Frontend B) Tickets

### S4-FE-24: Event Creation Dialog

- **Priority:** P1
- **Type:** Greenfield (New Frontend UI)
- **Assignee:** FE-B
- **Story Points:** 3
- **Description:** Form overlay allowing group members to schedule events. As a greenfield frontend task, this implements new Next.js UI components in `frontend/src/` utilizing shadcn/ui and Vanilla CSS, wiring them to the Next.js App Router.
- **Detailed Steps:**
  - _Greenfield Note:_ Use ESLint + Prettier for linting/formatting and ensure session cookies are handled securely without localStorage leakage.
  1. Fields: title, description, day/time, options input list.
- **Verification:** Check input validators triggers.

---

### S4-FE-25: Events List Component

- **Priority:** P1
- **Type:** Greenfield (New Frontend UI)
- **Assignee:** FE-B
- **Story Points:** 3
- **Description:** Render events widgets. As a greenfield frontend task, this implements new Next.js UI components in `frontend/src/` utilizing shadcn/ui and Vanilla CSS, wiring them to the Next.js App Router.
- **Detailed Steps:**
  - _Greenfield Note:_ Use ESLint + Prettier for linting/formatting and ensure session cookies are handled securely without localStorage leakage.
  1. Render widgets inside group sidebar showing scheduled events.
- **Verification:** Output listings.

---

### S4-FE-26: RSVP Switch Actions

- **Priority:** P1
- **Type:** Greenfield (New Frontend UI)
- **Assignee:** FE-B
- **Story Points:** 2
- **Description:** Selection checkboxes on events card showing going/not going toggling. As a greenfield frontend task, this implements new Next.js UI components in `frontend/src/` utilizing shadcn/ui and Vanilla CSS, wiring them to the Next.js App Router.
- **Detailed Steps:**
  - _Greenfield Note:_ Use ESLint + Prettier for linting/formatting and ensure session cookies are handled securely without localStorage leakage.
  1. Clicking posts option choose updates to `/api/events/:id/rsvp`.
- **Verification:** State tallies update verify.

---

### S4-FE-27: Group: Comment Components (Gap Fix)

- **Priority:** P1
- **Type:** Greenfield (New Frontend UI)
- **Assignee:** FE-B
- **Story Points:** 3
- **Dependencies:** S4-FE-22
- **Description:** Implement frontend layout dialog and accordion under group post cards to create and list comments. As a greenfield frontend task, this implements new Next.js UI components in `frontend/src/` utilizing shadcn/ui and Vanilla CSS, wiring them to the Next.js App Router.
- **Detailed Steps:**
  - _Greenfield Note:_ Use ESLint + Prettier for linting/formatting and ensure session cookies are handled securely without localStorage leakage.
  1. Create accordion component loading comments via `GET /api/group-posts/:id/comments` on expand.
  2. Create inline text form submitting commenting details to `/api/group-posts/:id/comments`.
- **Verification:** Confirm visual comments updating.

---

## SD-QA (System Design/QA) Tickets

### S4-SD-16: Platform: Group & Event Migrations (000006 & 000007)

- **Priority:** P0
- **Type:** Greenfield (New Module/Feature - DB Migrations)
- **Assignee:** SD-QA
- **Story Points:** 2
- **Dependencies:** S1-BE-06
- **Description:** Create the database migration files for the Group and Event vertical slices. As a greenfield backend feature, this implements the entirely new group ecosystem (creation, invites, join requests, group chat, group posts) under `internal/group/`, isolated from other domains.
- **Detailed Steps:**
  - _Greenfield Note:_ Follow TDD (Red-Green-Refactor). Ensure the slice adheres strictly to boundary rules (D5) without importing other slices' stores/transports.
  1. Create `db/migrations/000006_groups.up.sql` to create `groups`, `group_members`, `group_invitations`, `group_join_requests`, `group_chat_messages`, `group_posts`, and `group_post_comments` tables.
  2. Create `db/migrations/000006_groups.down.sql` to reverse.
  3. Create `db/migrations/000007_events.up.sql` to create `events`, `event_options`, and `event_rsvps` tables.
  4. Create `db/migrations/000007_events.down.sql` to reverse.
- **Verification:** Run up/down checks confirming schema updates execute.

---

### S4-SD-17: E2E: Complete Groups Workspace Journey

- **Priority:** P1
- **Type:** Testing/Verification
- **Assignee:** SD-QA
- **Story Points:** 5
- **Dependencies:** S4-BE-60
- **Description:** Core integration Playwright test validating group interaction flows.
- **Detailed Steps:**
  1. Script: Create Group -> Join members -> Post message -> Create Event -> Vote RSVP.
- **Verification:** Headless execution pass.
