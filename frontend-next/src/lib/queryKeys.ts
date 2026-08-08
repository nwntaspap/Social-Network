/**
 * lib/queryKeys.ts
 *
 * Central registry of TanStack Query keys.
 *
 * Why centralize: query keys are how the cache is indexed. Typos or
 * inconsistent shapes (e.g. ['feed'] in one file, ['Feed'] in another)
 * silently break cache invalidation. Keeping them here as functions also
 * means autocomplete works everywhere they're used, and invalidating a
 * whole feature (e.g. all group-related queries) is a one-liner.
 *
 * Convention: each key is a function, even when it takes no params, so the
 * usage is consistent (queryKeys.feed() not queryKeys.feed).
 */

export const queryKeys = {
  // ─── Home ───────────────────────────────────────────────────────────────
  feed: () => ['feed'] as const,
  suggestedUsers: () => ['suggestedUsers'] as const,
  userSearch: (query: string) => ['userSearch', query] as const,

  // ─── Profile ────────────────────────────────────────────────────────────
  profile: (userId: string) => ['profile', userId] as const,
  userPosts: (userId: string) => ['userPosts', userId] as const,
  followers: (userId: string) => ['followers', userId] as const,
  following: (userId: string) => ['following', userId] as const,
  pendingFollowRequests: () => ['pendingFollowRequests'] as const,

  // ─── Posts / Comments ───────────────────────────────────────────────────
  post: (postId: string) => ['post', postId] as const,
  comments: (topicId: string) => ['comments', topicId] as const,

  // ─── Groups ─────────────────────────────────────────────────────────────
  groups: (query?: string) => ['groups', query ?? ''] as const,
  group: (groupId: string) => ['group', groupId] as const,
  groupPosts: (groupId: string) => ['groupPosts', groupId] as const,
  groupMembers: (groupId: string) => ['groupMembers', groupId] as const,
  groupPendingRequests: (groupId: string) => ['groupPendingRequests', groupId] as const,
  groupEvents: (groupId: string) => ['groupEvents', groupId] as const,

  // ─── Chat ───────────────────────────────────────────────────────────────
  chats: () => ['chats'] as const,
  chatMessages: (chatId: string) => ['chatMessages', chatId] as const,

  // ─── Notifications ──────────────────────────────────────────────────────
  notifications: () => ['notifications'] as const,
  unreadNotificationCount: () => ['unreadNotificationCount'] as const,
} as const;
