# conventions.md Coverage Audit

**Model**: z-ai/glm-5.1  
**Timestamp**: 2026-06-19T12:04:37Z  
**Sources audited**: `docs/sprints/general-instructions.md`, `docs/architecture/target-architecture-with-phases.md`  
**Target**: `.agents/rules/conventions.md`

---

## GAPS — Info in source docs but missing from conventions.md

### From `general-instructions.md`

1. **Team structure & sprint cadence** (line 12-15): 5 devs, roles (SD-QA, BE-A, BE-B, FE-A, FE-B), 1-week sprints, ~7 weeks total. conventions.md has zero team/sprint info.

2. **Developer onboarding workflow** (line 44-55): "Pick ticket -> set up branch -> TDD cycle -> PR guidelines". Not in conventions.md.

3. **Frontend Feature-to-Audit Mapping (F1)** (line 157-174): Complete component/route mapping table (RegisterForm, LoginForm, ProfileCard, FollowButton, PostForm, GroupDirectory, NotificationBell, etc.). conventions.md has **none** of this.

4. **Frontend Interaction Patterns (F2)** (line 176-181): Confirmation dialogs, follow-gate feedback message ("At least one user must follow the other"), emoji support, WS reconnection with exponential backoff. conventions.md missing follow-gate feedback text and emoji support.

5. **Frontend State Management (F3)** (line 183-187): Auth persistence via cookies (partially in conventions.md security section), **session isolation** (Chrome/Firefox separate sessions), RSC recommendation. conventions.md missing session isolation rule.

6. **Bug list (Q1)** (line 219-231): B1.1-B1.8 specific bugs with file locations. Not in conventions.md. (Critical for Phase 1 reference.)

7. **Verification gates standalone commands** (line 254-261): `go vet`, `go build`, `go test -race -coverprofile`, `golangci-lint run`, `govulncheck ./...` as standalone commands without `make`. Not in conventions.md.

8. **Boundary verification grep command** (line 251): `grep -rn 'import' internal/*/transport/ internal/*/store/ | grep 'internal/' | grep -v 'platform/' | grep -v 'pkg/'`. conventions.md lacks this.

9. **Risk mitigation table (A5)** (line 344-352): 6 risks with mitigations. Not in conventions.md.

10. **Dependency map visual (Appendix B)** (line 369-399): Sprint-by-sprint dependency flow. Not in conventions.md.

11. **Ticket count summary (Appendix C)** (line 404-414): Per-sprint ticket counts (172 total). Not in conventions.md.

### From `target-architecture-with-phases.md`

12. **Current codebase pain points** (line 9-21): 32 handler dirs, 38 aliases, 3:1 overhead ratio. Not in conventions.md. (Useful context for _why_ vertical slices.)

13. **Feature overview table** (line 62-74): 10 features with new/migrated status. Not in conventions.md.

14. **Full directory tree — target state** (line 299-461): Complete `cmd/`, `internal/`, `pkg/`, `config/`, `bootstrap/` tree with every file listed. conventions.md has D1 layout summary but **no full tree**.

15. **D4 `DB` interface code** (line 227-245): `QueryContext`, `QueryRowContext`, `ExecContext` method signatures + factory `NewDB(cfg)`. conventions.md mentions D4 conceptually but omits the actual interface shape.

16. **D6 full dependency graph** (line 280-293): All 10 features with their exact dependencies listed. conventions.md has the chain summary but **missing individual arrows** (e.g. `comment -> user, topic`, `vote -> user, topic, comment, eventbus`, `oauth -> user`).

17. **Phases 1-5** (line 465-699): Entire phase sequence with per-feature migration steps, merge notes (user absorbs activity, topic absorbs category+vote), and old directory deletion list. conventions.md has none of this.

18. **Phases 6-10** (line 703-864): Frontend scaffold, Docker Compose, PostgreSQL, Redis, RabbitMQ optional phases. Not in conventions.md.

19. **Microservice promotion rules** (line 868-897): No shared storage, strict boundary imports, extraction path (move dir -> HTTP/gRPC transport -> replace in-memory calls). conventions.md has one bullet on this but **missing the 3-step extraction path**.

20. **Independent CQRS scaling** (line 892-897): Separate `cmd/commands/main.go` + `cmd/queries/main.go` binaries, asymmetric routing, database replication with write/read DSNs. Not in conventions.md.

21. **Message broker swappability** (line 887-891): EventBus interface, zero amqp import in features, Kafka swap = new `platform/kafka` + one-line bootstrap change. Not in conventions.md.

22. **Docker Compose config** (line 742-758): Two-service compose with volume mount, env vars. Not in conventions.md.

23. **Event bus event types** (line 852-858): `follow.requested`, `follow.accepted`, `group.invited`, `group.join_requested`, `event.created`. Not in conventions.md.

24. **Backend port 8080, Frontend port 3000** (line 48-49): Explicit port numbers. Not in conventions.md.

---

## OVERLAP — Already captured well in conventions.md

- D1 vertical slice layout
- D2 interface strategy
- D3 cross-slice communication (3 strategies)
- D5 boundary rules
- Strangler Fig 6-step process
- `/api/` vs `/api/v1/` prefix convention
- TDD red-green-refactor
- Contract test naming
- Testing pyramid numbers
- Migration safety (no drop in same migration)
- Branch naming + username resolution
- DoD checklist
- Security items (bcrypt, MIME, WS origin, ORDER BY whitelist)
- SSE for notifications + polling fallback
- K8s probes, graceful shutdown, 12-factor
- Event bus error isolation
- Feature toggle pattern
- WebSocket timeout constants
- Goroutine panic recovery
- RateLimiter ticker leak prevention

---

## SUGGESTIONS (priority-ranked)

| #   | Suggestion                                                                                                         | Priority |
| --- | ------------------------------------------------------------------------------------------------------------------ | -------- |
| 1   | Add **D6 full dependency graph** (all per-feature arrows, not just the chain) — critical for import discipline     | High     |
| 2   | Add **D4 `DB` interface shape** (`QueryContext`, `QueryRowContext`, `ExecContext`) — devs need exact signatures    | High     |
| 3   | Add **verification gates standalone commands** (`govulncheck`, `golangci-lint run`, `go test -race -coverprofile`) | High     |
| 4   | Add **boundary verification grep command**                                                                         | High     |
| 5   | Add **frontend F1 feature-to-audit mapping** (route -> component table)                                            | Medium   |
| 6   | Add **frontend session isolation rule** (Chrome/Firefox separate sessions)                                         | Medium   |
| 7   | Add **3-step microservice extraction path** (move dir -> transport -> replace calls)                               | Medium   |
| 8   | Add **event bus event types** (`follow.requested`, `follow.accepted`, etc.)                                        | Medium   |
| 9   | Add **port numbers** (BE 8080, FE 3000)                                                                            | Low      |
| 10  | Add **CQRS scaling path** (separate binaries, asymmetric routing, read replicas)                                   | Low      |
| 11  | Add **message broker swappability** (zero amqp in features, Kafka path)                                            | Low      |
| 12  | Add **risk mitigation table** from A5                                                                              | Low      |
| 13  | Consider adding **team/sprint structure** (5 devs, 1-week sprints)                                                 | Low      |

Items 1-4 are the most critical operational gaps — they affect daily coding decisions and CI verification. Items 5-8 are important for feature correctness. Items 9-13 are reference material that conventions.md currently defers to the source docs.
