---
model: deepseek-v4-flash-free
date: 2026-06-19
task: Audit conventions.md coverage vs general-instructions.md + target-architecture-with-phases.md
status: complete
---

# Audit: `conventions.md` Gap Report

**conventions.md** covers D1–D6, Strangler Fig, TDD, migrations, branching, security, and frontend basics well. 4 critical gaps found.

---

## Gap 1: Frontend Session Isolation

**Source:** `general-instructions.md` §F3 (State Management)

**Missing from conventions.md §8 (Frontend):**

> "Logging in from Chrome and Firefox with different users must keep sessions separate. Non-logged-in browsers must remain guest sessions."

**Why critical:** Security correctness requirement — without it, a dev might use localStorage or a shared cookie jar, failing audit.

---

## Gap 2: Follow-Gate Chat Validation Message

**Source:** `general-instructions.md` §F2 (Interaction Patterns)

**Missing from conventions.md §8 (Frontend):**

> "If a chat is attempted between non-followed users, show clear validation: 'At least one user must follow the other to initiate a chat.'"

**Why critical:** Defined UX requirement with a specific error string — easy to miss without documentation.

---

## Gap 3: Observable Prometheus Metrics Specification

**Source:** `target-architecture-with-phases.md` §3.3 (Middleware) + §A4 (Observability)

**Missing from conventions.md §9 (Infrastructure):**

> Expose Prometheus-compatible metrics for: request duration, error rate, DB query time.

conventions.md says "expose Prometheus-compatible metrics" but does not specify which 3 metric categories.

**Why critical:** Without guidance, developers will choose inconsistent metrics, creating observability gaps.

---

## Gap 4: CQRS Independent Scaling Path

**Source:** `target-architecture-with-phases.md` §Future Scale (item 4)

**Missing from conventions.md §2 (Refactoring & Slices):**

> - Separate entrypoints: `cmd/commands/main.go` and `cmd/queries/main.go` for asymmetric K8s scaling
> - Database read-replica wiring in the platform/database factory

**Why critical:** This is a future architectural constraint that affects today's wire-up choices (e.g., not hardcoding a single DB connection in main()).

---

## Items Correctly Excluded

The following from source docs are **sprint/planning material**, not standing conventions — correct to omit:

- Bug list (Q1, §Phase 1) — one-time fixes, not recurring rules
- Phase-by-phase migration steps — per-sprint work breakdown
- Risk mitigation table — project management, not coding convention
- Visual dependency map — sprint planning reference
- F1 component-to-feature mapping — too detailed for conventions; belongs in sprint specs
- Per-sprint ticket counts — irrelevant for coding conventions
