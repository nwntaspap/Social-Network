# Frontend ↔ Backend Synchronization Mismatches

> Audit of every request/response mismatch between `frontend-next/src/lib/api.ts`
> and the backend handlers in `internal/*/transport/`.
>
> Branch: `geoikonomou/front-backend-synchronization`
>
> Status: only **unresolved** mismatches are listed below, renumbered after pruning.
> Resolved so far: topic/comment delete & vote query params, event list creator,
> group routes, chat routes, register `dateOfBirth` format, login (full user + session cookie),
> `/me` (full user via `userResponse`), `GET /user/profile` (full user via `userResponse`
>
> - `followersCount`/`followingCount`), topic response (string id, nested `user`,
>   `imageUrl`, `commentsCount`), comment response (string id, `postId`, nested `user`,
>   `imageUrl`), group response (`membershipStatus` in list + detail),
>   comment votes (`GET /comments/topic/votes` — comments include `upvoteCount`,
>   `downvoteCount`, `voteScore`, `userVote`; `POST /comments/vote?id={id}` upsert-toggles),
>   group creation (`POST /groups` wired in `CreateGroupForm`, list refreshes via `refreshKey`),
>   chat users (`Chat[]` conversation envelopes, camelCase, `avatarUrl`),
>   chat history (`ChatMessage[]`: string id, camelCase, nested `sender`, `type: "private"`),
>   follow requests (`FollowRequest`: synthesized `id`, nested `requester`/`target`, `status: "pending"`,
>   `createdAt`),
>   user search (`GET /users` — SQL LIKE filter + pagination, flat `PaginatedResponse<User>`
>   with full `User` fields),
>   group search (`GET /groups` — SQL LIKE filter on title/description + flat `PaginatedResponse<Group>`),
>   pagination envelope (all paginated endpoints now return the flat `PaginatedResponse`
>   shape: `{ data, page, pageSize, totalCount, totalPages }` — `/users`, `/groups`,
>   group members/posts/comments, topic feeds),
>   group post votes (`POST /groups/posts/{postId}/vote` with `{ reactionType: 1 | -1 }`,
>   toggle semantics; group feed/posts include `dislikesCount`, `userVote`, real `likesCount`/`isLiked`),
>   group post comments (previously dead backend code — now routed via
>   `GET`/`POST /groups/posts/{postId}/comments`, paginated, response includes `imageUrl`),
>   comment creation (`POST /comments/create` now multipart with optional `image`; same for
>   `POST /groups/posts/{postId}/comments`; image files are written to `frontend/static/images/uploads/`
>   and served via `imageUrl`),
>   group invitation accept/decline (previously dead end-to-end: `RespondInvite` existed but was
>   unregistered and the invitee had no UI. Now routed via `POST /groups/{groupId}/invite/respond`
>   with `{ action: "accept" | "decline" }`, plus `GET /groups/invitations/pending` (auth) returning
>   pending invitations enriched with a group brief; frontend `GroupInvitations` component lists them
>   with Accept/Decline buttons).

---

## Summary: Recommended Strategy

**Backend adapts to frontend** — backend changes where data is genuinely missing; frontend only handles type conversions via a thin adapter layer.

| Category       | Changes                                                                                                   |
| -------------- | --------------------------------------------------------------------------------------------------------- |
| **Backend**    | pagination envelope aligned to frontend `PaginatedResponse` (all endpoints flat)                          |
| **Frontend**   | `api.ts` — login already unwraps `res.user`; new `transformers.ts` for field mappings / pagination unwrap |
| **No changes** | `types.ts` — backend will match frontend types                                                            |
