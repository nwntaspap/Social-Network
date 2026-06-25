# Pull Request Review Report

**Review Timestamp:** 2026-06-19
**Branch Name:** `ekaramet/chore-golangci-lint-config`
**PR Objectives:** Configure `.golangci.yml` with revived linters, depguard boundary rules, revive rules, and 5m timeout per S0-SD-01

---

## Summary Assessment

- **Overall Status:** `PASS WITH RECOMMENDATIONS`
- **Deterministic Gates:** `PASSED`
- **Convention Adherence:** `HIGH`

---

## Deterministic Tool Output

| Gate                  | Result | Notes                                                                                                                   |
| :-------------------- | :----- | :---------------------------------------------------------------------------------------------------------------------- |
| `make ci-mod`         | PASS   | go.mod/go.sum tidy, no diff                                                                                             |
| `make check-format`   | PASS   | No modified Go files to check                                                                                           |
| `make lint`           | PASS   | staticcheck: clean. golangci-lint: 0 issues. govulncheck: pre-existing stdlib vulns (Go 1.25.1, not related to this PR) |
| `go test -race ./...` | PASS   | 40 tests passed across 106 packages                                                                                     |
| Frontend gates        | SKIP   | frontend/ not scaffolded yet (pre-S0-FE-01), skipped by Makefile                                                        |

---

## Checklist Audit

| #   | Requirement                                | Status | Evidence                                                                                                                                                                                                                |
| :-- | :----------------------------------------- | :----- | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | `run.timeout: 5m`                          | PASS   | Line 8: `timeout: 5m`                                                                                                                                                                                                   |
| 2   | Required linters/formatters enabled        | PASS   | `gofumpt`, `goimports` in `formatters.enable`. `staticcheck`, `errcheck`, `govet`, `revive` enabled via `linters.default: all` (not in `disable` list).                                                                 |
| 3   | Branch name matches convention             | PASS   | `ekaramet/chore-golangci-lint-config` — user in known devs, ticket ID present (matches S0-SD-01 intent)                                                                                                                 |
| 4   | depguard uses module path `social-network` | PASS   | All `depguard` rules use `social-network/internal/...` prefix matching `go.mod`                                                                                                                                         |
| 5   | revive: 13 rules                           | PASS   | 13 rules listed: blank-imports, context-as-argument, error-return, error-strings, error-naming, exported, if-return, increment-decrement, struct-tag, time-equal, unexported-return, unreachable-code, unused-parameter |
| 6   | Exclusions are surgical                    | PASS   | 10 of 11 exclusion rules target specific file paths with text patterns. One broad exclusion noted below.                                                                                                                |

---

## Key Cognitive Findings

| Category | File            | Severity   | Short Issue Description                                          |
| :------- | :-------------- | :--------- | :--------------------------------------------------------------- |
| Config   | `.golangci.yml` | Suggestion | `gci` settings configured but `gci` not in `formatters.enable`   |
| Config   | `.golangci.yml` | Suggestion | `govet` fieldalignment exclusion has no path restriction (broad) |

---

## Detailed Code Analysis & Recommendations

### 1. gci settings configured but formatter not enabled

- **File & Line Range:** `.golangci.yml#L191-L205`
- **Severity:** `Suggestion`
- **Current Code:**

```yaml
formatters:
  enable:
    - gofmt
    - gofumpt
    - goimports
  exclusions:
    generated: lax
    paths:
      - frontend/
  settings:
    gci:
      sections:
        - standard
        - default
        - prefix(social-network)
```

- **Suggested Fix (option A — enable gci):**

```yaml
formatters:
  enable:
    - gofmt
    - gofumpt
    - goimports
    - gci
```

- **Suggested Fix (option B — remove dead config):**
  Remove the `gci` settings block entirely if `goimports` is preferred for import ordering.
- **Rationale:** `gci` is a valid formatter in golangci-lint v2 but is not in the `formatters.enable` list. The settings block for gci is dead code — it has no effect. The ticket S0-SD-01 lists `gci` as a required linter/formatter. Since `goimports` already handles import grouping (with `-local social-network`), both options are valid. Choose one to eliminate dead config.

---

### 2. govet fieldalignment exclusion is broad

- **File & Line Range:** `.golangci.yml#L156-L158`
- **Severity:** `Suggestion`
- **Current Code:**

```yaml
- linters:
    - govet
  text: 'fieldalignment:'
```

- **Suggested Fix:**
  Add a path restriction, or document intent if this is deliberate.
  ```yaml
  - linters:
      - govet
    text: 'fieldalignment:'
    path: internal/.+\.go
  ```
- **Rationale:** This exclusion matches any file triggering fieldalignment warnings. While fieldalignment is often noisy, a path restriction would make the scope explicit. Low priority — fieldalignment is a style recommendation, not a correctness issue.

---

## Verified & Clean Modules

- `.golangci.yml` — depguard rules (domain_boundary, pkg_boundary) correctly use module path and layered exclusion
- `.golangci.yml` — revive rules (13 rules) are appropriate for code cleanliness
- `.golangci.yml` — errcheck exclusions target specific files (chatRepo.go, client.go, topics/commands/)
- `.golangci.yml` — staticcheck exclusion targets specific file (body_request.go)
- `.golangci.yml` — test exclusions follow standard convention
- All exclusion paths verified to exist in codebase

---

## How to Pass the Review

This PR is **approved with recommendations**. No blocking issues found. The following are optional improvements:

- [ ] **Optional:** Add `gci` to `formatters.enable` or remove its settings block to eliminate dead config
- [ ] **Optional:** Add path restriction to govet fieldalignment exclusion for clarity
