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
> with full `User` fields),
> group search (`GET /groups` — SQL LIKE filter on title/description + flat `PaginatedResponse<Group>`),
> pagination envelope (all paginated endpoints now return the flat `PaginatedResponse`
> shape: `{ data, page, pageSize, totalCount, totalPages }` — `/users`, `/groups`,
> group members/posts/comments, topic feeds).

---

## Summary: Recommended Strategy

**Backend adapts to frontend** — backend changes where data is genuinely missing; frontend only handles type conversions via a thin adapter layer.

| Category       | Changes                                                                                                   |
| -------------- | --------------------------------------------------------------------------------------------------------- |
| **Backend**    | pagination envelope aligned to frontend `PaginatedResponse` (all endpoints flat)                          |
| **Frontend**   | `api.ts` — login already unwraps `res.user`; new `transformers.ts` for field mappings / pagination unwrap |
| **No changes** | `types.ts` — backend will match frontend types                                                            |
