# General-Instructions Gap Analysis: Frontend & Workflow Enforcement

> Cross-references: `general-instructions.md` vs `target-architecture-with-phases.md`, `docs/requirements/audit.md`, `docs/requirements/readme.md`, `.agents/workflows/`

---

## Part 1: Frontend Details Missing from general-instructions.md

### Problem

general-instructions.md is ~80% backend-focused. Frontend gets:

- Q2: 4 verification commands (ESLint, Prettier, tsc, vitest)
- A6: Two FE lines in Definition of Done
- Phase 6 of target-architecture: scaffold + pages summary

**Missing:** Every structural, process, and quality detail that the backend section specifies for Go is absent for Next.js/TS.

### 1.1 Required Additions (driven by audit.md + readme.md)

| Gap                                          | Source Requirement                                                                                                                                        | Proposed Section                                                                                                           |
| -------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| Registration form field mapping              | audit.md:39 — "correct form elements: Email, Password, First Name, Last Name, Date of Birth, Avatar (Optional), Nickname (Optional), About Me (Optional)" | **F1: UI Requirement Mapping** — explicit component/field list per audit checklist item                                    |
| Profile page content                         | audit.md:87-98 — display all registration info (except password), user posts, followers/following, privacy toggle                                         | **F1** — profile page must render all user data fields                                                                     |
| Follow/unfollow + confirmation popup         | audit.md:79, readme.md:229 — "unfollow confirmation pop-up" is bonus                                                                                      | **F2: Interaction Patterns** — follow/unfollow buttons, confirmation dialogs for privacy toggle and unfollow (bonus items) |
| Post privacy selector UI                     | readme.md:172-176 — three levels (public, almost_private, private+user picker)                                                                            | **F2** — visibility radio/select + user-picker for "private"                                                               |
| Group browse/join/invite                     | readme.md:181-198 — group discovery, invitation accept/decline, join request, event creation form                                                         | **F1** — group directory page, invitation UI, event form (title, description, day/time, 2+ options)                        |
| Chat: follow-gated, emoji, group chat        | readme.md:204-211 — private message only between followed users, emoji support, group chat                                                                | **F2** — chat input with emoji support, follow-gate UI feedback                                                            |
| Notifications: visible on every page         | readme.md:219 — "check notifications on every page"                                                                                                       | **F2** — notification bell/pill in global nav, always visible                                                              |
| Notifications vs messages: visually distinct | readme.md:220 — "different from new private messages"                                                                                                     | **F2** — separate icon, separate panel, visually distinct styling                                                          |
| Session persistence                          | audit.md:55-63 — login persists on refresh, two-browser independence                                                                                      | **F3: State Management** — session cookie handling, auth state persistence                                                 |
| Image upload: JPG/PNG/GIF                    | readme.md:55, audit.md:123-128 — support JPEG, PNG, GIF uploads                                                                                           | **F4: File Handling** — upload component, MIME validation, preview                                                         |
| Docker: 2 containers                         | readme.md:106, audit.md:213-219 — backend + frontend containers                                                                                           | **F6: Build & Deploy** — Dockerfile for Next.js, expose port 3000                                                          |

### 1.2 Proposed New Sections for general-instructions.md

#### F1: Frontend Feature-to-Audit Mapping (REQUIRED)

Maps each audit.md checklist item to a frontend component/page. Ensures no grading item is missed.

| Audit Item                                         | Frontend Route                  | Key Component                                                                                                        |
| -------------------------------------------------- | ------------------------------- | -------------------------------------------------------------------------------------------------------------------- |
| Registration form fields                           | `/register`                     | `RegisterForm` (email, password, firstName, lastName, dob, avatar[ optional], nickname[optional], aboutMe[optional]) |
| Login                                              | `/login`                        | `LoginForm` (email/username + password, OAuth buttons)                                                               |
| Profile display (all fields except password)       | `/profile/[id]`                 | `ProfileCard` + `ProfilePosts` + `FollowersList` + `FollowingList`                                                   |
| Privacy toggle                                     | `/profile/[id]`                 | `PrivacyToggle` with confirmation dialog                                                                             |
| Private profile lock screen                        | `/profile/[id]`                 | `PrivateProfileLock` (shown to non-followers)                                                                        |
| Follow/unfollow                                    | Profile page                    | `FollowButton` + `UnfollowConfirmDialog`                                                                             |
| Follow request (accept/decline)                    | Notifications                   | `FollowRequestActions`                                                                                               |
| Post creation + visibility selector                | `/post/new`                     | `PostForm` + `VisibilitySelector` + `AllowedUsersPicker`                                                             |
| Post image/GIF attachment                          | `/post/new`                     | `ImageUploader` (accept JPG, PNG, GIF)                                                                               |
| Comment with image                                 | Post detail                     | `CommentForm` + `ImageUploader`                                                                                      |
| Group creation                                     | `/groups/new`                   | `GroupForm` (title, description)                                                                                     |
| Group browse                                       | `/groups`                       | `GroupDirectory`                                                                                                     |
| Group invitation accept/decline                    | Notifications                   | `GroupInviteActions`                                                                                                 |
| Group join request                                 | `/groups/[id]`                  | `JoinRequestButton`                                                                                                  |
| Group posts/comments                               | `/groups/[id]`                  | `GroupFeed` + `GroupPostForm`                                                                                        |
| Event creation (title, desc, day/time, 2+ options) | `/groups/[id]/events/new`       | `EventForm` (title, description, datetime picker, min 2 options)                                                     |
| Event RSVP                                         | `/groups/[id]/events/[eventId]` | `RSVPOptions` (going/not going)                                                                                      |
| Chat: private + follow-gated                       | `/chat/[userId]`                | `ChatWindow` (follow-check, emoji support)                                                                           |
| Chat: group chat room                              | `/groups/[id]/chat`             | `GroupChatWindow`                                                                                                    |
| Notification bell (every page)                     | Global nav                      | `NotificationBell` + `NotificationPanel`                                                                             |
| Notification vs message distinction                | Global nav                      | Separate icons, separate panels, different styling                                                                   |
| OAuth (bonus)                                      | `/login`                        | `GitHubOAuthButton` + `GoogleOAuthButton`                                                                            |

#### F2: Frontend Interaction Patterns (REQUIRED)

- **Confirmation dialogs**: Unfollow, privacy toggle, group leave (maps to audit.md bonus items)
- **Follow-gate feedback**: When chat is denied due to no follow relationship, show explicit "You must follow this user to start a chat" message
- **Notification triggers**: Real-time via WebSocket/SSE — follow request, group invite, group join request, event creation (all 4 spec-required + any extras)
- **Emoji support**: Unicode text in chat input. No emoji picker required but recommended (bonus).
- **Typing indicators / online presence**: Listed in target-architecture Phase 6.4 but not in audit.md — mark as **optional/best practice**

#### F3: Frontend State Management (REQUIRED)

- **Auth state**: Session cookie (httpOnly, Secure, SameSite=Lax). Read via server-side session check or `/api/me` endpoint.
- **Persist on refresh**: Session is cookie-based (server manages). Frontend re-validates on mount via server component or middleware.
- **Two-browser independence**: No localStorage-only auth. Session is server-side.
- **Client state**: React context or zustand for UI state (selected tab, modal open/closed). NOT for auth or domain data.
- **Server state**: Use React Server Components (RSC) for data fetching. Client components only for interactivity.

#### F4: Frontend File Handling (REQUIRED)

- **Accepted types**: JPEG, PNG, GIF (audit.md requires all three for both posts and comments)
- **MIME validation**: Client-side check (accept attribute) + server-side magic-byte validation (existing `pkg/imgutil/`)
- **File size limit**: Display error if file exceeds server limit
- **Preview**: Show image preview before upload

#### F5: Frontend Project Structure (REQUIRED)

```
frontend/
  src/
    app/                          # Next.js App Router routes
      (auth)/                     # Route group: login, register
      (main)/                     # Route group: authenticated layout
        profile/[id]/
        post/
        groups/
          [id]/
            events/
            chat/
        chat/[userId]/
        notifications/
    components/
      ui/                         # shadcn/ui primitives
      features/                   # Domain-specific composites
        auth/
        profile/
        post/
        group/
        chat/
        notification/
    lib/
      api-client.ts               # Typed API client (fetch wrapper)
      auth.ts                     # Session helpers
      ws.ts                       # WebSocket connection manager
    styles/
      globals.css                 # Tailwind + HSL custom properties
    types/
      api.ts                      # API response types (from OpenAPI spec)
```

#### F6: Frontend Build & Deploy (REQUIRED)

- **Runtime**: Bun (package manager + script runner)
- **Dockerfile**: Multi-stage build. Output: standalone Next.js server on port 3000
- **Environment**: `NEXT_PUBLIC_API_URL` pointing to backend (default `http://localhost:8080`)
- **Verification commands** (must match Q2):
  ```bash
  bun run lint          # ESLint lint
  bun run format:check  # Prettier format check
  tsc --noEmit          # Type checking
  bun run test          # Vitest unit/component tests
  bun run test:e2e      # Playwright E2E tests
  ```

### 1.3 Optional Additions (Best Practices — NOT enforced)

| Topic                         | Content                                                                                                               |
| ----------------------------- | --------------------------------------------------------------------------------------------------------------------- |
| **Accessibility (a11y)**      | ARIA labels on interactive elements, keyboard navigation, focus management for modals/dialogs                         |
| **Dark mode / glassmorphism** | Tailwind HSL custom properties, `next-themes` for dark mode toggle. Listed in target-architecture Phase 6.1           |
| **Micro-animations**          | Framer Motion or CSS transitions. Listed in target-architecture Phase 6.1                                             |
| **Typography**                | Google Fonts (Inter or Outfit). Listed in target-architecture Phase 6.1                                               |
| **Error boundaries**          | Next.js `error.tsx` files per route segment                                                                           |
| **Loading states**            | Next.js `loading.tsx` skeleton components per route                                                                   |
| **Optimistic updates**        | For follow/unfollow, like/vote, RSVP — update UI before server confirmation                                           |
| **Image optimization**        | Next.js `<Image>` component for next/image auto-optimization                                                          |
| **Code splitting**            | Dynamic imports for heavy components (emoji picker, image editor)                                                     |
| **Testing patterns**          | Vitest: mock server actions, test component rendering with `@testing-library/react`. Playwright: page objects for E2E |

---

## Part 2: Frontend Onboarding Section (Proposed for general-instructions.md)

### 2.1 Onboarding: How to Pick a Ticket

```markdown
## Onboarding: Frontend Developers

### Step 1: Pick a Ticket

1. Open `docs/sprints/ticket-tracker.md` — find unchecked `FE-A` or `FE-B` items.
2. Check dependencies: does the ticket depend on a BE ticket not yet done?
   - If BE ticket is incomplete → pick a different FE ticket or ask if API mock exists (S1-SD006).
3. Assign yourself: update the checkbox in ticket-tracker.md with your name.
4. Open the sprint file for detailed steps.

### Step 2: Read the Context

- Read `docs/sprints/general-instructions.md` sections F1–F6.
- Read `docs/architecture/target-architecture-with-phases.md` Phase 6.
- Read `docs/requirements/audit.md` for the grading checklist items your ticket covers.
- Read the corresponding OpenAPI spec in `docs/api/<feature>.yaml` (if available).

### Step 3: Set Up Your Branch

- Branch name: `<username>/<type>-<detail>` (e.g. `fe-dev-a/feat-s2-fe-01-register-page`)
- Branches live ≤ 3 days.
- Use TDD: write component tests first, then implement.

### Step 4: Implement (TDD)

- RED: Write failing Vitest test for your component/hook.
- GREEN: Implement minimum code to pass.
- REFACTOR: Clean up, ensure ESLint + Prettier lint/format passes.
- Run: `bun run lint && bun run format:check && tsc --noEmit && bun run test`

### Step 5: Open a PR

- Use `pr-create` workflow or follow template below.
- FE reviews FE (cross-review with other frontend dev).
- PR must pass: ESLint + Prettier, tsc, Vitest, manual smoke test.
```

### 2.2 PR Template for Frontend

````markdown
# PR: [Ticket ID] — [Brief Title]

## Ticket Metadata

| Field     | Value           |
| --------- | --------------- |
| Ticket ID | [e.g. S2-FE006] |
| Assignee  | [Name]          |
| Sprint    | Sprint [N]      |
| Branch    | [branch-name]   |

## Audit Checklist Coverage

List which audit.md items this PR addresses:

- [ ] Registration form fields (audit.md:39)
- [ ] Profile display (audit.md:87-98)
- [ ] ...etc

## Changes

### [Component Name]

- **[NEW / MODIFY]** `path/to/component.tsx`
  - Description of changes

## Verification

```bash
bun run lint
bun run format:check
tsc --noEmit
bun run test
```
````

### Manual Smoke Tests

- [ ] Checked scenario [e.g. A1] from general-instructions.md Q3

## Definition of Done (Frontend)

- [ ] Component renders per F1 audit mapping
- [ ] ESLint + Prettier lint/format passes
- [ ] tsc --noEmit passes
- [ ] Vitest tests pass
- [ ] FE cross-review completed
- [ ] Manual smoke test relevant scenario passes

```

### 2.3 Required vs Optional for Frontend

| Required (enforced) | Optional (best practice) |
|---------------------|--------------------------|
| F1: Audit feature mapping complete | Accessibility (a11y) |
| F2: Interaction patterns per spec | Dark mode / glassmorphism |
| F3: Cookie-based auth, no localStorage-only | Micro-animations |
| F4: JPG/PNG/GIF upload, MIME validation | Optimistic updates |
| F5: Project structure (app/ routes, components/ui, components/features) | Error boundaries per route |
| F6: Bun, Dockerfile, verification commands | Loading skeletons (loading.tsx) |
| TDD for components (Vitest) | Image optimization (<Image>) |
| ESLint + Prettier + tsc + vitest green | Code splitting (dynamic imports) |
| Branch naming + conventional commits | Playwright E2E tests |

---

## Part 3: Workflow Enforcement Gap Analysis

### 3.1 pr-create.md

| General-Instructions Requirement | Enforced? | Gap |
|--------------------------------|-----------|-----|
| Branch naming `username/type-detail` | YES — Phase 1 Step 1 | ✅ |
| Conventional Commits | YES — Phase 1 Step 2 | ✅ |
| Cross-reference sprint ticket | YES — Phase 2 Step 1 | ✅ |
| Boundary rules (D5) | YES — Phase 2 Step 2 | ✅ |
| Interface rules (D2) | YES — Phase 2 Step 2 | ✅ |
| Cross-slice comm (D3) | YES — Phase 2 Step 2 | ✅ |
| DB factory (D4) | YES — Phase 2 Step 2 | ✅ |
| TDD requirement | YES — Phase 2 Step 2 | ✅ |
| Surgical changes | YES — Phase 2 Step 2 | ✅ |
| `make ci` for backend | YES — Phase 2 Step 3 | ✅ |
| **Frontend lint/format:check** | **PARTIAL** — says `npm run lint` and `npm run format:check` but should be `bun run` | **LOW** — commands use `npm` not `bun` |
| **Frontend specific checks (F1-F6)** | **NO** | **HIGH** — no verification of frontend audit feature mapping, component structure, or image handling |
| **Frontend DoD checklist** | **NO** | **HIGH** — DoD in PR template is backend-only (D5, D2, etc.) |
| **Manual smoke test traceability** | PARTIAL — mentions Q3 scenarios | **MEDIUM** — no mapping from FE ticket to specific smoke test scenarios |

**Proposed changes to pr-create.md:**
1. Phase 2 Step 3: Change `npm run lint` → `bun run lint`, `npm run format:check` → `bun run format:check`; add `bun run test` and `tsc --noEmit`
2. Phase 2 Step 2: Add frontend-specific checks:
   - Component renders per audit mapping (F1)
   - File upload supports JPG/PNG/GIF (F4)
   - Auth is cookie-based, no localStorage-only (F3)
   - Notification bell present in global nav (F2)
3. PR template DoD: Add frontend-specific items (ESLint + Prettier, tsc, Vitest, FE cross-review, audit checklist coverage)

### 3.2 pr-implement-qrspi.md

| General-Instructions Requirement | Enforced? | Gap |
|--------------------------------|-----------|-----|
| Read general-instructions.md | YES — Stage 2 Step 2 | ✅ |
| Boundary rules (D5) | YES — Stage 4 | ✅ |
| DB factory (D4) | YES — Stage 4 | ✅ |
| TDD (Red-Green-Refactor) | YES — Stage 7 | ✅ |
| Surgical changes | YES — Stage 7 Step 3 | ✅ |
| Branch naming | YES — Stage 6 | ✅ |
| **Frontend TDD pattern** | **NO** | **HIGH** — TDD described only with Go conventions (`commands/*_test.go`, `queries/*_test.go`, `store/*_test.go`). No frontend TDD (Vitest, component tests) |
| **Frontend structure constraints** | **NO** | **HIGH** — Structure stage (Stage 4) only mentions DB factory, SQLite constraints. No Next.js App Router, component hierarchy, or F1-F6 constraints |
| **Frontend verification commands** | **PARTIAL** — Stage 8 mentions `npm run lint` + `format:check` | **MEDIUM** — incomplete frontend verification chain |

**Proposed changes to pr-implement-qrspi.md:**
1. Stage 4: Add frontend structural constraints:
   - App Router routes per F1 mapping
   - Components in `components/ui/` (primitives) and `components/features/` (domain)
   - No client components for static pages (RSC-first)
   - Cookie-based auth (no localStorage)
2. Stage 5: For FE tickets, plan Vitest component tests + Playwright E2E scenarios
3. Stage 7: Add frontend TDD loop:
   - RED: Write failing Vitest test (component render, user interaction, hook output)
   - GREEN: Implement component/hook
   - REFACTOR: Run ESLint + Prettier lint/format
4. Stage 8: Change `npm run` → `bun run`. Add `tsc --noEmit` and `bun run test`

### 3.3 pr-implement.md

| General-Instructions Requirement | Enforced? | Gap |
|--------------------------------|-----------|-----|
| Read general-instructions.md | YES — Phase 1 Step 2 | ✅ |
| Boundary rules (D5) | YES — Phase 2 Step 2 | ✅ |
| TDD | YES — Phase 3 Step 1 | ✅ |
| Surgical scope | YES — Phase 3 Step 3 | ✅ |
| DB factory (D4) | YES — Phase 3 Step 1 | ✅ |
| **Frontend-specific anything** | **NO** | **HIGH** — Entirely backend-oriented. TDD is Go-only. Plan structure is Go-only. |

**Proposed changes to pr-implement.md:**
1. Phase 2: Add frontend plan path:
   - If ticket is `FE-*`: plan routes, components, API client calls, mock data
   - Reference F1 audit mapping for component list
2. Phase 3: Add frontend TDD:
   - Vitest for unit/component tests
   - Playwright for E2E
   - ESLint + Prettier for lint/format
3. Phase 4: Add frontend verification: `bun run lint && bun run format:check && tsc --noEmit && bun run test`

### 3.4 pr-review.md

| General-Instructions Requirement | Enforced? | Gap |
|--------------------------------|-----------|-----|
| Deterministic tools (lint, test) | YES — Phase 1 | ✅ |
| Architecture boundary rules | YES — Phase 2 Agent 3 | ✅ |
| Logic correctness (DB pooling, WAL) | YES — Phase 2 Agent 2 | ✅ |
| TDD / migration integrity | YES — Phase 2 Agent 5 | ✅ |
| **Frontend-specific subagent** | **NO** | **HIGH** — No subagent checks: RSC boundary correctness, client/server component split, cookie-based auth, image upload handling, notification visibility on every page |
| **Frontend tool commands** | **PARTIAL** — `npm run lint` + `format:check` | **LOW** — should be `bun run` |

**Proposed changes to pr-review.md:**
1. Phase 1: Change `npm run` → `bun run`. Add `tsc --noEmit` and `bun run test`
2. Phase 2: Add **Frontend & UI Compliance Agent** (or extend Agent 4):
   - RSC boundary: server components fetch data, client components handle interactivity only
   - Audit feature mapping: every required component from F1 exists
   - Image upload: accepts JPG/PNG/GIF with MIME validation
   - Notifications: bell visible in global nav on every page
   - Auth: session-based (cookies), no localStorage-only auth
   - Accessibility: ARIA on interactive elements (optional/best practice)

### 3.5 sn-code-audit.md

| General-Instructions Requirement | Enforced? | Gap |
|--------------------------------|-----------|-----|
| Allowed packages (Go) | YES — Phase 1 Step 4 | ✅ |
| All audit.md requirements | YES — Phase 2 Layer D | ✅ (comprehensive) |
| Docker 2 containers | YES — Layer F | ✅ |
| **Frontend framework audit** | **PARTIAL** — Layer A Step 5 mentions "verify frontend directory is well-organized" | **HIGH** — No frontend-specific code audit. No checks for: correct form fields, session persistence, notification visibility, image upload types, emoji support, privacy UI, confirmation popups |
| **Next.js best practices** | **NO** | **MEDIUM** — No RSC audit, no client/server component boundary check, no bundle size audit |

**Proposed changes to sn-code-audit.md:**
1. Add **Layer G — Frontend Compliance Audit**:
   - **G1. Registration form completeness**: Verify all 8 fields (5 required + 3 optional) in register page/component
   - **G2. Session persistence**: Verify auth is cookie-based (httpOnly session cookie). No localStorage-only auth. Confirm login survives page refresh.
   - **G3. Notification accessibility**: Notification UI present in global layout (visible on every page). Notifications visually distinct from chat messages.
   - **G4. Image handling**: Post and comment forms accept .jpg, .png, .gif. MIME validation on upload.
   - **G5. Privacy UI**: Profile lock screen for non-followers of private profiles. Privacy toggle with confirmation. Post visibility selector (3 levels).
   - **G6. Follow/chat gating**: Chat denied with message if no follow relationship. Follow button shows request flow for private profiles.
   - **G7. Group/event forms**: Event form has title, description, day/time, ≥2 options.
   - **G8. Emoji support**: Chat input accepts Unicode emoji characters.
   - **G9. RSC boundaries**: Data fetching in server components. Client components limited to interactivity.
   - **G10. Confirmation popups** (bonus): Present for unfollow, privacy toggle
2. Add to Bonus Feature Inventory:
   - Confirmation popups (audit.md bonus items)
   - OAuth UI
   - DB seeding UI

### 3.6 graphify.md

| General-Instructions Requirement | Enforced? | Gap |
|--------------------------------|-----------|-----|
| N/A — graphify is a knowledge graph tool, not a PR/audit workflow | N/A | No gaps — graphify.md delegates to the graphify skill. It's a utility, not an enforcement mechanism. |

**No changes needed.**

---

## Summary: Critical Gaps

| Workflow | Highest-Priority Gap | Severity |
|----------|---------------------|----------|
| pr-create.md | No frontend DoD checklist, uses `npm` not `bun` | HIGH |
| pr-implement-qrspi.md | No frontend TDD pattern, no FE structural constraints | HIGH |
| pr-implement.md | Entirely backend-oriented, zero FE support | HIGH |
| pr-review.md | No frontend compliance subagent | HIGH |
| sn-code-audit.md | No Layer G for frontend spec compliance | HIGH |
| graphify.md | N/A — utility workflow | NONE |

### Priority Actions

1. **general-instructions.md**: Add sections F1–F6 (frontend feature mapping, interaction patterns, state management, file handling, project structure, build/deploy)
2. **general-instructions.md**: Add Onboarding section for FE devs (pick ticket → read context → branch → TDD → PR)
3. **ALL PR workflows**: Replace `npm run` with `bun run`. Add `tsc --noEmit` and `bun run test`
4. **pr-create.md**: Add FE-specific DoD items to PR template
5. **pr-implement-qrspi.md**: Add FE TDD loop + FE structural constraints
6. **pr-implement.md**: Add FE plan/implement/validate path
7. **pr-review.md**: Add Frontend & UI Compliance subagent
8. **sn-code-audit.md**: Add Layer G (Frontend Compliance Audit) with 10 sub-checks
```
