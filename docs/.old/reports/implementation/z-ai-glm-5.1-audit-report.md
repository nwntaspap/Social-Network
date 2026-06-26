# Implementation Audit Report

**Date**: 2026-06-19
**Auditor**: z-ai/glm-5.1
**Plan**: `.opencode/implementation_plan.md` (v2)
**Scope**: Full correctness + optimality review of all 3 parts

---

## Overall Verdict

**Core architecture is correctly implemented.** The QRSPI split, gate scripts, lefthook, arch-lint, depguard, and Makefile targets all match the plan. 3 items require action.

---

## Part 1: Context Restructuring

### conventions.md Section Tags

| Plan Item                               | Status   | Evidence                                                        |
| --------------------------------------- | -------- | --------------------------------------------------------------- |
| `<!-- @section:rules-core -->` + `:end` | **PASS** | Lines 11, 76                                                    |
| `<!-- @section:rules-fe -->` + `:end`   | **PASS** | Lines 78, 90                                                    |
| `<!-- @section:rules-ci -->` + `:end`   | **PASS** | Lines 92, 111                                                   |
| `<!-- @section:rules-git -->` + `:end`  | **PASS** | Lines 113, 129                                                  |
| `<!-- @section:rules-dod -->` + `:end`  | **PASS** | Lines 131, 145                                                  |
| Content inside tags matches plan        | **PASS** | D1-D6 in core, frontend in fe, CI in ci, git in git, DoD in dod |

### Files Permanently Dropped From Agent Context

| File                                 | Plan                        | Status           | Finding                                                                                           |
| ------------------------------------ | --------------------------- | ---------------- | ------------------------------------------------------------------------------------------------- |
| `general-instructions.md`            | Drop from all agent context | **FAIL**         | Still referenced in `conventions.md:9` ("Refer to general-instructions.md...") and `remedy.md:46` |
| `target-architecture-with-phases.md` | Drop from all agent context | **PARTIAL FAIL** | Still referenced in `remedy.md:47` and `audit.md:47`                                              |

### Agent Context Allocation

| Agent              | Plan Context                                              | Actual Context                           | Match                     |
| ------------------ | --------------------------------------------------------- | ---------------------------------------- | ------------------------- |
| scout              | conventions.md `rules-core`                               | conventions.md `rules-core`              | **PASS**                  |
| architect          | conventions.md `rules-core` + `rules-ci`                  | conventions.md `rules-core` + `rules-ci` | **PASS**                  |
| forge              | conventions.md `rules-core` + `rules-ci`, AGENTS.md §1-§4 | Same                                     | **PASS**                  |
| review-conventions | conventions.md ALL                                        | Same                                     | **PASS**                  |
| review-analysis    | conventions.md `rules-core`                               | Same                                     | **PASS**                  |
| remedy             | conventions.md `rules-core` (per plan table)              | **4 files** including dropped ones       | **FAIL** — see finding #7 |
| publish            | conventions.md `rules-git` + `rules-dod`                  | Implicit via publish workflow            | **PASS**                  |

---

## Part 2: QRSPI Agent Architecture

### Agent File Verification

| Agent              | Plan File                                | Exists  | Mode     | Steps | Model                  | task  | Verdict  |
| ------------------ | ---------------------------------------- | ------- | -------- | ----- | ---------------------- | ----- | -------- |
| flowmaster         | `.opencode/agents/flowmaster.md`         | **YES** | primary  | 80    | deepseek-v4-flash-free | allow | **PASS** |
| scout              | `.opencode/agents/scout.md`              | **YES** | subagent | 20    | deepseek-v4-flash-free | deny  | **PASS** |
| architect          | `.opencode/agents/architect.md`          | **YES** | subagent | 15    | deepseek-v4-flash-free | deny  | **PASS** |
| forge              | `.opencode/agents/forge.md`              | **YES** | subagent | 45    | deepseek-v4-flash-free | deny  | **PASS** |
| review-gates       | `.opencode/agents/review-gates.md`       | **YES** | subagent | 10    | deepseek-v4-flash-free | deny  | **PASS** |
| review-conventions | `.opencode/agents/review-conventions.md` | **YES** | subagent | 20    | deepseek-v4-flash-free | deny  | **PASS** |
| review-analysis    | `.opencode/agents/review-analysis.md`    | **YES** | subagent | 25    | deepseek-v4-flash-free | deny  | **PASS** |
| remedy             | `.opencode/agents/remedy.md`             | **YES** | subagent | 35    | deepseek-v4-flash-free | deny  | **PASS** |
| publish            | `.opencode/agents/publish.md`            | **YES** | subagent | 30    | deepseek-v4-flash-free | deny  | **PASS** |

### Agent Roster Integrity

| Agent                         | Plan Action                          | Status                                            |
| ----------------------------- | ------------------------------------ | ------------------------------------------------- |
| `pr-implement → forge`        | Rename + rewrite                     | **PASS** — new forge.md written from scratch      |
| `pr-review → audit (DELETED)` | Delete — replaced by 3 review agents | **FAIL** — `audit.md` still exists (finding #1)   |
| `pr-fix → remedy`             | Rename + rewrite                     | **PASS** — new remedy.md written from scratch     |
| `pr-create → publish`         | Rename + rewrite                     | **PASS** — new publish.md written from scratch    |
| `ticket-to-pr → flowmaster`   | Rename + rewrite                     | **PASS** — new flowmaster.md written from scratch |

### Orphaned Agent

- **`audit.md`** — Plan explicitly says: "pr-review → audit (deleted — replaced by 3 review agents)". The file still exists at `.opencode/agents/audit.md`. flowmaster does not spawn it. It is dead code that could cause confusion if someone invokes it directly.

### QRSPI Flow Verification

The flowmaster orchestration matches plan:

1. scout → RESEARCH.md ✅
2. architect → PLAN.md ✅
3. forge → implements PLAN.md ✅
4. review-gates → JSON results ✅
5. review-conventions → compliance matrix ✅
6. review-analysis → findings list ✅
7. remedy → fix loop ✅
8. publish → PR_URL ✅

Fix loop logic correct: max 3 cycles, CHANGES_REQUESTED stops after 3, PASS_WITH_RECOMMENDATIONS proceeds after 3.

### Improvements Over Plan

Good additions not in the original plan:

- `forge.md:62-65` — self-check section before returning
- `flowmaster.md:45-74` — explicit fix loop pseudocode
- `flowmaster.md:76-86` — subagent invocation pattern with exact inputs
- `flowmaster.md:89-99` — rules section for orchestrator behavior

---

## Part 3: Deterministic Gates + Local Enforcement

### L0: Lefthook Git Hooks

| Plan Item                                   | Status   | Evidence                       |
| ------------------------------------------- | -------- | ------------------------------ |
| `lefthook.yml` exists                       | **PASS** | Root directory, 20 lines       |
| pre-commit: go-format (gofumpt + goimports) | **PASS** | Lines 4-8, `stage_fixed: true` |
| pre-commit: go-vet                          | **PASS** | Line 9-10                      |
| pre-push: go-test                           | **PASS** | Line 16                        |
| pre-push: go-build                          | **PASS** | Line 17                        |
| pre-push: arch-lint                         | **PASS** | Line 18-19                     |
| Both stages parallel                        | **PASS** | `parallel: true` on both       |

### L1: Gate Scripts

| Script                    | Plan Gate # | Exists  | Executable | Verdict                                            |
| ------------------------- | ----------- | ------- | ---------- | -------------------------------------------------- |
| `run-all.sh`              | Master      | **YES** | 755        | **PASS** — JSON output, all 10 gates               |
| `check-stack.sh`          | #1          | **YES** | 755        | **PASS** — Go version + module path                |
| `check-d1-layout.sh`      | #2          | **YES** | 755        | **PASS** — vertical slice dirs                     |
| `check-d5-boundaries.sh`  | #3          | **YES** | 755        | **PASS** — import boundary grep                    |
| `check-d6-dag.sh`         | #4          | **YES** | 755        | **PASS** — circular dep + notification check       |
| `check-tdd.sh`            | #6          | **YES** | 755        | **PASS** — test file existence                     |
| `check-migrations.sh`     | #7          | **YES** | 755        | **WARN** — delimiter regex broken (finding #4)     |
| `check-security.sh`       | #8          | **YES** | 755        | **PASS** — SQLi, WS origin, bcrypt                 |
| `check-branch.sh`         | #9          | **YES** | 755        | **PASS** — branch naming + conventional commits    |
| `check-scope-drift.sh`    | —           | **YES** | 755        | **WARN** — advisory only, always PASS (finding #6) |
| `check-coverage-delta.sh` | #13         | **YES** | 755        | **WARN** — destructive git checkout (finding #5)   |

### L2: go-arch-lint + Depguard

| Plan Item                               | Status   | Evidence                                                                        |
| --------------------------------------- | -------- | ------------------------------------------------------------------------------- |
| `.go-arch-lint.yml` exists              | **PASS** | Root directory, 40 lines                                                        |
| All 9 feature slices defined            | **PASS** | user through oauth                                                              |
| Cross-cutting components (5)            | **PASS** | core, platform, pkg, bootstrap, config                                          |
| D6 deps: slices → [core, platform, pkg] | **PASS** | All 9 slices identical rule                                                     |
| Shared layer deps correct               | **PASS** | core→[platform,pkg], platform→[pkg], pkg→[], config→[], bootstrap→anyVendorDeps |
| D5 depguard rules in `.golangci.yml`    | **PASS** | `d5_commands_queries`, `d5_transport`, `d5_store` all present (lines 91-116)    |

### Makefile Targets

| Target            | Plan | Status                              |
| ----------------- | ---- | ----------------------------------- |
| `check-arch`      | New  | **PASS** — calls go-arch-lint check |
| `review-gates`    | New  | **PASS** — calls run-all.sh         |
| `setup-hooks`     | New  | **PASS** — installs lefthook        |
| `setup-arch-lint` | New  | **PASS** — installs go-arch-lint    |

### Bonus Targets (not in plan but useful)

- `review-gates-fast` — quick gate subset (ci-mod + format + staticcheck)
- `review-gates-all` — gates + vulncheck

---

## Findings

### Finding #1 — HIGH: `audit.md` not deleted

**Plan**: "pr-review → audit (deleted — replaced by 3 review agents)"
**Reality**: `.opencode/agents/audit.md` still exists (102 lines, full 5-phase review agent)
**Impact**: Zombie agent. flowmaster never spawns it, but if invoked directly it duplicates the work of review-gates + review-conventions + review-analysis with a single larger context. Undermines the QRSPI context isolation goal.
**Fix**: Delete `.opencode/agents/audit.md`

### Finding #2 — HIGH: `general-instructions.md` not dropped from agent context

**Plan**: "Drop `general-instructions.md` and `target-architecture-with-phases.md` from all agent context lists"
**Reality**:

- `conventions.md:9` still says `Refer to general-instructions.md for detailed workflows.`
- `remedy.md:46` lists `docs/sprints/general-instructions.md` as context
  **Impact**: ~7,000 tokens consumed per remedy invocation for redundant content (per plan's own analysis).
  **Fix**:
- Remove line 9 from `conventions.md`
- Remove `docs/sprints/general-instructions.md` line from `remedy.md:46`

### Finding #3 — MEDIUM: `target-architecture-with-phases.md` not dropped from all agents

**Plan**: Drop from all agent context
**Reality**: Referenced in `remedy.md:47` and `audit.md:47`
**Impact**: When audit.md is deleted (finding #1), only remedy.md needs cleanup.
**Fix**: Remove `docs/architecture/target-architecture-with-phases.md` line from `remedy.md:47`

### Finding #4 — MEDIUM: `check-migrations.sh` delimiter check is broken

**File**: `scripts/gates/check-migrations.sh:23`
**Code**: `grep -rn '^\s*:' "$MIGRATION_DIR"/*.sql`
**Problem**: Regex matches **any** line starting with `:` after optional whitespace. This is not what the plan intended ("delimiter: `;` (never `:`)"). The regex should detect SQL statements ending with `:` as a statement delimiter. Current regex will false-positive on comments, default values, etc.
**Impact**: False positives in migration gate, causing spurious FAIL results.
**Fix**: Change to target actual statement-terminating colons:

```bash
BAD_DELIMITERS=$(grep -rnE ';\s*$' "$MIGRATION_DIR"/*.sql 2>/dev/null | grep -v "';" | head -5 || true)
```

Or, if the intent is specifically to catch `:` used as delimiter:

```bash
BAD_DELIMITERS=$(grep -rnE "^[^/-].*:\s*$" "$MIGRATION_DIR"/*.sql 2>/dev/null | head -5 || true)
```

### Finding #5 — MEDIUM: `check-coverage-delta.sh` is destructive and fragile

**File**: `scripts/gates/check-coverage-delta.sh:5-8`
**Code**: `git stash -q 2>/dev/null; git checkout main -q 2>/dev/null && go test ... && git checkout - -q 2>/dev/null; git stash pop -q 2>/dev/null`
**Problems**:

1. `set -euo pipefail` + git checkout will abort on any error, potentially leaving repo on wrong branch
2. `git stash` loses uncommitted changes if something goes wrong mid-pipeline
3. If `main` branch doesn't exist locally (e.g. fresh clone), the entire gate crashes
4. The `2>/dev/null` suppresses all error output, making failures silent
5. `git checkout main` could trigger pre-commit hooks from lefthook
   **Impact**: Gate can leave repo in broken state (detached HEAD, lost stash) during CI runs.
   **Fix**: Use `git stash --include-untracked` + trap for cleanup, or better: use `go test -cover`
   on both branches via worktrees or a separate checkout. Minimal fix — add trap:

```bash
cleanup() { git checkout - 2>/dev/null; git stash pop 2>/dev/null; }
trap cleanup EXIT
```

### Finding #6 — LOW: `check-scope-drift.sh` is advisory-only

**File**: `scripts/gates/check-scope-drift.sh`
**Code**: Always exits 0 (PASS). Just prints `git diff --stat`.
**Impact**: The gate count claims 10 gates but one is a no-op that never fails. Misleading in JSON output.
**Verdict**: Acceptable. Auto-detecting ticket scope is genuinely hard. The gate serves as a human-readable display that reviewing agents (review-analysis) can parse. Could be improved by passing the ticket ID and checking file paths against the ticket's feature scope, but that's beyond current scope.

### Finding #7 — LOW: `remedy.md` contradicts plan on context files

**Plan**: remedy context = `conventions.md` `rules-core` only (from agent context table)
**Reality**: `remedy.md:43-47` lists 4 context files:

1. `.agents/rules/conventions.md`
2. `AGENTS.md`
3. `docs/sprints/general-instructions.md`
4. `docs/architecture/target-architecture-with-phases.md`
   **Impact**: 2 of these are files the plan says to drop. Extra context = ~14,000 tokens wasted per remedy invocation.
   **Fix**: Remove items 3 and 4 from remedy.md context list (fixes findings #2 and #3 simultaneously)

### Finding #8 — LOW: `conventions.md:9` reference to dropped file

**File**: `.agents/rules/conventions.md:9`
**Code**: `Refer to [general-instructions.md](../../docs/sprints/general-instructions.md) for detailed workflows.`
**Impact**: Stale reference. Agents reading conventions.md will follow this link and load the dropped file.
**Fix**: Remove line 9 from conventions.md

---

## Summary Table

| #   | Severity | Item                                                             | Fix Effort |
| --- | -------- | ---------------------------------------------------------------- | ---------- |
| 1   | HIGH     | Delete `audit.md`                                                | 1 min      |
| 2   | HIGH     | Remove `general-instructions.md` from conventions.md + remedy.md | 2 min      |
| 3   | MEDIUM   | Remove `target-architecture-with-phases.md` from remedy.md       | 1 min      |
| 4   | MEDIUM   | Fix `check-migrations.sh` delimiter regex                        | 5 min      |
| 5   | MEDIUM   | Harden `check-coverage-delta.sh` with trap                       | 10 min     |
| 6   | LOW      | `check-scope-drift.sh` advisory — acceptable, no fix needed      | —          |
| 7   | LOW      | remedy.md context files list too broad (covered by #2+#3)        | —          |
| 8   | LOW      | conventions.md stale reference (covered by #2)                   | —          |

**Total fix time**: ~19 minutes for findings #1-#5. Findings #6-#8 are covered by #1-#5 or accepted.
