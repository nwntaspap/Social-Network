# Frontend ↔ Backend Synchronization Mismatches

> Audit of every request/response mismatch between `frontend-next/src/lib/api.ts`
> and the backend handlers in `internal/*/transport/`.
>
> Branch: `geoikonomou/front-backend-synchronization`
>
> Status: only **unresolved** mismatches are listed below, renumbered after pruning.
> Resolved so far: topic/comment delete & vote query params, event list creator,
> group routes, chat routes, register `dateOfBirth` format, login (full user + session cookie),
> `/me` (full user via `userResponse`), `GET /user/profile` (full user via `userResponse`>
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
>   with Accept/Decline buttons),
>   group join-request review (`GET /groups/{groupId}/requests/pending` was a stub returning `[]`
>   — now resolves real pending requests, enriched with requester + group brief, gated to
>   creator/admin (403 otherwise); `GroupSidebar` "Pending Requests" shows them with Accept/Decline),
>   group edit (`PUT /groups/{groupId}` — `EditGroupForm` dropdown in `GroupHeader` calls
>   `updateGroup`, then refreshes the group; creator/admin only, non-admin rejected),
>   group delete (`DELETE /groups/{groupId}` — `DeleteGroup` button now wired with a two-step inline
>   confirmation, deletes cascadingly and redirects to `/groups`; creator only).
>   group member count (`GetGroupResolver.Resolve` now populates `membersCount` via
>   `CountMembers` — the group page showed 0/blank while the list showed the real count),
>   group creator display (`GroupHeader` meta line and `GroupCard` now render "Created by {name}";
>   backend already returned `creator`), event creation (`CreateEventForm` now sends
>   `options: ['Going', 'Not going']` — the backend required ≥2 options so every creation 400'd;
>   real `ApiError.message` is surfaced instead of a generic string),
>   event attendee lists (`GET /events/{eventId}/rsvps` returns per-option users;
>   `EventsTab` shows option tallies, "View attendees" lazily fetches and renders
>   profile-linked attendee lists per option),
>   event edit (`PUT /groups/{groupId}/events/{eventId}` — creator-only `EditEventForm`
>   dropdown on each event card, enforced backend-side via `eventGroupRoleChecker`;
>   backend `UpdateEvent` command validates fields and group ownership),
>   group invites (`invite_member.go` no longer requires creator/admin — any member can invite),
>   join-request review (`get_pending_join_requests.go` + `respond_join.go` now gated to the
>   group creator only, 403 otherwise; admins rejected),
>   invite accept by non-creator → pending join request (`POST /groups/{groupId}/invite/respond`
>   now returns `{ "status": "ok" }` for direct accept/decline and `{ "status": "pending" }` when
>   the invitee's acceptance becomes a join request awaiting the creator's approval; only a
>   creator's invite accepts directly — `GroupInvitations` shows a pending notice to the invitee),
>   seed event RSVP options (the seeded "Community Hackathon 2025" event now exposes only
>   `Going`/`Not going` options with remapped RSVPs instead of three slots),
>   group post comments toggle (`PostCard` comments button toggles inline
>   `GET`/`POST /groups/posts/{postId}/comments` on the group feed; the separate
>   "Show comments/Hide comments" toggle was removed),
>   groups live search (`GroupSearchBar` is now a plain input that filters the groups grid via
>   the existing `GET /groups?query=` SQL LIKE filter, replacing the autocomplete dropdown).

---

## Summary: Recommended Strategy

**Backend adapts to frontend** — backend changes where data is genuinely missing; frontend only handles type conversions via a thin adapter layer.

| Category       | Changes                                                                                                   |
| -------------- | --------------------------------------------------------------------------------------------------------- |
| **Backend**    | pagination envelope aligned to frontend `PaginatedResponse` (all endpoints flat)                          |
| **Frontend**   | `api.ts` — login already unwraps `res.user`; new `transformers.ts` for field mappings / pagination unwrap |
| **No changes** | `types.ts` — backend will match frontend types                                                            |
