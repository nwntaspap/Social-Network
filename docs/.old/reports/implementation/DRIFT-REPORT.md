# Documentation Drift & Inconsistency Report

**Date:** 2026-06-20  
**Scope:** `docs/architecture/`, `docs/sprints/`, `README.md`, `.agents/rules/conventions.md`, `.opencode/agents/`  
**Sources:** 3 parallel subagent audits (architecture+README+conventions, sprints+architecture, opencode agents+rules)

---

## CRITICAL (8 findings)

### A1 | [stale] Verification gates path: `cmd/gates/` vs `internal/gates/` — neither exists

- **Files:** `README.md:259`, `architecture.md:214`, `conventions.md:118`, `sds.md:722`
- **Issue:** README and architecture.md reference `cmd/gates/main.go`. conventions.md and sds.md reference `internal/gates/`. Neither path exists in codebase. The actual gates exist at `cmd/gates/main.go` (checked — this DOES exist, but some docs point to `internal/gates/` instead).
- **Impact:** `make review-gates` and `go run cmd/gates/main.go --all` may confuse devs depending on which doc they follow.
- **Fix:** Pick one canonical path for gate code, update all docs consistently. Clarify: implementations in `internal/gates/`, runner at `cmd/gates/main.go`.

### A2 | [stale] `.github/PULL_REQUEST_TEMPLATE.md` referenced but doesn't exist

- **Files:** `README.md:392`, `DEVELOPMENT.md:222`, `conventions.md:151`
- **Issue:** Three docs reference this file. No `.github/` directory exists in repo.
- **Impact:** PR template copy step fails; devs/agents can't follow PR conventions.
- **Fix:** Create `.github/PULL_REQUEST_TEMPLATE.md` or remove/update references.

### A3 | [contradiction] Username source: `origin` remote vs own username

- **Files:** `README.md:369`, `DEVELOPMENT.md:202`, `conventions.md:141`
- **Issue:** README and DEVELOPMENT.md say "Gitea username from `origin` remote". conventions.md:141 explicitly says "Use your OWN username, NOT the `origin` remote owner." Opposite instructions.
- **Impact:** Devs use wrong username (remote owner `dkotsi`), branches fail PR validation.
- **Fix:** Align README and DEVELOPMENT.md with conventions.md — change to "Your own Gitea username (resolve via `tea whoami`)."

### C1 | [phase-misalignment] Architecture Phases 1-7 vs Sprints 0-6 — no mapping table

- **Files:** `target-architecture-with-phases.md`, `general-instructions.md`
- **Issue:** Architecture defines 7 phases. Sprints merge them differently (S0 = Bug Fixes + Scaffold + Docker dev, etc). Mapping is implicit, never documented.
- **Impact:** Devs reverse-engineer phase↔sprint mapping. No single source of truth for sequencing.
- **Fix:** Add Phase↔Sprint mapping table to `general-instructions.md` or `target-architecture-with-phases.md`.

### C2 | [sprint-drift] Sprint duration: "~7 weeks" vs actual 14 weeks

- **Files:** `general-instructions.md:14`, all sprint files
- **Issue:** General instructions says "~7 weeks (7 sprints)" + "Sprint length: 1 week". Sprint headers span Week 1–14 (2 weeks each).
- **Impact:** Schedule planning contradictory.
- **Fix:** Update `general-instructions.md` to "Sprint length: 2 weeks, Total: ~14 weeks" or rescale sprint headers.

### C3 | [terminology] SSE vs WebSocket for notifications — three-way contradiction

- **Files:** `sprint-3.md:7`, `README.md:5`, `architecture.md:6,142`
- **Issue:** Sprint 3 mandates SSE (`EventSource` + 15s polling fallback). README and architecture.md say "SSE or WebSockets" — ambiguous. Different transport protocols, different API contracts.
- **Impact:** Implementers may choose WebSocket, breaking S3-BE-57 endpoint and S3-FE-18 EventSource contract.
- **Fix:** Update README and architecture.md: SSE for notifications, WebSocket for chat only.

### C4 | [tracker] Appendix C ticket counts don't match actual sprint files

- **Files:** `general-instructions.md` Appendix C, sprint files, `ticket-tracker.md`
- **Issue:** Sprint 1 BE: Appendix says 11, actual = 9. Sprint 0 SD: Appendix says 3, actual = 4. Sprint 3 BE: Appendix says 26, actual = 25.
- **Impact:** Progress tracking unreliable.
- **Fix:** Recount all tickets, update Appendix C.

### A4 | [broken-ref] `tea pulls edit --add-reviewers` flag doesn't exist

- **Files:** `.opencode/agents/publish.md:93`
- **Issue:** The `--add-reviewers` flag doesn't exist on `tea pulls edit`. Only `--add-assignees` and `--add-labels` exist. Agent acknowledges this limitation for `tea pulls create` (line 80) but then uses the non-existent flag on `tea pulls edit`.
- **Impact:** Publish phase fails at runtime. No reviewers added to PR.
- **Fix:** Replace with Gitea API call `curl -X POST "<gitea-api>/pulls/$PR_NUMBER/requested_reviewers"` or find supported method.

---

## WARNING (21 findings)

### A5 | [contradiction] Vote feature: standalone slice vs absorbed into topic/comment

- **Files:** `architecture.md:119`, `target-architecture-with-phases.md:119`, `conventions.md:150`, sprints 2+3
- **Issue:** D6 dependency graph shows `vote` as standalone. conventions.md says vote absorbed into `topic/`. Sprints implement votes inside `topic/` and `comment/` slices.
- **Fix:** Remove `vote` from D6 as standalone, add note that vote lives in `topic/` and `comment/`.

### A6 | [terminology] `pkg/` contents listed differently across 3 docs

- **Files:** `README.md:88-126`, `target-architecture-with-phases.md:451-457`, `architecture.md:56`
- **Issue:** Three different listings of what `pkg/` contains. README: bcrypt, uuid, validators, imgutil. target-arch adds `oauth/` and `helpers/`. architecture.md omits `helpers/`.
- **Fix:** Unify `pkg/` contents across all docs.

### A7 | [terminology] "almost private" vs `almost_private`

- **Files:** `README.md:50`, `sds.md:55`, `target-architecture-with-phases.md:66`
- **Issue:** README uses human-readable "almost private" (space). Canonical enum value is `almost_private` (underscore).
- **Fix:** README should use `almost_private` or show "almost private (enum: `almost_private`)".

### A8 | [contradiction] `.env.example` referenced but doesn't exist

- **Files:** `DEVELOPMENT.md:39`, `README.md:176-187`
- **Issue:** DEVELOPMENT.md says `cp .env.example .env`. README defines `.env` fields inline without mentioning `.env.example`.
- **Fix:** Create `.env.example` or align DEVELOPMENT.md with README's inline `.env` content.

### A9 | [contradiction] Database location: 3 different paths

- **Files:** `DEVELOPMENT.md:107` (`db/data/forum.db`), `README.md:223` (`internal/platform/database/social.db`), `.env` (`data/social.db`)
- **Issue:** Three different DB paths documented.
- **Fix:** Verify actual DB path, update all docs to one canonical location.

### A10 | [contradiction] Commit format: with vs without ticket ID

- **Files:** `DEVELOPMENT.md:216`, `conventions.md:150`, `README.md:378`
- **Issue:** DEVELOPMENT.md shows `<type>(<scope>): <description>` (no ticket ID). conventions.md and README show `<type>(<scope>)[<ID>]: <description>` (with ticket ID).
- **Fix:** Unify on one format. If ticket ID required, DEVELOPMENT.md must include `[<ID>]`.

### A11 | [cross-ref] Smoke test scenarios duplicated

- **Files:** `conventions.md:133`, `target-architecture-with-phases.md:938-964`
- **Issue:** Both define the same smoke test scenarios A1–D3 independently. Drift risk.
- **Fix:** Single-source in general-instructions.md, others reference it.

### A12 | [stale] Vitest+Playwright status ambiguous

- **Files:** `conventions.md:17`, `architecture.md:178-180`, `README.md:148`
- **Issue:** conventions.md lists as if existing. architecture.md marks "(planned)". README lists without qualification.
- **Fix:** Clarify current vs target state in all docs.

### A13 | [stale] Migration numbering: flat `schema.sql` vs numbered migrations

- **Files:** `target-architecture-with-phases.md:514-524`, `db/migrations/`
- **Issue:** target-arch envisions numbered migrations. Current `db/migrations/` has only `schema.sql` and `indexes.sql`.
- **Fix:** Document current state vs target. Add note that existing SQL must be converted.

### A14 | [terminology] "Vertical Feature Slices" vs "Vertical Slices with CQRS"

- **Files:** `architecture.md:12`, `README.md:27`, `target-architecture-with-phases.md:149`
- **Issue:** README omits "CQRS" from heading. Others include it. CQRS is a hard requirement.
- **Fix:** Use "Vertical Slices with CQRS" consistently.

### A15 | [gap] Strangler Fig migration absent from README

- **Files:** `README.md`, `conventions.md:37-45`
- **Issue:** conventions.md has full Strangler Fig section. README has none. Critical workflow missing from entry point.
- **Fix:** Add Strangler Fig reference to README (even brief, linking to conventions.md).

### A16 | [gap] No documentation reading order in README

- **Files:** `README.md`, `conventions.md:191-211`
- **Issue:** conventions.md defines progressive disclosure reading order. README has no equivalent. New devs have no guided path.
- **Fix:** Add "Documentation Reading Order" or "Next Steps" section to README.

### A17 | [gap] DEVELOPMENT.md omits verification gates

- **Files:** `DEVELOPMENT.md`, `conventions.md`, `README.md`, `architecture.md`
- **Issue:** DEVELOPMENT.md is the contributor guide but omits `make review-gates`, coverage thresholds, gate descriptions.
- **Fix:** Add verification gates section or link to architecture.md's gate table.

### W6 | [scope] Notification type strings: dot notation vs underscore

- **Files:** `sprint-3.md:5`, `sds.md:241`
- **Issue:** SDS uses `follow.requested` (dot). Sprint-3 uses `follow_request` (underscore). Also `follow.accepted` vs `follow_accept`.
- **Impact:** Routing failures if wrong format used.
- **Fix:** Standardize on one format across all docs.

### W8 | [scope] Migration numbering fragmented across 3 docs

- **Files:** `target-architecture-with-phases.md:509-524`, `sds.md`, `sprint-5.md`
- **Issue:** Architecture lists migrations 000001-000007. Sprint 5 creates 000010. Gap (000008, 000009) implied but undocumented.
- **Fix:** Add consolidated migration numbering table to target-architecture or sds.

### W10 | [stale] `internal/pkg/` vs `pkg/` path for shared utilities

- **Files:** `sprint-0.md:23`, `target-architecture-with-phases.md:451-457`, `README.md:123`
- **Issue:** Sprint 0 creates `internal/pkg/`. Target arch specifies `pkg/` at repo root. Sprint 5 is migration point but undocumented.
- **Fix:** Document `internal/pkg/` → `pkg/` migration plan.

### W11 | [scope] Group roles: DB has 3 (creator, admin, member), sprint implements 2 (owner, member)

- **Files:** `sds.md:129`, `sprint-4.md:37`, `target-architecture-with-phases.md:617`
- **Issue:** "admin" role in DB has no implementation. "creator" vs "owner" naming drift.
- **Fix:** Remove `admin` from DB or add implementation tickets. Align "creator"/"owner".

### A18 | [convention-violation] Remedy agent references wrong section numbers

- **Files:** `.opencode/agents/remedy.md:43`, `conventions.md`
- **Issue:** Remedy references "security §7, TDD §3". Actual: Security is §6, TDD is §4, Strangler Fig is §3, Frontend is §7.
- **Fix:** Change to "security §6, TDD §4, Strangler Fig §3".

### A19 | [convention-violation] Gate names in conventions.md don't match actual `--gate=` flags

- **Files:** `conventions.md:129`, actual gate code
- **Issue:** Table lists "Coverage" and "ScopeDrift". Actual flags are `coverage-delta` and `scope-drift`. Using table names as `--gate=` values will error.
- **Fix:** Update gate table to use actual flag names with `--gate=` examples.

### A20 | [workflow-conflict] Flowmaster gate loop: steps vs pseudocode disagree

- **Files:** `.opencode/agents/flowmaster.md:36-41` vs `:46-74`
- **Issue:** Steps say "gates FAIL → spawn remedy → loop to step 5". Pseudocode shows different cycle. No `review_count` increment for gate-only failures = infinite loop possible.
- **Fix:** Align steps with pseudocode. Add max retry counter for gate-only failures.

### A21 | [workflow-conflict] Flowmaster APPROVED status never defined

- **Files:** `.opencode/agents/publish.md:47`, `.opencode/agents/flowmaster.md:59`
- **Issue:** Both check for `APPROVED` status but no subagent emits it. Only `PASS`, `FAIL`, `PASS_WITH_RECOMMENDATIONS`, `CHANGES_REQUESTED` are produced.
- **Fix:** Define `APPROVED` explicitly: "0 Critical + 0 Warning findings." Add to flowmaster rules.

---

## INFO (18 findings)

### I1 | `docs/requirements/audit.md` not cross-linked from architecture

- General-instructions.md and sprint-6 reference it. Architecture docs don't. **Fix:** Add link to architecture.md.

### I2 | Three DB paths: `db/data/forum.db`, `internal/platform/database/social.db`, `/app/data/social.db`

- **Fix:** Consolidate to one canonical reference (`.env`-driven).

### I3 | Strangler Fig steps: 6 steps (general-instructions) vs 5 steps (DEVELOPMENT.md)

- Not contradictory, different counts. **Fix:** Standardize on one version.

### I4 | Sprint-6 S6-SD-25/34 reference `audit.md` but not linked from architecture

- **Fix:** Add cross-link.

### I5 | `internal/gates/` vs `cmd/gates/` — both valid (impl vs runner)

- **Fix:** Clarify: gate code in `internal/gates/`, runner at `cmd/gates/main.go`.

### I6 | "Nickname" (Go) vs "username" (DB/API) naming tension

- **Fix:** Rename Go field to `Username` or add convention note.

### I7 | SDS `CREATE TABLE` blocks mix schema definition with migration staging

- **Fix:** Separate current tables from future migration additions.

### I8 | Seed migration: sprint says 000009, arch says 000007

- Sprint has correct newer number. **Fix:** Update architecture doc.

### I9 | Handler vs Resolver suffix convention undocumented

- Commands use `*Handler`, queries use `*Resolver`. **Fix:** Add naming convention to conventions.md.

### I10 | OAuth feature described 3 ways across docs

- **Fix:** Standardize description.

### I11 | "Non-blocking Event Bus" ambiguous (async vs non-blocking I/O)

- **Fix:** Replace with "asynchronous Event Bus" or "in-process channel-based Event Bus".

### I12 | Frontend spec duplicated in sds.md and architecture.md

- **Fix:** SDS as canonical, architecture.md references SDS §6.

### I13 | SDS §4 WebSocket doesn't reference `core/session/` for auth handshakes

- **Fix:** Add explicit reference.

### I14 | `make dev` alias not documented in README

- **Fix:** Add to README command table.

### I15 | Go version: "1.24" vs "1.24 (or 1.24.4+)"

- **Fix:** Standardize to "Go ≥1.24" everywhere.

### I16 | `cmd/migrate/main.go` referenced in conventions.md but doesn't exist

- **Fix:** Create migration runner or update path.

### I17 | Flowmaster runs `make ci` inside review-gates, mixing L1 gates and L2 CI

- **Fix:** Remove `make ci` from review-gates, add separate CI step.

### I18 | `rtk` prefix rule in antigravity-rtk-rules.md ignored by all agents

- **Fix:** Add `rtk*` permissions or downgrade rule to "recommended."

---

## Summary

| Severity | Count | Top Categories                                                                                                                    |
| -------- | ----- | --------------------------------------------------------------------------------------------------------------------------------- |
| CRITICAL | 8     | stale (2), contradiction (3), phase-misalignment (1), sprint-drift (1), broken-ref (1)                                            |
| WARNING  | 21    | scope (4), terminology (4), contradiction (3), workflow-conflict (2), convention-violation (2), gap (3), stale (2), cross-ref (1) |
| INFO     | 18    | stale (3), terminology (4), cross-ref (3), gap (2), broken-ref (2), workflow (2), scope (1), contradiction (1)                    |

### Priority Fix Order

| Priority | ID  | Issue                                | Blocks                      |
| -------- | --- | ------------------------------------ | --------------------------- |
| P0       | A3  | Username source contradiction        | PR validation               |
| P0       | C2  | Sprint duration 7 vs 14 weeks        | Scheduling                  |
| P0       | C3  | SSE vs WebSocket mandate             | Notification implementation |
| P0       | A4  | `tea --add-reviewers` non-existent   | Publish pipeline            |
| P1       | C1  | Phase↔Sprint mapping missing         | Sprint planning             |
| P1       | C4  | Ticket count mismatch                | Progress tracking           |
| P1       | A1  | Gates path: `cmd/` vs `internal/`    | Build/run commands          |
| P1       | A2  | PR template doesn't exist            | PR workflow                 |
| P1       | W6  | Notification type dot vs underscore  | Notification routing        |
| P1       | A18 | Remedy agent wrong section refs      | Wrong fixes applied         |
| P2       | A5  | Vote slice vs absorbed               | Architecture clarity        |
| P2       | A10 | Commit format with/without ticket ID | Commit consistency          |
| P2       | W8  | Migration numbering fragmented       | Schema evolution            |
| P2       | W11 | Group role DB vs implementation      | Group feature logic         |
| P2       | A9  | 3 different DB paths                 | Dev environment setup       |
| P2       | A19 | Gate names don't match flags         | Gate commands fail          |
