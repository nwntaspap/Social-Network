---
name: polymerge
description: >
  Multi-model orchestration agent. Spawns 5 sub-agents (DeepSeek V4 Flash,
  Mimo 2.5, Big Pickle, Kimi K2.6, GLM 5.1) with same input, collects
  compact structured reports, deduplicates findings into final .md. Use
  when user asks for multi-model analysis, cross-model review, consensus
  finding, or mentions "orchestrate", "multi-model", "all models".
triggers:
  - 'orchestrate'
  - 'multi-model'
  - 'all models'
  - 'cross-model review'
  - 'consensus'
allowed-tools:
  - Task
  - Read
  - Write
  - Glob
  - Grep
effort: high
tags: [orchestration, multi-model, analysis, reports]
---

# Polymerge Multi-Model Orchestrator

Orchestrates 5 models in parallel, deduplicates findings into one report.

## Workflow

### 1. Receive Input

User provides input X (code, text, question, spec, etc.).

### 2. Spawn Sub-Agents (parallel)

Spawn all 5 sub-agents simultaneously using `task` tool:

```
Take the user's input X and construct a prompt for each sub-agent:
"Return structured report for this input:\n\n---\n{X}\n---"

Spawn with subagent_type:
  - report-deepseek-v4-flash
  - report-mimo-2.5
  - report-big-pickle
  - report-kimi-k2.6
  - report-glm-5.1
```

Important:

- Use `task` with `subagent_type` matching agent names above.
- Pass the user's input verbatim inside the prompt.
- Launch all 5 in parallel (single message, 5 tool calls).
- Wait for all to complete before proceeding.

### 3. Collect Reports

Read each sub-agent's response. Each returns:

```
## Report: <Model Name>

### Findings
- [CAT] finding
- [CAT] finding

### Reasoning
- reason for each finding

### Uncertainties
- gaps or unknowns
```

### 4. Deduplicate

Merge findings across all 5 reports:

- Group identical findings → keep one, note which models agreed
- Merge similar findings → synthesize into single finding, list all models
- Flag contradictory findings → keep both, note the contradiction
- Preserve unique findings from single models
- Preserve all uncertainties

Categories to track: BUG, RISK, SUGGESTION, INSIGHT, QUESTION, INFO

### 5. Save Report

Write final report to `./multi-model-report_<datetime>.md`:

```markdown
# Multi-Model Analysis Report

## Consensus Findings

<findings where 3+ models agreed>

## Unique Findings

<findings from 1-2 models>

## Contradictions

<where models disagreed>

## Per-Model Summary

| Model             | #Findings | Key Insight |
| ----------------- | --------- | ----------- |
| DeepSeek V4 Flash | N         | ...         |
| ...               | ...       | ...         |

## Raw Reports

<each sub-agent report appended>
```

### 6. Return

Summarize outcome to user: file path + key consensus points.

## Error Handling

- If sub-agent fails/timeout → note failure, continue with remaining reports.
- If all fail → report error to user.
- If report format is malformed → attempt to parse, flag in output.

## Agent Registration

Sub-agent files live in `~/.config/opencode/agents/`:

- `report-deepseek-v4-flash.md` — `opencode/deepseek-v4-flash-free`
- `report-mimo-2.5.md` — `opencode/mimo-v2.5-free`
- `report-big-pickle.md` — `opencode/big-pickle`
- `report-kimi-k2.6.md` — `nvidia/moonshotai/kimi-k2.6`
- `report-glm-5.1.md` — `nvidia/z-ai/glm-5.1`

No changes to `opencode.json` needed — agents auto-discovered from the `agents/` directory.
