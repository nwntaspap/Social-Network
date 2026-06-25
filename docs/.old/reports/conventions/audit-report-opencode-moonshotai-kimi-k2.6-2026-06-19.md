# Audit Report — OpenCode (moonshotai/kimi-k2.6)

**Timestamp**: 2026-06-19

## Scope

Audit `conventions.md` for coverage of critical info from:

- `docs/sprints/general-instructions.md`
- `docs/architecture/target-architecture-with-phases.md`

No files modified.

---

## Missing from `conventions.md`

### From `general-instructions.md`

1. **Team structure & sprint cadence** (lines 10–17): 5-dev team (`SD-QA`, `BE-A`, `BE-B`, `FE-A`, `FE-B`), 1-week sprint, ~7 sprints total.
2. **Progressive reading chain** (lines 21–53): Stage 1–6 navigation map. `conventions.md` references `general-instructions.md` but skips the explicit stage flow.
3. **Full TDD cycle detail** (lines 78–98): RED/GREEN/REFACTOR steps exist in abbreviated form; missing the explicit per-use-case file conventions (`commands/<use_case>_test.go`, etc.) and the contract test deletion rule.
4. **Migration safety rule** (line 126): "Never drop column in same migration; first add new column, populate, then drop old in NEXT migration."
5. **Feature toggle examples** (lines 330–335, also in `target-architecture-with-phases.md`): `config.Features.Follow` toggle snippet.
6. **Frontend feature-to-audit mapping** (lines 159–175): Full component/page mapping for all audit checklist items (RegisterForm, ProfileCard, PostForm, GroupChatWindow, etc.).
7. **Frontend interaction patterns** (lines 177–182): Confirmation dialogs, follow-gate feedback string, emoji support, WebSocket reconnection.
8. **Frontend state management** (lines 183–188): `HttpOnly`/`Secure`/`SameSite=Lax` via cookies, session isolation, RSC recommendation.
9. **Bug tracking table** (lines 219–231): B1.1–B1.8 with file locations and assignees.
10. **Verification gates commands** (lines 242–261): `make ci`, `make be-ci`, `make fe-ci`, boundary grep, standalone `go vet`/`go build`/`go test`/`golangci-lint`/`govulncheck`.
11. **Testing pyramid** (lines 300–309): Visual pyramid with target counts. `conventions.md` has counts but no visual.
12. **Risk mitigation table** (lines 343–353): 5 risks with mitigations.

### From `target-architecture-with-phases.md`

13. **Architecture pain-point metrics** (lines 10–21): 32 handler dirs, 38 aliased imports, 182 files, 3:1 overhead ratio.
14. **System overview diagram** (lines 27–44): Browser → Go Backend → SQLite/Redis/RabbitMQ.
15. **Component descriptions** (lines 49–59): Descriptions of frontend, backend, and infrastructure services, including Redis and RabbitMQ as optional plug-ins.
16. **Technology & Tooling quick-reference** (lines 80–127): Full toolchain matrix for backend, frontend, and infrastructure. `conventions.md` has snippets but not the structured tables.
17. **Event bus start implementation** (lines 222, 499): Starts as in-process Go channels, later swappable for RabbitMQ.
18. **Feature directory tree** (lines 301–461): Detailed per-feature file listing (very long, may not belong in `conventions.md`).
19. **Phase-by-phase execution plan** (lines 465–699): Phases 1–10—critical for understanding migration order. Only bug fixes and Strangler Fig summary are in `conventions.md`.
20. **Per-feature migration steps** (lines 641–656): 9-step migration checklist for each feature.
21. **Special merge notes** (lines 658–691): `user/` absorbs `activity/`, `topic/` absorbs `category/` and `vote/`, `chat/` gets `transport/ws.go`, etc.
22. **Future phases (Optional)** (lines 765–821): PostgreSQL, Redis, RabbitMQ, Kafka, Microservice promotion, Kubernetes, CQRS scaling.
23. **Verification checklist commands** (lines 903–934): Backend/frontend manual command lists and boundary grep.
24. **Detailed manual test scenarios** (lines 938–964): A1–D3 with steps and expected results.

---

## Recommendations

1. **Add a "Progressive Disclosure" section** referencing the stage map.
2. **Expand "TDD & Idiomatic Go"** with per-use-case file naming and contract test deletion rule.
3. **Add "Frontend Feature Map"** (F1 from `general-instructions.md`) or link to it.
4. **Add "Bug Registry"** (B1.1–B1.8) or reference `target-architecture-with-phases.md` Phase 1.
5. **Add "Phase Roadmap"** summarizing Phases 1–6 at minimum.
6. **Add "Per-Feature Migration Steps"** (9-step checklist).
7. **Add "Optional Future Phases"** as a brief reference (PostgreSQL, Redis, RabbitMQ, Microservices, K8s, CQRS scaling).
8. **Expand "Verification"** with the manual command lists and boundary grep command.
9. **Add "Risk Mitigation" table** or reference it.
10. **Keep `conventions.md` bounded**—it may be better to link to the source docs for the full phase/roadmap detail rather than duplicating everything. Decision needed on whether `conventions.md` is a "living conventions doc" or also an "index to all project docs".
