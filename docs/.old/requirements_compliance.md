# Requirements Compliance Verification Matrix — arch_optimized_v2.md

This document provides a line-by-line cross-reference of every requirement from `docs/requirements/audit.md` and `docs/requirements/readme.md` against the finalized, optimized vertical slices architecture plan in `docs/plan/architecture/arch_optimized_v2.md`.

## Executive Summary

**Status:** 100% Compliant.
**Gap Analysis:** Zero missing requirements. Every specification from the README and the Audit checklist is explicitly addressed in the optimized architecture, including all bonus features.

---

## 1. Technical Stack & Infrastructure

| Requirement                                                       | Source                 | Addressed In Architecture                              | Status |
| :---------------------------------------------------------------- | :--------------------- | :----------------------------------------------------- | :----- |
| **JS Framework** (Next.js, Vue, Svelte, Mithril)                  | `readme.md` (Frontend) | Phase 6.1: Next.js App Router scaffold                 | ✅     |
| **Standard Go Packages Allowed**                                  | `readme.md` (Packages) | Compliant throughout backend codebase                  | ✅     |
| **Allowed Packages** (Gorilla WS, Migrate, SQLite3, Bcrypt, UUID) | `readme.md` (Packages) | Phase 3.2 (Gorilla), 2.4 (Migrate), 3.5 (Bcrypt, UUID) | ✅     |
| **Docker Compose** (Backend & Frontend containers)                | `readme.md` (Docker)   | Phase 7: 2 Services with ports 8080/3000               | ✅     |
| **Build Script for Docker**                                       | `audit.md` (Bonus)     | Phase 7: `scripts/docker-build.sh`                     | ✅     |
| **Architecture Separation** (Server, App, Database)               | `readme.md` (Backend)  | D1 (Vertical Slices: transport, core, platform)        | ✅     |
| **SQLite Usage**                                                  | `readme.md` (Sqlite)   | D4 (Factory Pattern), Phase 2.1                        | ✅     |
| **Database Migrations**                                           | `readme.md` (Migrate)  | Phase 2.4: Sequential migration runner                 | ✅     |
| **Migration Folder Structure** (Numbered up/down)                 | `readme.md` (Migrate)  | Phase 2.4: Numbered up/down SQL scripts                | ✅     |
| **DB Seeding Script**                                             | `audit.md` (Bonus)     | Phase 2.4: `000007_seed_data.up.sql`                   | ✅     |

## 2. Authentication & Security

| Requirement                                                                                 | Source             | Addressed In Architecture                               | Status |
| :------------------------------------------------------------------------------------------ | :----------------- | :------------------------------------------------------ | :----- |
| **Registration Form Elements** (Email, Pass, First/Last Name, DOB, Avatar, Nickname, About) | `readme.md` (Auth) | Phase 6.2, Phase 5 (user migration mapping fields)      | ✅     |
| **Duplicate User Prevention**                                                               | `audit.md` (Auth)  | Phase 5: `user/commands/register.go` handles validation | ✅     |
| **Session/Cookie Based Auth**                                                               | `readme.md` (Auth) | Phase 3.1: Session manager                              | ✅     |
| **Logout Functionality**                                                                    | `readme.md` (Auth) | Phase 5: `user/commands/logout.go`                      | ✅     |
| **Multi-browser Persistence**                                                               | `audit.md` (Auth)  | Phase 3.1: Server-side sessions mapped to cookies       | ✅     |
| **Third-party OAuth** (GitHub/Google)                                                       | `audit.md` (Bonus) | Phase 5: `oauth` vertical slice                         | ✅     |

## 3. Social & Profiles

| Requirement                                                        | Source                  | Addressed In Architecture                             | Status |
| :----------------------------------------------------------------- | :---------------------- | :---------------------------------------------------- | :----- |
| **Profile Display** (Info, Activity, Followers/Following)          | `readme.md` (Profile)   | Phase 5: `queries/get_profile.go`, `get_activity.go`  | ✅     |
| **Public vs Private Profiles**                                     | `readme.md` (Profile)   | Phase 5: `user/commands/toggle_privacy.go`            | ✅     |
| **Privacy Toggle Confirmation Popup**                              | `audit.md` (Bonus)      | Phase 6.3: Frontend confirmation popup                | ✅     |
| **Follow/Unfollow Mechanics**                                      | `readme.md` (Followers) | Phase 4.1: `commands/follow_user.go`                  | ✅     |
| **Public Follow** (Instant)                                        | `readme.md` (Followers) | Phase 4.1: Auto-follow + event publish                | ✅     |
| **Private Follow** (Request / Accept / Decline)                    | `readme.md` (Followers) | Phase 4.1: `accept_request.go` / `decline_request.go` | ✅     |
| **Unfollow Confirmation Popup**                                    | `audit.md` (Bonus)      | Phase 6.3: Frontend confirmation popup                | ✅     |
| **Profile Access Control** (Hide private profile if not following) | `audit.md` (Profile)    | Phase 5: `get_profile.go` lock screen check           | ✅     |

## 4. Content (Posts & Comments)

| Requirement                                                | Source              | Addressed In Architecture                              | Status |
| :--------------------------------------------------------- | :------------------ | :----------------------------------------------------- | :----- |
| **Create Posts & Comments**                                | `readme.md` (Posts) | Phase 5: `topic` and `comment` vertical slices         | ✅     |
| **Image & GIF Support**                                    | `readme.md` (Posts) | Phase 5: `create_topic.go`, MIME magic bytes check     | ✅     |
| **Post Privacy Options** (Public, Almost Private, Private) | `readme.md` (Posts) | Phase 5: `topic/commands/create_topic.go` enum         | ✅     |
| **Custom User Selection for Private Posts**                | `readme.md` (Posts) | Phase 5: `AllowedUser` entity mapped in topic creation | ✅     |

## 5. Groups & Events

| Requirement                                          | Source               | Addressed In Architecture                            | Status |
| :--------------------------------------------------- | :------------------- | :--------------------------------------------------- | :----- |
| **Create Group** (Title, Description)                | `readme.md` (Groups) | Phase 4.2: `commands/create_group.go`                | ✅     |
| **Invite Followers to Group**                        | `readme.md` (Groups) | Phase 4.2: `commands/invite_member.go`               | ✅     |
| **Join Request & Accept/Refuse**                     | `readme.md` (Groups) | Phase 4.2: `request_join.go`, `respond_join.go`      | ✅     |
| **Browse Groups**                                    | `readme.md` (Groups) | Phase 4.2: `queries/list_groups.go`                  | ✅     |
| **Group Posts & Comments**                           | `readme.md` (Groups) | Phase 4.2: `create_group_post.go` (membership check) | ✅     |
| **Group Events** (Title, Desc, Day/Time, 2+ Options) | `readme.md` (Groups) | Phase 4.3: `create_event.go`                         | ✅     |
| **Event Voting** (Going / Not Going)                 | `readme.md` (Groups) | Phase 4.3: `rsvp.go`                                 | ✅     |

## 6. Real-Time Chat & Notifications

| Requirement                                      | Source                     | Addressed In Architecture                                      | Status |
| :----------------------------------------------- | :------------------------- | :------------------------------------------------------------- | :----- |
| **Private Chat** (WebSockets)                    | `readme.md` (Chat)         | Phase 5: `chat/transport/ws.go`                                | ✅     |
| **Chat Authorization** (Must follow/be followed) | `readme.md` (Chat)         | Phase 5: `commands/send_private_msg.go` + `FollowChecker`      | ✅     |
| **Emojis in Chat**                               | `readme.md` (Chat)         | Phase 6.4: Unicode text support                                | ✅     |
| **Group Chat Room**                              | `readme.md` (Chat)         | Phase 4.2: `commands/send_group_message.go`                    | ✅     |
| **Notifications on all pages**                   | `readme.md` (Notification) | Phase 6.3: Notifications panel with WS updates                 | ✅     |
| **Event: Follow Request Notification**           | `readme.md` (Notification) | Phase 5: `consume_events.go` listens to `follow.requested`     | ✅     |
| **Event: Group Invite Notification**             | `readme.md` (Notification) | Phase 5: `consume_events.go` listens to `group.invited`        | ✅     |
| **Event: Group Join Request Notification**       | `readme.md` (Notification) | Phase 5: `consume_events.go` listens to `group.join_requested` | ✅     |
| **Event: Event Creation Notification**           | `readme.md` (Notification) | Phase 5: `consume_events.go` listens to `event.created`        | ✅     |

---

## Gap Analysis Conclusion

The `arch_optimized_v2.md` architecture is a **complete superset** of the specifications. It not only covers all baseline requirements and audit checklist items, but also gracefully integrates the required system stability fixes (from the Audit) into Phase 1, paving the way for the robust Vertical Slice feature implementation in Phase 4 and Phase 5.

---

## audit.md — Line-by-Line Verification

### Functional / Project Structure

| #   | Audit Question                                                                          | Verdict | Where in `docs/plan/architecture/arch_optimized_v2.md`                                                                                                                                                                                                                 |
| --- | --------------------------------------------------------------------------------------- | ------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| L3  | Allowed packages respected?                                                             | ✅      | Section "Allowed Packages". Standard Go packages, `gorilla/websocket`, `golang-migrate` (or custom migration logic), `sqlite3`, `bcrypt`, and `uuid` are used. Disallowed packages (Redis, RabbitMQ, PostgreSQL) are isolated in purely optional learning Phases 8-10. |
| L7  | Backend: well-organized structure, clear separation of packages and migrations folders? | ✅      | Section "Directory Tree — Final Target" (`internal/<feature>/` slices) and Phase 2.1 & 2.4 (`internal/platform/database/` and `db/migrations/`).                                                                                                                       |
| L9  | Frontend well organized?                                                                | ✅      | Phase 6.1 ("Phase 6: Next.js Frontend") defines directory tree targets for the Next.js frontend codebase (`src/app/`, `src/components/ui/`, `src/components/features/`, `src/styles/`).                                                                                |

### Backend 3-Tier Separation

| #   | Audit Question                                                | Verdict | Where in `docs/plan/architecture/arch_optimized_v2.md`                                                                                                                                                                            |
| --- | ------------------------------------------------------------- | ------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| L13 | Clear separation: **Server, App, Database**?                  | ✅      | Section "D1: Vertical Slices with CQRS" and "D5: Boundary Rules" detail the clean Separation of Concerns (SoC) within slices: Server (transport/HTTP/WS routing), App (commands/queries logic), Database (store/sqlite.go layer). |
| L15 | Server receives incoming requests, entry point?               | ✅      | Phase 3.4 (`internal/core/server/server.go`) defines the entry point that initializes routes and runs the HTTP server.                                                                                                            |
| L17 | App listens for requests, retrieves from DB, sends responses? | ✅      | Section "D1: Vertical Slices with CQRS". `transport/http.go` parses HTTP/WS request, delegates to `commands/` or `queries/`, retrieves from `store/sqlite.go`, and returns responses.                                             |
| L19 | Core logic in App, handling HTTP/other protocols?             | ✅      | Section "D1: Vertical Slices with CQRS" and target directory tree: all core logic (such as privacy checking, follow-rules, group membership validation) is in feature `commands/` and `queries/`.                                 |

### Database

| #   | Audit Question                                  | Verdict | Where in `docs/plan/architecture/arch_optimized_v2.md`                                                                                                                                  |
| --- | ----------------------------------------------- | ------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| L23 | SQLite being used?                              | ✅      | Section "D4: Database Access" and "Directory Tree — Final Target": `internal/platform/database/sqlite.go` implements primary SQLite database.                                           |
| L25 | Clients can request/submit data without errors? | ✅      | Feature `store/sqlite.go` files use safe SQL queries with parameterized statements. Phase 1 fixes pre-existing DB bugs (WAL mode, busy timeout, prepared statement execution contexts). |
| L27 | Migration system implemented?                   | ✅      | Phase 2.1 & 2.4: `internal/platform/database/migrations.go` reads and runs migrations sequentially.                                                                                     |
| L29 | Migration file system well organized?           | ✅      | Phase 2.4 details migrations in `db/migrations/` (or `backend/pkg/db/migrations/sqlite`) matching the numbered `*.up.sql` and `*.down.sql` format.                                      |
| L33 | Migrations applied by migration system?         | ✅      | Phase 2.1: Migrations are tracked using a `schema_migrations` table and executed on application startup.                                                                                |

### Authentication

| #   | Audit Question                                                                                                | Verdict | Where in `docs/plan/architecture/arch_optimized_v2.md`                                                                               |
| --- | ------------------------------------------------------------------------------------------------------------- | ------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| L37 | Sessions for authentication?                                                                                  | ✅      | Phase 3.1 & 3.3: `internal/core/session/` handles session logic, stored in SQLite database. Auth middleware handles session cookies. |
| L39 | Registration form: Email, Password, First Name, Last Name, DOB, Avatar (opt), Nickname (opt), About Me (opt)? | ✅      | Section "Directory Tree — Final Target" (`internal/user/commands/register.go`) and Phase 6.2 (Frontend Form).                        |
| L43 | Registration saves user to DB?                                                                                | ✅      | Feature `user/commands/register.go` saves the registered user to DB via `user/store/sqlite.go`.                                      |
| L47 | Login works?                                                                                                  | ✅      | `internal/user/commands/login.go` handles authentication.                                                                            |
| L51 | Wrong password/email detected?                                                                                | ✅      | `internal/user/commands/login.go` returns descriptive validation errors for wrong credentials.                                       |
| L55 | Duplicate email/user detected?                                                                                | ✅      | `internal/user/commands/register.go` checks database unique constraints on email and username/nickname.                              |
| L59 | Non-logged browser stays unregistered?                                                                        | ✅      | Phase 3.1: Session cookie-based authentication isolates sessions per browser client.                                                 |
| L63 | Both browsers keep right users?                                                                               | ✅      | Phase 3.1: Separate session cookies are managed per client connection/browser session.                                               |

### Followers

| #   | Audit Question                       | Verdict | Where in `docs/plan/architecture/arch_optimized_v2.md`                                                                                   |
| --- | ------------------------------------ | ------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| L69 | Send follow request to private user? | ✅      | `internal/follow/commands/follow_user.go` checks if profile `IsPrivate` and creates a pending follow request in `follow_requests` table. |
| L73 | Follow public user without request?  | ✅      | `internal/follow/commands/follow_user.go` automatically creates a follow relationship if target profile is public.                       |
| L77 | Accept/decline follow request?       | ✅      | `internal/follow/commands/accept_request.go` and `internal/follow/commands/decline_request.go`.                                          |
| L81 | Unfollow?                            | ✅      | `internal/follow/commands/unfollow_user.go`.                                                                                             |

### Profile

| #    | Audit Question                                       | Verdict | Where in `docs/plan/architecture/arch_optimized_v2.md`                                                                         |
| ---- | ---------------------------------------------------- | ------- | ------------------------------------------------------------------------------------------------------------------------------ |
| L87  | Profile shows all registration info except password? | ✅      | `internal/user/queries/get_profile.go` returning all profile fields.                                                           |
| L91  | Profile shows all posts by user?                     | ✅      | `internal/user/queries/get_activity.go` retrieves posts written by the user.                                                   |
| L95  | Profile shows followers and following?               | ✅      | `internal/user/queries/get_activity.go` delegates to follow store to retrieve follower/following lists.                        |
| L99  | Toggle private/public?                               | ✅      | `internal/user/commands/toggle_privacy.go` updates `is_private` flag in the users table.                                       |
| L103 | See followed private profile?                        | ✅      | `internal/user/queries/get_profile.go` checks follower status via `FollowChecker` interface before showing posts and activity. |
| L107 | Prevented from seeing non-followed private profile?  | ✅      | Same check — returns restricted view if not following.                                                                         |
| L111 | See non-followed public profile?                     | ✅      | `internal/user/queries/get_profile.go` bypasses follower check for public users.                                               |
| L115 | See followed public profile?                         | ✅      | `internal/user/queries/get_profile.go` bypasses follower check for public users.                                               |

### Posts

| #    | Audit Question                                 | Verdict | Where in `docs/plan/architecture/arch_optimized_v2.md`                                                                                                                                 |
| ---- | ---------------------------------------------- | ------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| L119 | Create post and comment?                       | ✅      | `internal/topic/commands/create_topic.go` and `internal/comment/commands/create_comment.go`.                                                                                           |
| L123 | Include image (JPG/PNG) or GIF in post?        | ✅      | `internal/topic/commands/create_topic.go` with magic-byte validation via `pkg/imgutil`.                                                                                                |
| L127 | Include image (JPG/PNG) or GIF in comment?     | ✅      | `internal/comment/commands/create_comment.go`.                                                                                                                                         |
| L131 | Post privacy: public, almost private, private? | ✅      | `internal/topic/commands/create_topic.go` supports setting visibility enum (`public`, `almost_private`, `private`) and inserting allowed followers to the `topic_allowed_users` table. |
| L133 | Private: specify allowed users?                | ✅      | Same as above. Allowed user list is verified before visibility.                                                                                                                        |

### Groups

| #    | Audit Question                                        | Verdict | Where in `docs/plan/architecture/arch_optimized_v2.md`                                                                       |
| ---- | ----------------------------------------------------- | ------- | ---------------------------------------------------------------------------------------------------------------------------- |
| L139 | Invite follower to group?                             | ✅      | `internal/group/commands/invite_member.go` and `internal/group/commands/respond_invite.go`.                                  |
| L143 | Group invitation accept/decline?                      | ✅      | Same as above.                                                                                                               |
| L147 | Group join request accept/decline?                    | ✅      | `internal/group/commands/request_join.go` and `internal/group/commands/respond_join.go`.                                     |
| L149 | Non-creator member can invite?                        | ✅      | `internal/group/commands/invite_member.go` permits any active member of the group to invite followers.                       |
| L151 | User can request to join?                             | ✅      | `internal/group/commands/request_join.go`.                                                                                   |
| L153 | Group member can create posts/comments?               | ✅      | `internal/group/commands/create_group_post.go` (includes membership validation checks).                                      |
| L157 | Event: title, description, day/time, going/not going? | ✅      | `internal/event/commands/create_event.go` validates title, description, time, and options (minimum "going" and "not going"). |
| L161 | Other user can see event and vote?                    | ✅      | `internal/event/commands/rsvp.go` and `internal/event/queries/list_group_events.go`.                                         |

### Chat

| #    | Audit Question                                     | Verdict | Where in `docs/plan/architecture/arch_optimized_v2.md`                                                                   |
| ---- | -------------------------------------------------- | ------- | ------------------------------------------------------------------------------------------------------------------------ |
| L167 | Private message received in realtime?              | ✅      | `internal/chat/commands/send_private_msg.go` dispatches via `internal/core/realtime` WebSocket hub.                      |
| L171 | Cannot chat between non-following users?           | ✅      | `internal/chat/commands/send_private_msg.go` performs a check against `FollowChecker` interface.                         |
| L175 | Chat doesn't crash server?                         | ✅      | Phase 1.7 (applies to `internal/core/realtime/client.go`) recovery middleware wraps goroutines.                          |
| L179 | Only targeted user receives PM?                    | ✅      | `internal/core/realtime/hub.go` delivers directly to connection associated with `recipientID`.                           |
| L183 | Group chat: all group members receive in realtime? | ✅      | `internal/group/commands/send_group_message.go` broadcasts via the WS hub to connections matching current group members. |
| L187 | Group chat doesn't crash?                          | ✅      | Same panic recovery as above.                                                                                            |
| L189 | Emojis via chat?                                   | ✅      | UTF-8 format text columns in SQLite database and JSON-payload WS transfer.                                               |

### Notifications

| #    | Audit Question                       | Verdict | Where in `docs/plan/architecture/arch_optimized_v2.md`                                                                                 |
| ---- | ------------------------------------ | ------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| L193 | Notifications visible on every page? | ✅      | Phase 6.3 (Navbar layout includes live notifications panel fetching from `/api/notifications` and listening via SSE or WS connection). |
| L197 | Notification: follow request?        | ✅      | `internal/notification/commands/consume_events.go` handles `follow.requested` event.                                                   |
| L201 | Notification: group invitation?      | ✅      | `internal/notification/commands/consume_events.go` handles `group.invited` event.                                                      |
| L205 | Notification: group join request?    | ✅      | `internal/notification/commands/consume_events.go` handles `group.join_requested` event.                                               |
| L209 | Notification: event created?         | ✅      | `internal/notification/commands/consume_events.go` handles `event.created` event.                                                      |

### Docker

| #    | Audit Question                                       | Verdict | Where in `docs/plan/architecture/arch_optimized_v2.md`                                                                                 |
| ---- | ---------------------------------------------------- | ------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| L215 | Two containers (backend + frontend), non-zero sizes? | ✅      | Phase 7 defines exactly 2 Docker services in `docker-compose.yml`: `backend` (exposing port 8080) and `frontend` (exposing port 3000). |
| L219 | Accessible via browser after docker?                 | ✅      | Phase 7: Ports 3000 and 8080 are exposed, enabling browser access at `http://localhost:3000`.                                          |

### Bonus

| #    | Audit Question                             | Verdict | Where in `docs/plan/architecture/arch_optimized_v2.md`                               |
| ---- | ------------------------------------------ | ------- | ------------------------------------------------------------------------------------ |
| L223 | OAuth (GitHub/external)?                   | ✅      | `internal/oauth/` slice supporting GitHub and Google OAuth providers.                |
| L225 | Migration to fill database (seed)?         | ✅      | Phase 2.1 migration manager supports loading schema and initial seed data SQL files. |
| L227 | Unfollow confirmation popup?               | ✅      | Phase 6.3 (frontend component actions).                                              |
| L229 | Profile privacy change confirmation popup? | ✅      | Phase 6.3 (frontend component actions).                                              |
| L231 | Additional notifications?                  | ✅      | Notification on follow request accepted (`follow.accepted`).                         |
| L233 | Build script for images/containers?        | ✅      | Phase 7 and "Verification Checklist" (`Makefile` and `docker-compose` wrappers).     |

---

## readme.md — Requirement Verification

### Frontend (L16-36)

| Requirement                                     | Verdict | Notes                                                |
| ----------------------------------------------- | ------- | ---------------------------------------------------- |
| Use a JS framework (Next.js/Vue/Svelte/Mithril) | ✅      | Next.js is configured and fully detailed in Phase 6. |

### Backend 3-Part Structure (L39-58)

| Requirement                                      | Verdict | Notes                                                                           |
| ------------------------------------------------ | ------- | ------------------------------------------------------------------------------- |
| Server: receives requests, entry point           | ✅      | `internal/core/server/server.go` serves as request entry point.                 |
| App: listens, retrieves from DB, sends responses | ✅      | Feature slices execute commands/queries on requests.                            |
| Database: organize and store data                | ✅      | `internal/platform/database/sqlite.go` + feature repositories.                  |
| Sessions and cookies for auth                    | ✅      | `internal/core/session/` stores sessions in DB and manages session cookies.     |
| Image handling (JPEG, PNG, GIF)                  | ✅      | Verified by checking magic bytes via `pkg/imgutil/` and storage in filesystems. |
| WebSocket for realtime                           | ✅      | `internal/core/realtime/` provides WebSocket connection hub.                    |

### SQLite (L60-66)

| Requirement            | Verdict | Notes                                                                  |
| ---------------------- | ------- | ---------------------------------------------------------------------- |
| Use SQLite as database | ✅      | Default primary database driver configured with WAL and busy timeouts. |

### Migrations (L67-99)

| Requirement                                  | Verdict | Notes                                                                      |
| -------------------------------------------- | ------- | -------------------------------------------------------------------------- |
| Migrations create tables on app run          | ✅      | Runs during bootstrap in `internal/platform/database/migrations.go`.       |
| Folder structure: numbered up/down SQL files | ✅      | Located at `db/migrations/sqlite/` or `backend/pkg/db/migrations/sqlite/`. |
| `sqlite.go` for connection + migration apply | ✅      | Handled by `internal/platform/database/sqlite.go` and `migrations.go`.     |

### Docker (L101-121)

| Requirement                                            | Verdict | Notes                                                          |
| ------------------------------------------------------ | ------- | -------------------------------------------------------------- |
| Two Docker images (backend + frontend)                 | ✅      | Addressed in Phase 7 (`Dockerfile` and `frontend/Dockerfile`). |
| Backend container: server logic, handle requests, DB   | ✅      | Custom Go backend image.                                       |
| Frontend container: serve HTML/CSS/JS, HTTP to backend | ✅      | Custom Next.js frontend image.                                 |
| Expose necessary ports                                 | ✅      | Port 8080 (backend) and 3000 (frontend).                       |

### Authentication (L123-141)

| Requirement                                                                                             | Verdict | Notes                                                                |
| ------------------------------------------------------------------------------------------------------- | ------- | -------------------------------------------------------------------- |
| Registration: Email, Password, First Name, Last Name, DOB, Avatar (opt), Nickname (opt), About Me (opt) | ✅      | Handled by `internal/user/commands/register.go` entity validation.   |
| Sessions and cookies                                                                                    | ✅      | Cookie-based session validation in auth middleware.                  |
| Stay logged in until logout                                                                             | ✅      | Persistent session validation in DB, with cookie clearing on logout. |

### Followers (L143-149)

| Requirement                                  | Verdict | Notes                                                    |
| -------------------------------------------- | ------- | -------------------------------------------------------- |
| Follow and unfollow                          | ✅      | Handled by `internal/follow/commands/`.                  |
| Follow request → accept/decline              | ✅      | Handled by follow request workflows in private profiles. |
| Public profile: auto-follow (bypass request) | ✅      | Auto-accept when target user is public.                  |

### Profile (L151-164)

| Requirement                                         | Verdict | Notes                                                  |
| --------------------------------------------------- | ------- | ------------------------------------------------------ |
| User info (all registration fields except password) | ✅      | Returned by `internal/user/queries/get_profile.go`.    |
| User activity (posts)                               | ✅      | Returned by `internal/user/queries/get_activity.go`.   |
| Followers and following lists                       | ✅      | Returned by follow repository queries.                 |
| Public vs private profile                           | ✅      | Controlled by `is_private` database flag.              |
| Toggle public/private                               | ✅      | Handled by `internal/user/commands/toggle_privacy.go`. |

### Posts (L166-177)

| Requirement                              | Verdict | Notes                                                        |
| ---------------------------------------- | ------- | ------------------------------------------------------------ |
| Create posts and comments                | ✅      | Handled by `internal/topic/` and `internal/comment/` slices. |
| Include image or GIF                     | ✅      | Image upload validation using magic bytes.                   |
| Privacy: public, almost private, private | ✅      | Visibility scopes stored and checked on queries.             |
| Private: choose followers who can see    | ✅      | Saved to `topic_allowed_users` table and evaluated.          |

### Groups (L179-199)

| Requirement                                                 | Verdict | Notes                                                 |
| ----------------------------------------------------------- | ------- | ----------------------------------------------------- |
| Create group (title, description)                           | ✅      | Handled by `internal/group/commands/create_group.go`. |
| Invite users, accept/decline                                | ✅      | Handled by `internal/group/commands/`.                |
| Members can invite others                                   | ✅      | Supported for all active group members.               |
| Request to join, creator accepts/declines                   | ✅      | Supported for the creator/owner of the group.         |
| Browse all groups                                           | ✅      | Handled by `internal/group/queries/list_groups.go`.   |
| Group posts and comments (members only)                     | ✅      | Membership check validation in queries and commands.  |
| Group events: title, description, day/time, going/not going | ✅      | Handled by `internal/event/commands/create_event.go`. |
| All users can choose event option                           | ✅      | RSVP votes updated in `event_rsvps` table.            |

### Chat (L201-211)

| Requirement                                  | Verdict | Notes                                                 |
| -------------------------------------------- | ------- | ----------------------------------------------------- |
| Private messages to followed/following users | ✅      | Validates active follow connections before chat init. |
| Realtime via WebSocket                       | ✅      | WS handlers coordinate with client hub.               |
| Emojis                                       | ✅      | Standard Unicode/UTF-8 support.                       |
| Group chat room                              | ✅      | Broadcasts messages to all connected group members.   |

### Chat / Realtime Requirements

| Requirement                                            | Verdict | Notes                                                                                        |
| ------------------------------------------------------ | ------- | -------------------------------------------------------------------------------------------- |
| Live messages through WebSocket on follow or if public | ✅      | Core chat feature uses WS connections, following checks are performed before initialization. |

### Notifications (L213-226)

| Requirement                                   | Verdict | Notes                                                              |
| --------------------------------------------- | ------- | ------------------------------------------------------------------ |
| Visible on every page                         | ✅      | Real-time navbar notification updates via SSE/WebSocket.           |
| Follow request notification (private profile) | ✅      | Event bus event `follow.requested` triggers notification save.     |
| Group invitation notification                 | ✅      | Event bus event `group.invited` triggers notification save.        |
| Group join request notification (creator)     | ✅      | Event bus event `group.join_requested` triggers notification save. |
| Event created notification (group members)    | ✅      | Event bus event `event.created` triggers notification save.        |

### Allowed Packages (L229-238)

| Package                                    | Verdict | Notes                                                                  |
| ------------------------------------------ | ------- | ---------------------------------------------------------------------- |
| Standard Go packages                       | ✅      | Used for core transport, validation, and structures.                   |
| gorilla/websocket                          | ✅      | Handles real-time transport client connections.                        |
| golang-migrate / sql-migration / migration | ✅      | Custom runner mimicking `golang-migrate` structure in platform folder. |
| sqlite3 (mattn/go-sqlite3)                 | ✅      | Standard DB driver utilized inside the SQLite platform service.        |
| bcrypt (golang.org/x/crypto/bcrypt)        | ✅      | Secures passwords inside the user service layer.                       |
| gofrs/uuid or google/uuid                  | ✅      | Generates entity UUID keys inside the application package.             |

---

## Resolved Architectural Gaps

Every gap identified in previous iterations has been resolved in the optimized architecture plan `docs/plan/architecture/arch_optimized_v2.md`:

### 1. Allowed Packages (Redis, RabbitMQ, PostgreSQL drivers)

- **Status**: **RESOLVED**
- **Details**: To strictly respect the "Allowed Packages" audit question, the core vertical slices (Phases 1-7) utilize only standard library tools, `go-sqlite3`, `bcrypt`, `uuid`, and `gorilla/websocket`. The advanced storage engine (PostgreSQL), message broker (RabbitMQ), and cache (Redis) layers are structured as purely **optional learning plug-ins** in Phases 8, 9, and 10. They are fully isolated from the core codebase.

### 2. "Server, App, Database" Architectural Mapping

- **Status**: **RESOLVED**
- **Details**: The vertical slices cleanly map to these three tiers per feature:
  - **Server**: The routing/transport layer is situated in `transport/http.go` and `internal/core/server/`.
  - **App**: Core business logic and rules are strictly isolated inside individual CQRS use-case command and query files.
  - **Database**: SQL execution is cleanly separated inside `store/sqlite.go`.

### 3. Docker "Two Containers" Limitation

- **Status**: **RESOLVED**
- **Details**: In Phase 7 (completion of core spec), the `docker-compose.yml` configures exactly **two containers** (`backend` and `frontend`). The auxiliary databases/queues are not required in the core composition, resolving any compliance check on container count.

---

## Summary

| Category                   | Total Questions | ✅ Pass | ❌ Gap |
| -------------------------- | :-------------: | :-----: | :----: |
| Functional / Structure     |        3        |    3    |   0    |
| Backend                    |        4        |    4    |   0    |
| Database                   |        5        |    5    |   0    |
| Authentication             |        9        |    9    |   0    |
| Followers                  |        4        |    4    |   0    |
| Profile                    |        8        |    8    |   0    |
| Posts                      |        5        |    5    |   0    |
| Groups                     |        7        |    7    |   0    |
| Chat                       |        7        |    7    |   0    |
| Notifications              |        5        |    5    |   0    |
| Docker                     |        2        |    2    |   0    |
| Bonus                      |        6        |    6    |   0    |
| **readme.md** Requirements |       all       |   all   |   0    |
| **Total**                  |     **65+**     | **65+** | **0**  |
