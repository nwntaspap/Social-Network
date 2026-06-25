# Sprint Plan Verification — Against Target Architecture

Date: 2026-06-18
Source: `docs/architecture/target-architecture-with-phases.md`

## Summary

All 7 sprints correctly cover all 7 architecture phases (Phases 8-10 optional, correctly excluded). Core mapping:

| Architecture Phase  | Sprint(s)                | Verdict                          |
| ------------------- | ------------------------ | -------------------------------- |
| Phase 1: Bugs (8)   | Sprint 0                 | Covered. 2 path errors.          |
| Phase 2: Platform   | Sprint 1                 | Covered.                         |
| Phase 3: Core       | Sprint 1                 | Covered.                         |
| Phase 4: Greenfield | Sprint 3, 4              | Covered.                         |
| Phase 5: Migration  | Sprint 2, 3, 5           | Covered.                         |
| Phase 6: Frontend   | All sprints (FE tickets) | Covered.                         |
| Phase 7: Docker     | Sprint 6                 | Covered but partly pre-existing. |

**18 issues found.** 6 blocking, 5 medium, 7 minor.

---

## Sprint 0: Foundation (Phase 1 — Bug Fixes)

### Coverage: 8/8 bugs, no gaps

| Bug ID                | Sprint Ticket | Architect Doc Path                        | Actual Path                                            | Match?          |
| --------------------- | ------------- | ----------------------------------------- | ------------------------------------------------------ | --------------- |
| 1.1 Migration `:`→`;` | S0-BE-04      | `infra/storage/sqlite/init.go`            | `internal/infra/storage/sqlite/init.go`                | Path fix needed |
| 1.2 SQLite WAL        | S0-BE-04      | `init.go`, `.env`                         | same (`internal/` prefixed)                            | Path fix needed |
| 1.3 OAuth Scan(ctx)   | S0-BE-05      | `infra/storage/sqlite/oauth/oauthRepo.go` | `internal/infra/storage/sqlite/oauth/oauthRepo.go`     | Path fix needed |
| 1.4 WS CheckOrigin    | S0-BE-05      | `infra/ws/handler.go`                     | `internal/infra/http/ws/handler.go`                    | **WRONG PATH**  |
| 1.5 SQL injection     | S0-BE-04      | `sqlite/topics/topicRepo.go`              | `internal/infra/storage/sqlite/topics/topicRepo.go`    | Path fix needed |
| 1.6 Prepared stmt     | S0-BE-05      | `sqlite/users/userRepo.go`                | `internal/infra/storage/sqlite/users/userRepo.go`      | Path fix needed |
| 1.7 WS panic recovery | S0-BE-05      | `infra/ws/client.go`                      | `internal/infra/ws/client.go`                          | Path fix needed |
| 1.8 RateLimiter leak  | S0-BE-05      | `middleware/ratelimiter/rateLimiter.go`   | `internal/infra/middleware/ratelimiter/rateLimiter.go` | Path fix needed |

### Issues

1. **[BLOCKING] Wrong WS path for 1.4**: Architecture doc and Sprint S0-BE-05 both say `infra/ws/handler.go`. Actual file is `internal/infra/http/ws/handler.go`. `infra/ws/` has `client.go`, not `handler.go`. Fix: update both architecture doc Phase 1 table and S0-BE-05 to correct path.

2. **[MEDIUM] All paths missing `internal/` prefix**: Architecture doc Phase 1 table drops the `internal/` prefix. Sprint 0 tickets copied this. Fix: normalize all paths to `internal/...` throughout.

3. **[MINOR] Extra scaffold tickets**: S0-BE-01 (Go scaffold), S0-BE-02 (Makefile), S0-BE-03 (golangci-lint) are not in Phase 1. These are reasonable setup prerequisites. Keep.

4. **[MINOR] Codebase already has structure**: `cmd/server/main.go`, `internal/bootstrap/bootstrap.go`, `internal/domain/`, `internal/app/`, `internal/infra/` already exist. S0-BE-01 says "Create the target structure" but target dirs already partially exist. The scaffold ticket should clarify it's adjusting existing layout, not greenfield.

---

## Sprint 1: Platform & Core (Phases 2+3)

### Coverage: Full

| Phase     | Architecture Items | Sprint Tickets | Status         |
| --------- | ------------------ | -------------- | -------------- |
| Phase 2.1 | DB Factory         | S1-BE-01       | Covered        |
| Phase 2.2 | EventBus           | S1-BE-02       | Covered        |
| Phase 2.3 | Cache              | S1-BE-03       | Covered        |
| Phase 2.4 | Migration system   | S1-BE-04       | Covered        |
| Phase 3.1 | Session            | S1-BE-05       | Covered        |
| Phase 3.2 | Realtime WS        | S1-BE-06       | Covered        |
| Phase 3.3 | Middleware         | S1-BE-07       | Covered        |
| Phase 3.4 | Server             | S1-BE-08       | Covered        |
| Phase 3.5 | OAuth rename       | S1-BE-09       | Covered        |
| Phase 3.5 | imgutil            | S1-BE-10       | Covered        |
| — (extra) | DB Seeding         | S1-BE-11       | Reasonable add |

### Codebase Verification

| New Dir               | Status                   | Existing Old Code                                                                                                           |
| --------------------- | ------------------------ | --------------------------------------------------------------------------------------------------------------------------- |
| `internal/platform/`  | Exists but EMPTY         | —                                                                                                                           |
| `internal/core/`      | Does NOT exist           | `internal/infra/middleware/`, `internal/infra/ws/`, `internal/infra/storage/sessionstore/`, `internal/infra/http/server.go` |
| `internal/pkg/oauth/` | Does NOT exist           | `internal/pkg/oAuth/` (camelCase)                                                                                           |
| `db/migrations/`      | Exists with wrong format | `schema.sql` + `indexes.sql`, NOT numbered up/down files                                                                    |

### Issues

5. **[BLOCKING] Migration format mismatch**: Codebase has `db/migrations/schema.sql` + `indexes.sql`. Architecture doc specifies numbered `000001_initial_schema.up.sql`/`.down.sql` format. S1-BE-04 must convert existing schema into the numbered format, not just create 000001 from scratch.

6. **[MEDIUM] S1-BE-01 SQLite pooling gap**: Sprint 1 adds `db.SetMaxOpenConns(1)` (Step 5). This is NOT in architecture Phase 2. Reasonable addition for SQLite safety. Keep.

7. **[MEDIUM] Migration files 000002-000007 not spread across sprints**: Sprint 1 only creates 000001 + 000007 (seed). Migration files 000002-000006 are described in Phase 2.4 but no sprint creates them. Architecture doc says 000002 (user profile fields), 000003 (topic privacy), 000004 (follow system), 000005 (groups), 000006 (events). Each should be created by the sprint that builds that feature.
   - **Fix**: Add migration creation tickets to Sprint 2 (000002+000003), Sprint 3 (000004), Sprint 4 (000005+000006). Or consolidate into S1-BE-04.

8. **[MINOR] S1-BE-09 OAuth rename timing**: Sprint 1 renames `pkg/oAuth/` → `pkg/oauth/`. But old code at `internal/pkg/oAuth/` is imported everywhere in the old bootstrap. Renaming means temporarily breaking old code OR renaming after old imports are removed. Safer: do rename in Sprint 5 alongside OAuth migration, not in Sprint 1.

---

## Sprint 2: User & Topic (Phase 5 — Migration)

### Coverage: Full

| Feature | Architecture Commands                                      | Sprint Tickets | Count Match |
| ------- | ---------------------------------------------------------- | -------------- | ----------- |
| User    | 5: register, login, logout, update_profile, toggle_privacy | S2-BE-03..07   | 5 ✓         |
| User    | 3 queries: get_profile, get_activity, list_users           | S2-BE-08..10   | 3 ✓         |
| Topic   | 2: create_topic, cast_vote                                 | S2-BE-15..16   | 2 ✓         |
| Topic   | 4 queries: get_feed, get_user_topics, get_topic, get_votes | S2-BE-17..20   | 4 ✓         |

### Codebase: Old Code to Migrate

| Feature | Domain                    | App                                 | Infra HTTP                       | Storage SQLite   | Absorbs                                                             |
| ------- | ------------------------- | ----------------------------------- | -------------------------------- | ---------------- | ------------------------------------------------------------------- |
| User    | `domain/user/` (2 files)  | `app/user/` (commands/, queries/)   | `infra/http/user/` (4 handlers)  | `sqlite/users/`  | `domain/activity/`, `app/activities/`                               |
| Topic   | `domain/topic/` (2 files) | `app/topics/` (commands/, queries/) | `infra/http/topic/` (5 handlers) | `sqlite/topics/` | `domain/category/`, `app/categories/`, `domain/vote/`, `app/votes/` |

### Issues

9. **[MEDIUM] Activity absorption**: Architecture Phase 5 says user absorbs `activity/`. Sprint 2 has S2-BE-09 (get_activity query) which handles this. Old activity code at `domain/activity/`, `app/activities/`, `infra/http/activity/`, `sqlite/activity/`. S2-BE-09 description says "retrieve list of posts" but activity in old code also includes follower/following lists. Clarify scope.

10. **[MINOR] Category/Vote absorption not explicit in tickets**: Architecture Phase 5 says topic absorbs category+vote. Sprint 2 tickets handle this implicitly (S2-BE-13 defines Vote entity, S2-BE-16 cast_vote command) but no ticket explicitly says "absorb category entities into Topic". The `Category` entity from old code should be merged into Topic or replaced by `Visibility` enum. Clarify.

11. **[MINOR] Missing migration file tickets**: 000002_user_profile_fields and 000003_topic_privacy migrations are in Phase 2.4 spec but no sprint ticket creates them. Should be in Sprint 2.

---

## Sprint 3: Follow, Comment, Notification (Phase 4+5)

### Coverage: Full

| Feature      | Type       | Arch Spec              | Sprint Tickets | Count |
| ------------ | ---------- | ---------------------- | -------------- | ----- |
| Follow       | Greenfield | 4 commands + 4 queries | S3-BE-01..12   | 12 ✓  |
| Comment      | Migration  | 1 command + 1 query    | S3-BE-13..18   | 6 ✓   |
| Notification | Migration  | 2 commands + 1 query   | S3-BE-19..24   | 6 ✓   |

### Codebase: Old Code to Migrate

| Feature      | Domain                 | App                  | Infra HTTP                 | Storage SQLite          | WS Handlers |
| ------------ | ---------------------- | -------------------- | -------------------------- | ----------------------- | ----------- |
| Comment      | `domain/comment/`      | `app/comments/`      | `infra/http/comment/`      | `sqlite/comments/`      | —           |
| Notification | `domain/notification/` | `app/notifications/` | `infra/http/notification/` | `sqlite/notifications/` | —           |
| Follow       | — (greenfield)         | —                    | —                          | —                       | —           |

### Issues

12. **[BLOCKING] Notification event subscription ordering**: Sprint 3 builds Notification which subscribes to `group.invited`, `group.join_requested`, `event.created` events (S3-BE-21). But those events are published by Group/Event features built in Sprint 4. The subscriber code must register handlers for event types that don't fire yet.

- **Not a blocker if**: the eventbus implementation simply registers subscriptions (no events flow until Sprint 4 publishes them).
- **Mitigation**: S3-BE-21 should write subscription registration code, test with synthetic/mock events, and note that real events arrive in Sprint 4.

13. **[MEDIUM] Comment contract tests reference old domain**: S3-BE-18 says "Ensure comments vertical slice compatibility with old domain". But by Sprint 3, old domain packages still exist (they're deleted in Sprint 6). This is correct Strangler Fig approach. Tests can reference both old and new code.

14. **[MINOR] Missing 000004_follow_system migration**: Should be created in Sprint 3. No ticket covers this.

---

## Sprint 4: Group & Event (Phase 4 — Greenfield)

### Coverage: Full

| Feature | Architecture Spec                  | Sprint Tickets | Count |
| ------- | ---------------------------------- | -------------- | ----- |
| Group   | 7 commands + 4 queries + HTTP + WS | S4-BE-01..15   | 15 ✓  |
| Event   | 2 commands + 1 query + HTTP        | S4-BE-16..21   | 6 ✓   |

### Codebase: No existing code (greenfield)

Both `internal/group/` and `internal/event/` do not exist. No old code to migrate. Correct.

### Issues

15. **[MEDIUM] Missing migration files**: 000005_groups and 000006_events not created in any sprint ticket. Should be in Sprint 4.

16. **[MINOR] Event depends on GroupChecker**: Architecture Phase 4 says Event "Requires GroupMemberChecker interface (defined locally, satisfied by group store)". Sprint 4 ordering is correct — Group is built first (BE-A), Event second (BE-B). Dependency handled. Good.

---

## Sprint 5: Chat & OAuth (Phase 5 — Migration)

### Coverage: Full

| Feature | Architecture Spec                 | Sprint Tickets | Count |
| ------- | --------------------------------- | -------------- | ----- |
| Chat    | 1 command + 2 queries + HTTP + WS | S5-BE-01..08   | 8 ✓   |
| OAuth   | 2 commands + HTTP                 | S5-BE-09..16   | 8 ✓   |

### Codebase: Old Code to Migrate

| Feature | Domain                    | App                               | Infra HTTP          | Storage SQLite  | WS/Other                                  |
| ------- | ------------------------- | --------------------------------- | ------------------- | --------------- | ----------------------------------------- |
| Chat    | `domain/chat/` (4 files)  | `app/chat/` (commands/, queries/) | `infra/http/chat/`  | `sqlite/chats/` | `infra/ws/handlers/` (8 WS handler files) |
| OAuth   | `domain/oauth/` (3 files) | `app/oauth/` (1 file)             | `infra/http/oauth/` | `sqlite/oauth/` | `pkg/oAuth/` (camelCase)                  |

### Issues

17. **[BLOCKING] OAuth package rename timing**: Architecture Phase 3 (Sprint 1) says rename `pkg/oAuth/` → `pkg/oauth/`. Sprint 1 S1-BE-09 does this. Sprint 5 S5-BE-09..16 creates OAuth slice using the new paths. But:
    - Old bootstrap.go still imports `internal/pkg/oAuth` (camelCase)
    - If Sprint 1 renames the package, old code breaks immediately
    - **Fix**: Do the rename in Sprint 5 alongside OAuth migration. Sprint 1 should NOT include S1-BE-09; move it to Sprint 5.

18. **[MINOR] GitHub/Google client tickets**: S5-BE-14 (GitHub) and S5-BE-15 (Google) create files at `pkg/oauth/github/client.go` and `pkg/oauth/google/client.go`. These already exist at `internal/pkg/oAuth/githubclient/` and `internal/pkg/oAuth/googleclient/`. Migration is just rename + path update. Tickets should clarify this is a move, not new code.

---

## Sprint 6: Cleanup, Docker, Verification (Phase 5 Cleanup + Phase 7)

### Coverage: Full

| Area                      | Sprint Tickets | Status  |
| ------------------------- | -------------- | ------- |
| Delete `internal/domain/` | S6-BE-01       | Covered |
| Delete `internal/app/`    | S6-BE-02       | Covered |
| Delete `internal/infra/`  | S6-BE-03       | Covered |
| Bootstrap wiring          | S6-BE-04       | Covered |
| Integration tests         | S6-BE-05       | Covered |
| Benchmark                 | S6-BE-06       | Covered |
| Boundary checks           | S6-BE-07       | Covered |
| Audit.md tests            | S6-BE-08       | Covered |
| Docker production         | S6-DEV-01..05  | Covered |

### Codebase: Pre-Existing Docker Setup

Dockerfile, docker-compose.yml, docker-compose.dev.yml already exist. Current setup is single-service "forum" combining frontend+backend. Phase 7 specifies 2 separate services (backend + frontend). Sprint 6 must rewrite Docker config, not create from scratch.

### Issues

19. **[BLOCKING] Docker already exists but wrong architecture**: Current `docker-compose.yml` has a single `forum` service serving both frontend (port 3001) and backend (port 8080). Architecture Phase 7 wants 2 services: `backend` and `frontend`. Sprint 6 S6-DEV-01 must rewrite, not create.

20. **[MEDIUM] Bootstrap complexity**: S6-BE-04 says "Complete final wiring of all 10 features". The old bootstrap.go wires 9 repos + notifier + hub + fileStorage + middleware + oAuth + session. New bootstrap must wire 10 feature slices each with their own store + transport + cross-slice interfaces. Single ticket at 5 SP seems undersized. Consider splitting into bootstrap per slice.

21. **[MINOR] Audit.md test scope**: S6-BE-08 and S6-FE-07 reference `docs/requirements/audit.md`. Verify this file exists. If not, tickets cannot be completed.

---

## Cross-Cutting Issues

### File Path Convention

Architecture doc uses paths without `internal/` prefix throughout:

- `infra/storage/sqlite/init.go` → actually `internal/infra/storage/sqlite/init.go`
- `infra/ws/handler.go` → actually `internal/infra/http/ws/handler.go`
- `middleware/ratelimiter/rateLimiter.go` → actually `internal/infra/middleware/ratelimiter/rateLimiter.go`

All sprint plan path references copied this. Fix: normalize ALL path references to include `internal/` prefix in both architecture doc and sprint plans.

### Migration File Distribution

| Migration                  | Phase   | Sprint That Should Create It | Currently Covered? |
| -------------------------- | ------- | ---------------------------- | ------------------ |
| 000001_initial_schema      | Phase 2 | Sprint 1 (S1-BE-04)          | ✓                  |
| 000002_user_profile_fields | Phase 2 | Sprint 2 (User migration)    | ✗                  |
| 000003_topic_privacy       | Phase 2 | Sprint 2 (Topic migration)   | ✗                  |
| 000004_follow_system       | Phase 2 | Sprint 3 (Follow)            | ✗                  |
| 000005_groups              | Phase 2 | Sprint 4 (Group)             | ✗                  |
| 000006_events              | Phase 2 | Sprint 4 (Event)             | ✗                  |
| 000007_seed_data           | Phase 2 | Sprint 1 (S1-BE-11)          | ✓                  |

5 of 7 migration files not assigned to any sprint. Fix: add migration creation subtasks to each feature sprint.

### Strangler Fig Ordering

The sprints correctly follow Strangler Fig pattern:

1. Sprint 1: Build new platform + core (new code alongside old)
2. Sprints 2-5: Build new slices (new code alongside old)
3. Sprint 6: Swap routing + delete old code

Ordering within feature sprints is correct: features that depend on others come later (Chat → Follow in Sprint 5, Event → Group in Sprint 4, Notification → all in Sprint 3 but handled via eventbus subscription).

---

## Conclusion

**Sprint plans correctly divide the architecture phases.** 18 issues identified: 6 blocking, 5 medium, 7 minor. All are path/ordering/documentation issues, not structural gaps. No features are missing. No duplicate work.

### Must Fix (Blocking)

1. Fix WS CheckOrigin path: `infra/ws/handler.go` → `internal/infra/http/ws/handler.go`
2. Fix all path references: add `internal/` prefix
3. Handle notification event subscription ordering (Sprint 3 before Sprint 4 events exist)
4. Merge existing migration files into numbered format in S1-BE-04
5. Move OAuth package rename from Sprint 1 to Sprint 5
6. Rewrite Docker Compose for 2-service architecture in Sprint 6

### Should Fix (Medium)

7. Add migration file creation tickets to Sprints 2, 3, 4
8. Clarify activity absorption scope in S2-BE-09
9. Clarify Category/Vote absorption into Topic
10. Split S6-BE-04 bootstrap into smaller tickets
11. Verify audit.md exists

### Optional (Minor)

12. Clarify scaffold is adjustment not greenfield
    13-18. Various documentation clarifications
