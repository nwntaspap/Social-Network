# Implementation Audit: `@.opencode/implementation_plan.md`

**Auditor:** minimaxai/minimax-m2.7
**Date:** 2026-06-19
**Overall Completion:** ~85%

---

## Summary

The implementation plan has been executed with strong fidelity on the core infrastructure (gate scripts, go-arch-lint, QRSPI agents, D5 depguard). Three critical gaps remain: two files the plan explicitly says to drop are still referenced in agent prompts, and an unplanned 11th agent (`audit.md`) exists that creates redundancy with the 3-agent review system.

---

## Part 1: Context Restructuring — PARTIAL

| Item                                      | Status | Notes                                                       |
| ----------------------------------------- | ------ | ----------------------------------------------------------- |
| `conventions.md` section tags             | ✅     | `@section:rules-core/fe/ci/git/dod` all present and correct |
| Agent context table (per plan spec)       | ⚠️     | Mostly aligned, gaps documented below                       |
| Drop `general-instructions.md`            | ❌     | Still referenced in `conventions.md:9`, `remedy.md:45`      |
| Drop `target-architecture-with-phases.md` | ❌     | Still referenced in `remedy.md:46`, `audit.md:48`           |

---

## Part 2: QRSPI Agent Architecture — MOSTLY COMPLETE

| Agent                   | Status | Notes                                                     |
| ----------------------- | ------ | --------------------------------------------------------- |
| `scout.md`              | ✅     | Matches spec exactly                                      |
| `architect.md`          | ✅     | Matches spec exactly                                      |
| `forge.md`              | ✅     | Reads PLAN.md, implement-only mode                        |
| `flowmaster.md`         | ✅     | QRSPI orchestration correct, all subagents accounted for  |
| `remedy.md`             | ✅     | Correct role, reads review report, surgical fixes         |
| `publish.md`            | ✅     | Tea CLI, reviewer assignment, tracker update              |
| `review-gates.md`       | ✅     | Runs `run-all.sh`, deterministic only                     |
| `review-conventions.md` | ✅     | Full compliance matrix against conventions.md             |
| `review-analysis.md`    | ✅     | 5-dimension analysis with hallucination filtering         |
| `audit.md`              | ❌     | **UNPLANNED** — duplicates review loop, should be deleted |
| Roster count            | ⚠️     | Plan says 10 agents, 11 exist (`audit.md` is extra)       |

---

## Part 3: Deterministic Gates + Local Enforcement — MOSTLY COMPLETE

| Item                                | Status | Notes                                                          |
| ----------------------------------- | ------ | -------------------------------------------------------------- |
| `check-stack.sh`                    | ✅     | Go 1.24 + module path `social-network`                         |
| `check-d1-layout.sh`                | ✅     | D1 vertical slice structure validated                          |
| `check-d5-boundaries.sh`            | ✅     | Grep-based D5 import checks                                    |
| `check-d6-dag.sh`                   | ✅     | `go list` DAG acyclicity + notification isolation              |
| `check-tdd.sh`                      | ✅     | Test file existence per feature                                |
| `check-migrations.sh`               | ✅     | Sequential naming + `;` delimiter                              |
| `check-security.sh`                 | ✅     | bcrypt, SQLi, WS origin patterns                               |
| `check-branch.sh`                   | ✅     | Naming + conventional commits                                  |
| `check-scope-drift.sh`              | ✅     | Advisory file listing                                          |
| `check-coverage-delta.sh`           | ✅     | Coverage regression >5% fails                                  |
| `run-all.sh`                        | ✅     | Runs all 10 gates, JSON output                                 |
| `lefthook.yml`                      | ⚠️     | `go-vet` runs on all files, not staged                         |
| `.go-arch-lint.yml`                 | ✅     | D6 DAG + all feature slices                                    |
| D5 depguard rules (`.golangci.yml`) | ✅     | Lines 91-116, exact match                                      |
| `Makefile` targets                  | ✅     | `review-gates`, `check-arch`, `setup-hooks`, `setup-arch-lint` |

---

## Critical Issues

### Issue 1: `general-instructions.md` still in agent context

**Plan says:** Drop from all agent context — _"80% redundant with conventions.md. Unique content only needed for manual QA."_

**Still referenced in:**

- `.agents/rules/conventions.md:9` — `Refer to general-instructions.md for detailed workflows`
- `.opencode/agents/remedy.md:45` — Context files list
- `.opencode/agents/audit.md:47` — Context files list

**Impact:** ~500-1000 tokens per agent invocation, defeating Part 1's 80% token savings goal.

### Issue 2: `target-architecture-with-phases.md` still in agent context

**Plan says:** Drop from all agent context — _"D5 rules already in conventions.md §2. Phases are per-ticket (read on demand by scout). ~7,000 tokens for ~500 tokens of value."_

**Still referenced in:**

- `.opencode/agents/remedy.md:46`
- `.opencode/agents/audit.md:48`

**Impact:** ~7,000 tokens per `remedy` invocation, significant context pollution.

### Issue 3: Unplanned `audit.md` agent

**Plan says:** `pr-review` → `audit` (deleted — replaced by 3 review agents)

**What happened:** `audit.md` (102 lines) was created but never deleted. It performs the same 5-phase review (gates + conventions + analysis + adversarial + synthesis) that `flowmaster` already orchestrates via `review-gates`, `review-conventions`, and `review-analysis` in its review loop.

**Why this is a problem:**

- `flowmaster` never spawns `audit` (not in its subagent invocation pattern)
- If manually invoked, `audit` duplicates the entire review loop
- The 3-agent split was designed for **context isolation** (fresh context per agent); `audit` defeats that by running everything in one context
- Plan explicitly says `pr-review` was deleted and replaced by 3 agents

**Fix:** Delete `.opencode/agents/audit.md`.

---

## Minor Issues

### Issue 4: `lefthook.yml` `go-vet` not scoped to staged files

```yaml
# Current (line 10):
- name: go-vet
  run: go vet ./...
```

Pre-commit hooks should ideally operate only on staged files. `gofumpt` and `goimports` correctly use `{staged_files}`; `go-vet` does not.

### Issue 5: `check-coverage-delta.sh` uses `git stash`

```bash
MAIN_COV=$(git stash -q 2>/dev/null; git checkout main -q 2>/dev/null && \
  go test -coverprofile=/tmp/main.cov ./... 2>/dev/null && ...
```

Uncommitted changes are stashed silently. If the stash fails (e.g., conflicting paths), `--stat` still succeeds. Minor risk since it's a coverage check, not a blocking gate.

---

## Files to Fix

| File                           | Change                                                                                                                                   |
| ------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------- |
| `.opencode/agents/audit.md`    | **DELETE** — not in plan roster, duplicates flowmaster's review loop                                                                     |
| `.opencode/agents/remedy.md`   | Remove `docs/sprints/general-instructions.md` and `docs/architecture/target-architecture-with-phases.md` from context list (lines 45-46) |
| `.agents/rules/conventions.md` | Remove line 9 `Refer to general-instructions.md...`                                                                                      |
| `lefthook.yml`                 | Change `go vet ./...` to `go vet {staged_files}`                                                                                         |

---

## Optimal Parts

The following are **well-implemented and match spec exactly:**

- All 10 gate scripts in `scripts/gates/`
- `run-all.sh` master runner with JSON output
- `.go-arch-lint.yml` with correct D6 DAG rules
- D5 depguard rules in `.golangci.yml` (lines 91-116)
- QRSPI agent files: `scout`, `architect`, `forge`, `flowmaster`
- Review agents: `review-gates`, `review-conventions`, `review-analysis`
- `publish.md` with tea CLI integration
- Permission model (all subagents `task: {"*": deny}`, only `flowmaster` has `task: {"*": allow}`)
- Section tags in `conventions.md`

---

_End of audit report._
