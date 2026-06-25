# Conventions.md Audit Report

**Model:** deepseek-ai/deepseek-v4-pro
**Timestamp:** 2026-06-19T12:04:33Z
**Scope:** Cross-reference `.agents/rules/conventions.md` against `docs/sprints/general-instructions.md` and `docs/architecture/target-architecture-with-phases.md`

---

## Summary

conventions.md **captures all D1-D6 design decisions, all security rules, all TDD workflow, and all boundary rules** from the source documents. It does its job as an always-on agent rule file. Five gaps identified below — none critical.

---

## Gaps: Missing from conventions.md

| #   | Missing Info                                                                                               | Source                                                  | Severity | Suggestion                                                                                                                                                                             |
| --- | ---------------------------------------------------------------------------------------------------------- | ------------------------------------------------------- | -------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | **Ticket format details** (component, priority, dependency, assignee, story points, acceptance criteria)   | general-instructions.md Meta                            | Medium   | Add to §6 Branching section. Devs need ticket anatomy when picking work items.                                                                                                         |
| 2   | **Contract test deletion rule** — "Delete contract tests AFTER old code removed."                          | general-instructions.md R1 step 6 + R1 Rule             | Medium   | §3 line 59 covers contract tests but not deletion timing. Add sentence.                                                                                                                |
| 3   | **Old code deletion constraint** — "No partial deletion. Old code exists until ALL its features migrated." | general-instructions.md R1 Rule                         | Medium   | §2 has 6-step Strangler Fig but missing this explicit constraint. Add bullet after step 6.                                                                                             |
| 4   | **Smoke test scenario table** (A1-D3 Steps/Expected)                                                       | general-instructions.md Q3                              | Medium   | §5 line 95 references smoke tests but only via doc link. Either reproduce the Q3 table inline or keep current cross-reference. Current state works but weakens standalone readability. |
| 5   | **BE linter stack enumeration** (`golangci-lint`, `staticcheck`, `govulncheck`, `gofumpt`)                 | target-architecture Phases 2+ / general-instructions Q2 | Low      | §1 mentions `make ci` but doesn't list individual tools. Devs troubleshooting lint failures would benefit.                                                                             |

---

## Verified: Already Covered

| Info                                                                                | conventions.md Location | Source Verified   |
| ----------------------------------------------------------------------------------- | ----------------------- | ----------------- |
| Technology stack (Go 1.24, SQLite, Next.js, Tailwind, shadcn/ui, ESLint + Prettier) | §1                      | ✓                 |
| Strangler Fig 6-step process                                                        | §2 lines 24-30          | ✓ matches R1      |
| Route prefix convention (`/api/` vs `/api/v1/`)                                     | §2 lines 31-34          | ✓                 |
| D1 vertical slice layout                                                            | §2 line 35              | ✓                 |
| D2 interface strategy (full within, narrow across)                                  | §2 lines 36-38          | ✓                 |
| D3 cross-slice communication (ID-only, consumer iface, eventbus)                    | §2 lines 39-42          | ✓                 |
| D4 DB factory (`DB` interface, not `*sql.DB`)                                       | §2 lines 43-44          | ✓                 |
| D5 boundary rules (full table)                                                      | §2 lines 45-50          | ✓                 |
| D6 dependency graph                                                                 | §2 lines 52-53          | ✓                 |
| Microservice readiness (no cross-slice SQL joins)                                   | §2 line 54              | ✓                 |
| Event bus error isolation (defer recover in subscribers)                            | §2 line 55              | ✓                 |
| Feature toggle pattern                                                              | §2 line 56              | ✓                 |
| TDD red-green-refactor                                                              | §3 lines 58-59          | ✓                 |
| Contract tests                                                                      | §3 lines 60-61          | ✓                 |
| Go test style (table-driven, `t.Run()`)                                             | §3 lines 61-62          | ✓                 |
| Test naming conventions                                                             | §3 line 62              | ✓                 |
| OpenAPI 3.0 contract testing (kin-openapi, msw)                                     | §3 lines 65-68          | ✓                 |
| Goroutine panic recovery                                                            | §3 line 69              | ✓                 |
| RateLimiter ticker leak prevention                                                  | §3 line 70              | ✓                 |
| Database migrations (seq, safety, delimiter, rollback, test)                        | §4                      | ✓                 |
| DoD checklist (11 items)                                                            | §5                      | ✓ matches R5 + A6 |
| Performance regression gate (`make ci-bench`, >10%)                                 | §5 line 98              | ✓                 |
| Branch naming + username resolution + verification                                  | §6                      | ✓                 |
| Commit format (Conventional Commits) + allowed scopes                               | §6                      | ✓                 |
| PR description template                                                             | §6 line 111             | ✓                 |
| Security: bcrypt cost 12                                                            | §7 line 113             | ✓                 |
| Security: plaintext password memory wiping                                          | §7 line 114             | ✓                 |
| Security: SQL parameterized queries                                                 | §7 line 115             | ✓                 |
| Security: ORDER BY whitelist                                                        | §7 line 116             | ✓                 |
| Security: MIME type validation (magic bytes)                                        | §7 line 117             | ✓                 |
| Security: WebSocket origin validation                                               | §7 line 118             | ✓                 |
| Security: WebSocket timeout constants                                               | §7 line 119             | ✓                 |
| Security: session cookie attributes (HttpOnly, Secure, SameSite=Lax)                | §7 line 120             | ✓                 |
| Frontend project structure                                                          | §8 lines 124-128        | ✓ matches F5      |
| Frontend build gates (bun lint, format:check, tsc --noEmit, test)                   | §8 line 131             | ✓ matches F6      |
| Frontend testing tools (Vitest, Playwright)                                         | §8 line 132             | ✓                 |
| File upload limits (10MB client-side)                                               | §8 line 133             | ✓ matches F4      |
| Design system (HSL, glassmorphism, Inter/Outfit)                                    | §8 line 134             | ✓                 |
| Destructive operation confirmation dialogs                                          | §8 line 135             | ✓ matches F2      |
| Notification panel vs chat visual distinction                                       | §8 line 136             | ✓ matches F1      |
| SSE for notifications (+ 15s polling fallback)                                      | §8 line 137             | ✓                 |
| Pre-commit/pre-push hooks (Husky/lefthook)                                          | §8 line 138             | ✓                 |
| Kubernetes probes (/healthz, /readyz)                                               | §9 line 139             | ✓                 |
| Graceful shutdown (SIGTERM/SIGINT)                                                  | §9 line 140             | ✓                 |
| 12-factor config (env vars only)                                                    | §9 line 141             | ✓                 |
| Observability (X-Request-ID, slog, Prometheus)                                      | §9 line 142             | ✓ matches A4      |
| Testing pyramid (20 E2E, 50 integration, 300+ unit)                                 | §3 line 64              | ✓ matches A1      |
| Surgical changes rule                                                               | §3 line 63              | ✓                 |
| Store method documentation comments (`// Used by:`)                                 | §2 line 50              | ✓                 |

---

## Recommendation

All 5 gaps are low-to-medium severity. conventions.md is adequate as-is for agent behavior rules. Apply suggestions #1-#3 for completeness. #4 and #5 are optional polish.
