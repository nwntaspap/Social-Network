# Multi-Model Analysis Report: Lefthook Gate Output Readability

## Consensus Findings

### [BUG] ANSI escape codes render as raw garbage in lefthook piped output

**Models**: DeepSeek V4, Mimo 2.5, Big Pickle, Kimi K2.6, GLM 5.1

Lefthook captures gate stdout via `tee -a .gates-report.log` (lefthook.yml line 22), which breaks `isatty` detection. ANSI escape sequences (`\033[32m` etc.) and emoji unicode chars appear as visible gibberish. Root cause: the pipe context means `os.Stdout` is not a terminal, but gates has no non-TTY auto-detect.

### [SUGGESTION] Add a plain/terse/simple ASCII output mode

**Models**: DeepSeek V4, Mimo 2.5, Big Pickle, Kimi K2.6, GLM 5.1

All models recommend a second output mode (flag name varies: `--simple`, `--plain`, `--terse`) that produces:

- No ANSI color codes
- No emoji / unicode box-drawing chars
- ASCII-only status markers: `[PASS]`, `[FAIL]`, `[SKIP]`
- Per-gate: gate name, failure reason, suggested fix
- Clean, grep-able, log-friendly output

### [SUGGESTION] Support env-var / auto-detect for non-TTY contexts

**Models**: DeepSeek V4, Mimo 2.5, Big Pickle

Auto-switch to plain mode when stdout is not a terminal (piped, CI, lefthook capture). Also support `NO_COLOR` env var (already partially implemented) and/or `GATES_PLAIN=1`.

### [INSIGHT] Existing infrastructure makes this a small change

**Models**: DeepSeek V4, Mimo 2.5, GLM 5.1

`cmd/gates/main.go` already has:

- `noColor` flag + env var detection
- `iconFor()` function with `[PASS]/[FAIL]/[SKIP]` fallback branch
- `Result.Message` with structured schema: `checked: X | why: Y | status: Z | debug: W`

Extending this to a `--plain` flag is minimal effort — add a new formatter function, no changes to gate internals.

### [SUGGESTION] Plain mode must include actionable suggestions, not just status

**Models**: DeepSeek V4, Mimo 2.5, Big Pickle, Kimi K2.6, GLM 5.1

Current `Result.Message` is a verbose diagnostic line. Plain mode should extract the actionable `debug:` / suggestion part and display it cleanly, e.g.:

```
[FAIL] format-check  Files not formatted  → run: make format
[FAIL] branch-check  Wrong branch name    → use feat/ or fix/ prefix
```

---

## Unique Findings

| Model       | Finding                                                                                                                              |
| ----------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| DeepSeek V4 | `Result.Message` has uniform pipe-delimited schema (`checked: X \| why: Y \| status: Z \| debug: W`) — makes simple renderer trivial |
| DeepSeek V4 | `--json` already exists for machine output but not human-readable simple mode                                                        |
| Mimo 2.5    | Suggest `--format=plain\|fancy\|json` tristate replacing current `--json` flag for forward-compatibility                             |
| Big Pickle  | Adding `printTerseResult()` function in `cmd/gates/main.go` keeps `internal/gates` package clean (no UI logic leak)                  |
| GLM 5.1     | Quickest fix: set `NO_COLOR=1` in lefthook.yml pre-push job — but still leaves emoji unicode chars                                   |
| GLM 5.1     | `--plain` should also replace `━` box-drawing with `=` for full ASCII safety                                                         |
| Kimi K2.6   | Need to locate actual gate scripts to determine output logic location                                                                |

---

## Contradictions

None — all models broadly agree on the problem and solution direction. Minor variations in flag naming (`--simple` vs `--plain` vs `--terse`) and implementation scope.

---

## Per-Model Summary

| Model             | #Findings | Key Insight                                                  |
| ----------------- | --------- | ------------------------------------------------------------ |
| DeepSeek V4 Flash | 6         | Uniform message schema makes terse renderer trivial          |
| Mimo 2.5          | 6         | `--format` tristate flag for future-proofing                 |
| Big Pickle        | 6         | `printTerseResult` keeps gate package clean                  |
| Kimi K2.6         | 4         | Need to locate script locations first                        |
| GLM 5.1           | 5         | Quickest fix = `NO_COLOR` in lefthook.yml, but emojis remain |

---

## Recommended Implementation

1. Add `--simple` flag to `cmd/gates/main.go` (naming per user's description)
2. When active: `noColor = true`, icons use `[PASS]/[FAIL]/[SKIP]`, box-chars → ASCII (e.g. `─` → `-`)
3. Auto-detect non-TTY on `os.Stdout` and switch to simple mode unless `--fancy` explicitly passed
4. Extract suggestion from `debug:` portion of `Result.Message` for clean display
5. In `lefthook.yml`, pass `--simple` to the gates command
6. Also respect `NO_COLOR` / `GATES_SIMPLE=1` env vars

---

## Raw Reports

### Report: DeepSeek V4 Flash

#### Findings

- [BUG] ANSI escape codes leak as raw "numbers" when gates output piped through lefthook (`tee`/pipe breaks terminal detection)
- [SUGGESTION] `--simple` flag would give ASCII-only output for pipe/CI contexts — no emoji, no ANSI, no unicode box-drawing
- [INSIGHT] `NO_COLOR` env var already exists but only strips ANSI codes — still emits unicode chars (✅❌➖━┃) that may render as boxes/question marks in non-UTF-8 terminals
- [SUGGESTION] Gate `Message` fields already embed debug hints (`| debug: run 'make format'...`) — simple mode could promote these to standalone suggestion lines
- [INFO] `--json` flag exists for structured output but not human-readable simple mode
- [SUGGESTION] Simple mode format suggestion: `[PASS|FAIL|SKIP] gate-name  reason  → suggestion`

#### Reasoning

- `main.go` lines 15-45: ANSI wrapping + emoji maps conditional on `noColor` flag — lefthook pipes output through `tee` which breaks isatty detection, raw escape codes appear
- Gate messages (e.g. `gate_format.go:58`, `gate_branch.go:60`) embed structured "why + debug + suggestion" in pipe-delimited string — simple mode parser can extract and display only the actionable parts
- `LintGate`, `BranchGate`, `FormatGate` all follow same message pattern: `checked: X | why: Y | status: Z | debug: W` — uniform schema makes a simple renderer trivial to implement
- Unicode emojis in non-UTF-8 or limited terminal environments (some CI runners) render as tofu boxes — `--simple` avoids this
- Existing `--json` satisfies machine readability but not "clean concise human-readable" requirement

#### Uncertainties

- Whether user wants `--simple` as a new CLI flag or as lefthook-specific config (e.g. env var `LEFTHOOK=1` auto-triggers simple mode)
- Unclear if `--json` + `jq` piping is acceptable as workaround or user specifically wants human-readable terse output
- Not sure which other gates exist (security, coverage, etc.) have messages that follow the same schema for suggestion extraction
- Unknown if lefthook version or config could be adjusted to force `NO_COLOR=1` + strip unicode as lighter fix

---

### Report: Mimo 2.5

#### Findings

- [BUG] Lefthook output pipes ANSI escape codes raw — terminal that can't parse them shows garbled numbers
- [SUGGESTION] Add `--plain` / `GATES_PLAIN=1` flag to `cmd/gates/main.go` for ASCII-only output (no emojis, no ANSI, just `[PASS]`, `[FAIL]`, `[SKIP]` with reason + suggestion)
- [SUGGESTION] Respect `NO_COLOR` env var for ANSI stripping (already partially implemented) but pipe through `cat` or detect non-TTY to auto-switch
- [SUGGESTION] Add `--format=plain|fancy|json` tristate replacing current `--json` flag for forward-compat
- [INSIGHT] `iconFor()` already has `noColor` branch returning `[PASS]/[FAIL]/[SKIP]` — just needs non-TTY detection to trigger automatically
- [INFO] `lefthook.yml` line 22 pipes gates output via `tee -a .gates-report.log` — pipe breaks TTY detection, ANSI always emitted
- [RISK] `printResult()` in plain mode must show actionable fix instructions, not just status — current `Message` field is verbose diagnostic, not fix hint

#### Reasoning

- Lefthook captures stdout of hook commands, disabling TTY — `isatty` would be false but gates tool has no auto-detect
- `.gates-report.log` accumulates ANSI escape codes making log file unreadable without `less -R`
- Adding `--plain` flag + env var gives CI/pipes clean output, local dev gets fancy by default
- `Result.Message` already contains "why" and "status" — plain mode should extract just the actionable suggestion, not dump the full message
- Existing `NO_COLOR` support is good foundation — extend with `isatty` check on `os.Stdout`

#### Uncertainties

- Whether lefthook has built-in `env` passthrough config for `GATES_PLAIN` or needs shell wrapper
- Whether user wants fix instructions per-gate or just clean status + failure reason
- Scope: only `cmd/gates/main.go` or also internal gate implementations need plain-friendly messages
- Whether to keep `--json` separate or unify under `--format`

---

### Report: Big Pickle

#### Findings

- [INSIGHT] Current `--json` output exists but no simple plain-text mode for lefthook/CI consumption
- [BUG] ANSI escape codes render as raw garbage in lefthook captured output (pipe/tee context)
- [SUGGESTION] Add `--terse` flag: ASCII-only, reject reason per gate, suggestion line
- [SUGGESTION] Auto-detect non-TTY output (lefthook pipe) and switch to terse mode
- [INFO] `NO_COLOR` env var removes ANSI codes but keeps emoji — still unreadable in pipes
- [SUGGESTION] LePhook `pre-push` job in `lefthook.yml` uses `tee` — explicitly triggers ANSI bleed

#### Reasoning

- `cmd/gates/main.go` uses raw ANSI codes (`\033[...`) in colored functions — lefthook captures stdout via pipe, terminal escape sequences become visible noise
- `NO_COLOR` strips ANSI codes but emojis (✅ ❌ ➖) and formatting boxes (━) survive — still noisy in log files
- `--json` exists but is too verbose for quick human scan; user wants middle ground between JSON and rich display
- Gate `Result` struct already has `Gate`, `Status`, `Message` fields — Message can carry reason+suggestion, just needs a terse formatter
- `lefthook.yml` line 22: `2>&1 | tee -a .gates-report.log` — tee forces pipe, which disables terminal detection even if gates checked `os.Stdout`
- Adding `--terse` flag in `cmd/gates/main.go` with a new `printTerseResult` func would keep `internal/gates` package clean (no UI logic leak)
- LePhook config could be extended to pass `--terse` conditionally, or gates could detect `os.Stdout` file descriptor via `term.IsTerminal`

#### Uncertainties

- Unknown if user wants terse mode auto-detected (non-TTY) or explicit `--terse` flag (or both)
- Unclear if suggestion strings should live in each Gate impl (`internal/gates/*.go`) or in the formatter layer (`cmd/gates/`)
- Not sure whether `--terse` should be a new renderer or just a format string replacement in the existing print path

---

### Report: Kimi K2.6

#### Findings

- [BUG] Lefthook output has unreadable ANSI escape codes/numbers instead of clean formatting.
- [SUGGESTION] Add dual-output mode for gates: existing colored/emoji format plus plain ASCII version.
- [SUGGESTION] Plain mode should list: hook name, failure reason, fix suggestion.
- [INFO] Lefthook is the hook runner; gate output is likely controlled by custom scripts/hooks.

#### Reasoning

- Lefthook prints ANSI color codes/terminal control sequences. In CI or non-TTY environments these render as numbers/garbage.
- User wants two output modes: (1) current rich format for dev TTY, (2) plain ASCII for CI/scripts/logging.
- Plain mode needs to be concise: what failed and why, no emojis or colors.
- Need to locate gate scripts in `Makefile` or `lefthook.yml` to implement this.

#### Uncertainties

- Where gate output logic lives (shell scripts, lefthook config, or application code).
- Whether lefthook version or configuration controls TTY detection.
- If CI environment variable is available to auto-switch modes.

---

### Report: GLM 5.1

#### Findings

- [BUG] Lefthook pipes raw ANSI/emoji output from gates binary; when lefthook captures stdout, escape sequences render as gibberish (e.g. `\033[32m`, `\u2705`)
- [SUGGESTION] Add `--plain` flag to `cmd/gates/main.go` for ASCII-only, no-color output (leverages existing `noColor` pattern + `iconFor` fallback)
- [SUGGESTION] In `lefthook.yml`, set `NO_COLOR=1` or pass `--plain` for the pre-push gates job to fix unreadable output
- [RISK] Current `--json` flag exists but is machine-oriented, not human-readable plain text — doesn't solve the "concise ASCII summary for humans" need
- [SUGGESTION] Add a dedicated `--plain` output mode that strips emojis, colors, unicode box-drawing chars and shows only gate name + status + failure reason + suggested fix
- [INSIGHT] The `Result.Message` field already contains structured `checked: X | why: Y | status: Z | debug: W` — plain mode just needs to suppress the ANSI/color layer, not reformat messages

#### Reasoning

- Lefthook runs `go run ./cmd/gates/ --all 2>&1 | tee -a .gates-report.log`; the pipe doesn't strip ANSI, and lefthook's own output framing may interfere with terminal escape handling
- `noColor` env-var and `iconFor()` fallback already exist in `main.go` (lines 13, 30-44), proving the pattern — extending to a `--plain` flag is minimal effort
- The `NO_COLOR` env var is already respected; setting it in lefthook.yml is the quickest fix but still leaves emojis (unicode chars like ✅ ❌ ━) that may also render poorly in some terminals/pipes
- A `--plain` flag that sets both `noColor=true` AND switching icons to `[PASS]`/`[FAIL]`/`[SKIP]` AND replacing `━` with `=` would give clean ASCII output suitable for lefthook, CI logs, and pipes
- The message format already includes actionable "debug:" hints — no content changes needed, just presentation layer

#### Uncertainties

- Whether lefthook has its own ANSI-stripping config option that could solve this without code changes
- Whether the user wants `--plain` to also shorten/simplify the `Result.Message` content itself (removing the `checked:|why:|status:|debug:` structure) or just strip visual formatting
- Whether `.gates-report.log` should also receive plain output or keep the fancy format for local review
