# Deterministic Code Quality Gates Reference

> **Purpose:** PASS/FAIL gates for CI that catch what linters + LLM review miss.
> **Principle:** Every gate exits 0 (pass) or non-zero (fail with structured output).
> **Target:** A `make review-gates` target that chains them cheapest-first.

---

## 1. Test Coverage Gate (Delta)

**Problem:** Absolute coverage (%) is meaningless on an existing codebase. A PR adding 1 line of uncovered code can show 80%+ project coverage. The gate must check **changed lines only**.

**Tools:**
| Tool | Scope | Run |
|------|-------|-----|
| `go test -coverprofile` | Go stdlib baseline | `go test -coverprofile=coverage.out -covermode=atomic ./...` |
| `gocovdiff` | Go, diff coverage | `gocovdiff -cov coverage.out -target-delta-cov 80` |
| `go-covercheck` | Go, diff mode | `go-covercheck --diff-from origin/main --block-threshold 80` |
| `diff-cover` | Cobertura/LCov → any lang | `diff-cover coverage.xml --fail-under=80` |

**Structured output:**

- `gocovdiff`: Markdown table + uncovered line annotations
- `go-covercheck`: JSON/YAML/MD (`--format json`)
- `diff-cover`: Text with per-file breakdown

**Exit code:** 0 = coverage ≥ threshold, 1 = below threshold

**What LLM misses:** An LLM cannot compute whether every changed line is covered. It guesses. gocovdiff checks byte-exact coverage blocks against `git diff`.

**Integration:**

```makefile
review-gates: coverage-delta

coverage-delta: ## Gate: changed-line coverage >= 80%
	go test -coverprofile=coverage.out -covermode=atomic ./...
	go install github.com/vearutop/gocovdiff@latest
	gocovdiff -cov coverage.out -target-delta-cov 80
```

---

## 2. Benchmark Regression Gate

**Problem:** Performance regressions invisible to linters. LLMs cannot run benchmarks or compute statistical significance.

**Tools:**
| Tool | Function | Run |
|------|----------|-----|
| `benchstat` | Statistical A/B comparison | `benchstat old.txt new.txt` |
| `go-verdict` | Wraps benchstat, structured verdict | `verdict --require new-wins --min-delta 5` |
| `benchdiff` | Automated checkout + benchstat | `benchdiff --old=main --threshold=5` |

**Structured output:**

- benchstat: Text table + p-value + % delta
- CSV format: `--format csv`
- go-verdict: JSON (`--format json`), single verdict line

**Exit code:**

- benchstat: **0 always** (no flag for fail-on-change — must wrap)
- go-verdict: 0 if `--require` matches, 1 otherwise
- Custom wrapper: parse benchstat output, exit 1 if delta > threshold AND p < 0.05

**What LLM misses:** Reproducibility. benchstat runs Mann-Whitney U-test with p-value. An LLM cannot measure wall-clock time or compute statistics from benchmark data.

**Integration:**

```makefile
review-gates: bench-regression

bench-regression: ## Gate: no >5% performance regression vs main
	git checkout main && go test -bench=. -count=10 ./... > /tmp/bench-main.txt
	git checkout - && go test -bench=. -count=10 ./... > /tmp/bench-head.txt
	go install golang.org/x/perf/cmd/benchstat@latest
	benchstat /tmp/bench-main.txt /tmp/bench-head.txt
	scripts/bench-gate.sh /tmp/benchstat.txt  # exits 1 if regression > 5%
```

**Reference:** RFC from sveltego (`binsarjr/sveltego#105`) — captures exact threshold (5% delta, p<0.05), baseline-on-main fetch pattern, and override mechanism via commit message.

---

## 3. Binary Size Gate

**Problem:** Dependencies, codegen, or dead code bloat the artefact silently. LLM cannot predict compile output size.

**Tools:**
| Tool | Function | Run |
|------|----------|-----|
| `godeps-guard` | Go binary + dep analysis | `godeps-guard check --base origin/main` |
| `sizecheck` (ad-hoc) | `stat` + threshold comparison | `scripts/binary-size-compare.sh` |
| Adobe `sizewatcher` | Any build artifact, multi-comparator | `sizewatcher` (Node) |
| `thresh` | JS artifact sizes, CI adapter | `thresh check` |

**Structured output:**

- godeps-guard: JSON summary, PR comment mode
- Custom: markdown table with base/PR/delta/%

**Exit code:** 0 = within threshold, 1 = exceeded

**What LLM misses:** Actual linker output size. LLM cannot simulate what the compiler includes.

**Integration:**

```makefile
review-gates: binary-size

binary-size: ## Gate: binary growth < 5% vs main
	go build -o /tmp/branch.bin ./cmd/server
	git stash && git checkout main
	go build -o /tmp/main.bin ./cmd/server
	git checkout - && git stash pop
	scripts/binary-size-gate.sh /tmp/main.bin /tmp/branch.bin 5
```

**Reference:** elastic/cloudbeat (#3611) — 5 MiB threshold, sticky PR comment, reusable script pattern.

---

## 4. API Compatibility Gate

**Problem:** Removing/changing exported API breaks downstream consumers silently. Linters don't check cross-version compatibility.

**Tools:**
| Tool | Scope | Run |
|------|-------|-----|
| `apidiff` (golang.org/x/exp) | Go packages/modules | `apidiff -m -incompatible old@tag .` |
| `go-apidiff` | Go, git-aware | `go-apidiff origin/main` |
| `oasdiff` | OpenAPI specs | `oasdiff breaking base.yaml revision.yaml` |
| `buf breaking` | Protobuf schemas | `buf breaking --against '.git#branch=main'` |
| `openapi-diff` (Java) | OpenAPI | `openapi-diff --fail-on-incompatible old new` |

**Structured output:**

- apidiff: Text (`TextIncompatible`), JSON via Go API
- oasdiff: JSON/YAML/MD/HTML/GitHub Annotations (`--format json`)
- buf breaking: JSON (`--error-format=json`), SARIF

**Exit code:**

- apidiff: 0 = compatible, 1 = breaking, 2 = error
- oasdiff: 0 = no breaking changes, 1 = breaking found
- openapi-diff: 0/1 via `--fail-on-incompatible`

**What LLM misses:** apidiff uses Go type-checker (full `go/types` analysis). It catches interface satisfaction, struct field removal, signature changes at the compiler level. LLMs hallucinate "this looks compatible."

**Integration:**

```makefile
review-gates: api-compat

api-compat: ## Gate: no breaking API changes vs main
	go install golang.org/x/exp/cmd/apidiff@latest
	git stash && git checkout main
	apidiff -m -w /tmp/api-baseline.out ./...
	git checkout - && git stash pop
	apidiff -m -incompatible /tmp/api-baseline.out ./...
```

---

## 5. License Compliance Gate

**Problem:** New dependency introduces GPL/AGPL → legal exposure. Linters don't check licenses.

**Tools:**
| Tool | Scope | Run |
|------|-------|-----|
| `go-licenses` | Go deps | `go-licenses check ./... --disallowed_types=forbidden` |
| `godeps-guard` | Go deps + license | `godeps-guard check --base origin/main` (in `.godepsguard.yaml`) |
| `license-checker` | npm | `license-checker --failOn GPL-3.0` |
| `go-license-audit` | Multi-ecosystem | `go-license-audit -strict -path .` |
| `Comply` | Multi-ecosystem + AI usage analysis | `comply scan . --ci --fail-on critical` |
| `SKY UK licence-compliance-checker` | Go vendor/modules | `licence-compliance-checker -r GPL -r AGPL --check-go-modules` |

**Structured output:**

- go-licenses: CSV to stdout
- go-license-audit: JSON, Markdown
- Comply: JSON with severity tiers, GitHub annotations
- godeps-guard: JSON, CycloneDX SBOM

**Exit code:** 0 = compliant, 1 = violation found

**What LLM misses:** Actual SPDX license detection from `LICENSE` files is a text-classification problem (go-license-detector uses TF-IDF + classifiers). LLMs hallucinate license types.

**Integration:**

```makefile
review-gates: license-check

license-check: ## Gate: no GPL/AGPL dependencies introduced
	go install github.com/google/go-licenses@latest
	go-licenses check ./... --disallowed_types=forbidden
```

---

## 6. Vulnerability Scanning Gate

**Problem:** New vulnerability in dependency tree. No linter checks CVE databases.

**Tools:**
| Tool | Scope | Run |
|------|-------|-----|
| `govulncheck` | Go vulns | `govulncheck ./...` |
| `go-vuln-gate` | Go vulns + CVSS filter | `go-vuln-gate --threshold 7.0 ./...` |
| `trivy` | Multi-ecosystem, containers, infra | `trivy fs --exit-code 1 --severity CRITICAL,HIGH .` |
| `snyk` | Multi-ecosystem | `snyk test --severity-threshold=high` |

**Structured output:**

- govulncheck: text, JSON, SARIF, OpenVEX
- trivy: table, JSON, SARIF, CycloneDX, template
- go-vuln-gate: JSON with CVSS scores, SARIF

**Exit code:**

- govulncheck: 0 = no vulns, 1 = vulns found (text mode only)
- trivy: configurable via `--exit-code`
- snyk: 0/1 based on severity threshold

**What LLM misses:** Live CVE database lookup. LLM has stale training data and no access to NVD. govulncheck uses `golang.org/x/vuln/vulncheck` — call-graph analysis to determine if the vulnerable symbol is actually reachable.

**Integration:**

```makefile
review-gates: vulncheck

vulncheck: ## Gate: no critical/high CVEs in deps
	govulncheck ./...
```

---

## 7. Generated Code Staleness Gate

**Problem:** Developer edits `.proto`/`.y`/`.rl` source, commits without regenerating → build breaks for everyone else. LLM cannot notice unless told.

**Tools:**
| Tool | Run |
|------|-----|
| `go generate` + `git diff` | `go generate ./... && git diff --exit-code` |
| `buf generate` + diff | `buf generate && git diff --exit-code` |
| Custom Makefile target | `make generate && git diff --exit-code` |

**Structured output:** `git diff` output showing changed files.

**Exit code:** `git diff --exit-code` exits 0 = clean, 1 = drift detected

**What LLM misses:** LLM sees the source and generated files but cannot verify which is authoritative. The only way to know is to re-run the generator and diff.

**Integration:**

```makefile
review-gates: check-generate

check-generate: ## Gate: generated code is up to date
	@echo "==> Checking generated code freshness..."
	go generate ./...
	git diff --exit-code || (echo "ERROR: Generated code out of date. Run 'go generate ./...' and commit."; exit 1)
```

**Reference:** microsoft/retina (#2146) — also checks for missing new files from `go generate` that aren't committed. cozystack (#2463) — adds pre-commit hook for same check.

---

## 8. TODOs/FIXMEs Gate

**Problem:** AI agents and developers leave TODO markers that become permanent technical debt. LLM reviewers rarely flag them.

**Tools:**
| Tool | Run |
|------|-----|
| `todo-scan` | `todo-scan check --max-new 0 --since origin/main` |
| `git-confirm` (pre-push hook) | `git config hooks.confirm.match TODO` |
| Custom grep + diff | `grep -rn 'TODO\|FIXME' --include='*.go' . \| sort > /tmp/new; diff ...` |
| `prevent-dangling-todos` | Pre-commit: TODOs must reference ticket |

**Structured output:**

- todo-scan: text, JSON (tags, authors, priorities, deadlines, per-package thresholds)
- git-confirm: interactive prompt per match

**Exit code:**

- todo-scan: 0 = within limits, 1 = exceeded
- grep: 0 = no matches, 1 = matches found

**What LLM misses:** LLM reads the diff and could see "TODO" but has no policy on whether it's allowed. `todo-scan check --max-new 0 --since main` is unambiguous — zero new TODOs, gate fails.

**Integration:**

```makefile
review-gates: todos

todos: ## Gate: no new TODOs/FIXMEs added vs main
	go install github.com/sotayamashita/todo-scan@latest
	todo-scan check --max-new 0 --since origin/main --block-tags FIXME,BUG
```

---

## 9. Dead Code Detection Gate

**Problem:** Unused functions, variables, types accumulate. LLMs are poor at whole-program reachability analysis.

**Tools:**
| Tool | Scope | Run |
|------|-------|-----|
| `deadcode` | Go unreachable functions | `deadcode -test ./...` |
| `staticcheck -checks U1000` | Go unused (single-package) | `staticcheck -checks U1000 ./...` |
| `staticcheck --unused.whole-program=true` | Go unused (cross-package) | `staticcheck --unused.whole-program=true -checks U1000 ./...` (deprecated — removed in newer versions) |
| `staticcheck -matrix` | Multi-build-tag dead code | feed build matrix via stdin |
| `unused` | Go unused (standalone, pre-2020) | Legacy |

**Structured output:**

- deadcode: `file:line: function F is dead` — machine-readable text
- staticcheck: SARIF, JSON, text, Checkstyle — configurable via `-f`

**Exit code:**

- deadcode: 0 = no dead code, 1 = dead code found
- staticcheck: 0 = clean, 1 = issues found (configurable via `-fail`)

**What LLM misses:** RTA (Rapid Type Analysis) is a whole-program, sound over-approximation. deadcode traces from `main()` + package initializers through interface dispatch using the type graph. LLMs cannot perform this analysis — they guess based on local context.

**Integration:**

```makefile
review-gates: dead-code

dead-code: ## Gate: no unreachable functions
	go install golang.org/x/tools/cmd/deadcode@latest
	deadcode -test ./...
```

**Note:** Use BOTH `deadcode` (whole-program RTA, for functions) AND `staticcheck -checks U1000` (per-package, for unused variables/types/constants).

---

## 10. OpenAPI/Protobuf Spec Drift Gate

**Problem:** Hand-edited generated spec files or implementation diverges from spec. Linters don't cross-reference.

**Tools:**
| Tool | Scope | Run |
|------|-------|-----|
| `oasdiff` | OpenAPI diff + breaking | `oasdiff breaking base.yaml revision.yaml` |
| `oasdiff changelog` | OpenAPI all changes | `oasdiff changelog base.yaml revision.yaml` |
| `oasdiff diff` | OpenAPI structural diff | `oasdiff diff base.yaml revision.yaml` |
| `oasdiff validate` | OpenAPI validity | `oasdiff validate spec.yaml` |
| `openapi-diff` (Java) | OpenAPI compatibility | `openapi-diff --fail-on-incompatible old new` |
| `buf breaking` | Protobuf | `buf breaking --against '.git#branch=main'` |
| `buf lint` | Protobuf style | `buf lint` |

**Structured output:**

- oasdiff: JSON, YAML, MD, HTML, GitHub Annotations, JUnit XML
- oasdiff breaking: Only breaking changes (exit 1 on any)
- oasdiff changelog: All changes (breaking + non-breaking)
- buf: JSON, text, SARIF

**Exit code:**

- oasdiff breaking: 0 = no breaking, 1 = breaking
- buf breaking: 0 = no breaking, non-zero = breaking

**What LLM misses:** Structural AST comparison of entire spec. oasdiff checks 470+ rules across all OpenAPI features. buf checks wire format compatibility (FIELD renaming ≠ PACKAGE-level move ≠ WIRE-level change). LLMs miss subtle compatibility breaks like enum value reordering or response code removal.

**Integration:**

```makefile
review-gates: openapi-drift proto-drift

proto-drift: ## Gate: no breaking protobuf changes vs main
	buf breaking --against '.git#branch=main'

openapi-drift: ## Gate: no breaking OpenAPI changes vs main
	git show main:openapi.yaml > /tmp/base-openapi.yaml
	oasdiff breaking /tmp/base-openapi.yaml openapi.yaml
```

---

## 11. SQL Migration Sequence Gate

**Problem:** Missing down migration, duplicate version numbers, gap in sequence, naming violation. LLMs don't verify file system invariants.

**Tools:**
| Approach | Run |
|----------|-----|
| Custom script | `scripts/check-migrations.sh` |
| `golang-migrate create -seq` | Generates correct pairs automatically |
| `migrate validate` | Validates migration files |

**Structure invariant (enforce with script):**

1. Every `<N>_*.up.sql` has a matching `<N>_*.down.sql`
2. No duplicate version numbers
3. Versions are sequential (no gaps in committed set)
4. Files in correct directory (e.g., `db/migrations/`)
5. Names contain only `{version}_{title}.up|down.sql`

**Exit code:** 0 = valid, 1 = violation

**What LLM misses:** LLM inspects one PR diff but cannot see the full migration directory state. A script enumerates all files and checks invariants deterministically.

**Integration:**

```makefile
review-gates: migration-check

migration-check: ## Gate: SQL migrations are properly sequenced and paired
	@scripts/check-migrations.sh
```

**Script pattern (check-migrations.sh):**

```bash
#!/bin/bash
set -euo pipefail
cd db/migrations
errors=0
for f in *.up.sql; do
  version="${f%%_*}"
  down="${version}_*.down.sql"
  compgen -G "$down" > /dev/null || { echo "Missing down for version $version"; errors=1; }
done
# Check no duplicate versions
ls *.sql | sed 's/_.*//' | sort | uniq -d | while read v; do
  echo "Duplicate version: $v"; errors=1
done
exit $errors
```

---

## 12. Secret Scanning Gate

**Problem:** AI agents or humans commit API keys, tokens, passwords. Linters don't detect high-entropy strings.

**Tools:**
| Tool | Run |
|------|-----|
| `gitleaks detect` | `gitleaks detect --source . --verbose --report-format sarif` |
| `trufflehog` | `trufflehog git file://. --since-commit HEAD~1` |
| `git secrets` | Pre-commit patterns |
| `ggshield` (GitGuardian) | `ggsecret scan ci` |
| `greengate scan` | Secrets + PII (26 patterns) + AST SAST |

**Structured output:** JSON, CSV, JUnit, SARIF (all supported by gitleaks), GitHub Annotations.

**Exit code:** 0 = no secrets, 1 = secrets found (configurable via `--exit-code`)

**What LLM misses:** Entropy analysis (Shannon entropy on base64/hex strings), regex pattern matching (AWS key format, GitHub tokens, RSA private keys). LLMs see "AKIAIOSFODNN7EXAMPLE" and may not flag it. gitleaks has 150+ built-in detectors.

**Integration:**

```makefile
review-gates: secret-scan

secret-scan: ## Gate: no secrets committed
	go install github.com/gitleaks/gitleaks@latest
	gitleaks detect --source . --verbose --report-format json
```

---

## 13. Misspelling Gate

**Problem:** Documentation, comments, error messages contain typos. Linters with spell-check disabled, LLM reviewers may miss.

**Tools:**
| Tool | Run |
|------|-----|
| `codespell` | `codespell --count --quiet-level=2 .` |
| `misspell` | `misspell -error .` |

**Structured output:** `file:line: column: misspelling` text.

**Exit code:**

- codespell: 0 = pass, 65 = misspellings found
- misspell: 0 = pass, 1 = errors

**What LLM misses:** `codespell` uses a curated dictionary of common misspellings (not a full dictionary — avoids false positives on domain terms). LLMs don't reliably catch "teh" vs "the", "recieve" vs "receive" in prose.

**Integration:**

```makefile
review-gates: spell-check

spell-check: ## Gate: no misspellings in docs/comments
	pip install codespell --quiet
	codespell --count --quiet-level=2 \
	  --skip="./vendor,./node_modules,./.git,./coverage.out" \
	  .
```

---

## 14. File Naming Convention Gate

**Problem:** Mixed naming conventions (snake_case, camelCase, kebab-case) across the project — aesthetics aside, this breaks tooling expectations (e.g., Go needs `snake_case` for test files).

**Tools:**
| Tool | Scope |
|------|-------|
| `pre-commit-filename-linter` | Multi-lang, configurable per pattern |
| `convention-lint` (Rust) | Cargo subcommand, glob patterns |
| `python_filename_linter` | Python PEP8 snake_case |
| `ESLint naming convention` | JS/TS |
| Pre-commit `fail` language | Custom regex pattern |
| Custom script | `scripts/check-filenames.sh` |

**Structured output:** `file: N does not follow S convention` per violation.

**Exit code:** 0 = all pass, 1 = violation

**What LLM misses:** LLM sees individual file names in a diff but cannot enforce project-wide convention. A script enumerates all files.

**Integration:**

```makefile
review-gates: file-naming

file-naming: ## Gate: filenames follow project convention (snake_case for Go)
	@scripts/check-filenames.sh
```

**Script pattern (check-filenames.sh):**

```bash
#!/bin/bash
set -euo pipefail
errors=0
find . -name '*.go' -not -path './vendor/*' | while read f; do
  base=$(basename "$f")
  case "$base" in
    *_test.go) ;; # _test.go permitted
    *.pb.go) ;; # generated
    *_generated.go) ;; # generated
    *_*)
      echo "FAIL: $f uses snake_case, expected kebab-case for Go files (except test/generated)"
      errors=1 ;;
  esac
done
exit $errors
```

---

## 15. File Size Gate

**Problem:** Monolithic files (1000+ lines) are unreviewable, untestable, and conceal complexity. LLMs do not flag "this file is too long."

**Tools:**
| Tool | Run |
|------|-----|
| `sloc-guard` | `sloc-guard check --diff main` |
| `check-added-large-files` (pre-commit) | Pre-commit `args: ['--maxkb=500']` |
| Archgate `max-file-length` | `archgate check` |
| Custom `wc -l` script + ratchet | `scripts/check-file-sizes.sh` |

**Structured output:**

- sloc-guard: JSON, text per-file with line count
- Custom: `file: N lines (max: M)` text

**Exit code:** 0 = all under limit, 1 = exceeded

**Critical feature — Ratchet mechanism:** Files that already exceed the limit should not grow further. Compare PR file size vs `git show main:file | wc -l`.

**What LLM misses:** LLM reads the whole file but does not measure or enforce line-count policy. A script provides hard enforcement.

**Integration:**

```makefile
review-gates: file-size

file-size: ## Gate: no file exceeds max lines, existing files don't grow
	@scripts/check-file-sizes.sh 500
```

**Script pattern (check-file-sizes.sh) with ratchet:**

```bash
#!/bin/bash
set -euo pipefail
MAX_LINES=${1:-500}
errors=0
git diff --name-only --diff-filter=ACMR origin/main...HEAD | grep '\.go$' | while read f; do
  lines=$(wc -l < "$f")
  # Ratchet: existing files cannot grow
  base=$(git show main:"$f" 2>/dev/null | wc -l || echo 0)
  if [ "$base" -gt 0 ] && [ "$lines" -gt "$base" ]; then
    echo "FAIL: $f grew from ${base} to ${lines} lines (ratchet violation)"
    errors=1
  elif [ "$base" -eq 0 ] && [ "$lines" -gt "$MAX_LINES" ]; then
    echo "FAIL: new file $f has ${lines} lines (max: ${MAX_LINES})"
    errors=1
  fi
done
exit $errors
```

---

## Master `review-gates` Target

Combined `make review-gates` target running all 15 checks, ordered by cost (fastest first):

```makefile
# ── Deterministic Review Gates ──────────────────────────────────────
# Ordered by speed: fast (seconds) → slow (minutes).
# Fail-fast: a cheap gate failing avoids running expensive ones.

REVIEW_GATES ?= \
	check-format \
	file-naming \
	file-size \
	spell-check \
	todos \
	migration-check \
	ci-mod \
	dead-code \
	staticcheck \
	golangci-lint \
	coverage-delta \
	license-check \
	vulncheck \
	secret-scan \
	api-compat \
	proto-drift \
	openapi-drift

review-gates: $(REVIEW_GATES) ## Run all deterministic quality gates
	@echo "✅ All review gates passed"

review-gates-fast: check-format file-naming file-size spell-check todos migration-check ci-mod dead-code ## Fast subset (~30s)
	@echo "✅ Fast review gates passed"

review-gates-slow: coverage-delta vulncheck secret-scan api-compat ## Expensive gates (on-demand)

review-gates-bench: bench-regression binary-size ## Performance/measurement gates (require dedicated runner)
```

---

## Case Studies: LLM Review → Automated Gates Replacement

### Lesson 1: The 45/30/25 Split (Aviator)

Team analyzed 1,000 PR review comments:

- **45%** deterministic (AST-checkable: naming, structure, coverage)
- **30%** execution-testable (run the code and assert)
- **25%** genuine judgment (architecture, tradeoffs)

Once codified, the deterministic bucket never needed a reviewer again.

### Lesson 2: False Positives Erode Trust (Salesforce/MuleSoft)

Every Golden Gate skill must pass a **deterministic validation pipeline**: schema validation → fixture tests → open-source eval → backtesting on real merged PRs. Skills promoted to merge-blocking status only after clearing a high detection-accuracy threshold. Result: developers trust the gates because they're right 100% of the time.

### Lesson 3: Deterministic Gate First, LLM Second (Autotomy)

Order of operations in CI:

1. `fuck-u-code` (AST-based score, 14 languages, < 1s) — catches structural disasters
2. If it passes → send to expensive LLM review for semantics only

Filter deterministic issues for free before paying for probabilistic review.

### Lesson 4: Three Independent Layers (Four Reviewers + Gauntlet)

Stack: AI Reviewer 1 → AI Reviewer 2 (different vendor) → AI Reviewer 3 (local, dual-model) → **CI gates** (hard pass/fail) → **Deterministic SAST** (SonarQube, same result every run).

Defense in depth: a bug survives a single AI miss. It does not survive all 5 layers at once.

### Lesson 5: Line-Level Correctness ≠ Architectural Understanding

A team using Claude Code for all reviews found junior engineers improved at passing automated review but **did not build system intuition**. Their conclusion: LLM as mandatory first pass (strips surface issues), human review remains mandatory for security, data model, compliance, and contract changes.

### Lesson 6: Evidence-Driven Gate Protocol

Five-dimension gate for LLM application releases:
| Dimension | Threshold |
|-----------|-----------|
| Task Success Rate | ≥ 90% |
| Precision | ≥ 85% |
| Recall | ≥ 80% |
| Latency (p95) | ≤ 2s |
| Evidence Coverage | ≥ min samples |

Verdict: PROMOTE / HOLD / ROLLBACK. No human in loop for routine releases.

---

## Which Gates Would an LLM Agent Miss or Misclassify?

| Gate                     | LLM Failure Mode                                       |
| ------------------------ | ------------------------------------------------------ |
| Coverage delta           | Cannot compute byte-exact coverage from diff           |
| Benchmark regression     | Cannot run benchmarks or compute Mann-Whitney p-value  |
| Binary size              | Cannot predict linker output size                      |
| API compat               | Cannot run `go/types` at compiler level                |
| License compliance       | Cannot detect SPDX identifier from LICENSE file text   |
| Vulnerability scanning   | No access to live NVD/CVE database                     |
| Generated code staleness | Cannot re-run generator and diff output                |
| Secret scanning          | Cannot compute Shannon entropy on strings              |
| Misspelling              | Unreliable at catching "teh" vs "the" in prose         |
| Dead code (RTA)          | Cannot trace whole-program reachability graph          |
| SQL migration sequence   | Cannot enumerate all files and verify invariants       |
| Spec drift               | Cannot perform full AST comparison of OpenAPI/Protobuf |
| File naming convention   | Cannot enforce project-wide naming policy              |
| File size                | Does not measure or enforce line-count policy          |
| TODOs/FIXMEs             | Has no policy on whether TODOs are allowed             |

**Rule of thumb:** If the check requires computation (p-values, AST analysis, reachability, byte-level comparison, file enumeration, live database access), an LLM cannot do it deterministically. Build a gate.
