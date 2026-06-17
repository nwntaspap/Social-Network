import { Category, Topic, Comment, ChatUser, Chat } from './types';
/**
 * lib/api.ts
 *
 * Centralised API client for the Next.js frontend.
 * Every fetch to the backend goes through here.
 *
 * The backend always wraps successful responses as:
 *   { "data": { ...payload... } }
 *
 * On error the backend returns:
 *   { "error": "message" }  (with a non-2xx status)
 */

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

  // Unwrap the { data: ... } envelope
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

// ─── Categories ───────────────────────────────────────────────────────────────

interface FetchCategoriesParams {
  order_by?: string;
  order?: string;
  search?: string;
  page?: number;
  page_size?: number;
}

export async function fetchCategories(params: FetchCategoriesParams = {}): Promise<Category[]> {
  const defaults: FetchCategoriesParams = {
    order_by: 'created_at',
    order: 'desc',
    search: '',
    page: 1,
    page_size: 20,
  };
  return api.get<Category[]>('/categories/all', { ...defaults, ...params });
}

// ─── Topics ───────────────────────────────────────────────────────────────────

export async function fetchTopics(params?: QueryParams): Promise<Topic[]> {
  return api.get<Topic[]>('/topics/all', params);
}

export async function fetchTopic(id: number): Promise<Topic> {
  return api.get<Topic>('/topic', { id });
}

export async function createTopic(body: Partial<Topic>): Promise<Topic> {
  return api.post<Topic>('/topics/create', body);
}

export async function updateTopic(body: Partial<Topic> & { id: number }): Promise<Topic> {
  return api.post<Topic>('/topics/update', body);
}

export async function deleteTopic(body: { id: number }): Promise<void> {
  return api.post<void>('/topics/delete', body);
}

// ─── Comments ─────────────────────────────────────────────────────────────────

export async function fetchComment(id: number): Promise<Comment> {
  return api.get<Comment>('/comments/get', { id });
}

export async function fetchCommentsByTopic(topicId: number): Promise<Comment[]> {
  return api.get<Comment[]>('/comments/topic', { topic_id: topicId });
}

export async function createComment(body: Partial<Comment>): Promise<Comment> {
  return api.post<Comment>('/comments/create', body);
}

export async function updateComment(body: Partial<Comment> & { id: number }): Promise<Comment> {
  return api.post<Comment>('/comments/update', body);
}

export async function deleteComment(body: { id: number }): Promise<void> {
  return api.post<void>('/comments/delete', body);
}

// ─── Votes ────────────────────────────────────────────────────────────────────

interface VoteBody {
  target_type: 'topic' | 'comment';
  target_id: number;
  vote_type: 'up' | 'down';
}

type VoteCountsParams = {
  target_type: 'topic' | 'comment';
  target_ids: number[];
};

interface VoteCountResult {
  target_id: number;
  upvotes: number;
  downvotes: number;
}

export async function castVote(body: VoteBody): Promise<void> {
  return api.post<void>('/vote/cast', body);
}

export async function deleteVote(body: { target_type: string; target_id: number }): Promise<void> {
  return api.delete<void>('/vote/delete', body);
}

export async function fetchVoteCounts(params: VoteCountsParams): Promise<VoteCountResult[]> {
  return api.get<VoteCountResult[]>('/vote/counts', params);
}

// ─── Notifications ────────────────────────────────────────────────────────────

export interface Notification {
  id: number;
  type: string;
  message: string;
  is_read: boolean;
  created_at: string;
}

export async function fetchNotifications(): Promise<Notification[]> {
  return api.get<Notification[]>('/notifications');
}

export async function fetchUnreadCount(): Promise<{ count: number }> {
  return api.get<{ count: number }>('/notifications/unread-count');
}

export async function markNotificationRead(id: number): Promise<void> {
  return api.post<void>(`/notifications/mark-read?id=${id}`);
}

export async function markAllNotificationsRead(): Promise<void> {
  return api.post<void>('/notifications/mark-all-read');
}

// ─── Activity ─────────────────────────────────────────────────────────────────

export interface ActivityItem {
  id: number;
  type: string;
  description: string;
  created_at: string;
}

export async function fetchUserActivity(): Promise<ActivityItem[]> {
  return api.get<ActivityItem[]>('/user/activity');
}

// ─── Chat ─────────────────────────────────────────────────────────────────────

export async function fetchChatUsers(): Promise<ChatUser[]> {
  return api.get<ChatUser[]>('/chat/users');
}

export async function initializeChat(userId: number): Promise<Chat> {
  return api.post<Chat>('/chat/init', { user_id: userId });
}
