# Frontend ↔ Backend Synchronization Mismatches

> Audit of every request/response mismatch between `frontend-next/src/lib/api.ts`
> and the backend handlers in `internal/*/transport/`.
>
> Branch: `geoikonomou/front-backend-synchronization`
>
> Status: only **unresolved** mismatches are listed below, renumbered after pruning.
> Resolved so far: topic/comment delete & vote query params, event list creator,
> group routes, chat routes, register `dateOfBirth` format, login (full user + session cookie),
> `/me` (full user via `userResponse`), topic response (string id, nested `user`,
> `imageUrl`, `commentsCount`), comment response (string id, `postId`, nested `user`,
> `imageUrl`), group response (`membershipStatus` in list + detail),
> chat users (`Chat[]` conversation envelopes, camelCase, `avatarUrl`),
> chat history (`ChatMessage[]`: string id, camelCase, nested `sender`, `type: "private"`),
> follow requests (`FollowRequest`: synthesized `id`, nested `requester`/`target`, `status: "pending"`,
> `createdAt`),
> user search (`GET /users` — SQL LIKE filter + pagination, flat `PaginatedResponse<User>`
> with full `User` fields).

---

## 1. SEARCH GROUPS — no-op

| Side         | Behavior                 |
| ------------ | ------------------------ |
| **Backend**  | Ignores `query` param    |
| **Frontend** | Expects filtered results |

| Option | Side     | Change              |
| ------ | -------- | ------------------- |
| **A**  | Backend  | Add SQL LIKE filter |
| **B**  | Frontend | Client-side filter  |

---

## 2. PAGINATION ENVELOPE — different structure

| Side                           | Shape                                                                        |
| ------------------------------ | ---------------------------------------------------------------------------- |
| **Backend**                    | `{ info: { totalRecords, currentPage, pageSize, totalPages }, data: [...] }` |
| **Frontend PaginatedResponse** | `{ data, page, pageSize, totalCount, totalPages }`                           |

**Issues:**

- `info.totalRecords` → `totalCount`
- `info.currentPage` → `page`
- `info` wrapper not expected by frontend

| Option | Side     | Change                                              |
| ------ | -------- | --------------------------------------------------- |
| **A**  | Backend  | Match frontend's `PaginatedResponse` shape directly |
| **B**  | Frontend | Adapter unwraps `info` → `PaginatedResponse` fields |

---

## Summary: Recommended Strategy

**Backend adapts to frontend** — backend changes where data is genuinely missing; frontend only handles type conversions via a thin adapter layer.

| Category       | Changes                                                                                                   |
| -------------- | --------------------------------------------------------------------------------------------------------- |
| **Backend**    | group search                                                                                              |
| **Frontend**   | `api.ts` — login already unwraps `res.user`; new `transformers.ts` for field mappings / pagination unwrap |
| **No changes** | `types.ts` — backend will match frontend types                                                            |
