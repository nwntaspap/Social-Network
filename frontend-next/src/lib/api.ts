/**
 * lib/api.ts
 *
 * Centralised API client for the Next.js frontend.
 * Every fetch to the backend goes through here.
 */

import { RegisterBody } from '@/app/register/page';
import type {
  User,
  Post,
  Comment,
  Group,
  Event,
  Chat,
  ChatMessage,
  Notification,
  FollowRequest,
  PaginatedResponse,
} from './types';

// ─── Configuration ────────────────────────────────────────────────────────────

const API_BASE = process.env.NEXT_PUBLIC_API_BASE || 'http://localhost:8080/api/v1';

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

// ─── Convenience methods ──────────────────────────────────────────────────────

type QueryParams = Record<
  string,
  string | number | boolean | null | undefined | Array<string | number>
>;

export const api = {
  get<T = unknown, P extends QueryParams = QueryParams>(path: string, params?: P): Promise<T> {
    if (!params) {
      return apiFetch<T>(path, { method: 'GET' });
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

    return apiFetch<T>(fullPath, { method: 'GET' });
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

export async function loginEmail(email: string, password: string): Promise<User> {
  return api.post<User>('/login/email', { email, password });
}

export async function loginUsername(username: string, password: string): Promise<User> {
  return api.post<User>('/login/username', { username, password });
}

export async function register(body: RegisterBody): Promise<void> {
  return api.post<void>('/register', body);
}

export async function logout(): Promise<void> {
  return api.post<void>('/logout');
}

export async function getCurrentUser(): Promise<User> {
  return api.get<User>('/auth/me');
}

// ─── Users / Profiles ─────────────────────────────────────────────────────────

export async function getUserProfile(userId: string): Promise<User> {
  return api.get<User>(`/users/${userId}`);
}

export async function updateProfile(body: Partial<User>): Promise<User> {
  return api.put<User>('/profile', body);
}

export async function toggleProfilePrivacy(): Promise<User> {
  return api.put<User>('/profile/privacy');
}

export async function searchUsers(query: string, page = 1): Promise<PaginatedResponse<User>> {
  return api.get<PaginatedResponse<User>>('/users/search', { query, page });
}

// ─── Follow ───────────────────────────────────────────────────────────────────

export async function sendFollowRequest(userId: string): Promise<FollowRequest> {
  return api.post<FollowRequest>(`/follow/request/${userId}`);
}

export async function handleFollowRequest(
  requestId: string,
  action: 'accept' | 'decline'
): Promise<void> {
  return api.put<void>(`/follow/request/${requestId}`, { action });
}

export async function unfollowUser(userId: string): Promise<void> {
  return api.delete<void>(`/follow/${userId}`);
}

export async function getFollowers(userId: string, page = 1): Promise<PaginatedResponse<User>> {
  return api.get<PaginatedResponse<User>>(`/users/${userId}/followers`, { page });
}

export async function getFollowing(userId: string, page = 1): Promise<PaginatedResponse<User>> {
  return api.get<PaginatedResponse<User>>(`/users/${userId}/following`, { page });
}

export async function getPendingFollowRequests(): Promise<FollowRequest[]> {
  return api.get<FollowRequest[]>('/follow/requests/pending');
}

// ─── Posts ────────────────────────────────────────────────────────────────────

export async function createPost(formData: FormData): Promise<Post> {
  const url = API_BASE + '/posts';
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

export async function getFeed(page = 1): Promise<PaginatedResponse<Post>> {
  return api.get<PaginatedResponse<Post>>('/posts/feed', { page });
}

export async function getUserPosts(userId: string, page = 1): Promise<PaginatedResponse<Post>> {
  return api.get<PaginatedResponse<Post>>(`/users/${userId}/posts`, { page });
}

export async function getPost(postId: string): Promise<Post> {
  return api.get<Post>(`/posts/${postId}`);
}

export async function deletePost(postId: string): Promise<void> {
  return api.delete<void>(`/posts/${postId}`);
}

export async function likePost(postId: string): Promise<void> {
  return api.post<void>(`/posts/${postId}/like`);
}

export async function unlikePost(postId: string): Promise<void> {
  return api.delete<void>(`/posts/${postId}/like`);
}

// ─── Comments ─────────────────────────────────────────────────────────────────

export async function createComment(
  postId: string,
  content: string,
  image?: File
): Promise<Comment> {
  if (image) {
    const formData = new FormData();
    formData.append('content', content);
    formData.append('image', image);

    const url = API_BASE + `/posts/${postId}/comments`;
    const response = await fetch(url, {
      method: 'POST',
      credentials: 'include',
      body: formData,
    });

    const text = await response.text();
    if (!response.ok) {
      const err = text ? JSON.parse(text) : {};
      throw new ApiError(response.status, err.error || err.message || 'Failed to create comment');
    }

    const body = JSON.parse(text);
    return body.data;
  }

  return api.post<Comment>(`/posts/${postId}/comments`, { content });
}

export async function getComments(postId: string, page = 1): Promise<PaginatedResponse<Comment>> {
  return api.get<PaginatedResponse<Comment>>(`/posts/${postId}/comments`, { page });
}

export async function deleteComment(commentId: string): Promise<void> {
  return api.delete<void>(`/comments/${commentId}`);
}

// ─── Groups ───────────────────────────────────────────────────────────────────

export async function createGroup(title: string, description: string): Promise<Group> {
  return api.post<Group>('/groups', { title, description });
}

export async function getGroup(groupId: string): Promise<Group> {
  return api.get<Group>(`/groups/${groupId}`);
}

export async function browseGroups(query?: string, page = 1): Promise<PaginatedResponse<Group>> {
  return api.get<PaginatedResponse<Group>>('/groups', { query, page });
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

export async function getGroupMembers(groupId: string, page = 1): Promise<PaginatedResponse<User>> {
  return api.get<PaginatedResponse<User>>(`/groups/${groupId}/members`, { page });
}

// ─── Events ───────────────────────────────────────────────────────────────────

export async function createEvent(
  groupId: string,
  data: { title: string; description: string; eventDate: string }
): Promise<Event> {
  return api.post<Event>(`/groups/${groupId}/events`, data);
}

export async function getGroupEvents(groupId: string): Promise<Event[]> {
  return api.get<Event[]>(`/groups/${groupId}/events`);
}

export async function respondToEvent(
  eventId: string,
  response: 'going' | 'notGoing'
): Promise<void> {
  return api.post<void>(`/events/${eventId}/respond`, { response });
}

// ─── Chat ─────────────────────────────────────────────────────────────────────

export async function getChats(): Promise<Chat[]> {
  return api.get<Chat[]>('/chats');
}

export async function getChatMessages(
  chatId: string,
  page = 1
): Promise<PaginatedResponse<ChatMessage>> {
  return api.get<PaginatedResponse<ChatMessage>>(`/chats/${chatId}/messages`, { page });
}

export async function sendPrivateMessage(
  receiverId: string,
  content: string
): Promise<ChatMessage> {
  return api.post<ChatMessage>(`/chats/private/${receiverId}`, { content });
}

export async function sendGroupMessage(groupId: string, content: string): Promise<ChatMessage> {
  return api.post<ChatMessage>(`/chats/group/${groupId}`, { content });
}

// ─── Notifications ────────────────────────────────────────────────────────────

export async function getNotifications(page = 1): Promise<PaginatedResponse<Notification>> {
  return api.get<PaginatedResponse<Notification>>('/notifications', { page });
}

export async function getUnreadNotificationCount(): Promise<{ count: number }> {
  return api.get<{ count: number }>('/notifications/unread-count');
}

export async function markNotificationRead(notificationId: string): Promise<void> {
  return api.put<void>(`/notifications/${notificationId}/read`);
}

export async function markAllNotificationsRead(): Promise<void> {
  return api.put<void>('/notifications/read-all');
}
