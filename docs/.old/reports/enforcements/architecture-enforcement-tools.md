# Architecture Enforcement Tools — Research

Goal: find patterns replicable in Go monorepo for vertical slice architecture boundaries.

---

## 1. ArchUnit (Java) — Gold Standard

**Ecosystem:** JVM (Java, Kotlin, Scala via bytecode)
**Maintainer:** TNG Technology Consulting
**Status:** **Very active** — v1.4.2 (Mar 2025), ~4K GitHub stars
**Repo:** https://github.com/TNG/ArchUnit

### How it works

Analyzes compiled bytecode (not source) via `ClassFileImporter`, builds in-memory graph of packages/classes/annotations/dependencies. Rules expressed as fluent Java DSL in plain JUnit tests.

### Rule types

| Category                | Examples                                                                                                     |
| ----------------------- | ------------------------------------------------------------------------------------------------------------ |
| **Layer**               | `layeredArchitecture().layer("UI")...whereLayer("Persistence").mayOnlyBeAccessedByLayers("Business Logic")`  |
| **Dependency**          | `noClasses().that().resideInAPackage("..service..").should().accessClassesThat().resideInAPackage("..ui..")` |
| **Naming**              | `classes().that().haveSimpleNameEndingWith("Controller").should().resideInAPackage("..controller..")`        |
| **Containment**         | `classes().that().areAnnotatedWith(@Service).should().resideInAPackage("..service..")`                       |
| **Cycle**               | `slices().matching("..(*)..").should().beFreeOfCycles()`                                                     |
| **Metrics**             | Lakos cumulative dependency, component metrics                                                               |
| **Architecture styles** | `layeredArchitecture()`, `onionArchitecture()`, custom module rules                                          |
| **PlantUML**            | Validate code against architecture diagrams                                                                  |

### CI/CD integration

Runs as standard JUnit tests. `FreezingArchRule` records known violations to a `ViolationStore` — only new violations fail the build. Gradual adoption in legacy codebases.

### Output

JUnit-native assertion errors with class + line numbers. `archunit_ignore_patterns.txt` for suppression.

### Key innovations

- **Empty-test protection**: fails if rule matches 0 files (typo prevention)
- **Freeze violations**: "get out of jail free card" — known violations ignored until touched code changes
- **Bytecode analysis**: works even without source, catches generated/compiled code
- **Modular architecture**: Core (bytecode import) / Lang (fluent DSL) / Library (prebuilt patterns)

### Verdict for vertical slices

★★★★★. Directly supports: layer isolation, dependency direction, cycle-free slices, hexagonal/onion architectures. The gold standard.

---

## 2. NetArchTest (.NET/C#)

**Ecosystem:** .NET Standard 2.0 (.NET Framework 4.6.1+, .NET Core 2.0+)
**Maintainer:** Ben Morris (original) → community fork `NetArchTest.eNhancedEdition`
**Status:** **Original stalled** (last release 2023), fork active
**Repo:** https://github.com/BenMorris/NetArchTest

### How it works

Uses Mono.Cecil (IL reader) to load compiled assemblies. Fluent API mirrors ArchUnit. Types loaded via `Types.InAssembly(typeof(X).Assembly)`.

### Rule types

| Category        | Examples                                                                |
| --------------- | ----------------------------------------------------------------------- |
| **Dependency**  | `HaveDependencyOn("Namespace")`, `ShouldNot().HaveDependencyOnAny(...)` |
| **Layer**       | Custom via `Policy` groups + namespace predicates                       |
| **Naming**      | `HaveNameEndingWith("Service")`, `ResideInNamespace(...)`               |
| **Inheritance** | `BeSealed()`, `ImplementInterface(...)`                                 |
| **Custom**      | `ICustomRule` interface + `MeetsCustomRule()`                           |

### CI/CD

Runs as xUnit/NUnit tests. `dotnet test` — trivial integration.

### Output

`PolicyResults` with `FailingTypes`. Standard test framework assertions.

### Fork (NetArchTest.eNhancedEdition)

- Fixes bugs, adds new rules, active as of 2025-2026
- API-compatible with original

### Verdict

★★★★. Solid ArchUnit port. Vertical slice enforcement via namespace predicates + Policy grouping.

---

## 3. Python Tools

Python has the richest ecosystem of ArchUnit alternatives (6+ tools). Summary:

### 3a. archetype-py (Most Complete)

**Status:** **Active** — v0.3.0 (May 2026), brand new
**Repo:** https://github.com/MossabArektout/archetype-py

- Forbidden/allowlisted imports, layer ordering, cycle detection, protected boundaries
- CLI: `archetype check .` + pytest plugin
- Output: JSON, text, GitHub Actions annotations
- **Baseline mode** (`--write-baseline`) for legacy adoption
- Changed-file enforcement (`--changed-from origin/main`)
- Pre-commit hook installer (`archetype install-hook`)
- Requires Python 3.11+

### 3b. ArchUnitPython (Most ArchUnit-like)

**Status:** **Active** — v1.0.1 (Apr 2026)
**Repo:** https://github.com/LukasNiessen/ArchUnitPython

- Port of ArchUnitTS, fluent API, zero dependencies
- pytest integration with `assert_passes()`
- Layer, dependency, naming, cycle rules
- Requires Python 3.10+

### 3c. pytest-archon (Lightweight)

**Status:** Stale (last v0.0.7 Sep 2025, prior 2023)

- Import test rules, simple API, Python 3.8+

### 3d. pyarchrules (Beta)

**Status:** Pre-release (v0.1.0b2, 2026)

- TOML-config focused, service isolation for monorepos
- Folder structure + directional import rules

### 3e. deply (YAML-defined)

**Status:** v0.8.2, active

- YAML config, layer-based analysis, Mermaid diagram output
- Suppress violations via `# deply:ignore`

### 3f. archy (Score/metrics)

**Status:** v0.18.0, active

- Treesitter-based, trended architecture score
- MCP server for AI agents (!)

### Verdict for vertical slices

Use **archetype-py** or **ArchUnitPython**. Both support layer isolation + dependency direction natively. archetype-py has better CI/legacy features.

---

## 4. TypeScript/JavaScript Tools

### 4a. ArchUnitTS (Most Complete)

**Status:** **Very active** — v2.3.0 (May 2026), ~200+ GitHub stars
**Repo:** https://github.com/LukasNiessen/ArchUnitTS
**npm:** `archunit`

- Fluent API: `projectFiles().inFolder('src/domain').shouldNot().dependOnFiles().inFolder('src/infrastructure')`
- Empty-test protection, cycle detection, layer rules
- Code metrics: LCOM cohesion, coupling, distance from main sequence
- **Nx monorepo support** — reads Nx project graph
- PlantUML diagram validation
- Jest/Vitest/Jasmine/Mocha — `toPassAsync()` matcher
- Custom rules, HTML reports, debug logging

### 4b. ts-arch

**Status:** Active
**Repo:** https://ts-arch.github.io/ts-arch/

- Files + slices API
- PlantUML adherence
- Nx project graph integration
- Jest-focused

### 4c. ts-archunit (Body-level analysis)

**Status:** Active (2026)
**Repo:** https://github.com/nielspeter/ts-archunit

- Goes beyond imports — analyzes function bodies, call patterns, types
- Baseline/gradual adoption, GitHub PR annotations
- Catches "service calls parseInt instead of extractCount()"

### 4d. dependency-cruiser

**Status:** **Very active** — v17.4.3 (2026), ~7K stars
**npm:** `dependency-cruiser`
**Repo:** https://github.com/sverweij/dependency-cruiser

- Rules-based dependency validation for JS/TS/CoffeeScript
- Visual dot/HTML output
- Rule config in `.dependency-cruiser.js`
- Circular, orphan, missing-dep detection
- Not strictly an "ArchUnit" — more of a dependency linter + visualizer

### Verdict for vertical slices

Use **ArchUnitTS** (richest feature set) or **dependency-cruiser** (mature, visual). arch-unit-ts also solid.

---

## 5. Rust Tools

### 5a. Rust Arkitect

**Status:** Active
**Repo:** https://github.com/pfazzi/rust_arkitect

- Fluent DSL: `rules_for_module("app::domain").it_must_not_depend_on_anything()`
- Baseline tracking for legacy code
- `cargo test` integration

### 5b. Archaven

**Status:** Active (v0.x, 2026)
**Repo:** https://github.com/mtanasiewicz/rust-archaven
**crates.io:** `archaven`

- Source-file scanner + module path patterns
- `Rule::between()`, `Rule::directories()`, custom `RuleSet`
- Layered, hexagonal, vertical-slice — pattern-agnostic
- Fast, test-integrated

### 5c. arch_validation_core (ArchTest)

**Status:** Active (v0.2.3)
**crates.io:** `arch_validation_core`

- Full rule set: `MayNotAccess`, `MayOnlyAccess`, `MayNotBeAccessedBy`, `MayOnlyBeAccessedBy`
- External crate whitelist/blacklist (`Available`/`Restricted`)
- Rule scoping (global, parent, subdomain)
- Conflict detection between rules
- CLI + JSON config

### 5d. cargo_pup

**Status:** Active (v0.1.0)

- `cargo pup` as `cargo` extension
- Config in `pup.ron`, compiles with Rust nightly
- Uses `rustc` interface for deep analysis

### 5e. layered-crate

**Status:** Active (v0.4.4)

- `Layerfile.toml` per crate to declare internal module dependencies
- Splits crate into virtual sub-crates for checking
- Practical for large single-crate projects

### 5f. arch-lint (tree-sitter)

**Status:** Active

- Dual engine: syn (Rust) + tree-sitter (Kotlin, cross-language)
- 21 built-in rules + custom
- `cargo test` integration via `check!()` macro

### 5g. intent (with TLA+)

**Status:** Active (v0.1.4)

- Custom language + CLI
- Structural constraints + behavioral specs → TLA+ verification

### Verdict for vertical slices

Use **Archaven** (simple, flexible) or **Rust Arkitect** (closest ArchUnit DSL). For strict enforcement: **arch_validation_core**.

---

## 6. DepCruft / dependency-cruiser

**Note:** The term "DepCruft" may refer to **dependency-cruiser** (JS/TS) or **dep-cruft** (older Python tool).

### dependency-cruiser (JS/TS) — v17.4.3

- **Stars:** ~7K
- **Status:** Very active
- CLI tool, config in `.dependency-cruiser.js`
- Rule-based validation of import/require graphs
- Output: text, DOT (GraphViz), HTML, CSV, JSON
- CI: `npx depcruise src`
- Rules: circular deps, orphans, forbidden modules, allowed modules, boundaries

### For Go monorepo

No direct equivalent. Pattern: scan import graph, apply allow/deny rules, fail on violation. Config-as-code approach is replicable.

---

## 7. JDepend (Java)

**Status:** **Minimal maintenance** — last release v2.10 (2020), last commit 2020
**Stars:** ~700
**Repo:** https://github.com/clarkware/jdepend

### What it does

Package-level dependency metrics (not class-level like ArchUnit):

- Afferent coupling (Ca), Efferent coupling (Ce)
- Abstractness (A), Instability (I)
- Distance from main sequence (D)
- Package dependency cycles

### Output

GUI, text, XML. ANT/Maven integration via plugins.

### Verdict

Historical significance. Superseded by ArchUnit. The _metrics_ (Ca, Ce, A, I, D) are useful concepts for any architecture enforcement tool.

---

## 8. Structure101

**Status:** Acquired by **Sonar** (2024). Being folded into SonarQube.
**Website:** https://structure101.com → now https://www.sonarsource.com/structure101/

### How it works

Desktop app (Studio) + Build (CI) + Workspace (IDE plugin). Visual architecture diagrams with hierarchical cells. Define layering, visibility (public/private), and dependency rules visually.

### Rule types

- **Layering**: cells arranged top-down, strict/relaxed
- **Visibility**: public/private cells (private = only siblings + parent can access)
- **Dependency overrides**: allow specific upward deps, block specific downward deps
- **Structure Spec**: physical container grouping + rules

### CI/CD

Structure101 Build CLI: `check-architecture` — CSV/XML output, fail build on violations, baseline support (`onlyNew`).

### Output

CSV, XML, diagram images. SonarQube integration incoming.

### Verdict

★★★★★ for visualization + enforcement, but **heavy** (licensed, desktop app). Sonar acquisition means architecture-as-code in SonarQube. Pattern: visual hierarchy → code rules → CI enforcement.

---

## 9. JQAssistant

**Status:** **Active** — v2.9.0 (Jan 2026), ~280 stars
**Repo:** https://github.com/jQAssistant/jqassistant

### How it works

Scans Java bytecode (and other artifacts via plugins) into a **Neo4j graph database**. Rules expressed in **Cypher** query language. Embedded Neo4j — no server setup.

### Key features

- Scan: bytecode, XML, properties, Git history, Maven deps, JaCoCo coverage
- **Concepts**: label nodes with domain meaning (e.g., `:Service`, `:Entity`)
- **Constraints**: Cypher queries that detect violations
- Severity levels: blocker/critical/major/minor/info — configurable fail-on threshold
- Plugins: Maven, CLI, Structurizr, JaCoCo, Git

### Example Cypher constraint

```cypher
MATCH (s:Service)
WHERE NOT EXISTS {
  MATCH (s)-[:ANNOTATED_BY]->(ann:Type)
  WHERE ann.fqn = 'jakarta.enterprise.context.ApplicationScoped'
}
RETURN s.fqn
```

### CI/CD

Maven plugin (`jqassistant:scan`, `jqassistant:analyze`). `failOnSeverity` config.

### Output

HTML report (searchable), Neo4j browser, XML, console. Can integrate with AsciiDoc/arc42 for living documentation.

### Verdict

★★★★. Extremely flexible (graph query is Turing-complete). Heavy dependency (Neo4j embedded). Unique value: cross-artifact analysis (bytecode + git + coverage in one graph). Pattern: graph database → arbitrary queries → architectural constraints.

---

## 10. Hooks (Git hooks + pre-commit)

### 10a. pre-commit (framework)

**Status:** **Very active**
**Website:** https://pre-commit.com/

Multi-language hook manager. YAML config, community hook repos. Can run any executable on staged files.

### 10b. prehook

**Status:** Active (2026)

- Single binary, pre-commit + pre-push guards
- Delegate to gitleaks/trufflehog/semgrep/osv-scanner/trivy
- YAML config, `prehook doctor` for setup validation

### 10c. intent-audit-harness

**Status:** Active (2026)

- Hash-pins testing config; AI-proof policy enforcement
- Architecture checker: delegates to dependency-cruiser / import-linter / ArchUnit / deptrac / arch-go
- CRAP scorer, escape-scan detects AI attempts to weaken thresholds

### 10d. trustlock

**Status:** Active (2026)

- Dependency admission controller (Git hook + CI)
- Evaluates cooldown, provenance, pinning, install scripts per new dependency

### Verdict

Hooks are the **enforcement layer**, not the rule engine. Pattern: pre-commit runs lightweight checks, CI runs full rule set. Pre-commit + pre-push + CI = multi-gate enforcement.

---

## Cross-Cutting Patterns for Go Monorepo

### Common Architecture

```
Rules (DSL/YAML/Code)
  → Scanner (AST/bytecode/import-graph)
    → Graph (in-memory / DB)
      → Evaluator (pattern matching)
        → Reporter (text/JSON/sarif)
```

### Rule Categories All Tools Support

1. **Layer/Dependency direction** — A must not import B
2. **Naming conventions** — files/classes ending with "Service"
3. **Containment** — types with @Annotation must be in certain package
4. **Cycle freedom** — no circular dependencies between slices
5. **Visibility** — private internals not accessible from outside
6. **Isolation** — modules don't import each other's internals

### Successful Patterns

| Pattern                  | Tools                                               |
| ------------------------ | --------------------------------------------------- |
| **Tests as enforcement** | ArchUnit, NetArchTest, Rust Arkitect, go-arch-guard |
| **Config as rules**      | deply, arch-go, goverhaul, go-arch-lint             |
| **Visual hierarchy**     | Structure101, JDepend GUI                           |
| **Graph queries**        | JQAssistant (Neo4j + Cypher)                        |
| **Build proxy**          | goarch (replaces `go build`)                        |

### Key Trends (2025-2026)

1. **AI agent guardrails** — ts-archunit, archy MCP server, intent-audit-harness. Architecture tests as AI containment.
2. **Baseline adoption** — FreezingArchRule, archetype-py `--write-baseline`, intent-audit-harness hash-pinning. Critical for legacy codebases.
3. **Visualization** — dependency-cruiser DOT output, deply Mermaid diagrams, Structure101 Structure Map, dep-insight D3.js.
4. **Monorepo awareness** — ArchUnitTS Nx support, ts-arch Nx slices, layered-crate for Rust internal deps, archetype-py service isolation.

---

## Go-Specific Tool Landscape

### Built in this project

- `/cmd/` — entry points
- `/internal/` — application code
- `/frontend/` — TypeScript UI

### Go-native tools available

| Tool                         | Approach                    | Layer/Dependency | Naming             | Cycle | CI                  |
| ---------------------------- | --------------------------- | ---------------- | ------------------ | ----- | ------------------- |
| **arch-go**                  | Config + test               | ✅               | ✅                 | ✅    | `go test`           |
| **go-arch-lint** (fe3dback)  | YAML config                 | ✅               | ❌ component-based | ✅    | CLI                 |
| **go-arch-lint** (coderhyme) | YAML groups                 | ✅               | ❌                 | ✅    | CLI                 |
| **goverhaul**                | YAML config + caching       | ✅               | ❌                 | ❌    | CLI                 |
| **goarch** (ksanderer)       | Build proxy                 | ✅               | ✅ (limited)       | ❌    | Replaces `go build` |
| **go-arch-guard**            | Presets + tests             | ✅               | ✅                 | ✅    | `go test`           |
| **goarctest**                | Fluent API tests            | ✅               | ✅                 | ❌    | `go test`           |
| **GoArchTest** (solrac97gr)  | Fluent API tests            | ✅               | ❌                 | ❌    | `go test`           |
| **archunit** (kcmvp)         | Fluent API + prebuilt rules | ✅               | ✅                 | ❌    | `go test`           |
| **cht-go-lint**              | 41 rules, presets           | ✅               | ✅                 | ✅    | `go test` / CLI     |

### Top recommendations for vertical slices in Go monorepo

1. **cht-go-lint** (channel-io) — most comprehensive (41 rules), clean-arch preset, layer-aware rules, JSON/GitHub annotations output, `go test` integration
2. **go-arch-guard** (NamhaeSusan) — preset-based (DDD, Clean Architecture, Hexagonal), test-native, isolation + layer direction + naming + blast radius
3. **arch-go** — config + `go test`, dependency/naming/contents rules, simple YAML, battle-tested
4. **go-arch-lint** (fe3dback) — YAML-defined layers, component dependency graphs, visual output, CI-ready
5. **goverhaul** — blazing fast (MUS binary cache), YAML rules, visual dependency graphs

### For multi-language (Go + TS): combine

- **Go side**: arch-go or cht-go-lint in `go test`
- **TS side**: ArchUnitTS in Vitest/Jest
- **Shared enforcement**: pre-commit hooks + CI parallel steps
- **Cross-language**: intent (syn for Rust, regex for TS — but not Go yet)

### Strategy for this repos

1. Config-as-code layer definitions (`internal/`, `cmd/`, `frontend/`)
2. Go test suite with arch-go or cht-go-lint for `internal/` boundaries
3. ArchUnitTS vitest suite for `frontend/` boundaries
4. `make arch-check` target in CI
5. Pre-commit hook for quick local feedback
6. Gradual adoption via baseline mode (suppress known violations, fail only new)
