/**
 * lib/api.ts
 *
 * Typed API client — mirrors frontend/static/js/api.js
 *
 * All requests go to the Go backend at /api/v1/...
 * Cookies are sent automatically (credentials: 'include').
 *
 * The backend base URL is read from NEXT_PUBLIC_API_URL (set in .env.local).
 * Falls back to '' (same origin) so the dev proxy in next.config.ts can
 * forward /api/* to http://localhost:8080.
 */

const API_BASE = (process.env.NEXT_PUBLIC_API_URL ?? '') + '/api/v1';

// ─── Types ────────────────────────────────────────────────────────────────────

export interface User {
  id: string;
  username: string;
  email: string;
  avatar_url?: string;
  AvatarURL?: string;
  first_name?: string;
  last_name?: string;
}

export interface ApiError extends Error {
  status: number;
}

// ─── Core fetch wrapper ───────────────────────────────────────────────────────

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const url = `${API_BASE}${path}`;

  const res = await fetch(url, {
    method,
    credentials: 'include',
    headers: body ? { 'Content-Type': 'application/json' } : {},
    body: body ? JSON.stringify(body) : undefined,
  });

  // Try to parse JSON regardless of status so we can surface error messages
  let data: unknown;
  try {
    data = await res.json();
  } catch {
    data = null;
  }

  if (!res.ok) {
    const message =
      (data as Record<string, string>)?.error ||
      (data as Record<string, string>)?.message ||
      `Request failed: ${res.status}`;
    const err = new Error(message) as ApiError;
    err.status = res.status;
    throw err;
  }

  return data as T;
}

// ─── Public API surface ───────────────────────────────────────────────────────

export const api = {
  get: <T>(path: string) => request<T>('GET', path),
  post: <T>(path: string, body?: unknown) => request<T>('POST', path, body),
  put: <T>(path: string, body?: unknown) => request<T>('PUT', path, body),
  delete: <T>(path: string) => request<T>('DELETE', path),
};

// ─── Named fetch helpers (mirrors the named exports in api.js) ────────────────

export async function fetchCurrentUser(): Promise<User> {
  return api.get<User>('/me');
}

export async function fetchCategories() {
  return api.get('/categories/all');
}

export async function fetchNotifications() {
  return api.get('/notifications');
}

export async function fetchUnreadCount() {
  return api.get<{ count: number; unread_count?: number }>('/notifications/unread-count');
}

export async function markNotificationRead(id: string) {
  return api.put(`/notifications/${id}/read`);
}

export async function markAllNotificationsRead() {
  return api.put('/notifications/read-all');
}

export async function fetchComment(id: string) {
  return api.get(`/comments/${id}`);
}
