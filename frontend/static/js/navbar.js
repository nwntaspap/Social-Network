/**
 * navbar.js
 *
 * Renders the navbar into #navbar-root and manages the notification bell.
 *
 * This replaces:
 *   - frontend/html/partials/navbar.html  (the template)
 *   - frontend/static/js/notifications.js (the notification bell JS)
 *   - The Go BFF's AuthMiddleware passing .User into templates
 *
 * Call renderNavbar(user) on every route change so the nav reflects the
 * current auth state without a full page reload.
 */

import { escapeHTML } from './helpers.js';
import {
  fetchNotifications,
  fetchUnreadCount,
  fetchComment,
  markNotificationRead,
  markAllNotificationsRead,
} from './api.js';
import { initChat } from './chat.js';

// ─── Public API ───────────────────────────────────────────────────────────────

/**
 * Renders the navbar for the given user (or null for guests).
 * Attaches notification bell logic if a user is present.
 *
 * @param {object|null} user
 */
export function renderNavbar(user) {
  const root = document.getElementById('navbar-root');
  if (!root) return;

  root.innerHTML = buildNavbarHTML(user);

  if (user) {
    initUserMenu();
    initNotifications(user);
    initChat(user);
    initActiveNavLink();
  }
}

// ─── HTML builders ────────────────────────────────────────────────────────────

function buildNavbarHTML(user) {
  return /* html */ `
    <header>
      <nav class="navbar">
        <div class="nav-container">
          <div class="logo">
            <a class="logo-link" href="/" data-link>
              <div class="logo-icon">
                <img src="/static/images/icons/logo-icon.png" alt="Logo Icon" />
              </div>
              <span class="logo-title">Forum</span>
            </a>
          </div>

          ${user ? buildLoggedInNav(user) : buildGuestNav()}
        </div>
      </nav>
    </header>
  `;
}

function buildLoggedInNav(user) {
  console.log(user.avatar_url)
  const avatarSrc =
    user.avatar_url || user.AvatarURL
      ? escapeHTML(user.avatar_url || user.AvatarURL)
      : '/static/images/user-avatar.png';

  const username = escapeHTML(user.username || user.Username || '');

  return /* html */ `
    <div class="welcome-box">
      <div class="welcome-user-box">
        <div class="welcome-user-text">
          <span class="welcome-user">Welcome,</span>
          <span class="welcome-user-name">${username}</span>
        </div>
        <div class="user-menu-wrapper">
          <button
            type="button"
            class="user-avatar-link user-avatar-trigger"
            id="userMenuTrigger"
            aria-label="Open account menu"
            aria-expanded="false"
          >
            <span class="user-avatar">
              <img src="${avatarSrc}" alt="User Avatar" />
            </span>
          </button>
          <div class="user-menu-dropdown" id="userMenuDropdown" style="display:none">
            <a href="/activity" class="user-menu-item" data-link>Profile</a>
            <a href="/account/settings" class="user-menu-item" data-link>Account Settings</a>
          </div>
        </div>
      </div>

      <ul class="nav-links">
        <!-- Notification bell -->
        <li class="nav-icon notification-wrapper">
          <button
            class="nav-icon-link notification-bell"
            id="notificationBell"
            aria-label="Notifications"
          >
            <svg width="24" height="24" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
              <path
                d="M12 22C13.1 22 14 21.1 14 20H10C10 21.1 10.9 22 12 22ZM18 16V11C18 7.93 16.37 5.36 13.5 4.68V4C13.5 3.17 12.83 2.5 12 2.5C11.17 2.5 10.5 3.17 10.5 4V4.68C7.64 5.36 6 7.92 6 11V16L4 18V19H20V18L18 16Z"
                fill="currentColor"
              />
            </svg>
            <span class="notification-badge" id="notificationBadge" style="display:none">0</span>
          </button>

          <div class="notification-dropdown" id="notificationDropdown" style="display:none">
            <div class="notification-header">
              <h3>Notifications</h3>
              <button class="mark-all-read-btn" id="markAllReadBtn">Mark all as read</button>
            </div>
            <div class="notification-list" id="notificationList">
              <p class="notification-empty">No notifications yet</p>
            </div>
          </div>
        </li>

        <li class="nav-link nav-link-create">
          <a href="/topics/create" data-link>New Post</a>
        </li>
        <li class="nav-link">
          <a href="/logout" data-link>Logout</a>
        </li>
      </ul>
    </div>
  `;
}

function buildGuestNav() {
  return /* html */ `
    <ul class="nav-links">
      <li class="nav-link">
        <a href="/login" data-link>Login</a>
      </li>
      <li class="nav-link">
        <a href="/register" data-link>Register</a>
      </li>
    </ul>
  `;
}

// ─── Active nav link highlighting ────────────────────────────────────────────

function initActiveNavLink() {
  const path = window.location.pathname;
  document.querySelectorAll('.nav-link a[data-link]').forEach((a) => {
    a.closest('.nav-link')?.classList.toggle('active', a.getAttribute('href') === path);
  });
}

function initUserMenu() {
  const trigger = document.getElementById('userMenuTrigger');
  const dropdown = document.getElementById('userMenuDropdown');
  const notificationDropdown = document.getElementById('notificationDropdown');

  if (!trigger || !dropdown) return;

  trigger.addEventListener('click', (e) => {
    e.stopPropagation();
    const isOpen = dropdown.style.display !== 'none';

    if (notificationDropdown) {
      notificationDropdown.style.display = 'none';
    }

    dropdown.style.display = isOpen ? 'none' : 'block';
    trigger.setAttribute('aria-expanded', String(!isOpen));
  });

  dropdown.querySelectorAll('a[data-link]').forEach((link) => {
    link.addEventListener('click', () => {
      dropdown.style.display = 'none';
      trigger.setAttribute('aria-expanded', 'false');
    });
  });

  document.addEventListener('click', (e) => {
    if (!trigger.contains(e.target) && !dropdown.contains(e.target)) {
      dropdown.style.display = 'none';
      trigger.setAttribute('aria-expanded', 'false');
    }
  });
}

// ─── Notification bell (mirrors frontend/static/js/notifications.js) ─────────

let _notificationSSE = null;

function initNotifications() {
  const bell = document.getElementById('notificationBell');
  const dropdown = document.getElementById('notificationDropdown');
  const markAllBtn = document.getElementById('markAllReadBtn');
  const userMenuDropdown = document.getElementById('userMenuDropdown');
  const userMenuTrigger = document.getElementById('userMenuTrigger');

  if (!bell || !dropdown) return;

  // Load initial unread count
  refreshUnreadCount();

  // Toggle dropdown on bell click
  bell.addEventListener('click', (e) => {
    e.stopPropagation();
    const isOpen = dropdown.style.display !== 'none';
    if (isOpen) {
      dropdown.style.display = 'none';
    } else {
      if (userMenuDropdown) {
        userMenuDropdown.style.display = 'none';
      }
      userMenuTrigger?.setAttribute('aria-expanded', 'false');
      dropdown.style.display = 'flex';
      dropdown.style.flexDirection = 'column';
      loadNotifications();
    }
  });

  // Close dropdown on outside click
  document.addEventListener('click', (e) => {
    if (!bell.contains(e.target) && !dropdown.contains(e.target)) {
      dropdown.style.display = 'none';
    }
  });

  // Mark all as read
  markAllBtn?.addEventListener('click', async () => {
    try {
      await markAllNotificationsRead();

      // Update every unread item in the list immediately
      document.querySelectorAll('.notification-item.unread').forEach((item) => {
        item.classList.remove('unread');
        item.dataset.read = 'true';
        item.querySelector('.notification-unread-dot')?.remove();
      });

      updateBadge(0);
    } catch (err) {
      console.error('Failed to mark all as read:', err);
    }
  });

  // Start SSE stream for live notification count updates
  startNotificationStream();
}

async function refreshUnreadCount() {
  try {
    const data = await fetchUnreadCount();
    const count = data?.count ?? data?.unread_count ?? 0;
    updateBadge(count);
  } catch {
    // Not logged in or network error — badge stays hidden
  }
}

function updateBadge(count) {
  const badge = document.getElementById('notificationBadge');
  if (!badge) return;
  if (count > 0) {
    badge.textContent = count > 99 ? '99+' : String(count);
    badge.style.display = 'block';
  } else {
    badge.style.display = 'none';
  }
}

async function loadNotifications() {
  const list = document.getElementById('notificationList');
  if (!list) return;

  list.innerHTML = `<p class="notification-empty">Loading…</p>`;

  try {
    const data = await fetchNotifications();
    const notifications = data?.notifications ?? data ?? [];

    if (!notifications.length) {
      list.innerHTML = `<p class="notification-empty">No notifications yet</p>`;
      return;
    }

    list.innerHTML = notifications.map(buildNotificationItemHTML).join('');

    // Click to mark as read + navigate
    list.querySelectorAll('.notification-item').forEach((item) => {
      item.addEventListener('click', async () => {
        const id = item.dataset.id;
        const isRead = item.dataset.read === 'true';
        const relatedType = item.dataset.relatedType;
        const relatedId = item.dataset.relatedId;

        // Update UI immediately
        if (!isRead) {
          item.classList.remove('unread');
          item.dataset.read = 'true';
          item.querySelector('.notification-unread-dot')?.remove();

          try {
            await markNotificationRead(id);
            refreshUnreadCount();
          } catch (err) {
            console.error('Failed to mark notification as read:', err);
          }
        }

        // Navigate to related content
        if (relatedId) {
          const { navigate } = await import('./router.js');

          if (relatedType === 'topic') {
            navigate(`/topic/${relatedId}`);
          } else if (relatedType === 'comment') {
            try {
              const comment = await fetchComment(relatedId);
              const topicId = comment?.topicId;
              if (topicId) {
                navigate(`/topic/${topicId}`);
              }
            } catch (err) {
              console.error('Failed to resolve comment topic:', err);
            }
          }

          document.getElementById('notificationDropdown').style.display = 'none';
        }
      });
    });
  } catch (err) {
    list.innerHTML = `<p class="notification-empty">Failed to load notifications</p>`;
    console.error(err);
  }
}

function buildNotificationItemHTML(n) {
  const isUnread = !n.isRead;
  const icon = notificationIcon(n.type);
  const title = escapeHTML(n.title || '');
  const message = escapeHTML(n.message || '');
  const time = formatNotificationTime(n.createdAt);
  const id = escapeHTML(String(n.id ?? ''));
  const relatedType = escapeHTML(n.relatedType ?? '');
  const relatedId = escapeHTML(n.relatedId ?? '');

  return /* html */ `
    <div class="notification-item ${isUnread ? 'unread' : ''}"
         data-id="${id}"
         data-read="${!isUnread}"
         data-related-type="${relatedType}"
         data-related-id="${relatedId}">
      <div class="notification-icon ${escapeHTML(n.type || '')}">
        ${icon}
      </div>
      <div class="notification-content">
        <div class="notification-title">${title}</div>
        <div class="notification-message">${message}</div>
        <div class="notification-time">${time}</div>
      </div>
      ${isUnread ? `<div class="notification-unread-dot"></div>` : ''}
    </div>
  `;
}

function notificationIcon(type) {
  switch (type) {
    case 'reply':
      return '💬';
    case 'like':
      return '💚';
    case 'dislike':
      return '🤮';
    case 'mention':
      return '📣';
    default:
      return '🔔';
  }
}

function formatNotificationTime(isoString) {
  if (!isoString) return '';
  const date = new Date(isoString);
  if (isNaN(date)) return '';
  const diffMin = Math.floor((Date.now() - date) / 60000);
  if (diffMin < 1) return 'just now';
  if (diffMin < 60) return `${diffMin}m ago`;
  const diffHr = Math.floor(diffMin / 60);
  if (diffHr < 24) return `${diffHr}h ago`;
  return date.toLocaleDateString('en-GB', { day: 'numeric', month: 'short' });
}

function startNotificationStream() {
  if (_notificationSSE) {
    _notificationSSE.close();
    _notificationSSE = null;
  }

  try {
    const es = new EventSource('/api/v1/notifications/stream', {
      withCredentials: true,
    });

    // Backend sends: event: connected
    es.addEventListener('connected', () => {
      // stream is live — nothing to do, unread_count arrives next
    });

    // Backend sends: event: unread_count  data: {"type":"unread_count","count":N}
    es.addEventListener('unread_count', (e) => {
      try {
        const data = JSON.parse(e.data);
        updateBadge(data.count ?? 0);
      } catch {
        // malformed frame — ignore
      }
    });

    // Backend sends: event: notification  data: {...notification object...}
    es.addEventListener('notification', () => {
      // A new notification arrived — refresh the badge count from the API
      // (same as before — avoids stale in-memory counts)
      refreshUnreadCount();
    });

    es.onerror = () => {
      es.close();
      _notificationSSE = null;
      // Browser will NOT auto-reconnect after we call close().
      // Retry after 5 s so the user doesn't lose live updates.
      setTimeout(startNotificationStream, 5000);
    };

    _notificationSSE = es;
  } catch (err) {
    console.warn('Could not start notification stream:', err);
  }
}

/**
 * Closes the SSE stream. Call this before logging out.
 */
export function closeNotificationStream() {
  if (_notificationSSE) {
    _notificationSSE.close();
    _notificationSSE = null;
  }
}
