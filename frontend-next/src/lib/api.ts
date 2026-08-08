/**
 * lib/api.ts
 *
 * Centralised API client for the Next.js frontend.
 * Every fetch to the backend goes through here.
 */

import { RegisterBody } from '@/app/register/page';
import type {
  User,
  Profile,
  LoginResponse,
  Post,
  Comment,
  FollowRequest,
  Group,
  GroupMember,
  GroupJoinRequest,
  Event,
  Chat,
  ChatMessage,
  Notification,
  PaginatedResponse,
} from './types';

// ─── Configuration ────────────────────────────────────────────────────────────

const API_BASE = process.env.NEXT_PUBLIC_API_BASE || '/api/v1';

// ─── Error class ─────────────────────────────────────────────────────────────

export class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }

  get isUnauthorized(): boolean {
    return this.status === 401;
  }

  get isForbidden(): boolean {
    return this.status === 403;
  }

  get isNotFound(): boolean {
    return this.status === 404;
  }

  get isTooManyRequests(): boolean {
    return this.status === 429;
  }
}

// ─── Core fetch wrapper ──────────────────────────────────────────────────────

interface RequestOptions extends Omit<RequestInit, 'body'> {
  body?: unknown;
}

async function apiFetch<T = unknown>(path: string, options: RequestOptions = {}): Promise<T> {
  const url = API_BASE + path;
  const hasBody = options.body !== undefined && options.body !== null;

  const defaultHeaders: Record<string, string> = hasBody
    ? { 'Content-Type': 'application/json' }
    : {};

  const mergedOptions: RequestInit = {
    ...options,
    headers: {
      ...defaultHeaders,
      ...(options.headers as Record<string, string>),
    },
    credentials: 'include',
    body: hasBody ? JSON.stringify(options.body) : undefined,
  };

  const response = await fetch(url, mergedOptions);
  const text = await response.text();

  // Check first response header
  if (response.status === 401) {
    handleUnauthorized();
    throw new ApiError(401, 'Session expired');
  }

  if (!response.ok) {
    let message: string;
    try {
      const errBody = text ? JSON.parse(text) : {};
      message = errBody?.error || errBody?.message || `HTTP ${response.status}`;
    } catch {
      message = text || `HTTP ${response.status}`;
    }
    throw new ApiError(response.status, message);
  }

  if (!text) return {} as T;

  let body: unknown;
  try {
    body = JSON.parse(text);
  } catch {
    throw new ApiError(response.status, 'Failed to parse server response');
  }

  if (body && typeof body === 'object' && 'data' in body) {
    return (body as { data: T }).data;
  }

  return body as T;
}

// ─── Check Backend Session Expiration ──────────────────────────────────────────────────────

function handleUnauthorized() {
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new Event('auth:logout'));
  }
}

// ─── Convenience methods ──────────────────────────────────────────────────────

type QueryParams = Record<
  string,
  string | number | boolean | null | undefined | Array<string | number>
>;

export const api = {
  /**
   * `signal` is optional and threaded all the way to fetch(). TanStack
   * Query passes a signal into every queryFn automatically — GET functions
   * in this file accept and forward it so in-flight requests get cancelled
   * when a component unmounts or a query is refetched/superseded (e.g. fast
   * typing in a search box).
   */
  get<T = unknown, P extends QueryParams = QueryParams>(
    path: string,
    params?: P,
    signal?: AbortSignal
  ): Promise<T> {
    if (!params) {
      return apiFetch<T>(path, { method: 'GET', signal });
    }

    const qs = new URLSearchParams();

    Object.entries(params).forEach(([key, value]) => {
      if (value === undefined || value === null || value === '') return;

      if (Array.isArray(value)) {
        value.forEach((v) => qs.append(key, String(v)));
      } else {
        qs.append(key, String(value));
      }
    });

    const query = qs.toString();
    const fullPath = query ? `${path}?${query}` : path;

    return apiFetch<T>(fullPath, { method: 'GET', signal });
  },

  post<T = unknown>(path: string, body?: unknown): Promise<T> {
    return apiFetch<T>(path, {
      method: 'POST',
      body,
    });
  },

  put<T = unknown>(path: string, body?: unknown): Promise<T> {
    return apiFetch<T>(path, {
      method: 'PUT',
      body,
    });
  },

  delete<T = unknown>(path: string, body?: unknown): Promise<T> {
    return apiFetch<T>(path, {
      method: 'DELETE',
      ...(body !== undefined && { body }),
    });
  },
};

// ─── Auth ─────────────────────────────────────────────────────────────────────
// (mutations — no signal needed, useMutation doesn't provide one)

export async function loginEmail(email: string, password: string): Promise<LoginResponse> {
  return api.post<LoginResponse>('/login/email', { identifier: email, password });
}

export async function loginUsername(username: string, password: string): Promise<LoginResponse> {
  return api.post<LoginResponse>('/login/username', { identifier: username, password });
}

export async function register(body: RegisterBody): Promise<void> {
  return api.post<void>('/register', body);
}

export async function logout(): Promise<void> {
  return api.post<void>('/logout');
}

export async function getCurrentUser(signal?: AbortSignal): Promise<User> {
  return api.get<User>('/me', undefined, signal);
}

// ─── Users / Profiles ─────────────────────────────────────────────────────────

export async function getUserProfile(userId: string, signal?: AbortSignal): Promise<Profile> {
  return api.get<Profile>(`/user/profile`, { user_id: userId }, signal);
}

export async function updateProfile(body: Partial<User>): Promise<User> {
  return api.put<User>('/user/update', body);
}

export async function toggleProfilePrivacy(): Promise<void> {
  return api.post<void>('/user/privacy');
}

export async function searchUsers(
  query: string,
  page = 1,
  signal?: AbortSignal
): Promise<PaginatedResponse<User>> {
  return api.get<PaginatedResponse<User>>('/users', { query, page, pageSize: 10 }, signal);
}

// ─── Follow ───────────────────────────────────────────────────────────────────

export async function sendFollowRequest(userId: string): Promise<void> {
  return api.post<void>('/follow', { targetId: userId });
}

export async function handleFollowRequest(
  followerId: string,
  action: 'accept' | 'decline'
): Promise<void> {
  if (action === 'accept') {
    return api.post<void>('/follow/accept', { followerId });
  }
  return api.post<void>('/follow/decline', { followerId });
}

export async function unfollowUser(userId: string): Promise<void> {
  return api.post<void>('/follow/unfollow', { targetId: userId });
}

export async function getFollowers(userId: string, signal?: AbortSignal): Promise<User[]> {
  return api.get<User[]>('/follow/followers', { userId }, signal);
}

export async function getFollowing(userId: string, signal?: AbortSignal): Promise<User[]> {
  return api.get<User[]>('/follow/following', { userId }, signal);
}

export async function getPendingFollowRequests(signal?: AbortSignal): Promise<FollowRequest[]> {
  return api.get<FollowRequest[]>('/follow/requests', undefined, signal);
}

// ─── Posts (Topics) ───────────────────────────────────────────────────────────

/**
 * Uses FormData (image upload) — deliberately bypasses apiFetch's JSON
 * body handling and calls fetch() directly. This is fine as-is; apiFetch
 * always JSON.stringifies options.body and sets Content-Type: application/json,
 * neither of which is correct for multipart/form-data (the browser needs
 * to set that header itself, with the multipart boundary).
 */
export async function createPost(formData: FormData): Promise<Post> {
  const url = API_BASE + '/topics/create';
  const response = await fetch(url, {
    method: 'POST',
    credentials: 'include',
    body: formData,
  });

  const text = await response.text();
  if (!response.ok) {
    const err = text ? JSON.parse(text) : {};
    throw new ApiError(response.status, err.error || err.message || 'Failed to create post');
  }

  const body = JSON.parse(text);
  return body.data;
}

export async function getFeed(
  page = 1,
  size = 10,
  signal?: AbortSignal
): Promise<PaginatedResponse<Post>> {
  return api.get<PaginatedResponse<Post>>('/topics/feed', { page, limit: size }, signal);
}

export async function getUserPosts(
  userId: string,
  page = 1,
  size = 10,
  signal?: AbortSignal
): Promise<PaginatedResponse<Post>> {
  return api.get<PaginatedResponse<Post>>('/topics/user', { userId, page, limit: size }, signal);
}

export async function getPost(postId: string, signal?: AbortSignal): Promise<Post> {
  return api.get<Post>('/topics/get', { id: postId }, signal);
}

export async function deletePost(postId: string): Promise<void> {
  return api.delete<void>(`/topics/delete?id=${postId}`);
}

export async function likePost(postId: string): Promise<void> {
  return api.post<void>(`/topics/vote?id=${postId}`);
}

export async function unlikePost(postId: string): Promise<void> {
  return api.delete<void>(`/topics/vote?id=${postId}`);
}

// ─── Comments ─────────────────────────────────────────────────────────────────

export async function createComment(topicId: string, content: string): Promise<Comment> {
  return api.post<Comment>('/comments/create', { topicId, content });
}

export async function getComments(topicId: string, signal?: AbortSignal): Promise<Comment[]> {
  return api.get<Comment[]>('/comments/topic/votes', { topicId }, signal);
}

export async function voteComment(commentId: string, reactionType: 1 | -1): Promise<void> {
  return api.post<void>(`/comments/vote?id=${commentId}`, { reactionType });
}

export async function deleteComment(commentId: string): Promise<void> {
  return api.delete<void>(`/comments/delete?id=${commentId}`);
}

// ─── Groups ───────────────────────────────────────────────────────────────────

export async function createGroup(title: string, description: string): Promise<Group> {
  return api.post<Group>('/groups', { title, description });
}

export async function getGroup(groupId: string, signal?: AbortSignal): Promise<Group> {
  return api.get<Group>(`/groups/${groupId}`, undefined, signal);
}

export async function updateGroup(groupId: string, data: Partial<Group>): Promise<Group> {
  return api.put<Group>(`/groups/${groupId}`, data);
}

export async function deleteGroup(groupId: string): Promise<void> {
  return api.delete<void>(`/groups/${groupId}`);
}

export async function browseGroups(
  query?: string,
  page = 1,
  signal?: AbortSignal
): Promise<PaginatedResponse<Group>> {
  return api.get<PaginatedResponse<Group>>('/groups', { query, page, limit: 20 }, signal);
}

export async function inviteToGroup(groupId: string, userId: string): Promise<void> {
  return api.post<void>(`/groups/${groupId}/invite`, { userId });
}

export async function requestToJoinGroup(groupId: string): Promise<void> {
  return api.post<void>(`/groups/${groupId}/request`);
}

export async function handleJoinRequest(
  requestId: string,
  action: 'accept' | 'decline'
): Promise<void> {
  return api.put<void>(`/groups/requests/${requestId}`, { action });
}

export async function leaveGroup(groupId: string): Promise<void> {
  return api.delete<void>(`/groups/${groupId}/leave`);
}

export async function getGroupPosts(
  groupId: string,
  page = 1,
  size = 10,
  signal?: AbortSignal
): Promise<PaginatedResponse<Post>> {
  return api.get<PaginatedResponse<Post>>(
    `/groups/${groupId}/posts`,
    { page, limit: size },
    signal
  );
}

export async function getGroupMembers(
  groupId: string,
  page = 1,
  size = 20,
  signal?: AbortSignal
): Promise<PaginatedResponse<GroupMember>> {
  return api.get<PaginatedResponse<GroupMember>>(
    `/groups/${groupId}/members`,
    { page, limit: size },
    signal
  );
}

export async function getPendingJoinRequests(
  groupId: string,
  signal?: AbortSignal
): Promise<GroupJoinRequest[]> {
  return api.get<GroupJoinRequest[]>(`/groups/${groupId}/requests/pending`, undefined, signal);
}

// ─── Events ───────────────────────────────────────────────────────────────────

export async function createEvent(
  groupId: string,
  data: { title: string; description: string; eventDate: string; options?: string[] }
): Promise<Event> {
  return api.post<Event>(`/groups/${groupId}/events`, data);
}

export async function getGroupEvents(
  groupId: string,
  page = 1,
  size = 10,
  signal?: AbortSignal
): Promise<PaginatedResponse<Event>> {
  return api.get<PaginatedResponse<Event>>(
    `/groups/${groupId}/events`,
    { page, limit: size },
    signal
  );
}

export async function respondToEvent(eventId: string, response: string): Promise<void> {
  return api.post<void>(`/events/${eventId}/respond`, { response });
}

// ─── Chat ─────────────────────────────────────────────────────────────────────

export async function getChats(signal?: AbortSignal): Promise<Chat[]> {
  return api.get<Chat[]>('/chat/users', undefined, signal);
}

export async function getChatMessages(
  chatId: string,
  signal?: AbortSignal
): Promise<ChatMessage[]> {
  return api.get<ChatMessage[]>('/chat/history', { chatId }, signal);
}

// ─── Notifications ────────────────────────────────────────────────────────────

export async function getNotifications(
  page = 1,
  signal?: AbortSignal
): Promise<PaginatedResponse<Notification>> {
  return api.get<PaginatedResponse<Notification>>('/notifications', { page }, signal);
}

export async function getUnreadNotificationCount(signal?: AbortSignal): Promise<{ count: number }> {
  return api.get<{ count: number }>('/notifications/unread-count', undefined, signal);
}

export async function markNotificationRead(notificationId: string): Promise<void> {
  return api.put<void>(`/notifications/${notificationId}/read`);
}

export async function markAllNotificationsRead(): Promise<void> {
  return api.put<void>('/notifications/read-all');
}
