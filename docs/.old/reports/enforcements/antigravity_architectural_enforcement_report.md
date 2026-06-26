# The Antigravity Architectural Enforcement Report: Deterministic Gating, Boundaries, and Governance in Modern CI/CD

This report presents a comprehensive synthesis of the tools, custom scripts, and architectural methodologies utilized by senior software architects to enforce codebase boundaries, code quality, and Architectural Decision Record (ADR) compliance.

The findings are structured into nine distinct architectural domains, detailing exact configurations, code snippets, and automated integration patterns.

---

## Table of Contents

1. [Codebase Boundary Enforcement (JVM, .NET, Rust, Go, TS)](#1-codebase-boundary-enforcement-jvm-net-rust-go-ts)
2. [Monorepo & Micro-Frontend Gating (Nx, Turborepo, Bazel, MFE)](#2-monorepo--micro-frontend-gating-nx-turborepo-bazel-mfe)
3. [Architecture Decision Records (ADRs) Automation](#3-architecture-decision-records-adrs-automation)
4. [AST-Based Rules & Custom Scripting](#4-ast-based-rules--custom-scripting)
5. [Database Boundaries & Query Enforcement](#5-database-boundaries--query-enforcement)
6. [API and Schema Drift Detection](#6-api-and-schema-drift-detection)
7. [Git Gating & Policy-as-Code (Lefthook, pre-commit, OPA)](#7-git-gating--policy-as-code-lefthook-pre-commit-opa)
8. [Code Quality Gates & Cognitive Complexity](#8-code-quality-gates--cognitive-complexity)
9. [Synthesis: The Unified Shift-Left Pipeline Blueprint](#9-synthesis-the-unified-shift-left-pipeline-blueprint)

---

## 1. Codebase Boundary Enforcement (JVM, .NET, Rust, Go, TS)

Architects use compile-time checks or test-driven static analysis to enforce dependency paths (e.g., preventing the core domain layer from importing infrastructure or delivery layers).

### A. JVM (Java & Kotlin)

On the JVM, **ArchUnit** and **Spring Modulith** are the standards.

- **ArchUnit:** Runs as a standard unit test. If a developer introduces a forbidden import, the build fails.

  ```java
  import com.tngtech.archunit.core.domain.JavaClasses;
  import com.tngtech.archunit.core.importer.ClassFileImporter;
  import com.tngtech.archunit.core.importer.ImportOption;
  import org.junit.jupiter.api.Test;
  import static com.tngtech.archunit.library.Architectures.onionArchitecture;

  class ArchitectureTest {
      private final JavaClasses classes = new ClassFileImporter()
          .withImportOption(ImportOption.Predefined.DO_NOT_INCLUDE_TESTS)
          .importPackages("com.example.myapp");

      @Test
      void onionArchitectureShouldBeRespected() {
          onionArchitecture()
              .domainModels("com.example.myapp.domain.model..")
              .domainServices("com.example.myapp.domain.service..")
              .applicationServices("com.example.myapp.application..")
              .adapter("persistence", "com.example.myapp.infrastructure.persistence..")
              .adapter("web", "com.example.myapp.infrastructure.web..")
              .check(classes);
      }
  }
  ```

- **Spring Modulith:** Verifies package-based modular monolith encapsulation (e.g., preventing packages outside a module from importing package segments designated as `internal`).

  ```java
  import org.junit.jupiter.api.Test;
  import org.springframework.modulith.core.ApplicationModules;

  class ModulithTest {
      static ApplicationModules modules = ApplicationModules.of(YourApplication.class);

      @Test
      void verifyModules() {
          modules.verify(); // Detects cycle and encapsulation violations
      }
  }
  ```

### B. .NET (C#)

**NetArchTest** allows C# architects to verify dependency constraints within standard xUnit/NUnit test setups.

```csharp
using Xunit;
using NetArchTest.Rules;

public class BoundaryTests {
    [Fact]
    public void Domain_ShouldNotDependOnInfrastructureOrApi() {
        var domainAssembly = typeof(Domain.Entities.User).Assembly;
        var forbidden = new[] { "MyApp.Infrastructure", "MyApp.Api" };

        var result = Types.InAssembly(domainAssembly)
            .ShouldNot()
            .HaveDependencyOnAny(forbidden)
            .GetResult();

        Assert.True(result.IsSuccessful);
    }
}
```

### C. Rust

Rust guarantees boundaries natively through **cargo workspaces** (cyclic crate dependencies are physically blocked by the compiler) and module visibility (`pub(crate)`). For internal module trees within a single crate, architects use:

- **`boundary`:** A tree-sitter based static analyzer configured via `.boundary.toml` to restrict dependencies across directory structures in CI.
- **`layered-crate`:** Enforces layer dependency limits inside a single crate (e.g. `domain` bottom, `infra` top).

### D. TypeScript & Go

- **TS/JS:** **`dependency-cruiser`** matches import statements against JSON/JS schemas to block circular runs or unauthorized folder imports. Alternatively, **`eslint-plugin-boundaries`** maps structures at the editor level.
- **Go:** **`go-arch-lint`** validates Go dependencies against a YAML file (`.go-arch-lint.yml`), while package names like `internal/` natively block import requests outside their parent module tree at compiler level.

---

## 2. Monorepo & Micro-Frontend Gating (Nx, Turborepo, Bazel, MFE)

Monorepos and Micro-Frontends (MFEs) require dependency isolation to prevent them from devolving into a distributed monolith.

```
       +--------------------------------------------+
       |           Shell / Host App                 |
       +--------------------------------------------+
                              |
                +-------------+-------------+
                |                           |
                v                           v
       +-----------------+         +-----------------+
       |  Order MFE      |         |  Billing MFE    |
       +-----------------+         +-----------------+
                |                           |
                +-------------+-------------+ (Forbidden direct import)
                              |
                              v
                      Shared Design System
```

### A. Nx Workspace Boundary Tags

Nx utilizes project metadata tags inside `project.json` files and enforces restrictions via `@nx/enforce-module-boundaries` in ESLint.

```json
// libs/admin/feature-dashboard/project.json
{
  "name": "admin-feature-dashboard",
  "tags": ["scope:admin", "type:feature"]
}
```

In `.eslintrc.json`:

```json
"@nx/enforce-module-boundaries": [
  "error",
  {
    "depConstraints": [
      {
        "sourceTag": "type:feature",
        "onlyDependOnLibsWithTags": ["type:ui", "type:util", "type:data-access"]
      },
      {
        "sourceTag": "scope:admin",
        "onlyDependOnLibsWithTags": ["scope:admin", "scope:shared"]
      }
    ]
  }
]
```

### B. Turborepo Boundaries

Turborepo features native **experimental boundaries** defined inside `turbo.json`:

```json
{
  "boundaries": {
    "tags": {
      "public-api": { "dependencies": { "allow": ["shared-utils"] } }
    }
  }
}
```

### C. Bazel Target Visibility Rules

Bazel enforces boundaries using the **default-deny** model at the analysis phase. It stops compilation before it begins if packages reference private rules.

```python
# //packages/auth/BUILD
package(default_visibility = ["//visibility:private"])

cc_library(
    name = "auth_api",
    srcs = ["public_api.cc"],
    visibility = ["//visibility:public"],
)

cc_library(
    name = "auth_core",
    srcs = ["auth_core.cc"],
    visibility = ["//packages/auth:__subpackages__", "//packages/billing:__subpackages__"],
)
```

### D. Micro-Frontend Boundaries

To prevent runtime pollution and ensure independent deployability:

1.  **Strict Directed Acyclic Graph (DAG):** App Container imports MFEs; MFEs **never** import other MFEs.
2.  **Asynchronous Communication:** MFEs communicate via custom DOM Events (`window.dispatchEvent`) rather than referencing shared stores.
3.  **Module Federation Encapsulation:** Architects configure webpack federation plugins to expose _only_ the entry shell and bundle runtime packages as singletons.
    ```javascript
    new ModuleFederationPlugin({
      name: 'order_mfe',
      filename: 'remoteEntry.js',
      exposes: { './OrderApp': './src/bootstrap.tsx' },
      shared: { react: { singleton: true } },
    });
    ```

---

## 3. Architecture Decision Records (ADRs) Automation

Architects apply "Docs-as-Code" principles to ADRs. Documentation is checked in along with code changes, linted for structure, and compiled dynamically into searchable index files.

### A. Lifecycle Management & Formatting

- **`adr-tools`:** Command line interface for initializing, drafting (`adr new`), linking, and superseding records.
- **`adr-log`:** Scans files to build or refresh the index directory (`docs/adr/index.md`).
- **`markdownlint`:** Checks style structure, ensuring that headings and lists adhere to Nygard or MADR patterns.
- **`Spectral`:** Utilized to validate frontmatter structure (e.g. status constraints, deciders, dates) by parsing frontmatter into JSON.

### B. Automated CI Verification Workflows

A common pattern checks if a PR touches core architecture folders or changes dependencies (e.g. `go.mod`, `Cargo.toml`, `package.json`) and **fails the build** if no accompanying ADR is added or modified in the same PR.

```yaml
# .github/workflows/adr-enforcer.yml
name: ADR Presence Gate
on:
  pull_request:
    types: [opened, synchronize, reopened]

jobs:
  verify-adr:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Detect Architectural & Dependency Changes
        id: change_detector
        run: |
          TRIGGER_PATTERN="^(go\.mod|package\.json|Cargo\.toml|internal/core/|db/migrations/)"
          CHANGED_FILES=$(git diff --name-only origin/${{ github.base_ref }}...HEAD)
          TRIGGERED_CHANGES=$(echo "$CHANGED_FILES" | grep -E "$TRIGGER_PATTERN" || true)

          if [ -n "$TRIGGERED_CHANGES" ]; then
            echo "requires_adr=true" >> $GITHUB_OUTPUT
          else
            echo "requires_adr=false" >> $GITHUB_OUTPUT
          fi

      - name: Ensure ADR File exists
        if: steps.change_detector.outputs.requires_adr == 'true'
        run: |
          ADR_CHANGES=$(git diff --name-only origin/${{ github.base_ref }}...HEAD | grep "^docs/adr/" || true)
          if [ -z "$ADR_CHANGES" ]; then
            echo "::error::Architectural modifications made but no ADR found under docs/adr/."
            exit 1
          fi
```

---

## 4. AST-Based Rules & Custom Scripting

When declarative rules are insufficient, architects write custom scripts that parse the Abstract Syntax Tree (AST) of source code files to enforce semantic invariants.

### A. Python (`ast`)

Enforces that controller files do not perform direct database queries.

```python
import ast

class ControllerDbAccessChecker(ast.NodeVisitor):
    def __init__(self, filename):
        self.filename = filename
        self.violations = []

    def visit_Import(self, node):
        for alias in node.names:
            if alias.name.startswith("db") or "models" in alias.name:
                self.violations.append(f"[{self.filename}:{node.lineno}] Forbidden direct import in controller.")
        self.generic_visit(node)

    def visit_Call(self, node):
        if isinstance(node.func, ast.Attribute):
            curr = node.func
            while isinstance(curr, ast.Attribute):
                curr = curr.value
            if isinstance(curr, ast.Name) and curr.id == "db":
                self.violations.append(f"[{self.filename}:{node.lineno}] Forbidden direct database call on 'db'.")
        self.generic_visit(node)
```

### B. Go (`go/analysis`)

Catches database sql imports in delivery layers.

```go
package nocontrollerdb

import (
	"go/ast"
	"strings"
	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "nocontrollerdb",
	Doc:  "prevents importing database packages directly in controllers",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	pkgPath := pass.Pkg.Path()
	if !strings.Contains(pkgPath, "controller") && !strings.Contains(pkgPath, "handler") {
		return nil, nil
	}

	forbidden := map[string]bool{
		"database/sql":            true,
		"gorm.io/gorm":            true,
		"github.com/jmoiron/sqlx": true,
	}

	for _, file := range pass.Files {
		for _, imp := range file.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			if forbidden[path] {
				pass.Reportf(imp.Pos(), "architectural violation: forbidden import %s in controller", path)
			}
		}
	}
	return nil, nil
}
```

### C. TypeScript ESLint Custom Rules

Using `@typescript-eslint/utils` to check if methods matching route decorators return API envelope models.

```typescript
import { ESLintUtils } from '@typescript-eslint/utils';

export default ESLintUtils.RuleCreator((name) => `https://rules.myorg.com/${name}`)({
  name: 'controller-response-envelope',
  meta: {
    type: 'problem',
    docs: { description: 'Enforce that all route handlers return ApiResponse<T>' },
    messages: { missingEnvelope: 'Route handler method must return an ApiResponse envelope.' },
    schema: [],
  },
  defaultOptions: [],
  create(context) {
    const services = ESLintUtils.getParserServices(context);
    const checker = services.program.getTypeChecker();
    return {
      MethodDefinition(node) {
        const hasRouteDecorator = node.decorators?.some((dec) => {
          const text = dec.expression.getText();
          return /^(Get|Post|Put|Delete)\(/.test(text);
        });
        if (!hasRouteDecorator) return;

        const tsNode = services.esTreeNodeToTSNodeMap.get(node);
        const signature = checker.getSignatureFromDeclaration(tsNode);
        if (!signature) return;

        const returnType = checker.getReturnTypeOfSignature(signature);
        const typeStr = checker.typeToString(returnType);

        if (!typeStr.includes('ApiResponse<') && !typeStr.includes('Promise<ApiResponse<')) {
          context.report({ node, messageId: 'missingEnvelope' });
        }
      },
    };
  },
});
```

### D. Rust Compiler Macros

A custom procedural macro attribute checks return types at compile-time.

```rust
#[proc_macro_attribute]
pub fn controller_handler(_attr: TokenStream, item: TokenStream) -> TokenStream {
    let input_fn = parse_macro_input!(item as ItemFn);

    let return_type_valid = match &input_fn.sig.output {
        ReturnType::Default => false,
        ReturnType::Type(_, ty) => {
            let type_str = quote! { #ty }.to_string().replace(" ", "");
            type_str.contains("ResponseEnvelope<") || type_str.contains("Result<ResponseEnvelope<")
        }
    };

    if !return_type_valid {
        return syn::Error::new_spanned(&input_fn.sig.output, "Must return ResponseEnvelope")
            .to_compile_error()
            .into();
    }
    quote! { #input_fn }.into()
}
```

---

## 5. Database Boundaries & Query Enforcement

Architects block database design flaws, cross-domain SQL joins in modular monoliths, and lazy-loading $N+1$ query patterns dynamically in CI.

### A. SQLFluff (SQL Linting)

Prevents `SELECT *` patterns and implicit outer/cross joins that degrade query planning performance.

```ini
# .sqlfluff
[sqlfluff:rules:ambiguous.column_count]
enabled = True # Denies SELECT *

[sqlfluff:rules:ambiguous.join_condition]
enabled = True # Denies implicit joins / Cartesian products
```

### B. Compile-Time Verification (`sqlc vet`)

`sqlc` compiles SQL queries to Go and checks queries against live migrations. The `sqlc vet` command supports **Common Expression Language (CEL)** validation rules:

```yaml
# sqlc.yaml
rules:
  - name: no-delete-without-where
    message: 'DELETE statements must include a WHERE clause.'
    rule: |
      query.sql.contains("DELETE") && !query.sql.contains("WHERE")
  - name: enforce-limit
    message: 'Queries fetching collections must specify a LIMIT.'
    rule: |
      query.sql.startsWith("SELECT") && !query.sql.contains("LIMIT")
```

### C. N+1 Query Prevention & Budgeting

To detect $N+1$ query patterns at runtime, architects inject interceptors into test runner loops:

- **Java (QuickPerf):** Enforces query counts on database transactions via unit test annotations.
  ```java
  @QuickPerfTest
  class FeedRepoTest {
      @Test
      @ExpectSelect(1) // Fails if Hibernate executes more than one select (N+1 catch)
      void getUserFeed() { feedService.getUserFeed(1L); }
  }
  ```
- **Python (Django):** Enforces limits within Django test cases using `assertNumQueries`.
  ```python
  with self.assertNumQueries(3):
      self.client.get('/api/feed/')
  ```
- **Ruby on Rails (Bullet):** Fails unit tests if lazy loading is detected.
  ```ruby
  Bullet.enable = true
  Bullet.raise = true # Raise exception in CI
  ```

### D. Custom Go SQL Cross-Domain Join Linter

For modular monoliths where modules share a database but must remain logically isolated (e.g. `user` and `billing`), architects parse migration files into a AST (using `pg_query_go`) and fail if a query executes a `JOIN` across table prefixes belonging to different modules.

```go
// Enforces that package 'follow' does not join table 'users_' directly, respecting D5/D6 rules
if !strings.HasPrefix(tableName, "follows_") {
    log.Fatalf("Boundary violation: cross-slice join detected.")
}
```

---

## 6. API and Schema Drift Detection

API and schema drift checks ensure that changes to code do not break external contracts (REST, DB, GraphQL) without triggering build flags.

```
+------------+        +---------------+        +---------------+
| Spectral   | -----> | oasdiff       | -----> | Schemathesis  |
| Lint Spec  |        | Breaking Chg. |        | Runtime Fuzz  |
+------------+        +---------------+        +---------------+
```

### A. OpenAPI Linting (`Spectral`)

Forces consistency, path casing, and mandatory authentication definition checks.

```bash
spectral lint docs/api/openapi.yaml --fail-severity=error
```

### B. Breaking Change Gating (`oasdiff`)

`oasdiff` runs a backward-compatibility check by comparing the API schema in the current pull request against the base branch (e.g. `main` or latest release tag) and exits with code `1` if a breaking change is detected.

```bash
oasdiff breaking origin/main:openapi.yaml HEAD:openapi.yaml --fail-on ERR
```

### C. Schema Migration & Drift Gating (Atlas & Prisma)

- **Prisma Diff:** Compares the local model file definition against generated migrations using a temporary database shadow, verifying that no model changes were made without creating a migration script.
  ```bash
  npx prisma migrate diff --exit-code --from-migrations ./prisma/migrations --to-schema-datamodel ./prisma/schema.prisma --shadow-database-url "$SHADOW_DATABASE_URL"
  ```
- **Atlas Migrate Lint:** Analyzes SQL migration files to check for table locks, column drops, or non-concurrent index additions.
  ```bash
  atlas migrate lint --env ci --git-base origin/main
  ```

### D. Consumer-Driven Contract Testing (Pact)

Using Pact, consumers register their expectation contracts in a Pact Broker. Providers run these contracts as mock integrations. In the CD deployment stage, `can-i-deploy` blocks deployment if compatibility is unknown.

```bash
pact-broker can-i-deploy --pacticipant MyService --version "$GIT_COMMIT" --to-environment production
```

### E. Runtime Fuzzing (Schemathesis)

Performs property-based testing directly against local Docker execution instances in CI, verifying that all real endpoints conform exactly to the OpenAPI schema types.

```bash
schemathesis run http://localhost:8080/openapi.json --report junit
```

---

## 7. Git Gating & Policy-as-Code (Lefthook, pre-commit, OPA)

To avoid developer friction, architects run selective local checks within a 3-second limit. Heavyweight or system-wide checks are run as policy engines in CI.

### A. Git Hook Gating (Lefthook vs. pre-commit)

- **Lefthook (Go-based, parallel, zero runtime boot overhead):**
  ```yaml
  # lefthook.yml
  pre-commit:
    parallel: true
    commands:
      linter:
        glob: '*.{js,ts,jsx,tsx}'
        run: npx eslint --fix --cache {staged_files}
        stage_fixed: true
      type-check:
        glob: '*.{ts,tsx}'
        run: npx tsc-files --noEmit
      unit-tests:
        glob: '*.{js,ts,jsx,tsx}'
        run: npx vitest related --run {staged_files}
  ```
- **pre-commit (Python-based, virtualenv sandbox environments):**
  Allows managing tool dependencies inside the hook framework, locking versions. Use `pass_filenames: false` to force project-wide tools to evaluate the codebase exactly once instead of invoking them once per changed file.

### B. Policy-as-Code (OPA & Checkov)

- **Checkov:** Scans IaC templates (Terraform, Kubernetes) for misconfigurations based on pre-built rules.
- **Open Policy Agent (OPA/Rego):** Allows architects to author custom declarative rules to evaluate plan changes or admission limits.

#### Example Rego Policy: Enforcing Private S3 Buckets (Terraform Plan)

```rego
package terraform.s3

import rego.v1

s3_buckets contains resource if {
    some resource in input.resource_changes
    resource.type == "aws_s3_bucket"
    resource.change.actions[_] in ["create", "update"]
}

deny contains msg if {
    some bucket in s3_buckets
    acl := bucket.change.after.acl
    acl != "private"
    msg := sprintf("S3 Bucket '%s' violates security policy: ACL must be 'private'", [bucket.address])
}
```

---

## 8. Code Quality Gates & Cognitive Complexity

Architects define quality gates to reject structural code additions that exceed readability thresholds.

### A. SonarQube Quality Profile & Gate Configurations

In SonarQube, architects activate the rule `Cognitive Complexity of methods should not be too high` (typically thresholded at 15) and categorize it as a **Blocker**.

Using Terraform, they declare the Quality Gate:

```hcl
resource "sonarqube_qualitygate" "strict_gate" {
  name       = "deterministic-quality-gate"
  is_default = true

  condition {
    metric    = "new_coverage"
    op        = "LT"
    threshold = "80"
  }
  condition {
    metric    = "new_blocker_violations"
    op        = "GT"
    threshold = "0"
  }
}
```

To fail the build in CI, architects run the scanner with:

```bash
sonar-scanner -Dsonar.qualitygate.wait=true
```

### B. CodeClimate Configurations

Defined in `.codeclimate.yml`:

```yaml
version: '2'
checks:
  complex-logic: # Cognitive complexity check
    enabled: true
    config:
      threshold: 4
  similar-code:
    enabled: true
    config:
      threshold: 15
```

### C. Go Linter (`golangci-lint`)

Integrates complexity check tools directly into `.golangci.yml`:

```yaml
linters-settings:
  gocyclo:
    min-complexity: 15 # Cyclomatic paths
  gocognit:
    min-complexity: 15 # Cognitive layout
linters:
  enable:
    - gocyclo
    - gocognit
```

---

## 9. Synthesis: The Unified Shift-Left Pipeline Blueprint

To maintain developer velocity while guaranteeing architectural constraints, senior architects organize validation gates into a multi-tiered pipeline:

```
               [ Developer Loop ]
              +------------------+
              | Local Pre-commit | (Fast, file-specific checks)
              |  - Lefthook      | (Lint-staged, secrets check, unit tests)
              +------------------+
                       |
                       v
               [ Git Pull Request ]
              +------------------+
              |    CI Gateway    | (Static boundary & schema verification)
              |  - go-arch-lint  | (Verify layers and circular imports)
              |  - oasdiff       | (Verify backward compatibility)
              |  - ADR Enforcer  | (Ensure architecture docs present)
              |  - SQLFluff / AST| (Prevent database joins & SQL anti-patterns)
              +------------------+
                       |
                       v
               [ Dynamic Testing ]
              +------------------+
              | Integration Gate | (Interceptors & dynamic validation)
              |  - Schemathesis  | (Live OpenAPI compliance verification)
              |  - QuickPerf     | (Catch N+1 query patterns)
              +------------------+
                       |
                       v
               [ Delivery Gate ]
              +------------------+
              |    CD Release    | (Policy-as-Code and contract matrices)
              |  - Pact (can-i)  | (Check provider-consumer matrices)
              |  - OPA (Rego)    | (Check Terraform infrastructure plans)
              +------------------+
```

By decoupling boundary enforcement from manual code reviews, architects ensure that architectural guidelines are maintained deterministically at every phase of the software development lifecycle.
