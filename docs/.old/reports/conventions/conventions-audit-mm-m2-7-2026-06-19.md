# Audit: `conventions.md` vs Source Documents

**Model**: minimaxai/minimax-m2.7
**Date**: Fri Jun 19 2026

---

**TL;DR**: Conventions covers ~85% of critical info. Notable gaps below.

---

## MISSING from `general-instructions.md`

| Gap                                                                                    | Severity | Location in source                |
| -------------------------------------------------------------------------------------- | -------- | --------------------------------- |
| **Onboarding Guide** (pick ticket → set up branch → TDD cycle → PR)                    | Medium   | `general-instructions.md:44-51`   |
| **Q1 Bug list** (8 specific bugs with file paths, fixes, and reproducer-first process) | High     | `general-instructions.md:219-237` |
| **Smoke test scenarios A1–D3** (the actual step-by-step tables)                        | High     | `general-instructions.md:267-282` |
| **Risk mitigation table** (A5)                                                         | Low      | `general-instructions.md:345-352` |
| **Sprint/team meta** (5 devs, 1-week sprints, 7 sprints, ~172 tickets)                 | Low      | `general-instructions.md:10-17`   |
| **Definition of Done** (A6 — partially covered in DoD checklist)                       | Medium   | `general-instructions.md:356-363` |

---

## MISSING from `target-architecture-with-phases.md`

| Gap                                                                                                                        | Severity | Location in source                           |
| -------------------------------------------------------------------------------------------------------------------------- | -------- | -------------------------------------------- |
| **System overview diagram** (browser → Go → SQLite/Redis/RabbitMQ)                                                         | Medium   | `target-architecture-with-phases.md:27-45`   |
| **Component descriptions** (what each layer does, port numbers 8080/3000)                                                  | Medium   | `target-architecture-with-phases.md:47-59`   |
| **Feature overview table** (which features are new vs migrated, descriptions)                                              | Low      | `target-architecture-with-phases.md:63-75`   |
| **Complete tooling table** (golangci-lint v2.2.1, staticcheck, govulncheck, benchstat, pprof, goimports, gci, go mod tidy) | High     | `target-architecture-with-phases.md:78-127`  |
| **CI pipeline** (exact `make ci` flow: `ci-mod → format → check-format → lint → test`)                                     | High     | `target-architecture-with-phases.md:128-139` |
| **Phase 1–7 descriptions** (which work happens in each phase)                                                              | Medium   | `target-architecture-with-phases.md:465-700` |
| **Verification checklist** (exact `go vet`, `go build`, `golangci-lint`, `govulncheck`, `tsc --noEmit` commands)           | High     | `target-architecture-with-phases.md:900-934` |
| **Graceful shutdown** (SIGTERM/SIGINT handling, drain in-flight)                                                           | Medium   | `target-architecture-with-phases.md:881-884` |
| **Kubernetes readiness** (`/healthz`, `/readyz` probe behavior)                                                            | Medium   | `target-architecture-with-phases.md:880-884` |
| **CQRS independent scaling path** (separate `cmd/commands/main.go` and `cmd/queries/main.go` entrypoints)                  | Low      | `target-architecture-with-phases.md:891-897` |
| **RabbitMQ exchange/queue topology table**                                                                                 | Low      | `target-architecture-with-phases.md:850-859` |

---

## PARTIAL coverage (correct but could be richer)

- **Boundary verification grep**: Exists in conventions `:88` but `target-architecture-with-phases.md:931-934` shows a slightly different (and more correct) version: adds `grep -v 'infra/'`
- **Testing pyramid**: Conventions says "Aim for ~20 E2E, ~50 integration, ~300+ unit" — matches general-instructions, but lacks the ASCII art diagram which some prefer for quick visual reference
- **Smoke test scenarios**: Conventions references them (`general-instructions.md` link) but doesn't reproduce the A1–D3 tables inline — if the link breaks, the data is gone
- **DoD checklist**: Conventions has it but it's missing the "merged to main via squash merge" and "deployed to dev environment" items from general-instructions A6

---

## Suggestions

1. **Add Q1 bug list** — the 8 specific bugs with file paths are critical institutional knowledge that shouldn't live only in a sprint doc
2. **Add smoke test scenarios A1–D3 inline** — don't rely on a cross-link
3. **Add the complete tooling table** — golangci-lint version, staticcheck, govulncheck, benchstat, pprof, gci are all mentioned in target-architecture but not spelled out in conventions
4. **Add exact `make ci` pipeline steps** — backend: `ci-mod → format → check-format → lint → test`, frontend: `lint → format:check → tsc --noEmit → test`
5. **Add `/healthz` and `/readyz` probe behavior** — currently only "Graceful Shutdown" is mentioned, not the K8s probe specifics
6. **Add boundary verification grep with `grep -v 'infra/'`** — the version in target-architecture-with-phases is slightly more correct
7. **Add onboarding guide bullet points** — the 4-step onboarding (pick ticket → branch → TDD → PR) is the most important onboarding content
