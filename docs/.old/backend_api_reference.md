# Backend API Reference

## Scope

This document describes the backend HTTP API registered in the server router.

Base path: `/api/v1`

## Authentication

Protected endpoints require authenticated user context (session/access token middleware).

## Response format

Most handlers return JSON through the shared helper:

- Success (`< 400`):

```json
{
  "info": null,
  "data": { "...": "payload" }
}
```

- Error (`>= 400`):

```json
{
  "error": "message"
}
```

Notes:

- Some endpoints do not use the shared helper and return raw JSON arrays/objects or plain text errors.
- OAuth endpoints use redirects, not JSON payloads.
- Notification stream is Server-Sent Events (`text/event-stream`).

## Endpoint index

| Method | Path                                  | Auth     |
| ------ | ------------------------------------- | -------- |
| GET    | `/api/v1/health`                      | Optional |
| POST   | `/api/v1/register`                    | No       |
| POST   | `/api/v1/login/email`                 | No       |
| POST   | `/api/v1/login/username`              | No       |
| POST   | `/api/v1/logout`                      | Required |
| GET    | `/api/v1/me`                          | Required |
| GET    | `/api/v1/auth/github/login`           | No       |
| GET    | `/api/v1/auth/github/callback`        | No       |
| GET    | `/api/v1/auth/google/login`           | No       |
| GET    | `/api/v1/auth/google/callback`        | No       |
| POST   | `/api/v1/topics/create`               | Required |
| PUT    | `/api/v1/topics/update`               | Required |
| DELETE | `/api/v1/topics/delete`               | Required |
| GET    | `/api/v1/topic`                       | Optional |
| GET    | `/api/v1/topics/all`                  | Optional |
| POST   | `/api/v1/comments/create`             | Required |
| PUT    | `/api/v1/comments/update`             | Required |
| DELETE | `/api/v1/comments/delete`             | Required |
| GET    | `/api/v1/comments/get`                | No       |
| GET    | `/api/v1/comments/topic`              | No       |
| POST   | `/api/v1/category/create`             | Required |
| PUT    | `/api/v1/category/update`             | Required |
| DELETE | `/api/v1/category/delete`             | Required |
| GET    | `/api/v1/category`                    | Optional |
| GET    | `/api/v1/categories/all`              | No       |
| POST   | `/api/v1/vote/cast`                   | Required |
| DELETE | `/api/v1/vote/delete`                 | Required |
| GET    | `/api/v1/vote/counts`                 | Optional |
| GET    | `/api/v1/user/activity`               | Required |
| GET    | `/api/v1/notifications/stream`        | Required |
| GET    | `/api/v1/notifications/unread-count`  | Required |
| GET    | `/api/v1/notifications`               | Required |
| POST   | `/api/v1/notifications/mark-read`     | Required |
| POST   | `/api/v1/notifications/mark-all-read` | Required |

## Health

### GET `/api/v1/health`

- Input:
  - None.
- Output:
  - `200 OK`
  - `data.status` (string)
  - `data.timestamp` (RFC3339 string)
- Errors:
  - `405` invalid method
  - `500` failed to create test notification

## Auth and User

### POST `/api/v1/register`

- Input (JSON):

```json
{
  "nickname": "string",
  "password": "string",
  "email": "string",
  "firstname": "string",
  "lastname": "string",
  "age": int,
  "gender": "string"
}
```

- Output:
  - `201 Created`
  - `data.userId` (string)
  - `data.message` (string)
- Errors:
  - `400` invalid payload or validation errors
  - `405` invalid method
  - `500` registration failure

### POST `/api/v1/login/email`

- Input (JSON):

```json
{
  "email": "string",
  "password": "string"
}
```

- Output:
  - `200 OK`
  - `data.userId` (string)
  - `data.username` (string)
  - `data.accessToken` (string)
  - `data.refreshToken` (string)
- Errors:
  - `400` invalid payload or validation errors
  - `405` invalid method
  - `500` login/session creation failure

### POST `/api/v1/login/username`

- Input (JSON):

```json
{
  "username": "string",
  "password": "string"
}
```

- Output:
  - `200 OK`
  - `data.userId` (string)
  - `data.username` (string)
  - `data.accessToken` (string)
  - `data.refreshToken` (string)
- Errors:
  - `400` invalid payload or validation errors
  - `405` invalid method
  - `500` login/session creation failure

### POST `/api/v1/logout`

- Input:
  - No body.
  - Auth/session token required.
- Output:
  - `200 OK`
  - `data.message` = `Logged out successfully`
- Errors:
  - `401` unauthorized or no session
  - `405` invalid method
  - `500` logout failure

### GET `/api/v1/me`

- Input:
  - No body.
  - Auth/session required.
- Output:
  - `200 OK`
  - `data.id` (string)
  - `data.username` (string)
  - `data.email` (string)
- Errors:
  - `401` user not found in context
  - `405` invalid method

## OAuth

### GET `/api/v1/auth/github/login`

### GET `/api/v1/auth/google/login`

- Input:
  - No body.
- Output:
  - `307 Temporary Redirect` to provider auth URL.
- Errors:
  - `500` state generation/internal failure

### GET `/api/v1/auth/github/callback`

### GET `/api/v1/auth/google/callback`

- Input (query):
  - `code` (required)
  - `state` (required)
  - `error` (optional, provider error)
- Output:
  - `307 Temporary Redirect` to frontend callback URL with:
    - `access_token`
    - `refresh_token`
- Errors:
  - `405` invalid method
  - `500` oauth state/code/provider/session failures (plain text error body)

## Topics

### POST `/api/v1/topics/create`

- Input (JSON):

```json
{
  "title": "string",
  "content": "string",
  "imagePath": "string",
  "categoryIds": [1, 2]
}
```

- Output:
  - `201 Created`
  - `data.userId` (string)
  - `data.message` (string)
- Errors:
  - `400` invalid payload/validation
  - `401` unauthenticated
  - `405` invalid method
  - `500` create failure

### PUT `/api/v1/topics/update`

- Input (JSON):

```json
{
  "topicId": 123,
  "title": "string",
  "content": "string",
  "imagePath": "string",
  "categoryIds": [1, 2]
}
```

- Output:
  - `201 Created` (handler currently returns 201 for update)
  - `data.userId` (string)
  - `data.message` (string)
- Errors:
  - `400` invalid payload/validation
  - `401` unauthenticated
  - `405` invalid method
  - `500` update failure

### DELETE `/api/v1/topics/delete`

- Input (query):
  - `id` (int, required)
- Output:
  - `200 OK`
  - `data.userId` (string)
  - `data.topicId` (int)
  - `data.message` (string)
- Errors:
  - `400` invalid/missing id or validation errors
  - `401` unauthenticated
  - `405` invalid method
  - `500` delete failure

### GET `/api/v1/topic`

- Input (query):
  - `id` (int, required)
- Output:
  - `200 OK`
  - `data.topicId` (int)
  - `data.title` (string)
  - `data.content` (string)
  - `data.imagePath` (string)
  - `data.userId` (string)
  - `data.ownerUsername` (string)
  - `data.createdAt` (string)
  - `data.updatedAt` (string)
  - `data.categoryIds` (int[])
  - `data.categoryNames` (string[])
  - `data.categoryColors` (string[])
  - `data.comments` (array)
  - `data.upvotes` (int)
  - `data.downvotes` (int)
  - `data.score` (int)
  - `data.userVote` (int or null)
- Errors:
  - `400` invalid/missing id or validation errors
  - `404` topic not found
  - `405` invalid method
  - `500` internal server error

### GET `/api/v1/topics/all`

- Input (query, all optional):
  - `page` (int, default from helper)
  - `page_size` (int, default from helper)
  - `order_by` (default `created_at`)
  - `order` (default `desc`)
  - `search` (string)
  - `category` (int)
- Output:
  - `200 OK`
  - `data.topics` (array)
  - `data.categories` (array)
  - `data.filters` object with applied search/order values
  - `data.pagination` object (`page`, `limit`, `total`, `total_pages`, `has_next`, `has_prev`, `next_page`, `prev_page`)
- Errors:
  - `400` validation errors
  - `405` invalid method
  - `500` retrieval failure

## Comments

### POST `/api/v1/comments/create`

- Input (JSON):

```json
{
  "topicId": 123,
  "content": "string"
}
```

- Output:
  - `201 Created`
  - `data.commentId` (int)
  - `data.message` (string)
- Errors:
  - `400` invalid payload/validation
  - `401` unauthenticated
  - `405` invalid method
  - `500` create failure

### PUT `/api/v1/comments/update`

- Input (JSON):

```json
{
  "id": 456,
  "content": "string"
}
```

- Output:
  - `200 OK`
  - `data.message` (string)
- Errors:
  - `400` invalid payload/validation
  - `401` unauthenticated
  - `405` invalid method
  - `500` update failure

### DELETE `/api/v1/comments/delete`

- Input (query):
  - `id` (int, required)
- Output:
  - `200 OK`
  - `data.message` (string)
- Errors:
  - `400` invalid/missing id or validation errors
  - `401` unauthenticated
  - `405` invalid method
  - `500` delete failure

### GET `/api/v1/comments/get`

- Input (query):
  - `id` (int, required)
- Output:
  - `200 OK`
  - `data.id` (int)
  - `data.topicId` (int)
  - `data.userId` (string)
  - `data.username` (string)
  - `data.content` (string)
  - `data.createdAt` (string)
  - `data.updatedAt` (string)
- Errors:
  - `400` invalid/missing id or validation errors
  - `404` comment not found
  - `405` invalid method
  - `500` internal server error

### GET `/api/v1/comments/topic`

- Input (query):
  - `id` (topic id, int, required)
- Output:
  - `200 OK`
  - `data.comments` (array)
- Errors:
  - `400` invalid/missing id or validation errors
  - `405` invalid method
  - `500` retrieval failure

## Categories

### POST `/api/v1/category/create`

- Input (JSON):

```json
{
  "name": "string",
  "description": "string"
}
```

- Output:
  - `201 Created`
  - `data.categoryName` (string)
  - `data.message` (string)
- Errors:
  - `400` invalid payload/validation
  - `401` unauthenticated
  - `405` invalid method
  - `500` create failure

### PUT `/api/v1/category/update`

- Input (JSON):

```json
{
  "id": 12,
  "name": "string",
  "description": "string"
}
```

- Output:
  - `200 OK`
  - `data.categoryId` (int)
  - `data.message` (string)
- Errors:
  - `400` invalid payload/validation
  - `401` unauthenticated
  - `405` invalid method
  - `500` update failure

### DELETE `/api/v1/category/delete`

- Input (query):
  - `id` (int, required)
- Output:
  - `200 OK`
  - `data.categoryId` (int)
  - `data.message` (string)
- Errors:
  - `400` invalid/missing id or validation errors
  - `401` unauthenticated
  - `405` invalid method
  - `500` delete failure

### GET `/api/v1/category`

- Input (query):
  - `id` (int, required)
- Output:
  - `200 OK`
  - `data.categoryId` (int)
  - `data.categoryName` (string)
- Errors:
  - `400` invalid/missing id or validation errors
  - `405` invalid method
  - `500` retrieval failure

### GET `/api/v1/categories/all`

- Input (query, optional):
  - `page` (int)
  - `page_size` (int)
  - `order_by` (default `created_at`)
  - `order` (default `desc`)
  - `search` (string)
- Output:
  - `200 OK`
  - `data.categories` (array)
  - `data.filters` object
  - `data.pagination` object (`page`, `limit`, `totalPages`, `totalItems`, `has_next`, `has_prev`, `next_page`, `prev_page`)
- Errors:
  - `405` invalid method
  - `500` retrieval failure

## Votes

### POST `/api/v1/vote/cast`

- Input (JSON):

```json
{
  "topicId": 123,
  "commentId": null,
  "reactionType": 1
}
```

Rules:

- Vote target can be a topic or comment (`topicId` or `commentId`).
- `reactionType` is used in notifications (`1` like, `-1` dislike).

- Output:
  - `200 OK`
  - `data.message` (string)
- Errors:
  - `400` invalid JSON
  - `401` unauthorized
  - `405` invalid method
  - `500` cast failure

### DELETE `/api/v1/vote/delete`

- Input (JSON):

```json
{
  "topicId": 123,
  "commentId": null
}
```

- Output:
  - `200 OK`
  - `data.message` (string)
- Errors:
  - `401` unauthenticated
  - `405` invalid method
  - `500` body parse/delete failure

### GET `/api/v1/vote/counts`

- Input (query):
  - `topic_id` (int) or `comment_id` (int)
- Rules:
  - Provide exactly one of `topic_id` or `comment_id`.
- Output:
  - `200 OK`
  - `data.upvotes` (int)
  - `data.downvotes` (int)
  - `data.score` (int)
- Errors:
  - `400` both IDs provided or neither provided
  - `405` invalid method
  - `500` retrieval failure

## Activity

### GET `/api/v1/user/activity`

- Input:
  - No body.
  - Auth required.
- Output:
  - `200 OK`
  - `data.createdTopics` (array)
  - `data.likedTopics` (array)
  - `data.dislikedTopics` (array)
  - `data.likedComments` (array)
  - `data.dislikedComments` (array)
  - `data.userComments` (array)
- Errors:
  - `401` unauthenticated
  - `405` invalid method
  - `500` retrieval failure

## Notifications

### GET `/api/v1/notifications/stream`

- Input:
  - No body.
  - Auth required.
- Output:
  - `200 OK`, `Content-Type: text/event-stream`
  - SSE events:
    - `{"type":"connected"}`
    - `{"type":"unread_count","count":number}`
    - notification JSON payloads
    - heartbeat comments every 10s
- Errors:
  - `401` unauthorized
  - `500` streaming not supported

### GET `/api/v1/notifications/unread-count`

- Input:
  - No body.
  - Auth required.
- Output (raw JSON, not wrapped):

```json
{
  "count": 3
}
```

- Errors:
  - `401` unauthorized
  - `500` read/encode failure

### GET `/api/v1/notifications`

- Input (query):
  - `limit` optional (1..100, default `50`)
- Output (raw JSON array, not wrapped):

```json
[
  {
    "id": 1,
    "user_id": "...",
    "type": "..."
  }
]
```

- Errors:
  - `401` unauthorized
  - `500` fetch/encode failure

### POST `/api/v1/notifications/mark-read`

- Input (query):
  - `id` (notification id, int, required)
- Output:
  - `200 OK`
  - empty body
- Errors:
  - `400` invalid notification id
  - `401` unauthorized
  - `500` mark failure

### POST `/api/v1/notifications/mark-all-read`

- Input:
  - No body.
- Output:
  - `200 OK`
  - empty body
- Errors:
  - `401` unauthorized
  - `500` mark failure

## Notes for frontend migration to JavaScript

- Use the response wrapper `data` for most endpoints.
- Handle raw responses for notification list/count and SSE stream.
- OAuth login/callback flows are redirect-based, not JSON API calls.
