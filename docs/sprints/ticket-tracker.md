# Social Network Refactoring — Ticket Tracker Checklist

For project details, refactoring strategies, TDD methodologies, QA plans, and appendices, see [general-instructions.md](general-instructions.md).

For step-by-step ticket instructions, see individual sprint files:

- [Sprint 0: Foundation](sprint-0.md)
- [Sprint 1: Platform & Core Infrastructure](sprint-1.md)
- [Sprint 2: User & Topic Features](sprint-2.md)
- [Sprint 3: Follow, Comment & Notification](sprint-3.md)
- [Sprint 4: Group & Event Features](sprint-4.md)
- [Sprint 5: Chat & OAuth](sprint-5.md)
- [Sprint 6: Integration, Cleanup & Polish](sprint-6.md)

---

## Sprint 0: Foundation (Week 1–2)

### BE-A (Backend A)

- [x] **S0-BE-01:** Go Project Scaffold
- [x] **S0-BE-02:** Bug Fixes (B1.1, B1.2, B1.5)

### BE-B (Backend B)

- [x] **S0-BE-03:** Makefile + CI Pipeline
- [x] **S0-BE-04:** Bug Fixes (B1.3, B1.4, B1.6, B1.7, B1.8)

### FE-A (Frontend A)

- [ ] **S0-FE-01:** Next.js Scaffold + Tooling (Frontend Setup)

### FE-B (Frontend B)

- [ ] **S0-FE-02:** shadcn/ui Components + Layout

### SD-QA (System Design/QA)

- [x] **S0-SD-01:** golangci-lint Config
- [ ] **S0-SD-02:** Docker Compose Development Environment
- [ ] **S0-SD-03:** CI Pipeline, Go Gates & Pre-commit Hooks
- [x] **S0-SD-04:** Dev Environment Docs

---

## Sprint 1: Platform & Core Infrastructure (Week 3–4)

### BE-A (Backend A)

- [x] **S1-BE-05:** Platform: DB Factory
- [x] **S1-BE-06:** Custom Migration System
- [x] **S1-BE-07:** Core: Session Management
- [x] **S1-BE-08:** Core: Middlewares
- [x] **S1-BE-09:** Shared: Image Type Verification Utility

### BE-B (Backend B)

- [x] **S1-BE-10:** Platform: Event Bus
- [x] **S1-BE-11:** Platform: Cache
- [x] **S1-BE-12:** Core: Realtime WebSocket Hub
- [ ] **S1-BE-13:** Core: HTTP Server Bootstrap

### FE-A (Frontend A)

- [ ] **S1-FE-03:** Auth Pages (Login & Registration UI)
- [ ] **S1-FE-04:** API Client Wrapper

### FE-B (Frontend B)

- [ ] **S1-FE-05:** Nav Layout Shell

### SD-QA (System Design/QA)

- [ ] **S1-SD-05:** Platform: Database Seeding (000009) (Gap Fix)
- [ ] **S1-SD-06:** API Mocking Service

---

## Sprint 2: User & Topic Features (Week 5–6)

### Joint BE-A & BE-B

- [ ] **S2-BE-14:** Wire User & Topic bootstrap routes

### BE-A (Backend A)

- [ ] **S2-BE-15:** User: Entity & Repository Interface
- [x] **S2-BE-16:** User: SQLite Store
- [x] **S2-BE-17:** User: Register Command
- [x] **S2-BE-18:** User: Login Command
- [x] **S2-BE-19:** User: Logout Command
- [x] **S2-BE-20:** User: Update Profile Command
- [x] **S2-BE-21:** User: Toggle Privacy Command
- [x] **S2-BE-22:** User: Get Profile Query
- [x] **S2-BE-23:** User: Get Activity Query
- [x] **S2-BE-24:** User: List Users Query
- [ ] **S2-BE-25:** User: HTTP Transport Routing

### BE-B (Backend B)

- [ ] **S2-BE-26:** Topic: Entity & Repository Interface
- [ ] **S2-BE-27:** Topic: SQLite Store
- [ ] **S2-BE-28:** Topic: Create Topic Command
- [ ] **S2-BE-29:** Topic: Cast Vote Command
- [ ] **S2-BE-30:** Topic: Get Feed Query
- [ ] **S2-BE-31:** Topic: Get User Topics Query
- [ ] **S2-BE-32:** Topic: Get Topic Query
- [ ] **S2-BE-33:** Topic: Get Votes Query
- [ ] **S2-BE-34:** Topic: HTTP Transport Routing

### FE-A (Frontend A)

- [ ] **S2-FE-06:** Registration Form
- [ ] **S2-FE-07:** Login Page
- [ ] **S2-FE-08:** Profile Page
- [ ] **S2-FE-09:** Privacy Toggle with Confirmation Popup (Bonus)

### FE-B (Frontend B)

- [ ] **S2-FE-10:** Home Feed Page
- [ ] **S2-FE-11:** Post Creation Form
- [ ] **S2-FE-12:** Post Card Component

### SD-QA (System Design/QA)

- [ ] **S2-SD-07:** User Slice: Migration Verification Contract Tests
- [ ] **S2-SD-08:** Topic Slice: Migration Verification Contract Tests
- [ ] **S2-SD-09:** Platform: User & Topic Migrations (000002 & 000003)
- [ ] **S2-SD-10:** E2E: User Signup to Feed Journey

---

## Sprint 3: Follow, Comment & Notification (Week 7–8)

### BE-A (Backend A)

- [ ] **S3-BE-35:** Follow: Entities & Repository Interface
- [x] **S3-BE-36:** Follow: SQLite Store
- [x] **S3-BE-37:** Follow: Follow User Command
- [ ] **S3-BE-38:** Follow: Unfollow User Command
- [x] **S3-BE-39:** Follow: Accept Request Command
- [x] **S3-BE-40:** Follow: Decline Request Command
- [x] **S3-BE-41:** Follow: Get Followers Query
- [x] **S3-BE-42:** Follow: Get Following Query
- [x] **S3-BE-43:** Follow: Get Pending Requests Query
- [x] **S3-BE-44:** Follow: Are Connected Query **P0**
- [x] **S3-BE-45:** Follow: HTTP Transport Routing

### BE-B (Backend B)

- [x] **S3-BE-46:** Comment: Entity & Repository Interface
- [x] **S3-BE-47:** Comment: SQLite Store
- [x] **S3-BE-48:** Comment: Create Comment Command
- [ ] **S3-BE-49:** Comment: Get Comments Query
- [ ] **S3-BE-50:** Comment: HTTP Transport Routing
- [ ] **S3-BE-51:** Comment: Cast Vote Command & Queries (Gap Fix) **P1**
- [ ] **S3-BE-52:** Notification: Entity & Repository Interface
- [ ] **S3-BE-53:** Notification: SQLite Store
- [ ] **S3-BE-54:** Notification: Event Bus Consumer
- [ ] **S3-BE-55:** Notification: Mark Read Command
- [ ] **S3-BE-56:** Notification: List Notifications Query
- [ ] **S3-BE-57:** Notification: HTTP Transport Routing
- [ ] **S3-BE-58:** Notification: Old Schema→New Schema Migration

### Joint BE-A & BE-B

- [ ] **S3-BE-59:** Wire Follow, Comment & Notification bootstrap routes

### FE-A (Frontend A)

- [ ] **S3-FE-13:** Follow Button with Popup
- [ ] **S3-FE-14:** Followers List Pages
- [ ] **S3-FE-15:** Follow Request Notifications

### FE-B (Frontend B)

- [ ] **S3-FE-16:** Comment Section Components
- [ ] **S3-FE-17:** Notifications Panel
- [ ] **S3-FE-18:** Notifications Live Stream
- [ ] **S3-FE-19:** Comment Card Vote Buttons (Gap Fix) **P1**

### SD-QA (System Design/QA)

- [ ] **S3-SD-11:** Follow: Event Publishing Verification
- [ ] **S3-SD-12:** Comment Slice: Contract Tests
- [ ] **S3-SD-13:** Platform: Follow System & Notifications Migrations (000004 & 000005)
- [ ] **S3-SD-14:** E2E: Relationships Notifications Flow
- [ ] **S3-SD-15:** E2E: Posts Comments Notification Flow

---

## Sprint 4: Group & Event Features (Week 9–10)

### Joint BE-A & BE-B

- [ ] **S4-BE-60:** Wire Group & Event bootstrap routes

### BE-A (Backend A)

- [ ] **S4-BE-61:** Group: Entities & Repository Interface
- [ ] **S4-BE-62:** Group: SQLite Store
- [ ] **S4-BE-63:** Group: Create Group Command
- [ ] **S4-BE-64:** Group: Invite Member Command
- [ ] **S4-BE-65:** Group: Respond Invite Command
- [ ] **S4-BE-66:** Group: Request Join Command
- [ ] **S4-BE-67:** Group: Respond Join Command
- [ ] **S4-BE-68:** Group: Create Post Command
- [ ] **S4-BE-69:** Group: Send Group Message Command
- [ ] **S4-BE-70:** Group: List Groups Query
- [ ] **S4-BE-71:** Group: Get Group Detail Query
- [ ] **S4-BE-72:** Group: Get Group Feed Query
- [ ] **S4-BE-73:** Group: Get Group Chat History Query
- [ ] **S4-BE-74:** Group: HTTP Transport Routing
- [ ] **S4-BE-75:** Group: WS Transport Routing
- [ ] **S4-BE-76:** Group: Post Comments (Gap Fix) **P1**

### BE-B (Backend B)

- [ ] **S4-BE-77:** Event: Entities & Repository Interface
- [ ] **S4-BE-78:** Event: SQLite Store
- [ ] **S4-BE-79:** Event: Create Event Command
- [ ] **S4-BE-80:** Event: RSVP Command
- [ ] **S4-BE-81:** Event: List Group Events Query
- [ ] **S4-BE-82:** Event: HTTP Transport Routing

### FE-A (Frontend A)

- [ ] **S4-FE-20:** Groups Directory Page
- [ ] **S4-FE-21:** Group Profile Page
- [ ] **S4-FE-22:** Group Posts Feed
- [ ] **S4-FE-23:** Group Chat Workspace

### FE-B (Frontend B)

- [ ] **S4-FE-24:** Event Creation Dialog
- [ ] **S4-FE-25:** Events List Component
- [ ] **S4-FE-26:** RSVP Switch Actions
- [ ] **S4-FE-27:** Group: Comment Components (Gap Fix) **P1**

### SD-QA (System Design/QA)

- [ ] **S4-SD-16:** Platform: Group & Event Migrations (000006 & 000007)
- [ ] **S4-SD-17:** E2E: Complete Groups Workspace Journey

---

## Sprint 5: Chat & OAuth (Week 11–12)

### Joint BE-A & BE-B

- [ ] **S5-BE-83:** Wire Chat & OAuth bootstrap routes

### BE-A (Backend A)

- [ ] **S5-BE-84:** Chat: Entity & Repository Interface
- [ ] **S5-BE-85:** Chat: SQLite Store
- [ ] **S5-BE-86:** Chat: Send Private Message Command
- [ ] **S5-BE-87:** Chat: Get Chat History Query
- [ ] **S5-BE-88:** Chat: List Conversations Query
- [ ] **S5-BE-89:** Chat: HTTP Transport Routing
- [ ] **S5-BE-90:** Chat: WS Transport Routing
- [ ] **S5-BE-91:** Platform: Chat Migrations (000008) (Gap Fix) **P0**

### BE-B (Backend B)

- [ ] **S5-BE-92:** OAuth: Entity & Repository Interface
- [ ] **S5-BE-93:** OAuth: SQLite Store
- [ ] **S5-BE-94:** OAuth: Initiate Login Command
- [ ] **S5-BE-95:** OAuth: Callback Processor Command
- [ ] **S5-BE-96:** OAuth: HTTP Transport Routing
- [ ] **S5-BE-97:** OAuth Client: GitHub Implementation
- [ ] **S5-BE-98:** OAuth Client: Google Implementation
- [ ] **S5-BE-99:** Shared: Refactor OAuth Packages

### FE-A (Frontend A)

- [ ] **S5-FE-28:** Chat Feed View
- [ ] **S5-FE-29:** Realtime Live Sockets Hook
- [ ] **S5-FE-30:** Chat Message Bubble Component

### FE-B (Frontend B)

- [ ] **S5-FE-31:** GitHub OAuth Button Integration
- [ ] **S5-FE-32:** Google OAuth Button Integration

### SD-QA (System Design/QA)

- [ ] **S5-SD-18:** Chat Slice: Contract Tests
- [ ] **S5-SD-19:** OAuth Slice: Contract Tests
- [ ] **S5-SD-20:** E2E: Messaging Real-Time Delivery Journey
- [ ] **S5-SD-21:** E2E: GitHub OAuth Sign In

---

## Sprint 6: Integration, Cleanup & Polish (Week 13–14)

### BE-A (Backend A)

- [ ] **S6-BE-100:** Clean Legacy Slices: Domain
- [ ] **S6-BE-101:** Clean Legacy Slices: App

### BE-B (Backend B)

- [ ] **S6-BE-102:** Clean Legacy Slices: Infra

### Joint BE-A & BE-B

- [ ] **S6-BE-103:** Bootstrap Wiring

### FE-A (Frontend A)

- [ ] **S6-FE-33:** Responsive Design Check

### FE-B (Frontend B)

- [ ] **S6-FE-34:** Components Error Boundaries & Loading States

### Joint FE-A & FE-B

- [ ] **S6-FE-35:** Production Build Validation

### SD-QA (System Design/QA)

- [ ] **S6-SD-22:** Full Integration Test Suite
- [ ] **S6-SD-23:** Performance Benchmarks
- [ ] **S6-SD-24:** Vertical Slice Boundary Checks
- [ ] **S6-SD-25:** Audit.md Automation Test Suite (Gap Fix)
- [ ] **S6-SD-26:** Production Docker Setup
- [ ] **S6-SD-27:** Health Check Endpoints
- [ ] **S6-SD-28:** Graceful Server Shutdowns
- [ ] **S6-SD-29:** Twelve-Factor Configurations Mappings
- [ ] **S6-SD-30:** Docker Smoke Verification Script
- [ ] **S6-SD-31:** Full E2E Test Suite
- [ ] **S6-SD-32:** Accessibility (a11y) Audit
- [ ] **S6-SD-33:** Frontend Performance Audits
- [ ] **S6-SD-34:** E2E Audit.md Playwright Suite (Gap Fix)

---
