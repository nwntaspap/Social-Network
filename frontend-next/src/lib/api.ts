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
  Group,
  GroupMember,
  GroupJoinRequest,
  GroupInvitation,
  Event,
  GroupEventsResponse,
  EventRSVPsResponse,
  Chat,
  ChatMessage,
  GroupChatMessageWire,
  GroupPresence,
  NotificationsResponse,
  FollowRequest,
  PaginatedResponse,
} from './types';

// ─── Configuration ────────────────────────────────────────────────────────────

const API_BASE = process.env.NEXT_PUBLIC_API_BASE || '/api/v1';

/**
 * Base URL for the standalone notifications service. The Next dev proxy routes
 * /api/* to the main backend (8080), which does not serve /notifications/*, so
 * the notifications UI talks to this service directly (CORS-enabled).
 */
// Proxied through Next.js (see next.config.ts rewrites) so the session
// cookie stays first-party and no CORS setup is required.
const NOTIFICATIONS_BASE = '/notifications-api/api/v1';

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

export async function loginEmail(email: string, password: string): Promise<LoginResponse> {
  return api.post<LoginResponse>('/login/email', { identifier: email, password });
}

export async function loginUsername(username: string, password: string): Promise<LoginResponse> {
  return api.post<LoginResponse>('/login/username', { identifier: username, password });
}

export async function register(body: RegisterBody, avatar?: File | null): Promise<void> {
  const form = new FormData();
  form.append('email', body.email);
  form.append('password', body.password);
  form.append('firstName', body.firstName);
  form.append('lastName', body.lastName);
  form.append('nickname', body.nickname);
  form.append('dateOfBirth', body.dateOfBirth);
  form.append('gender', body.gender);
  if (avatar) {
    form.append('avatar', avatar);
  }
  const url = API_BASE + '/register';
  const response = await fetch(url, {
    method: 'POST',
    credentials: 'include',
    body: form,
  });
  if (!response.ok) {
    const text = await response.text();
    let message = `HTTP ${response.status}`;
    try {
      const err = text ? JSON.parse(text) : {};
      message = err.error || err.message || message;
    } catch {
      if (text) message = text;
    }
    throw new ApiError(response.status, message);
  }
}

export async function logout(): Promise<void> {
  return api.post<void>('/logout');
}

export async function getCurrentUser(): Promise<User> {
  return api.get<User>('/me');
}

// ─── Users / Profiles ─────────────────────────────────────────────────────────

export async function getUserProfile(userId: string): Promise<Profile> {
  return api.get<Profile>(`/user/profile`, { user_id: userId });
}

export async function updateProfile(body: Partial<User>): Promise<User> {
  return api.put<User>('/user/update', body);
}

export async function toggleProfilePrivacy(isPrivate: boolean): Promise<void> {
  return api.post<void>('/user/privacy', { isPrivate });
}

export async function searchUsers(query: string, page = 1): Promise<PaginatedResponse<User>> {
  return api.get<PaginatedResponse<User>>('/users', { query, page, pageSize: 10 });
}

export async function getSuggestedUsers(
  excludeFollowedOf: string,
  page = 1
): Promise<PaginatedResponse<User>> {
  return api.get<PaginatedResponse<User>>('/users', { excludeFollowedOf, page, pageSize: 10 });
}

// ─── Follow ───────────────────────────────────────────────────────────────────

export interface FollowResponse {
  status: 'following' | 'pending';
}

export async function sendFollowRequest(userId: string): Promise<FollowResponse> {
  return api.post<FollowResponse>('/follow', { targetId: userId });
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

export async function getFollowers(userId: string): Promise<User[]> {
  return api.get<User[]>('/follow/followers', { userId });
}

export async function getFollowing(userId: string): Promise<User[]> {
  return api.get<User[]>('/follow/following', { userId });
}

export async function getPendingFollowRequests(): Promise<FollowRequest[]> {
  return api.get<FollowRequest[]>('/follow/requests');
}

// ─── Posts (Topics) ───────────────────────────────────────────────────────────

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

export async function createGroupPost(formData: FormData, groupId: string): Promise<Post> {
  const url = API_BASE + `/groups/${groupId}/posts`;
  const response = await fetch(url, {
    method: 'POST',
    credentials: 'include',
    body: formData,
  });

  const text = await response.text();
  if (!response.ok) {
    const err = text ? JSON.parse(text) : {};
    throw new ApiError(response.status, err.error || err.message || 'Failed to create group post');
  }

  const body = JSON.parse(text);
  return body.data;
}

export async function getFeed(page = 1, size = 10): Promise<PaginatedResponse<Post>> {
  return api.get<PaginatedResponse<Post>>('/topics/feed', { page, limit: size });
}

export async function getUserPosts(
  userId: string,
  page = 1,
  size = 10
): Promise<PaginatedResponse<Post>> {
  return api.get<PaginatedResponse<Post>>('/topics/user', { userId, page, limit: size });
}

export async function getPost(postId: number): Promise<Post> {
  return api.get<Post>('/topics/get', { id: postId });
}

export async function deletePost(postId: number): Promise<void> {
  return api.delete<void>(`/topics/delete?id=${postId}`);
}

export async function likePost(postId: number): Promise<void> {
  return api.post<void>(`/topics/vote?id=${postId}`, { reactionType: 1 });
}

export async function dislikePost(postId: number): Promise<void> {
  return api.post<void>(`/topics/vote?id=${postId}`, { reactionType: -1 });
}

export async function voteGroupPost(postId: string, reactionType: 1 | -1): Promise<void> {
  return api.post<void>(`/groups/posts/${postId}/vote`, { reactionType });
}

// ─── Comments ─────────────────────────────────────────────────────────────────

export async function createComment(
  topicId: number,
  content: string,
  image?: File | null
): Promise<Comment> {
  const formData = new FormData();
  formData.append('topicId', String(topicId));
  formData.append('content', content);
  if (image) {
    formData.append('image', image);
  }
  return postMultipart<Comment>('/comments/create', formData, 'Failed to create comment');
}

export async function getComments(topicId: number): Promise<Comment[]> {
  return api.get<Comment[]>('/comments/topic/votes', { topicId });
}

/** Fetch a single comment (used to resolve a comment notification's post). */
export async function getComment(commentId: number): Promise<Comment> {
  return api.get<Comment>('/comments/get', { id: commentId });
}

export async function voteComment(commentId: number, reactionType: 1 | -1): Promise<void> {
  return api.post<void>(`/comments/vote?id=${commentId}`, { reactionType });
}

export async function deleteComment(commentId: number): Promise<void> {
  return api.delete<void>(`/comments/delete?id=${commentId}`);
}

// ─── Group post comments ──────────────────────────────────────────────────────

export async function getGroupPostComments(postId: string): Promise<PaginatedResponse<Comment>> {
  return api.get<PaginatedResponse<Comment>>(`/groups/posts/${postId}/comments`);
}

export async function createGroupPostComment(
  postId: string,
  content: string,
  image?: File | null
): Promise<Comment> {
  const formData = new FormData();
  formData.append('content', content);
  if (image) {
    formData.append('image', image);
  }
  return postMultipart<Comment>(
    `/groups/posts/${postId}/comments`,
    formData,
    'Failed to create comment'
  );
}

function postMultipart<T>(path: string, formData: FormData, fallbackMessage: string): Promise<T> {
  const url = API_BASE + path;
  return fetch(url, {
    method: 'POST',
    credentials: 'include',
    body: formData,
  }).then(async (response) => {
    const text = await response.text();
    if (!response.ok) {
      const err = text ? JSON.parse(text) : {};
      throw new ApiError(response.status, err.error || err.message || fallbackMessage);
    }
    const body = JSON.parse(text);
    return body.data as T;
  });
}

// ─── Groups ───────────────────────────────────────────────────────────────────

export async function createGroup(title: string, description: string): Promise<Group> {
  return api.post<Group>('/groups', { title, description });
}

export async function getGroup(groupId: string): Promise<Group> {
  return api.get<Group>(`/groups/${groupId}`);
}

export async function updateGroup(groupId: string, data: Partial<Group>): Promise<Group> {
  return api.put<Group>(`/groups/${groupId}`, data);
}

export async function deleteGroup(groupId: string): Promise<void> {
  return api.delete<void>(`/groups/${groupId}`);
}

export async function browseGroups(query?: string, page = 1): Promise<PaginatedResponse<Group>> {
  return api.get<PaginatedResponse<Group>>('/groups', { query, page, limit: 20 });
}

export async function inviteToGroup(groupId: string, userId: string): Promise<void> {
  return api.post<void>(`/groups/${groupId}/invite`, { userId });
}

export async function getMyGroupInvitations(): Promise<GroupInvitation[]> {
  return api.get<GroupInvitation[]>('/groups/invitations/pending');
}

export interface InviteResponse {
  status: 'member' | 'pending' | 'ok';
}

export async function respondToGroupInvitation(
  groupId: string,
  action: 'accept' | 'decline'
): Promise<InviteResponse> {
  return api.post<InviteResponse>(`/groups/${groupId}/invite/respond`, { action });
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
  size = 10
): Promise<PaginatedResponse<Post>> {
  return api.get<PaginatedResponse<Post>>(`/groups/${groupId}/posts`, { page, limit: size });
}

export async function getGroupMembers(
  groupId: string,
  page = 1,
  size = 20
): Promise<PaginatedResponse<GroupMember>> {
  return api.get<PaginatedResponse<GroupMember>>(`/groups/${groupId}/members`, {
    page,
    limit: size,
  });
}

export async function getPendingJoinRequests(groupId: string): Promise<GroupJoinRequest[]> {
  return api.get<GroupJoinRequest[]>(`/groups/${groupId}/requests/pending`);
}

// ─── Events ───────────────────────────────────────────────────────────────────

export async function createEvent(
  groupId: string,
  data: { title: string; description: string; eventDate: string; options?: string[] }
): Promise<Event> {
  return api.post<Event>(`/groups/${groupId}/events`, data);
}

export async function updateEvent(
  groupId: string,
  eventId: string,
  data: { title: string; description: string; eventDate: string }
): Promise<Event> {
  return api.put<Event>(`/groups/${groupId}/events/${eventId}`, data);
}

export async function getGroupEvents(
  groupId: string,
  cursor?: string
): Promise<GroupEventsResponse> {
  return api.get<GroupEventsResponse>(`/groups/${groupId}/events`, { cursor, size: 10 });
}

export async function respondToEvent(eventId: string, response: string): Promise<void> {
  return api.post<void>(`/events/${eventId}/respond`, { response });
}

export async function getEventRSVPs(eventId: string): Promise<EventRSVPsResponse> {
  return api.get<EventRSVPsResponse>(`/events/${eventId}/rsvps`);
}

// ─── Chat ─────────────────────────────────────────────────────────────────────

export async function getChats(): Promise<Chat[]> {
  return api.get<Chat[]>('/chat/users');
}

/**
 * /chat/start returns the raw chat row ({id, user_one_id, ...}), which has no
 * participants. Callers must resolve the full conversation via getChats().
 */
export async function startChat(userId: string): Promise<{ id: string }> {
  return api.post<{ id: string }>('/chat/start', { userId });
}

export async function getChatMessages(chatId: string): Promise<ChatMessage[]> {
  return api.get<ChatMessage[]>('/chat/history', { chatId });
}

export async function getGroupChatHistory(
  groupId: string,
  limit = 50
): Promise<GroupChatMessageWire[]> {
  return api.get<GroupChatMessageWire[]>(`/groups/${groupId}/chat/messages`, { limit });
}

/** Groups the authenticated user belongs to (for the chat widget's Groups tab). */
export async function getMyGroups(page = 1): Promise<PaginatedResponse<Group>> {
  return api.get<PaginatedResponse<Group>>('/groups/mine', { page, limit: 20 });
}

/** How many of a group's members are currently online. Only members may call it. */
export async function getGroupPresence(groupId: string): Promise<GroupPresence> {
  return api.get<GroupPresence>(`/groups/${groupId}/presence`);
}

// ─── Notifications ───────────────────────────────────────────────────────────

async function notifFetch<T = unknown>(path: string, options: RequestInit = {}): Promise<T> {
  const response = await fetch(NOTIFICATIONS_BASE + path, {
    ...options,
    credentials: 'include',
    headers: {
      ...(options.body ? { 'Content-Type': 'application/json' } : {}),
      ...(options.headers as Record<string, string>),
    },
  });

  if (response.status === 401) {
    handleUnauthorized();
    throw new ApiError(401, 'Session expired');
  }

  if (!response.ok) {
    let message: string;
    const text = await response.text();
    try {
      const errBody = text ? JSON.parse(text) : {};
      message = errBody?.error || errBody?.message || `HTTP ${response.status}`;
    } catch {
      message = text || `HTTP ${response.status}`;
    }
    throw new ApiError(response.status, message);
  }

  const text = await response.text();
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

export async function getNotifications(limit = 50): Promise<NotificationsResponse> {
  return notifFetch<NotificationsResponse>(`/notifications?limit=${limit}`);
}

export async function getUnreadNotificationCount(): Promise<{ count: number }> {
  return notifFetch<{ count: number }>('/notifications/unread-count');
}

export async function markNotificationRead(notificationId: number): Promise<void> {
  return notifFetch<void>(`/notifications/read?id=${notificationId}`, { method: 'PATCH' });
}

export async function markAllNotificationsRead(): Promise<void> {
  return notifFetch<void>('/notifications/read-all', { method: 'PATCH' });
}
