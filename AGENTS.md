# AGENTS.md

Guidelines, rules, and configuration tracking for AI agents operating in this repository.

---

## 1. RTK - Rust Token Killer

All shell commands MUST be prefixed with the `rtk` command to optimize context size and minimize token consumption:

- **Rule**: Use `rtk <command>` instead of raw shell commands.
- **Examples**:
  - `rtk git status`
  - `rtk bash hello_devops_test.sh`
  - `rtk ls -la`
- **Savings**: Run `rtk gain` to view cumulative token savings in this session.

---

## 2. Karpathy Guidelines

Follow the 4 principles of software engineering to minimize coding errors (see details in the local karpathy-guidelines skill) (@.agents/skills/karpathy-guidelines/SKILL.md):

1. **Think Before Coding**: Explicitly state assumptions, surface tradeoffs, and clarify ambiguity.
2. **Simplicity First**: Write the minimum amount of code to solve the problem. Nothing speculative.
3. **Surgical Changes**: Touch only what you must. Match existing style.
4. **Goal-Driven Execution**: Define clear success criteria and verify changes systematically.

---

## 3. Bug Protocol

**If you find a bug, first write a test that reproduces it, then fix the bug and verify the test passes.**

---

## 4. Gitea CLI (tea)

Use the Gitea CLI (`tea`) to manage issues, pull requests, releases, and repository configuration under the `zone01` login.

---

## 5. Git push last changes

Always push after you made changes. Never use `--force` or `--force-with-lease`. Use softer ways to resolve conflicts, rebasing etc.

---

## 6. Verification Gates

Before completing a task, run verification gates to ensure code quality and that all requirements are met. Use `make gates` to run all gates or `make test` for basic tests.

---

## 7. Documentation Drifts

Whenever you make changes or updates to files (docs), you should also update all other docs need update.
You should run a validation and verfication subagents before closing a task to ensure there is no drift in documentation.

- Critical folders to always check are :
  - `README.md`
  - `docs/architecture/architecture.md`
  - `docs/architecture/DEVELOPMENT.md`
  - `docs/architecture/sds.md`
  - `docs/architecture/target-architecture-with-phases.md`
  - `.agents/rules/conventions.md`
