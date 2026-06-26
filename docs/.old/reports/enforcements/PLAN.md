# .opencode Agent Configuration — Change Plan

## Root Cause Fixed

- `{file:.opencode/agents/*.md}` paths resolved relative to config directory `.opencode/`, producing double-nested path `.opencode/.opencode/agents/*.md` → opencode crashed on startup.
- **Fix**: Removed `"agent"` block from `opencode.json`. All 5 agents now standalone markdown files in `.opencode/agents/` with YAML frontmatter.

## All Changes

### 1. Convert agents to standalone markdown frontmatter format

**Files**: `.opencode/agents/publish.md`, `remedy.md`, `forge.md`, `audit.md`, `flowmaster.md`

Each `.md` file now has YAML frontmatter with full config. Filename = agent name (opencode auto-discovers).

Config moved from `opencode.json` into frontmatter:

- `description`, `mode`, `model`, `color`, `steps`, `temperature`, `permission`

### 2. Removed explicit `"agent"` block from `opencode.json`

**File**: `.opencode/opencode.json`

- `"agent": { ... }` removed entirely — agents auto-discovered from `.opencode/agents/*.md`
- `"plugin"` key also removed — plugin auto-discovered from `.opencode/plugins/`

### 3. Subagents now deny spawning subagents

**Scope**: all 4 subagents (`publish`, `remedy`, `forge`, `audit`)

```yaml
task:
  '*': deny
```

Prevents infinite recursion. Only `flowmaster` (primary orchestrator) has `task: {"*": allow}`.

### 4. Distinct colors per agent type

| Agent        | Color     | Role               |
| ------------ | --------- | ------------------ |
| `publish`    | `primary` | Push & PR creation |
| `remedy`     | `warning` | Bugfix loop        |
| `forge`      | `success` | Implementation     |
| `audit`      | `accent`  | Code review        |
| `flowmaster` | `primary` | Orchestrator       |

### 5. Added step limits (cost guardrails)

| Agent        | Steps |
| ------------ | ----- |
| `publish`    | 30    |
| `remedy`     | 25    |
| `forge`      | 50    |
| `audit`      | 40    |
| `flowmaster` | 60    |

### 6. `audit` edit permission — granular allow for report file

```yaml
edit:
  '*': deny
  'docs/reviews/PR_REVIEW_REPORT.md': allow
```

Agent needs `edit` to write the review report but should not modify code.

### 7. `audit` bash — added catch-all `"*": "ask"`

Missing `"*": "ask"` meant unknown bash commands had undefined permission. Added as first rule.

### 8. Removed stray Chinese characters in `flowmaster.md`

Line 6 had `枪口` ("muzzle") — removed.

### 9. Graphify plugin — session reset for reminder

**File**: `.opencode/plugins/graphify.js`

- `reminded` flag now resets on `"session.created"` event, so each new session gets the graph hint.

### 10. `forge.md` — removed stale "HumanLayer RPI" reference

Description now says "RPI framework" consistently; "HumanLayer" prefix dropped.

## Verification

- `opencode` starts with 0 config errors
- All 5 agents auto-discovered from `.opencode/agents/*.md`
- Prompt content loads correctly (1211-1577 chars each)
- Permissions, colors, steps, models match frontmatter config

## Files Modified

```
.opencode/opencode.json              — trimmed to just $schema (agents now .md)
.opencode/agents/publish.md        — standalone frontmatter agent
.opencode/agents/remedy.md           — standalone frontmatter agent
.opencode/agents/forge.md     — standalone frontmatter agent, removed HumanLayer ref
.opencode/agents/audit.md        — standalone frontmatter agent, fine-grained edit perm
.opencode/agents/flowmaster.md     — standalone frontmatter agent, removed stray chars
.opencode/plugins/graphify.js        — session.created resets reminded flag
.opencode/PLAN.md                    — this file
```
