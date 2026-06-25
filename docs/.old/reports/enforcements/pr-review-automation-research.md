# PR Review Automation: Deterministic Tools Research

## 1. What SHOULD Be Automated vs What Needs Human/LLM Judgment

### Automate (deterministic, zero false positives acceptable):

| Category         | Examples                                       | Rationale                        |
| ---------------- | ---------------------------------------------- | -------------------------------- |
| Style/formatting | `gofmt`, `prettier`, `eslint --fix`            | Unambiguous; machine-perfect     |
| Typo/misspelling | `misspell`, `codespell`                        | Deterministic dictionary match   |
| Obvious bugs     | nil deref, unused vars, shadowing              | Type system + dataflow; zero FP  |
| Security scan    | SQL injection patterns, XSS, hardcoded secrets | Signature match; SARIF output    |
| PR metadata      | Size, title format, template filled, WIP flag  | Simple string/number checks      |
| Infra linting    | `actionlint`, `hadolint`, `yamllint`           | Schema validation; deterministic |
| Binary/bytecode  | Dep version check, lockfile diff               | Hash comparison                  |
| CI/CD config     | `actionlint` for GH Actions                    | Syntax+semantic check of YAML    |

### Needs Human or LLM Judgment:

| Category            | Examples                            | Why                          |
| ------------------- | ----------------------------------- | ---------------------------- |
| Architecture/design | "Should this be a microservice?"    | Tradeoffs, context-dependent |
| Logic correctness   | "Is this algorithm right?"          | Intent understanding         |
| Edge cases          | "What happens when input is empty?" | Domain knowledge             |
| Naming/API design   | "Is this function name clear?"      | Subjective; reader empathy   |
| Cross-file impact   | "Will this break consumer X?"       | Full-codebase awareness      |
| Business logic      | "Does this match the spec?"         | Product knowledge            |
| Test quality        | "Are these tests meaningful?"       | Hard to formalize well       |

### The Hybrid Pattern (best practice):

```
PR pipeline: [deterministic gates] → [agent (LLM) review] → [human review]
```

- Deterministic gates fail **fast**, block obviously broken PRs
- LLM agent reviews as **first pass** (summarize, flag suspicious patterns)
- Human reviews **only what needs judgment** (architecture, tradeoffs)

Source: G-Research blog, Tanagram, repowise comparisons (2025-2026)

---

## 2. Detailed Tool Analysis

### 2.1 reviewdog

- **URL**: <https://github.com/reviewdog/reviewdog>
- **Stars**: 9.4k
- **Language**: Go
- **What**: Universal bridge between any linter and code hosting platform PR comments
- **How it works**: Pipes linter output (any tool via `errorformat`), filters by diff, posts as PR comments
- **Output**: Structured per-file, per-line annotations inline in PR diff
- **Agent-consumable**: Yes — output is line+file+severity+message. JSON format available via `-reporter=json`
- **Go ecosystem**: Written in Go, native `golangci-lint` action available (`reviewdog/action-golangci-lint`)
- **Key insight**: reviewdog solves the **false positive problem** by only reporting on changed lines. This is its killer feature — teams can enable strict linters that normally produce too many warnings.

### 2.2 Danger / Danger-JS

- **URL**: <https://danger.systems> / <https://github.com/danger/danger>
- **Stars**: 5k+ (Ruby), 4.5k+ (JS)
- **Language**: Ruby / TypeScript (JS)
- **What**: Codify team norms as code. Runs rules in CI, comments on PRs
- **Output**: Custom message, warn, fail, markdown
- **Agent-consumable**: Yes — output is PR comment text; can produce structured JSON with custom plugins
- **Go ecosystem**: No native Go port. Danger-JS can be used in any CI. For Go-native alternative, use `reviewdog` + shell scripts.
- **Common rules**: CHANGELOG enforcement, ticket link check, file-based auto-assignment, binary file detection, test file addition check
- **Stats**: Teams report 20-30% faster review times (ExpertBeacon 2024)
- **Key insight**: Danger is the most expressive; you write arbitrary JavaScript/Ruby. But Go projects have no native Danger — they use reviewdog + shell or custom Go tools.

### 2.3 golangci-lint in PR Review Mode

- **URL**: <https://golangci-lint.run>
- **Language**: Go
- **What**: Meta-linter aggregating 50+ linters (staticcheck, govet, gosec, unused, etc.)
- **Output formats**: JSON, Checkstyle, SARIF (standard), tab, colored-line, CodeClimate, JUnit XML
- **Agent-consumable**: Yes — JSON/SARIF output is machine-parseable. Used by reviewdog, GitHub CodeQL, etc.
- **PR integration**: Via `golangci/golangci-lint-action@v6` - annotates PR with findings
- **Key insight**: Runs linters in parallel with caching. Completes in seconds on large codebases. The integrated `golangci-lint-action` posts annotations directly to PR diffs without needing reviewdog.

### 2.4 CodeClimate

- **URL**: <https://codeclimate.com>
- **What**: Hosted platform for code quality + test coverage trends
- **Output**: Per-commit delta report, maintainability index, coverage changes
- **Agent-consumable**: Yes — API returns JSON. But it's a SaaS platform, not a CLI tool.
- **Go ecosystem**: Supports Go natively (gocyclo, govet, etc.)
- **Key insight**: Good for **trend visibility** over time (is quality improving?). Less good as a PR gate because it's opinionated and can have false positives.

### 2.5 SonarQube

- **URL**: <https://www.sonarsource.com/products/sonarqube>
- **What**: Quality gates — predefined thresholds (bugs, vulns, coverage, duplication, maintainability rating)
- **Output**: Quality Gate status (PASS/FAIL) + SARIF/JSON detailed reports
- **Agent-consumable**: Yes — Web API returns structured data. SARIF for CI consumption.
- **Go ecosystem**: SonarGo plugin available; `golangci-lint` SARIF output feeds into SonarQube
- **Key insight**: The **quality gate** concept is the strongest pattern — you define "code must be A-rated on new code" and it auto-blocks. But it's heavyweight for small projects.

### 2.6 Codacy

- **URL**: <https://www.codacy.com>
- **What**: Automated code review platform. 40+ languages, static analysis + AI (Gemini)
- **Output**: PR inline comments + dashboard + CLI
- **Agent-consumable**: Yes — API, webhooks, VS Code extension
- **Go ecosystem**: Supports Go natively
- **Pricing**: Free for small teams (~50 users with limited features), paid from $15/user/mo
- **Key insight**: Codacy is the most mature rule library (since 2012). The hybrid approach (deterministic + AI) is now the industry standard. Strongest option for polyglot codebases.

### 2.7 actionlint

- **URL**: <https://github.com/rhysd/actionlint>
- **Stars**: 4k
- **Language**: Go
- **What**: Static checker for GitHub Actions workflow YAML files
- **Checks**: Syntax, strong `${{ }}` expression type checking, action input/output validation, shellcheck integration for inline scripts, glob validation, cron syntax
- **Output**: Structured per-file, per-line errors. JSON output available.
- **Agent-consumable**: Yes — JSON output, designed for CI integration
- **Go ecosystem**: Written in Go. Integrates with reviewdog natively.
- **Key insight**: Catches errors that LLMs routinely generate in workflow files (wrong `branch:` vs `branches:`, type mismatches in expressions). Zero false positives in syntax/type checks.

### 2.8 hadolint

- **URL**: <https://github.com/hadolint/hadolint>
- **Language**: Haskell
- **What**: Dockerfile linter following Docker best practices
- **Checks**: Pin image versions, no `:latest`, no `sudo`, `COPY` over `ADD`, layer consolidation, shellcheck for RUN commands
- **Output**: Structured DL-coded errors + line info
- **Agent-consumable**: Yes — JSON output via `--format json`
- **Go ecosystem**: Can be integrated via reviewdog. For Go-native, see `dockerfile_lint` but hadolint is the standard.
- **Key insight**: Hadolint catches the exact Dockerfile mistakes LLM agents make (unpinned versions, missing `--no-cache`, multi-stage layering issues).

### 2.9 yamllint

- **URL**: <https://github.com/adrienverge/yamllint>
- **Language**: Python
- **What**: YAML file linter (syntax + style rules)
- **Output**: Structured per-line errors
- **Agent-consumable**: Yes — parsable output
- **Go ecosystem**: Reviewdog supports yamllint. Go-native: `yamllint` itself is fine (language-agnostic).

---

## 3. PR Size Gates

### Tools:

- **`CodelyTV/pr-size-labeler`** (GitHub Action, 390 stars): Labels XS/S/M/L/XL/XXL by changed lines. Can `fail_if_xl`.
- **`noqcks/pull-request-size`** (GitHub App, 172 stars): Labels + optionally blocks. Strips generated files.
- **`actions/labeler`** (GitHub official, 2.5k stars): `max-files-changed: 100` skips file-based labeling if too many files.
- **`priyanceo/prcheck`** (Python): Diff threshold + path-based labeling.

### Deterministic output:

Yes — line count is an integer. Label assignment is deterministic.

### Agent-consumable:

Yes — labels are first-class GitHub metadata. Agents can query them via API.

### Go ecosystem:

No Go-specific tool needed for this (it's just line counting). Write a 50-line Go program using `go-git` if needed.

---

## 4. PR Description Quality Gates

### What to check:

- PR template hasn't been deleted (compare with canonical template via diff)
- Required sections filled (checking for placeholder text like `[description]`)
- Ticket/issue referenced (regex for `#123`, `PROJ-123`, URL patterns)
- Minimum description length (e.g., >10 words)
- Screenshots for UI changes (markdown image syntax)

### Tools:

- **`rohitjmathew/pr-template-enforcer`**: Compares PR body against template using diff library
- **Danger/CI script**: Custom rule checking `danger.github.pr.body`
- **`Conventional Pull Request`** action (conde-nast): Lints PR title + body

### Deterministic output:

Yes — string matching against template. Pass/fail is unambiguous.

### Agent-consumable:

Yes — CI status check (pass/fail) is a boolean.

### Go ecosystem:

No specialized Go tool. Can use `regexp` + `strings` in a 30-line Go CI script.

---

## 5. Conventional Commit PR Title Enforcement

### Tools:

- **`CondeNast/conventional-pull-request-action`** (16 stars): Uses commitlint + config-conventional. Validates PR title + commit messages.
- **`lw-ci/action-conventional-pull-request`** (1 star, v4): Lightweight, uses commitlint for PR title.
- **`commitlint`** + **`husky`**: Local + CI enforcement.
- **`amannn/action-semantic-pull-request`**: Most popular (700+ stars). Enforces Conventional Commits on PR titles only.

### Deterministic output:

Yes — regex-based type extraction. `feat:`, `fix:`, `chore:`, etc. Pass/fail is unambiguous.

### Agent-consumable:

Yes — CI status check. Agents can also read the failure message.

### Go ecosystem:

No Go-native equivalent needed. commitlint is Node.js but runs in any CI. `amannn/action-semantic-pull-request` is the best-in-class GitHub Action.

---

## 6. WIP Detection

### Mechanisms:

- **GitHub**: Draft PRs cannot be merged by default. `gh pr view --json isDraft`
- **Title prefix**: Block PRs with `[WIP]`, `WIP:`, `[DRAFT]` in title
- **GitLab**: Native `WIP:` / `Draft:` prefix prevents merge
- **Label-based**: Block PRs with `do-not-merge` label

### Tools:

- **`samholmes/block-wip-pr-action`**: Blocks PRs with fixup!/squash! commits
- **Custom CI**: `if: github.event.pull_request.draft == true` then skip CI or `exit 1`

### Deterministic output:

Yes — boolean. Is draft? Yes/no.

### Agent-consumable:

Yes — CI status check.

### Go ecosystem:

Trivial. `gh pr view --json isDraft --jq .isDraft` piped to Go program.

---

## 7. Auto-Labeling Based on Changed Files

### Tools:

- **`actions/labeler`** (GitHub official, 2.5k stars): V6 supports glob patterns per label. Example: `frontend:` → `src/frontend/**`. Supports `max-files-changed` limit.
- **`rkneela0912/auto-label-by-path`**: JSON path→label mapping. Team-based, component-based, file-type-based.
- **`priyanceo/prcheck`**: Python-based, combines path + diff threshold.

### Deterministic output:

Yes — glob match against file paths. Deterministic per commit.

### Agent-consumable:

Yes — labels are GitHub API entities. Agents can read/write.

### Go ecosystem:

Use `path/filepath.Match` or `doublestar.Match` for glob matching. A 100-line Go program if you don't want GitHub Actions.

---

## 8. Change Impact Analysis

### What it produces:

"This PR changes 3 files in 2 packages, affecting 12 downstream consumers across 4 services."

### Tools:

- **Static dependency graph**: `go list -json ./...` gives import graph. Parse to find reverse dependencies.
- **`go-callvis`**: Visual call graph
- **Moderne / OpenRewrite**: Java-focused, but the pattern applies
- **Custom Go tool**: `depgraph` style — walk imports both directions
- **Graph-based tools** (PuppyGraph, OneUptime): For microservice-level dependency mapping

### Deterministic output:

Yes — import graph is a compile-time property. For Go, `go list -json ./...` returns deterministic import data.

### Agent-consumable:

Yes — JSON. Provide `changed_files`, `affected_packages`, `downstream_consumers` to agent.

### Go ecosystem:

Excellent. `go list -json -deps ./...` gives full dependency tree. Build a `//go:generate` tool that outputs a `dependency_map.json`.

---

## 9. Merge Conflict Detection Beyond Git

### Types:

- **Textual conflicts**: Git catches these (line-level merge markers)
- **Semantic conflicts**: Code merges cleanly but behavior changes break. Example: function rename in one branch, new caller in another.
- **Logical conflicts**: Both branches change the same logic differently, no syntactic conflict.

### Tools:

- **`git merge-tree`**: More advanced than `git merge` for previewing conflicts without touching working tree. Available since Git 2.38+.
- **MergeGuard** (VS Code extension): Uses `git merge-tree` to predict conflicts before opening PR
- **ConflictLens** (research): LLM-based semantic conflict detection, 0.91 precision / 0.76 recall
- **Syncwright** (Go CLI): AI-powered merge conflict resolution. Written in Go, has GitHub Action.
- **Unit test generation approach**: Generate regression tests from both branches, see if they disagree.

### Deterministic output:

`git merge-tree` output is deterministic (same input → same output). Semantic conflict detection is probabilistic.

### Agent-consumable:

Yes — `git merge-tree` outputs JSON-parsable conflict regions. Semantic tools produce structured reports.

### Go ecosystem:

- `Syncwright` is Go-based (<https://github.com/NeuBlink/syncwright>)
- `go-git` library can do merge simulations programmatically

---

## 10. The Boundary: Deterministic Tools vs LLMs

### What deterministic tools catch that LLMs miss:

| Category            | Tool                        | LLM Failure Mode                                |
| ------------------- | --------------------------- | ----------------------------------------------- |
| Syntax errors       | `go vet`, `golangci-lint`   | LLMs hallucinate valid syntax in wrong contexts |
| Type mismatches     | `actionlint` for GH Actions | LLMs generate wrong `${{ }}` types              |
| Schema violations   | `yamllint`, `hadolint`      | LLMs invent config keys that don't exist        |
| Regressions         | `staticcheck`, `errcheck`   | LLMs miss silent error swallowing               |
| Deprecated APIs     | `staticcheck` SA1019        | LLMs use outdated APIs (training cutoff)        |
| Secrets in code     | `gitleaks`, `truffleHog`    | LLMs sometimes generate tokens/keys inline      |
| Binary file changes | `danger` / diff check       | LLMs don't visualize binary impacts             |
| CI config           | `actionlint`                | LLMs generate non-working GH Actions YAML       |

### What LLMs catch that deterministic tools miss:

| Category        | Example                                         | Tool Equivalent            |
| --------------- | ----------------------------------------------- | -------------------------- |
| Logic intent    | "This loop has off-by-one"                      | None — needs understanding |
| Cross-context   | "This change breaks module X two packages away" | Static analysis limited    |
| Naming          | "Variable name is misleading"                   | None                       |
| Security nuance | "This auth bypass is exploitable in edge case"  | SAST has false negatives   |
| API design      | "This function signature is confusing"          | None                       |
| Edge cases      | "What if user is null here?"                    | Dataflow analysis partial  |
| Code smells     | "Too many responsibilities in this function"    | Cyclomatic complexity only |

### The win: Deterministic gates as "pre-filter" for LLM agents

```
LLM produces code
  → golangci-lint catches syntax + type errors
  → actionlint catches CI config errors
  → hadolint catches Dockerfile errors
  → gitleaks catches secrets
  → PR size gate catches too-large changes
  → conventional commit check
  → description quality check
  → WIP detection
  → auto-label for routing
  → THEN: dedicated LLM review agent (narrower scope, less noise)
  → THEN: human review (fewer things to check)
```

### Source references:

- <https://www.repowise.dev/blog/comparisons/best-ai-code-review-tools> (2026)
- <https://www.tanagram.ai/blog/ai-agent-architecture-patterns-for-code-review-automation> (2025)
- <https://www.gresearch.com/news/building-a-code-review-tool-the-llm-patterns-that-actually-work> (2026)
- <https://www.codeant.ai/blogs/best-pull-request-automation-tools-in-2026> (2026)
- <https://dev.to/rahulxsingh/how-to-automate-code-reviews-in-2026-complete-setup-guide-16b5> (2026)
- <https://www.deployhq.com/blog/the-perfect-pull-request-best-practices-for-collaborative-development> (2026)
- <https://github.com/alibaba/open-code-review> (battle-tested at Alibaba scale)
