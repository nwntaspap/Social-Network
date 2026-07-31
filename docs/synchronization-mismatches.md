# Frontend ↔ Backend Synchronization Mismatches

> Audit of every request/response mismatch between `frontend-next/src/lib/api.ts`
> and the backend handlers in `internal/*/transport/`.
>
> Branch: `geoikonomou/front-backend-synchronization`
>
> Status: only **unresolved** mismatches are listed below, renumbered after pruning.
> Resolved so far: topic/comment delete & vote query params, event list creator,
> group routes, chat routes, register `dateOfBirth` format, login (full user + session cookie),
> `/me` (full user via `userResponse`).

---

## 1. TOPIC RESPONSE — int IDs, missing user object

| Side                      | Shape                                                                                             |
| ------------------------- | ------------------------------------------------------------------------------------------------- |
| **Backend TopicResponse** | `{ id (int), userId, imagePath, ownerUsername, privacy, likesCount, isLiked, ... }`               |
| **Frontend Post type**    | `{ id (string), userId, user: User, imageUrl, privacy, commentsCount, likesCount, isLiked, ... }` |

**Issues:**

- `id` is `int` → should be `string`
- `imagePath` → should be `imageUrl`
- `ownerUsername` → should be `user: User` (nested object)
- Missing: `commentsCount`, `user`

| Option | Side     | Change                                                                    |
| ------ | -------- | ------------------------------------------------------------------------- |
| **A**  | Backend  | Add `user` object, convert `id` to string, rename fields                  |
| **B**  | Frontend | Adapter maps `ownerUsername`→`user`, `imagePath`→`imageUrl`, `String(id)` |

---

## 2. COMMENT RESPONSE — int IDs, wrong field names

| Side                        | Shape                                                                |
| --------------------------- | -------------------------------------------------------------------- |
| **Backend CommentResponse** | `{ id (int), userId, topicId (int), content, ... }`                  |
| **Frontend Comment type**   | `{ id (string), userId, postId (string), user: User, content, ... }` |

**Issues:**

- `id` is `int` → should be `string`
- `topicId` → should be `postId`
- Missing: `user: User`

| Option | Side     | Change                                                              |
| ------ | -------- | ------------------------------------------------------------------- |
| **A**  | Backend  | Add `user` object, convert IDs to string, rename `topicId`→`postId` |
| **B**  | Frontend | Adapter transforms field names and types                            |

---

## 3. GROUP RESPONSE — missing `membershipStatus`

| Side                      | Shape                                                                                |
| ------------------------- | ------------------------------------------------------------------------------------ |
| **Backend GroupResponse** | `{ id, title, description, creatorId, creator, membersCount, createdAt, updatedAt }` |
| **Frontend Group type**   | `{ ..., membershipStatus?: 'none' \| 'pending' \| 'member' }`                        |

**Issue:** `membershipStatus` only exists in `GroupDetailResponse`, not in list response.

| Option | Side     | Change                                     |
| ------ | -------- | ------------------------------------------ |
| **A**  | Backend  | Add membership check query to `ListGroups` |
| **B**  | Frontend | Default to `'none'` when missing           |

---

## 4. CHAT USER — snake_case, wrong field names

| Side                  | Shape                                                                      |
| --------------------- | -------------------------------------------------------------------------- |
| **Backend ChatUser**  | `{ user_id, nickname, chat_id, unread_count, is_online, last_message_at }` |
| **Frontend ChatUser** | `{ id, username, avatarUrl, isOnline, lastMessageAt }`                     |

**Issues:**

- snake_case → camelCase
- `user_id` → `id`
- `nickname` → `username`
- Missing: `avatarUrl`

| Option | Side     | Change                                               |
| ------ | -------- | ---------------------------------------------------- |
| **A**  | Backend  | Convert to camelCase, rename fields, add `avatarUrl` |
| **B**  | Frontend | Adapter transforms shape                             |

---

## 5. CHAT MESSAGE — snake_case, int ID, missing fields

| Side                     | Shape                                                               |
| ------------------------ | ------------------------------------------------------------------- |
| **Backend Message**      | `{ id (int), chat_id, sender_id, content, created_at }`             |
| **Frontend ChatMessage** | `{ id (string), senderId, sender: User, content, type, createdAt }` |

**Issues:**

- snake_case → camelCase
- `id` is `int` → should be `string`
- Missing: `sender: User`, `type`

| Option | Side     | Change                                               |
| ------ | -------- | ---------------------------------------------------- |
| **A**  | Backend  | Convert to camelCase, string ID, add `sender` object |
| **B**  | Frontend | Adapter transforms shape                             |

---

## 6. FOLLOW / REQUEST — flat, no nested User

| Side                       | Shape                                                                             |
| -------------------------- | --------------------------------------------------------------------------------- |
| **Backend Follow**         | `{ followerId, followeeId, createdAt }`                                           |
| **Frontend FollowRequest** | `{ id, requesterId, requester: User, targetId, target: User, status, createdAt }` |

**Issues:**

- Missing: `id`, `requester: User`, `target: User`, `status`

| Option | Side     | Change                                             |
| ------ | -------- | -------------------------------------------------- |
| **A**  | Backend  | Enrich with nested User objects, add `id`/`status` |
| **B**  | Frontend | Adapter synthesizes missing fields                 |

---

## 7. SEARCH USERS — no-op, no pagination

| Side         | Behavior                                                                             |
| ------------ | ------------------------------------------------------------------------------------ |
| **Backend**  | Ignores `query` param, returns flat `[{ id, email, firstName, lastName, nickname }]` |
| **Frontend** | Expects `PaginatedResponse<User>` with `username`, `avatarUrl`                       |

**Issues:**

- Search is no-op
- No pagination
- Missing fields in response

| Option | Side     | Change                             |
| ------ | -------- | ---------------------------------- |
| **A**  | Backend  | Add SQL LIKE filter + pagination   |
| **B**  | Frontend | Client-side filter (bad for scale) |

---

## 8. SEARCH GROUPS — no-op

| Side         | Behavior                 |
| ------------ | ------------------------ |
| **Backend**  | Ignores `query` param    |
| **Frontend** | Expects filtered results |

| Option | Side     | Change              |
| ------ | -------- | ------------------- |
| **A**  | Backend  | Add SQL LIKE filter |
| **B**  | Frontend | Client-side filter  |

---

## 9. PAGINATION ENVELOPE — different structure

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
| **Backend**    | topic response, comment response, group list, chat queries, follow queries, list users/groups search      |
| **Frontend**   | `api.ts` — login already unwraps `res.user`; new `transformers.ts` for field mappings / pagination unwrap |
| **No changes** | `types.ts` — backend will match frontend types                                                            |
