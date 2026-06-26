# Deterministic Architecture Enforcement — Research Report

**Author:** opencode (deepseek-v4-flash-free)
**Date:** 2026-06-19
**Context:** Agent Architecture Optimization Plan (context reduction, subagent decomposition, deterministic gates)

---

## Part 1: Go Architecture Boundary Enforcement (D1-D6)

### Tool Landscape

| Tool                           | Stars | Approach                              | Best For                         | Output                          |
| ------------------------------ | ----- | ------------------------------------- | -------------------------------- | ------------------------------- |
| **go-arch-lint**               | ~500  | YAML layer rules, component deps      | VSA boundary enforcement         | `--json`, exit codes            |
| **golangci-lint + depguard**   | ~19k  | Import allowlist/blocklist            | Banning forbidden pkgs           | `--out-format json`             |
| **cht-go-lint**                | new   | 41 rules incl DDD, naming, structure  | Naming, structure, layers        | `--format json`, GA annotations |
| **arch-go**                    | ~255  | `go test`-style architecture tests    | Unit-test-integrated rules       | Go test output                  |
| **archlint**                   | new   | 229 checks, SOLID metrics, MCP server | Deep analysis, agent integration | JSON + MCP                      |
| **go-ruleguard**               | ~500  | Pattern-based custom linting DSL      | Dynamic rules, no recompilation  | Via golangci-lint               |
| **go-critic**                  | ~2k   | ~100 checks, extensible               | Code style + diagnostics         | Via golangci-lint               |
| **go-consistent**              | ~364  | Auto-detect most-used pattern         | Enforce consistency              | Text, JSON                      |
| **Custom go/analysis**         | —     | AST-based analyzers                   | Project-specific sharp edges     | `go vet -json` format           |
| **40-line go list -deps bash** | —     | Transitive import check               | Quick single-boundary gate       | stdout, exit code               |
| **loov/goda**                  | ~1.7k | Query-based dep graph toolkit         | Analysis + visualization         | Graphviz DOT, Mermaid           |
| **dep-tree**                   | ~1.7k | File-level dep rules, glob patterns   | Fine-grained per-file rules      | CLI, 3D viz                     |
| **go-cyclic**                  | small | Circular dep chain display            | Cycle debugging                  | CLI chain output                |

### Winning Stack for VSA

```
internal/ (compiler-enforced) + go-arch-lint (YAML rules) + golangci-lint depguard (import bans)
+ custom go/analysis analyzers (sharp edges)
```

**Real-world reference:** CyberAgent (Japan) runs `go-arch-lint` across 50+ Go services in a monorepo, dynamically validating only changed services via `yq` + GitHub Actions + Reviewdog for inline PR comments.

### Key Architecture Patterns

**Hexagonal architecture rules:**

```
domain  →  (no platform deps, only stdlib)
ports   ←  domain (interfaces defined in domain or ports)
adapters → ports, domain
app      → adapters, domain
```

**Standard 3-layer:**

```
transport → service → store
```

**Vertical slice isolation:**

```
internal/orders/... → may NOT import internal/billing/...
internal/billing/... → may NOT import internal/orders/...
```

---

## Part 2: ADR Enforcement Tools & Patterns

### Management Tools

| Tool               | Language | Deterministic   | Structured Output    | Status               |
| ------------------ | -------- | --------------- | -------------------- | -------------------- |
| adr-tools (npryce) | Bash     | Yes (numbering) | No (Markdown prose)  | Unmaintained, 5.5k★  |
| phodal/adr         | Node.js  | Yes             | **JSON, HTML, CSV**  | Active               |
| dotnet-adr         | .NET     | Yes             | Limited              | Active               |
| MADR v4            | Template | Yes (structure) | Optional YAML FM     | Active, 2.3k★        |
| Structured MADR    | Template | Yes             | **Required YAML FM** | Active, AI-parseable |

### ADR Linting & Quality Gates

- **markdownlint** — enforce required headings, line length, links. JSON output.
- **YAML frontmatter schema** — validate `title`, `status`, `author`, `tags` against JSON Schema
- **Filename pattern** — enforce `NNNN-title.md` convention
- **Status lifecycle** — validate transitions (proposed → accepted → deprecated → superseded)

### ADR Drift Detection

| Tool         | Approach                                                                                 | Status                   |
| ------------ | ---------------------------------------------------------------------------------------- | ------------------------ |
| **Erode**    | Architecture-as-code (LikeC4) vs code diff. Multi-stage AI pipeline.                     | Open source, active 2026 |
| **SquiglOS** | ADR as ground truth, invariant indexing. Drift with severity × spread × recency.         | Private beta 2026        |
| **Archyl**   | C4 model in YAML + ADR linking. AI analysis for SPOFs, circular deps.                    | Active product           |
| **Archgate** | Turns ADRs into `.rules.ts` companion files. `archgate check` reports file + line + ADR. | Open source              |

### Key Finding: ADR Format Matters

- **Prose-format ADRs** → ~40% AI compliance
- **Constraint-oriented ADRs** (code blocks, tables, enumerated rules) → **100% compliance** with pre-commit hooks
- **YAML frontmatter** enables agent consumption: filtering, searching, drift detection

### Monorepo ADR Layout

```
monorepo/
├── apps/web/docs/adrs/         # web-specific decisions
├── apps/api/docs/adrs/         # api-specific decisions
├── packages/shared/docs/adrs/  # shared library decisions
├── docs/adrs/                  # org-wide / cross-cutting decisions
```

---

## Part 3: Deterministic CI Gates (Architecture Fitness Functions)

### The Golden Pipeline (convergent across Google, Netflix, Uber, Shopify, Monzo)

```
PR → Deterministic Gate → Structured Report → LLM Deep Analysis → Human
         (blocking)           (JSON)            (advisory)        (owns merge)
```

### 14 Gate Categories

| #   | Gate                              | Tool/Approach                    | Output          | Blocking?  |
| --- | --------------------------------- | -------------------------------- | --------------- | ---------- |
| 1   | Technology stack                  | `grep '^go ' go.mod`             | pass/fail       | ✅         |
| 2   | Vertical slice layout (D1)        | Directory existence checks       | pass/fail       | ✅         |
| 3   | D5 boundary rules                 | `go list -json ./...` + jq       | violations list | ✅         |
| 4   | D6 dependency DAG                 | Import graph analyzer            | cycle report    | ✅         |
| 5   | No cross-slice SQL joins          | `grep -rn 'JOIN' store/`         | locations       | ⚠️ Partial |
| 6   | TDD (tests present)               | File existence per new `.go`     | pass/fail       | ✅         |
| 7   | DB migrations                     | Sequential naming, up/down pairs | pass/fail       | ✅         |
| 8   | Security (bcrypt, SQLi)           | `grep` patterns                  | locations       | ⚠️ Partial |
| 9   | Branch & commits                  | `git branch` regex, commitlint   | pass/fail       | ✅         |
| 10  | DoD checklist                     | Composite of above               | summary         | ✅         |
| 11  | Frontend (ESLint + Prettier, tsc) | `make fe-ci`                     | pass/fail       | ✅         |
| 12  | Infrastructure (healthz)          | Code reading                     | ❌ LLM only     | ❌         |
| 13  | Test coverage delta               | `go test -coverprofile`, diff    | % change        | ✅         |
| 14  | Binary size regression            | `ls -l`, threshold compare       | pass/fail       | ✅         |

### Q3: Shell Scripts vs Go Binary — Best Practice

**Both, layered by complexity:**

| Tier          | Approach                  | Checks                                                          | Why                                        |
| ------------- | ------------------------- | --------------------------------------------------------------- | ------------------------------------------ |
| Simple (80%)  | Shell scripts             | Directory structure, naming, migration sequence, file existence | Trivially expressed, zero deps, sub-second |
| Complex (20%) | Go binary (`go/analysis`) | Import graph, DAG enforcement, AST-based rules, type-checking   | Needs whole-program analysis               |

**Proposed scripts/gates/ layout:**

```
scripts/gates/
├── run-all.sh                    # Runs all gates, outputs JSON report
├── check-stack.sh                # Go version, module path
├── check-d1-layout.sh            # Vertical slice structure validation
├── check-d5-boundaries.sh        # Import boundary violations
├── check-d6-dag.sh               # Dependency DAG acyclicity
├── check-tdd-coverage.sh         # Test file existence for new code
├── check-migrations.sh           # Sequential naming, up/down pairs
├── check-security.sh             # Bcrypt cost, SQL injection patterns
├── check-branch.sh               # Branch naming, conventional commits
└── check-scope-drift.sh          # Only ticket-related files changed
```

### Architecture Fitness Function Hierarchy

From _Building Evolutionary Architectures_ (Ford, Parsons, Kua):

| Dimension  | Types                  | Example                                                  |
| ---------- | ---------------------- | -------------------------------------------------------- |
| Scope      | Atomic vs Holistic     | Complexity of one class vs security×scalability tradeoff |
| Cadence    | Triggered vs Continual | PR gate vs production monitoring                         |
| Evaluation | Static vs Dynamic      | Import analysis vs response time                         |
| Automation | Automated vs Manual    | CI gate vs architecture review board                     |

**Properties of good deterministic gates:**

- **Objective** — same code always gets same result
- **Automated** — no human judgement required
- **Fast** — completes in seconds/minutes
- **Actionable** — clear failure message + remediation hint
- **Versioned** — rules live in repo alongside code

---

## Part 4: Hybrid LLM + Deterministic Review

### The Consensus Pattern

```
┌──────────────────────────────────────────────────────┐
│  L1: Syntax & Format (golangci-lint, Prettier)        │ ← zero cost, blocking
│  L2: Static Analysis (Semgrep, CodeQL)                │ ← deterministic, blocking
│  L3: Security Scan (Trivy, Snyk, govulncheck)         │ ← blocking on critical
│  L4: AI Semantic Review (LLM)                         │ ← advisory, deep reasoning
└──────────────────────────────────────────────────────┘
```

**Rule:** If pipeline failure blocks release, keep it deterministic. LLM only where human judgment was already required.

### What Each Layer Owns

| Check Type                          | Owner                      | Why                          |
| ----------------------------------- | -------------------------- | ---------------------------- |
| Formatting, imports, dead code      | Linter                     | Zero FP, millisecond         |
| Known bug patterns (NPE, SQLi, XSS) | Deterministic ruleset      | High precision, auditable    |
| Dependency CVEs                     | Snyk/Trivy                 | Compliance requirement       |
| Type errors, nil pointer derefs     | Compiler + static analysis | Must be correct              |
| Business logic correctness          | LLM                        | Needs semantic understanding |
| Architecture violations             | LLM                        | Context-dependent            |
| Missing edge cases, test gaps       | LLM                        | Judgement call               |
| Security logic flaws                | LLM                        | Pattern of attack changes    |

### Token Savings with Deterministic Pre-processing

| Team/Tool                       | Savings                      | Context                            |
| ------------------------------- | ---------------------------- | ---------------------------------- |
| code-review-graph (Next.js)     | **49x** (739k → 15k tokens)  | 27,732-file repo                   |
| code-review-graph (avg 6 repos) | **8.2x**                     | 13 heterogeneous commits           |
| Contexly (logic skeletons)      | **25x** (197k → 7.7k tokens) | Function signatures + impact paths |
| Local-splitter (research)       | **45-79%**                   | Cloud token savings via routing    |

**Rule of thumb:** Expect **6-10x** for most codebases, **30-50x** for monorepos.

### Failure Modes of Pure LLM Review

| Failure                                 | Mitigation                                                   |
| --------------------------------------- | ------------------------------------------------------------ |
| Hallucination (flags non-issues)        | Verification pass (AST verification, deterministic re-check) |
| Line drift (wrong line refs)            | Deterministic positioning via pre-processor                  |
| Skipped files (doesn't read everything) | Deterministic file selection + bundling (pre-model)          |
| High cost (full codebase every time)    | Blast radius / code maps / incremental review                |

---

## Part 5: Industry References

### Blog Posts & Case Studies

| Company              | Topic                                             | URL                                                                                    |
| -------------------- | ------------------------------------------------- | -------------------------------------------------------------------------------------- |
| CyberAgent           | go-arch-lint across 50+ services                  | `developers.cyberagent.co.jp/blog/archives/59647`                                      |
| lastminute.com       | dependency-cruiser + danger.js for MR comments    | `technology.lastminute.com/how-we-enforce-architecture-boundaries-at-scale-on-our-app` |
| DEV (Gabriel Anhaia) | The Dependency Rule as Go CI check (3 approaches) | `dev.to/gabrielanhaia/the-dependency-rule-written-as-a-ci-check-in-go-51co`            |
| Banandre             | Why your architecture should fail the build       | `www.banandre.com/blog/why-your-architecture-should-fail-the-build`                    |
| Monzo                | Migrating 2800 microservices + semgrep            | `monzo.com/blog/how-we-run-migrations-across-2800-microservices`                       |
| Monzo                | Go monorepo consistency                           | `monzo.com/blog/2022/09/29/migrating-our-monorepo-seamlessly-from-dep-to-go-modules`   |

### Key Numbers from Industry

- **45%** of PR review comments are codifiable into deterministic gates (Shopify case study)
- **30-day Rust platform study**: 4 constrained ADRs → 67 compliant commits, zero drift incidents, **70% code review reduction**
- **Alibaba open-code-review**: 200+ deterministic rules in first layer, 8 goroutines for parallel sub-agent review
- **Google AutoCommenter**: deployed to 10k+ developers daily, focuses on best-practice violations
- **Meta RADAR**: risk-calibrated auto-approval pipeline, 3 validation layers (static heuristics, Diff Risk Scoring, Automated Code Review)

### GitHub Repos of Interest

| Repo                             | Stars | Notes                                  |
| -------------------------------- | ----- | -------------------------------------- |
| `fe3dback/go-arch-lint`          | ~500  | YAML architecture linting              |
| `loov/goda`                      | ~1.7k | Go dependency analysis toolkit         |
| `channel-io/cht-go-lint`         | new   | 41 rules, JSON output, GA annotations  |
| `fdaines/arch-go`                | ~255  | `go test`-style architecture tests     |
| `mshogin/archlint`               | new   | 229 checks, MCP server for Claude Code |
| `quasilyte/go-ruleguard`         | ~500  | Pattern-based custom linting DSL       |
| `rogpeppe/modcop`                | small | Explicit dependency whitelisting       |
| `aethiopicuschan/goimportsruler` | small | Glob-based import rules, GitHub Action |

---

## Part 6: Recommendations for This Codebase

### Priority Order

1. **Part 3 first (deterministic gates)** — biggest ROI: 45% of review work automated, agents get cheaper + more reliable
2. **Part 1 second (context reduction)** — simplest change: edit agent markdown files, 72-83% token reduction per agent
3. **Part 2 third (subagent decomposition)** — most architectural, verify nested spawning support first

### Recommended Gate Stack

```makefile
# Makefile targets
check-arch: check-layers check-imports check-circular check-d1 check-d5 check-d6
check-ci: check-arch check-migrations check-branch check-security check-coverage
review-gates: make check-ci 2>&1 | tee .build/gates-report.json
```

### Script vs Binary Decision (Q3)

Implement as **shell scripts** (fast, zero deps) for checks 1-9 (directory structure, naming, imports via grep).
Implement as **Go binary** for checks 3/4 (import graph, DAG enforcement) using `go list -json ./...` + `go/analysis`.

### ADR Recommendation (from Q2/Option C)

Restructure `conventions.md` with clear section markers (`§1`, `§2`, etc.) so agents can reference specific sections. Add YAML frontmatter block for machine parsing.

---

_Generated by opencode (deepseek-v4-flash-free) from 11 parallel research subagents covering Go architecture tools, ArchUnit ecosystem, ADR enforcement, CI gates, import graph analysis, code quality gates, architecture fitness functions, PR review automation, structure-based linting, build system visibility, and hybrid LLM+deterministic review patterns._
