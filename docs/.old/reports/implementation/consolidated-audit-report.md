# Consolidated Implementation Audit & Validation Report

**Date**: 2026-06-19
**Scope**: Context Engineering & Agent Architecture Optimization (v2)
**Sources**: Deduplicated findings from 5 independent AI audits (moonshotai/kimi-k2.6, minimaxai/minimax-m2.7, deepseek-v4-flash-free, flash, z-ai/glm-5.1)
**Validation Auditor**: Antigravity AI (Second Validation Audit)

---

## Executive Summary & Second Validation Audit Status

The core architectural restructuring—splitting the monolithic audit agent into a modular QRSPI workflow and setting up the CI gating foundation—is successfully scaffolded. However, the execution mechanism is critically broken.

During the **Second Validation Audit**, we executed the gate-verification scripts and reviewed the repository configuration. We confirmed that the master gate runner (`make review-gates` / `run-all.sh`) crashes immediately with **Error 141 (SIGPIPE)**, the dependency gate `check-d6-dag.sh` crashes with **exit code 1** due to a pre-migration empty directory layout, `check-coverage-delta.sh` fails on clean trees and uses dangerous checkouts, and `check-branch.sh` floods the outputs with hundreds of false-positive historical commit failures due to unrelated remote Git history.

This report consolidates all verified findings, resolves a major false-positive finding from the initial audits (`opencode.json`), and provides the exact remediation changes to make the gate pipeline operational.

---

## 1. Critical Issues (Pipeline & Architecture Blockers)

### 1.1 `run-all.sh` Master Gate Runner SIGPIPE Crash (Error 141)

- **Root Cause**: The script runs with `set -euo pipefail`. Inside the `run_gate` function, output is piped to `head -5` to truncate:
  ```bash
  output=$(echo "$output" | head -5 | tr '\n' ' ' | sed 's/"/\\"/g')
  ```
  If a gate script generates more than 5 lines of output (e.g., listing non-conventional commits), `head -5` terminates and closes the read end of the pipe. This triggers a `SIGPIPE` (exit code 141) on `echo`. Under `pipefail`, this crashes the entire pipeline and prevents the final JSON summary from being printed.
- **Validation Status**: **VERIFIED**. Running `make review-gates` fails with `Error 141` on line 275 of the Makefile.
- **Remediation**: Disable `pipefail` temporarily during output truncation, or avoid piping `echo` to `head`.
  ```diff
  -  output=$(echo "$output" | head -5 | tr '\n' ' ' | sed 's/"/\\"/g')
  +  # Temporarily disable pipefail to avoid SIGPIPE (exit code 141) when truncating
  +  set +o pipefail
  +  output=$(echo "$output" | head -n 5 | tr '\n' ' ' | sed 's/"/\\"/g')
  +  set -o pipefail
  ```

### 1.2 Unrelated Git Histories (Merge-Base Mismatch) & Rebase Strategy Mismatch

- **Root Cause**: The local `main` branch tracks `github/main` (which has a completely different commit history and root commit from Gitea's `origin/main` remote). As a result:
  - `git merge-base main HEAD` fails.
  - `git log main..HEAD` lists the entire history of the repository (hundreds of commits) as "non-conventional".
  - `git diff main..HEAD` floods `check-scope-drift.sh`, `remedy.md`, and `publish.md` with irrelevant tree diffs.
  - Rebase instructions in `publish.md` (`Ensure branch is rebased on main`) will fail with massive conflicts.
- **Validation Status**: **VERIFIED**. Running `check-branch.sh` directly returns hundreds of pages of `Non-conventional commit` errors dating back to the repository's initial commits.
- **Remediation**: Resolve the merge-base dynamically in both `check-branch.sh` and `check-scope-drift.sh` by falling back to `origin/main` if local `main` is unrelated.
  ```diff
  +# Find the correct merge-base or default to origin/main if local main is unrelated
  +BASE="main"
  +if ! git merge-base main HEAD &>/dev/null; then
  +  BASE="origin/main"
  +fi
  +
  -COMMITS=$(git log main..HEAD --format='%s' 2>/dev/null || true)
  +COMMITS=$(git log "$BASE"..HEAD --format='%s' 2>/dev/null || true)
  ```
  Additionally, update `publish.md` to specify rebasing onto `origin/main` rather than `main`.

### 1.3 `check-coverage-delta.sh` is Fragile and Destructive

- **Root Cause**: The script runs with `set -e` and attempts to use `git stash -q` and `git stash pop`.
  1. On a clean working tree, `git stash` creates no stash. When the script completes, `git stash pop` fails with exit code 1 because the stash stack is empty, crashing the script.
  2. Running `git checkout main` and testing coverage in the active workspace is destructive. If interrupted, the working tree is left in a detached state or on the wrong branch.
  3. If `main` doesn't exist locally (e.g. fresh clone), it crashes.
- **Validation Status**: **VERIFIED**. The script exits with status 1 on clean trees and suppresses diagnostics.
- **Remediation**: Avoid stashing if the worktree is clean, dynamically determine the base branch, and safely recover the branch state:
  ```diff
  -MAIN_COV=$(git stash -q 2>/dev/null; git checkout main -q 2>/dev/null && \
  -  go test -coverprofile=/tmp/main.cov ./... 2>/dev/null && \
  -  go tool cover -func=/tmp/main.cov | tail -1 | awk '{print $3}' | tr -d '%'; \
  -  git checkout - -q 2>/dev/null; git stash pop -q 2>/dev/null)
  +STASHED=false
  +if [ -n "$(git status --porcelain)" ]; then
  +  git stash -q
  +  STASHED=true
  +fi
  +
  +BASE_BRANCH="main"
  +if ! git merge-base main HEAD &>/dev/null; then
  +  BASE_BRANCH="origin/main"
  +fi
  +
  +MAIN_COV=""
  +if git checkout "$BASE_BRANCH" -q 2>/dev/null; then
  +  if go test -coverprofile=/tmp/main.cov ./... 2>/dev/null; then
  +    MAIN_COV=$(go tool cover -func=/tmp/main.cov | tail -n 1 | awk '{print $3}' | tr -d '%')
  +  fi
  +  git checkout - -q
  +fi
  +
  +if [ "$STASHED" = true ]; then
  +  git stash pop -q 2>/dev/null || true
  +fi
  ```

### 1.4 Empty Feature List Crash in `check-d6-dag.sh`

- **Root Cause**: The script lists features using:
  ```bash
  FEATURES=$(ls -d internal/*/ 2>/dev/null | xargs -I{} basename {} | grep -v -E '^(core|platform|pkg|config|bootstrap|domain|infra|app)$')
  ```
  Because the repository is currently in a pre-migration/layered state, the `grep -v -E` command filters out all directory listings. When `grep` finds no matches, it exits with code 1. Because the script runs under `set -e`, this causes the script to crash immediately without output.
- **Validation Status**: **VERIFIED**. Running `check-d6-dag.sh` directly exits with code 1 and yields zero output.
- **Remediation**: Append `|| true` to tolerate empty grep results:
  ```diff
  -FEATURES=$(ls -d internal/*/ 2>/dev/null | xargs -I{} basename {} | grep -v -E '^(core|platform|pkg|config|bootstrap|domain|infra|app)$')
  +FEATURES=$(ls -d internal/*/ 2>/dev/null | xargs -I{} basename {} | grep -v -E '^(core|platform|pkg|config|bootstrap|domain|infra|app)$' || true)
  ```

### 1.5 Orphaned Legacy `audit.md` Agent

- **Root Cause**: The implementation plan explicitly mandated deleting the monolithic `audit` agent and replacing it with three isolated review subagents (`review-gates`, `review-conventions`, and `review-analysis`). However, `.opencode/agents/audit.md` was left in the repository.
- **Validation Status**: **VERIFIED**. The file `.opencode/agents/audit.md` (5.5K) is still present. It contains the old monolithic review loop, violating context isolation and creating duplicate agents.
- **Remediation**: Remove the orphaned file:
  ```bash
  rm .opencode/agents/audit.md
  ```

---

## 2. Medium Issues & Warnings (Context Bloat & Delimiter Errors)

### 2.1 Stale Context References in `remedy.md`

- **Root Cause**: The prompt for `remedy.md` (lines 45-46) lists `general-instructions.md` and `target-architecture-with-phases.md` under its `Context Files` section. This contradicts Phase 1's directive to drop these files to save ~14,000 tokens of redundant context per invocation.
- **Validation Status**: **VERIFIED**. Stale references exist in `.opencode/agents/remedy.md`.
- **Remediation**: Remove these lines from `.opencode/agents/remedy.md:45-46`.

### 2.2 Stale Context Reference in `conventions.md`

- **Root Cause**: Line 9 of `.agents/rules/conventions.md` still contains:
  ```markdown
  Refer to [general-instructions.md](../../docs/sprints/general-instructions.md) for detailed workflows.
  ```
  Agents parsing the rules will try to load this dropped file, negating token optimizations.
- **Validation Status**: **VERIFIED** in `.agents/rules/conventions.md:9`.
- **Remediation**: Remove line 9 from the file.

### 2.3 Broken Delimiter Check in `check-migrations.sh`

- **Root Cause**: The delimiter check uses the regex:
  ```bash
  BAD_DELIMITERS=$(grep -rn '^\s*:' "$MIGRATION_DIR"/*.sql 2>/dev/null | head -5 || true)
  ```
  This matches any line starting with a colon (`:`), which triggers false positives on comments or default arguments, rather than matching lines ending with a colon `:` representing bad separators.
- **Validation Status**: **VERIFIED**. The regex does not validate statement-terminating semicolons properly.
- **Remediation**: Change the query to target actual statement-terminating colons:
  ```bash
  BAD_DELIMITERS=$(grep -rnE "^[^/-].*:\s*$" "$MIGRATION_DIR"/*.sql 2>/dev/null | head -5 || true)
  ```

### 2.4 Stale Agent Reference in `publish.md`

- **Root Cause**: The error messaging inside `publish.md` (line 47) instructs that `flowmaster` must run `remedy`/`audit` again. It fails to recognize that `audit` has been replaced.
- **Validation Status**: **VERIFIED** in `.opencode/agents/publish.md`.
- **Remediation**: Update `publish.md` to remove `/audit` from the error message.

### 2.5 Scope List Mismatch in `check-branch.sh`

- **Root Cause**: The branch validation script lists `tracker` as an allowed conventional commit scope:
  ```bash
  ALLOWED_SCOPES="user|topic|follow|group|event|chat|notification|oauth|core|platform|comment|tracker"
  ```
  However, `tracker` is not defined as an allowed scope in `conventions.md` (which only lists `user, topic, follow, group, event, chat, notification, oauth, core, platform, comment`).
- **Validation Status**: **VERIFIED**.
- **Remediation**: Remove `tracker` from the script's scope list or document it in `conventions.md`.

---

## 3. Low Issues & Suggestions (Minor Tweaks)

### 3.1 `lefthook.yml` `go-vet` Scope Optimization

- **Root Cause**: The `go-vet` hook runs on the entire repository (`go vet ./...`) instead of running only on staged files (`{staged_files}`).
- **Validation Status**: **VERIFIED** on line 10 of `lefthook.yml`.
- **Remediation**: Change target execution from `./...` to `{staged_files}`.

### 3.2 `check-d5-boundaries.sh` Broad Grep

- **Root Cause**: The regex `grep -rn 'import'` is too generic, matching comments or strings that contain "import" and failing to exclude test files (`*_test.go`).
- **Validation Status**: **VERIFIED** on line 8 of `scripts/gates/check-d5-boundaries.sh`.
- **Remediation**: Refine regex to `grep -rn '^\s*import'` and add `--exclude='*_test.go'`.

### 3.3 Undocumented Directory Skips in `check-d1-layout.sh` and `check-d6-dag.sh` & Silenced Errors

- **Root Cause**: Both `check-d1-layout.sh` (line 9) and `check-d6-dag.sh` (lines 5 and 15) skip the `domain`, `infra`, and `app` directories, which was not specified in the implementation plan. While this is a reasonable adaptation for the pre-migration repository layout, it constitutes an undocumented layout drift. Furthermore, `check-d6-dag.sh` redirects `go list` stderr to `/dev/null` (line 9), silencing compilation/syntax errors.
- **Validation Status**: **VERIFIED** in both gate scripts.
- **Remediation**: Document the directory skip adaptation and remove `2>/dev/null` (or capture stderr) in `check-d6-dag.sh` to ensure package syntax errors fail the gate.

### 3.4 `check-security.sh` Bcrypt Detection Fragility

- **Root Cause**: The script only catches inline assignments like `cost = N`. It misses constant-based cost definitions or package default reference variables (`bcrypt.DefaultCost`).
- **Validation Status**: **VERIFIED**.
- **Remediation**: Standardize on checking for low-cost definitions or enforce `bcrypt.DefaultCost` or constants with values < 12.

### 3.5 Missing Context Sections in `publish.md`

- **Root Cause**: The `publish.md` agent file does not explicitly declare its requirement for the `rules-git` and `rules-dod` sections of `conventions.md` in its header, relying on implicit loaded context.
- **Validation Status**: **VERIFIED**.
- **Remediation**: Add explicit context documentation to the `publish.md` file header.

### 3.6 `check-scope-drift.sh` is Advisory-Only

- **Root Cause**: The script always prints `PASS` and exits with code 0. It is decorative and provides no enforcement mechanism.
- **Validation Status**: **VERIFIED**.
- **Remediation**: Keep as advisory (useful for reviewing agents) but label clearly in outputs.

### 3.7 Extra Self-Check in `forge.md`

- **Root Cause**: The `forge.md` agent includes an added "Self-check" section (lines 62-65) before returning. While this is a technically positive scope-drift guardrail, it deviates from the plan's exact structure.
- **Validation Status**: **VERIFIED**.
- **Remediation**: Retain it as it acts as an excellent additional guardrail.

---

## 4. Reconciled False Positives (Disproven Initial Findings)

### 4.1 Finding: `opencode.json` is Empty (🔴 CRITICAL in Kimi Audit)

- **Initial Audit Claim**: Kimi's audit claimed that `opencode.json` being empty is a critical issue that makes agents undiscoverable.
- **Reconciliation Verdict**: **DISPROVEN / FALSE POSITIVE**.
- **Details**: According to `docs/reports/enforcements/PLAN.md`, this was a deliberate architectural fix. Config was intentionally migrated from `opencode.json` into frontmatter on standalone markdown files in `.opencode/agents/` to fix startup crashes where paths resolved relative to the config directory `.opencode/` resulting in double-nested paths. The opencode system auto-discovers agents from `.opencode/agents/*.md` dynamically. Therefore, `opencode.json` is **correctly** trimmed to just `$schema`. No action is required.

### 4.2 Legacy Agent Deletion Verification (Except `audit.md`)

- **Finding**: The implementation plan required deleting the legacy agent files `pr-implement.md`, `pr-review.md`, `pr-fix.md`, `pr-create.md`, and `ticket-to-pr.md`.
- **Validation Status**: **VERIFIED**. None of these legacy files exist in the repository; they have been successfully deleted. Only the orphaned `audit.md` (detailed in 1.5) remains to be cleaned up.
