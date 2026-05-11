/**
 * api.js
 *
 * Centralised API client. Every fetch to the backend goes through here.
 * Mirrors what the Go BFF's newRequest / newRequestWithCookies helpers did,
 * but running in the browser instead.
 *
 * The backend always wraps successful responses as:
 *   { "data": { ...payload... } }
 *
 * On error the backend returns:
 *   { "error": "message" }  (with a non-2xx status)
 */

const API_BASE = '/api/v1';

/**
 * Generic fetch wrapper.
 * Returns the unwrapped `data` field from the backend envelope.
 * Throws an ApiError on non-2xx responses.
 */
async function apiFetch(path, options = {}) {
  const url = API_BASE + path;
  const hasBody = options.body !== undefined && options.body !== null;

  const defaultOptions = {
    headers: hasBody ? { 'Content-Type': 'application/json' } : {},
    credentials: 'include',
  };

  const mergedOptions = {
    ...defaultOptions,
    ...options,
    headers: { ...defaultOptions.headers, ...(options.headers || {}) },
  };

  const response = await fetch(url, mergedOptions);

  const text = await response.text();

  if (!response.ok) {
    let message;
    try {
      const errBody = text ? JSON.parse(text) : {};
      message = errBody?.error || errBody?.message || `HTTP ${response.status}`;
    } catch {
      message = text || `HTTP ${response.status}`;
    }
    throw new ApiError(response.status, message);
  }

  if (!text) return {};

  let body;
  try {
    body = JSON.parse(text);
  } catch {
    throw new ApiError(response.status, 'Failed to parse server response');
  }

  return body?.data !== undefined ? body.data : body;
}

// ─── Convenience methods ──────────────────────────────────────────────────────

export const api = {
  get(path, params) {
    let fullPath = path;
    if (params) {
      const qs = new URLSearchParams(
        // Filter out empty values so we don't send ?search=&page=0 noise
        Object.fromEntries(
          Object.entries(params).filter(([, v]) => v !== '' && v !== null && v !== undefined)
        )
      ).toString();
      if (qs) fullPath += '?' + qs;
    }
    return apiFetch(fullPath, { method: 'GET' });
  },

  post(path, body) {
    return apiFetch(path, {
      method: 'POST',
      body: JSON.stringify(body),
    });
  },

  put(path, body) {
    return apiFetch(path, {
      method: 'PUT',
      body: JSON.stringify(body),
    });
  },

  delete(path, body) {
    return apiFetch(path, {
      method: 'DELETE',
      ...(body !== undefined && { body: JSON.stringify(body) }),
    });
  },
};

// ─── Error class ─────────────────────────────────────────────────────────────

export class ApiError extends Error {
  constructor(status, message) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }

  get isUnauthorized() {
    return this.status === 401;
  }

  get isForbidden() {
    return this.status === 403;
  }

  get isNotFound() {
    return this.status === 404;
  }

  get isTooManyRequests() {
    return this.status === 429;
  }
}

// ─── Auth ─────────────────────────────────────────────────────────────────────

/**
 * Fetches the currently logged-in user from /me.
 * Returns the user object, or null if not authenticated.
 * This replaces the Go BFF's AuthMiddleware + getCurrentUser logic.
 */
export async function fetchCurrentUser() {
  try {
    return await api.get('/me');
  } catch (err) {
    if (err instanceof ApiError && err.isUnauthorized) {
      return null;
    }
    // Any other error (network, 500, etc.) — treat as not logged in
    // but log it so it's visible during development.
    console.warn('Could not resolve current user:', err.message);
    return null;
  }
}

// ─── Categories ───────────────────────────────────────────────────────────────

/**
 * Fetches all categories with their latest topics.
 * Mirrors the BFF's defaultCategoriesOptions + createURLWithParams logic.
 */
export async function fetchCategories(params = {}) {
  const defaults = {
    order_by: 'created_at',
    order: 'desc',
    search: '',
    page: 1,
    page_size: 20,
  };
  return api.get('/categories/all', { ...defaults, ...params });
}

// ─── Topics ───────────────────────────────────────────────────────────────────

export async function fetchTopics(params = {}) {
  return api.get('/topics/all', params);
}

export async function fetchTopic(id) {
  return api.get('/topic', { id });
}

export async function createTopic(body) {
  return api.post('/topics/create', body);
}

export async function updateTopic(body) {
  return api.post('/topics/update', body);
}

export async function deleteTopic(body) {
  return api.post('/topics/delete', body);
}

// ─── Comments ─────────────────────────────────────────────────────────────────
export async function fetchComment(id) {
  return api.get('/comments/get', { id });
}

export async function fetchCommentsByTopic(topicId) {
  return api.get('/comments/topic', { topic_id: topicId });
}

export async function createComment(body) {
  return api.post('/comments/create', body);
}

export async function updateComment(body) {
  return api.post('/comments/update', body);
}

export async function deleteComment(body) {
  return api.post('/comments/delete', body);
}

// ─── Votes ────────────────────────────────────────────────────────────────────

export async function castVote(body) {
  return api.post('/vote/cast', body);
}

export async function deleteVote(body) {
  return api.delete('/vote/delete', body);
}

export async function fetchVoteCounts(params) {
  return api.get('/vote/counts', params);
}

// ─── Notifications ────────────────────────────────────────────────────────────

export async function fetchNotifications() {
  return api.get('/notifications');
}

export async function fetchUnreadCount() {
  return api.get('/notifications/unread-count');
}

export async function markNotificationRead(id) {
  return api.post(`/notifications/mark-read?id=${id}`, null);
}

export async function markAllNotificationsRead() {
  return api.post('/notifications/mark-all-read', {});
}

// ─── Activity ─────────────────────────────────────────────────────────────────

export async function fetchUserActivity() {
  return api.get('/user/activity');
}

// ─── Chat ─────────────────────────────────────────────────────────────────────

/**
 * Fetches list of users available for chat, sorted by last message activity.
 * Returns array of ChatUser objects with online status and last message time.
 */
export async function fetchChatUsers() {
  return api.get('/chat/users');
}

/**
 * Initializes or retrieves an existing chat with another user.
 * Returns the Chat object with chat_id that should be used for messaging.
 */
export async function initializeChat(userId) {
  return api.post('/chat/init', { user_id: userId });
}
