# Audit: `conventions.md` vs `general-instructions.md` + `target-architecture-with-phases.md`

**Model:** big-pickle
**Date:** 2026-06-19
**Scope:** Does `.agents/rules/conventions.md` contain all critical info from `docs/sprints/general-instructions.md` and `docs/architecture/target-architecture-with-phases.md`?

---

## What conventions.md Already Has (Correct)

| Topic                                                                 | Source Doc           | Coverage     |
| --------------------------------------------------------------------- | -------------------- | ------------ |
| Tech stack (Go 1.24, SQLite, Next.js, shadcn, ESLint + Prettier)      | Both                 | ✅ Complete  |
| Strangler Fig 6-step process                                          | Both                 | ✅ Complete  |
| D1–D6 (layout, interfaces, cross-slice, DB factory, boundaries, deps) | Both                 | ✅ Complete  |
| TDD (Red-Green-Refactor, table-driven, naming)                        | general-instructions | ✅ Adequate  |
| DB migrations (up/down, safety, rollback)                             | Both                 | ✅ Complete  |
| Branch naming + commit format                                         | general-instructions | ✅ Complete  |
| Security (bcrypt, param queries, WS origin, cookies)                  | target-architecture  | ✅ Complete  |
| Frontend structure + build gates                                      | general-instructions | ✅ Complete  |
| Smoke test reference (A1–D3)                                          | general-instructions | ✅ Links out |
| K8s probes, graceful shutdown, 12-factor                              | target-architecture  | ✅ Complete  |

---

## Missing from conventions.md — Critical

### 1. Definition of Done — Incomplete

`general-instructions.md:356-363` lists 7 items. `conventions.md:80-92` has 12 but **misses 3 process gates**:

- **Merged to main via squash merge** — no PR is done until the branch is actually merged.
- **Deployed to dev environment (Docker Compose)** — dev env verification is a hard gate.
- **Manual smoke test passes** — conventions.md only says "verify smoke scenarios" but doesn't list it as a DoD checkbox item.

### 2. Risk Mitigation Table — Entirely Missing

`general-instructions.md:345-352` has a risk mitigation table (6 rows):

| Risk                                   | Mitigation                                                        |
| -------------------------------------- | ----------------------------------------------------------------- |
| Old code breaks during migration       | Contract tests verify old API → new API behavior match            |
| Migration takes too long per feature   | Strangler Fig — ship one feature at a time, old code still runs   |
| Breaking API change for FE             | OpenAPI spec defined before BE implementation; FE mocks from spec |
| Database migration fails in production | Every migration has `.down.sql`; tested in CI                     |
| Performance regression                 | `make ci-bench` every sprint; flag > 10% degradation              |
| Dev blocked waiting for other dev      | Independent slices per BE dev; FE mocks BE APIs                   |

Critical for developers to understand _why_ the rules exist.

### 3. CI Pipeline Breakdown — Missing Detail

`conventions.md` mentions `make ci` but never explains what it runs. `target-architecture-with-phases.md:129-139` defines:

```
make be-ci:
  ci-mod → format → check-format → lint (staticcheck + golangci-lint + govulncheck) → test

make fe-ci:
  bun run lint → bun run format:check → tsc --noEmit → bun run test
```

Without this, a developer running `make ci` doesn't know what they're invoking.

### 4. Full Verification Gate Commands — Missing

`general-instructions.md:254-261` lists standalone commands for running without `make`:

```bash
go vet ./...
go build ./...
go test -race -coverprofile=coverage.out ./...
golangci-lint run
govulncheck ./...
```

`conventions.md:86` only says `go vet` for BE and `tsc --noEmit` for FE, missing `go build`, `go test -race`, `golangci-lint`, and `govulncheck`.

### 5. TDD Detailed Walkthrough — Missing

`conventions.md:58-62` gives a brief summary. `general-instructions.md:80-116` has the full RED→GREEN→REFACTOR loop with **concrete code examples** showing where test files live and how contract test files are named.

```go
// commands/register_test.go
func TestRegisterHandler_ValidInput(t *testing.T) { ... }
func TestRegisterHandler_InvalidEmail(t *testing.T) { ... }
func TestRegisterHandler_UnderAge(t *testing.T) { ... }

// internal/user/store/sqlite_migration_test.go
func TestUserStore_Migrated_SameAsOld_RegisterUser(t *testing.T) { ... }
```

---

## Missing — Moderate

### 6. Boundary Verification Command

`general-instructions.md:251` / `target-architecture-with-phases.md:933` share this grep:

```bash
grep -rn 'import' internal/*/transport/ internal/*/store/ | grep 'internal/' | \
  grep -v 'platform/' | grep -v 'pkg/'
```

`conventions.md:81` states the rule ("Conforms to D5 boundary rules") but doesn't give the tool to check it.

### 7. Guiding Principle

`target-architecture-with-phases.md:5` — **"One pattern, everywhere."** When there are two ways to do something, pick one and use it consistently. Simplicity and readability beat optimization.

This is the philosophical anchor for D1–D6.

### 8. Feature Overview (New vs Migrated)

`target-architecture-with-phases.md:63-74` — Which features are greenfield vs migrate-from-old. Helpful context for devs picking tickets.

### 9. Meta / Team Context

`general-instructions.md:11-17` — 5 devs, 1-week sprints, TDD + Strangler Fig + Trunk-Based Development. Lower criticality but useful onboarding context.

### 10. Microservice & CQRS Scaling Notes

`target-architecture-with-phases.md:867-896` — conventions.md has a one-liner on microservice promotion (line 53) but misses:

- Message broker swappability (RabbitMQ → Kafka)
- Separate entrypoints for commands vs queries
- Asymmetric scaling via Ingress routing

---

## Recommendations (Priority Order)

1. **Add missing DoD items** — `Merged to main via squash merge`, `Deployed to dev environment (Docker Compose)`, `Manual smoke test passes`
2. **Add Risk Mitigation table** — helps new devs understand "why these rules exist"
3. **Add CI pipeline breakdown** — document what `make ci` / `make be-ci` / `make fe-ci` actually run
4. **Add standalone verification commands** — so devs know what to run without `make`
5. **Add TDD detailed example** — the test file name patterns from general-instructions
6. **Add boundary verification grep command** — concrete tool for D5 rule checking
7. **Consider adding** "One pattern, everywhere" guiding principle + feature overview table as concise context blocks
