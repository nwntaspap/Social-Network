# Build System Visibility Rules for Architecture Enforcement

**Research for D6 Dependency DAG Enforcement** — June 2026

---

## 1. Core Concept: Visibility as Architecture Enforcement

All major monorepo build systems enforce a common pattern: **every target explicitly declares its dependencies**, and the build system rejects undeclared or disallowed edges at build time. This turns architectural rules from convention into hard failures.

### Why it works

- Boundary violations == **build break** (not just lint warning)
- Enforcement happens **before code is merged** (CI gate)
- Visibility is **declarative**, co-located with code in BUILD files
- Default deny: private by default, explicitly grant access

---

## 2. Bazel (`blaze`)

### 2.1 Target Visibility

**Mechanism:** `visibility` attribute on every rule target. Governs who may depend on the target.

| Specification               | Meaning                                  |
| --------------------------- | ---------------------------------------- |
| `//visibility:public`       | Anyone can depend                        |
| `//visibility:private`      | Only same package (default)              |
| `//foo/bar:__pkg__`         | Grant to `//foo/bar` only                |
| `//foo/bar:__subpackages__` | Grant to `//foo/bar` and all subpackages |
| `//foo/bar:my_group`        | Grant to `package_group` targets         |

**Key behaviors:**

- Default visibility is `private` unless `default_visibility` is set in `package()`.
- **Transitive visibility (Bazel 8.x+):** Propagates restrictions through the dep chain — if `A` depends on `B` and `B` has transitive visibility rules, `A` inherits them.
- **Load visibility (Bazel 6.0+):** Controls which packages can `load()` a `.bzl` file. Default: public. Call `visibility("private")` in `.bzl` files to restrict.
- Violations fail at **analysis phase** (before build executes).

```python
# //myapp/internal:impl BUILD
cc_library(
    name = "impl",
    visibility = ["//visibility:private"],  # only same package
)

# //myapp:api BUILD
cc_library(
    name = "api",
    visibility = ["//myapp/...", "//other_team:__subpackages__"],
)
```

### 2.2 Package Groups

Named groups of packages for reusable visibility:

```python
package_group(
    name = "bounded_context_clients",
    packages = [
        "//orders/...",
        "//billing/...",
    ],
)
```

### 2.3 `tags` and `target_compatible_with`

Not primary enforcement — tags are metadata for querying/filtering. `target_compatible_with` constrains platform compatibility, not dependency boundaries.

### 2.4 Violation Behavior

```
Error: target '//myapp:binary' depends on '//internal:helper'
  which is not visible. Use 'bazel build --check_visibility=false'
  to temporarily disable.
```

Build **fails entirely** — hard error, not warning.

### 2.5 `buildozer` — Auditing & Bulk Editing

```
# Print all targets with public visibility
buildozer 'print visibility' '//...:%cc_library'

# Set visibility on all targets in a package
buildozer 'set visibility //visibility:private' '//myapp/...:all'

# Add a dep to all matching targets
buildozer 'add deps //shared:lib' '//myapp/...:%cc_library'
```

Also supports piping from `bazel query` for complex transformations.

### 2.6 `gazelle` — Auto-generating BUILD files

- Parses Go/Python/Proto source, generates correct BUILD files
- Can infer `deps` from imports
- Does NOT set visibility by default — generates `default_visibility = ["//visibility:public"]`
- Teams must layer their own visibility rules after gazelle generation
- Pattern: run gazelle, then run a script/buildozer to lock down visibility

### 2.7 `rules_go` — Go-specific Visibility

- Go binaries typically get `visibility = ["//visibility:public"]`
- Go libraries default to `private` when explicit
- Gazelle generates `go_library` with `visibility = ["//visibility:private"]` and `go_binary` public
- Go `internal` packages at the source level are redundant with Bazel visibility (but still recommended for tooling compatibility)

### 2.8 Salesforce `bazel-visibility-tool`

Open-source tool from Salesforce for managing visibility at scale:

- Defines **Visibility Groups** (layers/components)
- Each group defines which other groups it is visible to
- Starlark-based group definitions with membership lookup at analysis time
- Enforces that every target belongs to exactly one group
- Architecture is declared once, not scattered across BUILD files

---

## 3. Buck2

### 3.1 Visibility + `within_view`

Buck2 has **two-directional** visibility:

- `visibility` — who can depend on **me** (allowlist of target patterns)
- `within_view` — who **I** can depend on (restricts my deps)

When both conflict, `within_view` takes precedence.

```python
# Buck2
java_library(
    name = 'example',
    visibility = ['PUBLIC'],
    within_view = ['//foo:bar', '//hello:world'],
)
```

**Key difference from Bazel:** Buck2 has `within_view` as an explicit constraint on outgoing edges. Bazel only constrains incoming edges.

### 3.2 PACKAGE Files for Hierarchical Visibility

`PACKAGE` files apply visibility rules to **all targets in a directory**:

```python
# PACKAGE file
package(
    visibility = ["//allowed/..."],
    within_view = ["//permitted/..."],
)
```

These inherit through `inherit = True` to subpackages.

### 3.3 Audit Command

```
buck2 audit visibility <target>
```

Shows effective visibility and why a dependency is allowed/blocked.

### 3.4 Violation Behavior

Build fails with clear error showing which visibility rule was violated.

---

## 4. Please Build (`plz`)

### 4.1 Visibility Model

Please uses a `visibility` attribute on build rules, similar to Bazel but simpler:

```python
go_library(
    name = "mylib",
    visibility = ["//myapp/..."],
)
```

**Key characteristics:**

- `visibility` takes a list of build label patterns
- No `__pkg__` / `__subpackages__` syntax — uses `//path/...` semantics directly
- Default is **visible to all** (different from Bazel!)
- No `within_view` concept (Buck2 feature)
- No `package_group` abstraction

### 4.2 Testing

```
plz test //myapp/...
```

Dependency validation happens during target graph construction.

### 4.3 Strength

Please's visibility is **adequate** but less sophisticated than Bazel/Buck2. Better suited for mid-sized repos.

---

## 5. Pants

### 5.1 Dependency Rules API

Pants 2.16+ has a rich visibility system via `__dependencies_rules__` and `__dependents_rules__`:

```python
# BUILD file
__dependencies_rules__(
    (
        {"type": python_sources},
        "src/a/**",
        "src/b/lib.py",
        "!*",  # deny all else
    ),
    ("*", "*"),  # default: allow everything
)

__dependents_rules__(
    (
        ({"type": "python_*", "tags": ["any-python"]},),
        ("!tests/**", "!src/*/*/**"),
    ),
    ("*", "*"),
)
```

**Key features:**

- Rules can match on: `type`, `tags`, `path` (glob)
- Both incoming and outgoing dependency rules
- Rules propagate to subdirectories unless overridden
- Default: all dependencies allowed
- No matching rule = error

### 5.2 Enforcement

```
pants lint ::
```

Enforced during dependency calculation. Configurable with `[visibility].enforce`.

### 5.3 Expressiveness

Pants rules are the most expressive (regex/glob/type/tag matching) but also the most complex to maintain.

---

## 6. DDD Bounded Contexts → Build System Visibility

### 6.1 Canonical Mapping Pattern

| DDD Concept           | Build System Mapping                                  |
| --------------------- | ----------------------------------------------------- |
| Bounded Context       | Top-level directory (e.g., `//orders/`)               |
| Aggregate Root        | Public target with controlled visibility              |
| Domain/Internal       | Private targets, `__subpackages__` visibility         |
| Shared Kernel         | Package group for shared types                        |
| Anti-Corruption Layer | Dedicated targets with specific visibility grants     |
| Context Map           | Package groups defining allowed dependency directions |

### 6.2 Example: E-Commerce Monorepo

```
//orders/
  BUILD          # package(default_visibility = ["//orders/..."])
  api/           # public interfaces
  internal/      # private implementation
  tests/
//billing/
  BUILD          # package(default_visibility = ["//billing/..."])
  api/
  internal/
//shared/
  BUILD          # package_group for shared-kernel clients
```

### 6.3 Google's Approach

From Bazel docs and dependency management guide:

- **Most targets stay private** — only 1 public target per BUILD file typically
- **API targets** get explicit allowlist visibility to known consumers
- **Team boundaries** mapped to package groups
- **Code review gates** enforce that visibility changes go through OWNERS approval
- **TAP** (Test Automation Platform) runs affected tests; a visibility change triggers all downstream consumers' tests

### 6.4 Go `internal` Pattern

Go's `internal/` packages provide language-level boundary enforcement — complementary, not competitive:

- `internal/` packages can only be imported by code rooted in the parent of `internal/`
- Teams often use **both** Go `internal` AND Bazel visibility (defense in depth)
- Bazel visibility catches cross-repo / cross-target violations earlier

---

## 7. Real-World: Google's Piper + TAP

### 7.1 Scale

- 2+ billion lines of code, single monorepo (Piper)
- ~9 million files
- 25,000+ engineers committing daily

### 7.2 How Visibility Works at Google

- **Target-level visibility** is mandatory — no target can be public without explicit approval
- Teams OWN their packages via OWNERS files
- Visibility changes require **review by the owning team**
- TAP automatically selects and runs tests affected by a change
- A visibility reduction immediately prevents new dependencies; existing dependents are grandfathered (or break)
- Tools like `buildozer` enable bulk visibility audits

### 7.3 Key Insight

Google's architecture enforcement works because:

1. The **build system enforces it** (not convention)
2. Changes to visibility require **human review** (OWNERS)
3. The **CI system catches violations** before merge
4. **Granular targets** make boundaries meaningful

---

## 8. Can This Be Replicated Without Bazel?

### 8.1 Yes — With Tradeoffs

#### Option A: Script-Based Dependency Checking

```python
# ci/check-boundaries.py
# Scans import statements, enforces allowed edges
ALLOWED = {
    "orders": {"shared", "orders"},
    "billing": {"shared", "billing"},
    "shared": {"shared"},
}
```

**Pros:** Simple, language-specific, no build system migration
**Cons:** Doesn't understand transitive deps, needs maintenance, false negatives

#### Option B: `dependency-cruiser` (JS/TS)

```javascript
// .dependency-cruiser.js
module.exports = {
  forbidden: [
    {
      name: 'orders-not-to-billing',
      from: { path: '^src/orders' },
      to: { path: '^src/billing' },
      severity: 'error',
    },
  ],
};
```

**Pros:** Graph-aware, supports patterns/glob, CI-ready, generates visual graphs
**Cons:** JS/TS only (by default), per-file not per-target granularity

#### Option C: ESLint Plugin (Nx-style)

```
@nx/enforce-module-boundaries
```

**Pros:** Editor integration (red squiggles), CI ready, tag-based rules
**Cons:** ESLint ecosystem only, not build-system aware

#### Option D: Custom Makefile Targets

```makefile
check-boundaries:
	@for dep in $(shell grep -r "^import.*from '\.\." src/); do \
		validate_dep $$dep; \
	done
```

**Pros:** Zero dependencies
**Cons:** Fragile, doesn't scale, no graph awareness

#### Option E: TypeScript Project References

```json
{
  "references": [{ "path": "../shared" }, { "path": "../orders" }]
}
```

TypeScript compiler itself enforces that you only import from referenced projects. Combined with `composite: true`, this provides build-level boundary enforcement with zero extra tooling.

**Pros:** No extra tools, compiler-enforced, vs-code aware
**Cons:** TS-only, limited expressiveness

### 8.2 Comparison Matrix

| Approach             | Enforcement Level | Config Complexity | Scale         | Language Support  |
| -------------------- | ----------------- | ----------------- | ------------- | ----------------- |
| Bazel visibility     | Build break       | High              | 100K+ targets | All (via rules)   |
| Buck2 visibility     | Build break       | High              | 100K+ targets | All (via rules)   |
| Please visibility    | Build break       | Medium            | 10K+ targets  | All (via plugins) |
| Pants dep rules      | Lint/build        | Medium            | 10K+ targets  | Python/Go/Java/TS |
| dependency-cruiser   | CI gate           | Low               | 1K+ modules   | JS/TS             |
| ESLint nx-boundaries | Lint + IDE        | Low               | 1K+ modules   | JS/TS             |
| Go internal pkgs     | Compile error     | Zero              | N/A           | Go only           |
| TS project refs      | Compile error     | Low               | N/A           | TS only           |
| Custom script        | CI gate           | Low               | Small         | Any               |

### 8.3 Recommendation for D6

**For a Rust codebase** (no built-in package visibility like Go `internal`):

1. **If already using Cargo workspace:** Add a boundary-checking script in CI that parses `use` statements and enforces allowed dependencies between workspace members. ~100 lines of Rust or Python.

2. **If considering Bazel migration:** Bazel visibility is the gold standard but has significant setup cost (BUILD files for every target, toolchain config, migration effort).

3. **Hybrid approach:** Use `cargo-deny` for license/vulnerability checks AND a custom boundary checker for DDD constraints.

4. **If using Nx/Cargo workspace:** Nx's `enforce-module-boundaries` ESLint rule is the lightest viable option for polyglot repos.

**For D6 specifically:** A **simple script-based approach** in CI is sufficient for enforcing DDD bounded context boundaries. Only migrate to Bazel/Buck2 when the repo grows beyond ~50 crates or when mult-language support is needed.

---

## 9. Summary

### How visibility rules become architecture enforcement

- Targets declare who can depend on them (incoming) and/or who they can depend on (outgoing)
- Violations are **hard build errors** at analysis time
- Default-deny forces explicit opt-in for cross-boundary dependencies

### Patterns for bounded context / vertical slice visibility

- One directory = one bounded context
- Internal implementation: private visibility
- Public API surface: explicitly granted to consumers
- Package groups as context maps

### CI integration

- Pre-merge checks run `bazel build //...` (or equivalent)
- Visibility violation == build failure == blocked merge
- Google-style: visibility changes need OWNERS approval

### What happens on violation

- **Bazel:** `Error: target is not visible from...` — build aborts during analysis
- **Buck2:** Build fails with visibility/within_view conflict message
- **Pants:** `DependencyRuleActionDeniedError` with from→to: DENY
- **dependency-cruiser:** Exits non-zero with eslint-style report

### Lighter alternatives

- `dependency-cruiser` (JS/TS) — graph-aware, CI-ready, very expressive rules
- `cargo-deny` + custom script (Rust) — lightest weight for single-language
- TypeScript project references — compiler-enforced for TS-only
- Nx module boundaries — tags + ESLint + IDE integration
- Go `internal/` — language-level, zero-config, for Go only
