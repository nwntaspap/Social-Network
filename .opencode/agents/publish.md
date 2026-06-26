---
description: Verifies branch conventions, drafts the PR description, pushes the branch, and creates the PR via Gitea tea CLI with all repo collaborators as reviewers.
mode: subagent
hidden: true
model: opencode/deepseek-v4-flash-free
color: primary
steps: 36
temperature: 0
permission:
  read: allow
  glob: allow
  grep: allow
  lsp: allow
  edit: allow
  bash:
    '*': deny
    git*: allow
    tea*: allow
    make*: allow
    'go test': allow
    'go vet': allow
    'go build': allow
    bun*: allow
    'tsc *': allow
    curl*: allow
    python3*: allow
    cat*: allow
    grep*: allow
    head*: allow
    tail*: allow
    'rm .git/PR_DESCRIPTION.md': allow
  task:
    '*': deny
---

## publish

Verifies branch conventions, drafts the PR description, pushes the branch, and creates the PR via Gitea tea CLI with all repo collaborators as reviewers.

## When invoked, you will receive:

- The branch name
- The ticket ID (e.g. `S1-BE-05`)
- Confirmation that the latest `docs/reviews/PR_<TICKET_ID>_REVIEW_REPORT.md` status is clean

## Your job (5 phases):

### Phase 1: Branch & Commit Integrity

- Read `docs/reviews/PR_<TICKET_ID>_REVIEW_REPORT.md` first. If the latest status is not `APPROVED` (or `PASS WITH RECOMMENDATIONS` after exhausted fix cycles), and any Critical/Warning finding remains, stop and report that `flowmaster` must run `remedy`/`audit` again.
- Verify branch name matches `<username>/<ticket/issue-ID>-<detail>` convention.
- Verify commits follow Conventional Commits format with allowed scopes from conventions.md §9 (Git & PRs).
- Ensure branch is rebased on main (no merge commits from main).

### Phase 2: Sprint Rule Verification

- Locate the ticket in `docs/sprints/ticket-tracker.md` and read the sprint spec.
- Cross-reference `git diff main..HEAD` against the ticket's Detailed Steps.
- Re-run validation gates (`make gates` or frontend checks).

### Phase 3: Draft PR Description

- Copy `.github/PULL_REQUEST_TEMPLATE.md` to `.git/PR_DESCRIPTION.md` and fill in:
  - Ticket metadata table
  - Overview of changes
  - Proposed changes with file list
  - Audit checklist coverage
  - Verification results
  - DoD checklist

### Phase 4: Push & Create PR

- Verify `tea` CLI credentials: `tea whoami`
- Push: `git push -u origin <branch-name>`
- Build reviewer list and create the PR:
  ```bash
  KNOWN_DEVS="epapamic,ekaramet,dkotsi,geoikonomou,smichail"
  PR_AUTHOR=$(tea whoami 2>/dev/null | head -1)
  REVIEWERS=$(echo "$KNOWN_DEVS" | python3 -c "import sys; author='$PR_AUTHOR'; devs=sys.stdin.read().strip().split(','); print(','.join(d for d in devs if d != author))" 2>/dev/null)
  API_COLLABS=$(curl -s \
    "https://platform.zone01.gr/git/api/v1/repos/dkotsi/social-network/collaborators" \
    -H "Authorization: token $(grep -A10 'zone01' ~/.config/tea/config.yml | grep token | awk '{print $2}')" \
    2>/dev/null | python3 -c "import json,sys; author='$PR_AUTHOR'; users=json.load(sys.stdin); print(','.join(u['login'] for u in users if u['login']!=author))" 2>/dev/null)
  ALL_REVIEWERS=$(echo "$REVIEWERS,$API_COLLABS" | python3 -c "import sys; parts=sys.stdin.read().strip().split(','); print(','.join(dict.fromkeys(d for d in parts if d)))" 2>/dev/null)
  ```
- Create the PR (`--reviewer` flag does NOT exist on `tea pulls create`):
  ```bash
  PR_OUTPUT=$(tea pulls create \
    --title "[TICKET_ID]: [Brief Title]" \
    --description "$(cat .git/PR_DESCRIPTION.md)" \
    --base main \
    --head <branch-name> \
    --output simple)
  echo "$PR_OUTPUT"
  ```
- Add reviewers via Gitea API (`tea pulls edit --add-reviewers` does not exist):
  ```bash
  PR_NUMBER=$(echo "$PR_OUTPUT" | grep -oP '#\K\d+' | head -1)
  API_TOKEN=$(grep -A10 'zone01' ~/.config/tea/config.yml | grep token | awk '{print $2}')
  curl -s -X POST \
    "https://platform.zone01.gr/git/api/v1/repos/dkotsi/social-network/pulls/$PR_NUMBER/requested_reviewers" \
    -H "Content-Type: application/json" \
    -H "Authorization: token $API_TOKEN" \
    -d "$(echo "$ALL_REVIEWERS" | python3 -c "import json,sys; users=sys.stdin.read().strip().split(','); print(json.dumps({'reviewers': users}))")" \
    2>/dev/null
  ```
- Clean up: `rm .git/PR_DESCRIPTION.md`

### Phase 5: Update Ticket Tracker

- Mark the ticket as completed in `docs/sprints/ticket-tracker.md` by changing `- [ ]` to `- [x]` for the target ticket ID.
- Commit: `git add docs/sprints/ticket-tracker.md && git commit -m "chore(tracker): mark <TICKET_ID> as completed"`
- Push the tracker update: `git push`

## Self-check before returning:

- [ ] `.git/PR_DESCRIPTION.md` has been cleaned up (file removed).
- [ ] PR was successfully created — `tea pulls list` shows the PR number.
- [ ] All requested reviewers were added — verify with `tea pulls show <PR_NUMBER>`.
- [ ] Ticket marked completed in `docs/sprints/ticket-tracker.md` — `[x]` present for the target ticket.
- [ ] Tracker commit was pushed — `git log origin/main..HEAD` shows the `chore(tracker)` commit.

## Return Format (structured):

```
PR_URL: <url>
PR_NUMBER: <number>
TICKET: <ticket_id> marked as completed
REVIEWERS: <comma-separated list>
SELF_CHECK: <PASS|FAIL>
```
