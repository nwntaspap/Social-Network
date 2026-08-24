'use client';

/**
 * components/features/notifications/NotificationBell.tsx
 *
 * Notification bell + dropdown for the navbar.
 *
 * Behavior:
 *  - Renders a live list of notifications; the DOM is the only store.
 *  - SSE "notification" events: deleted=false → prepend an item;
 *    deleted=true → remove the item by its data-notification-id.
 *  - Clicking an item marks it read (optimistically) and navigates to its
 *    target (comments are resolved to their post first).
 *  - follow_request / group_invite / group_join_request items show
 *    Accept/Decline buttons that call the backend and remove the item.
 */

import { useCallback, useEffect, useRef, useState } from 'react';
import { useRouter } from 'next/navigation';
import Image from 'next/image';
import { useAuth } from '@/context/AuthContext';
import {
  getNotifications,
  getUnreadNotificationCount,
  getComment,
  markNotificationRead,
  markAllNotificationsRead,
  handleFollowRequest,
  respondToGroupInvitation,
  handleJoinRequest,
} from '@/lib/api';
import { notificationStream } from '@/lib/notificationStream';
import {
  getNotificationHref,
  getNotificationIcon,
  getNotificationMessage,
  hasNotificationActions,
} from '@/lib/notificationFormat';
import { formatNotificationTime } from '@/lib/helpers';
import type { Notification } from '@/lib/types';

const ICON_EMOJI: Record<string, string> = {
  like: '💚',
  dislike: '🤮',
  follow: '👥',
  group: '👨‍👩‍👧‍👦',
  event: '📅',
  comment: '💬',
  mention: '📣',
};

interface ItemActions {
  onClick: (n: Notification) => void;
  onAction: (n: Notification, action: 'accept' | 'decline') => void;
}

export default function NotificationBell() {
  const { user } = useAuth();
  const router = useRouter();

  const [open, setOpen] = useState(false);
  const [unread, setUnread] = useState(0);
  const [loaded, setLoaded] = useState(false);

  const wrapperRef = useRef<HTMLDivElement>(null);
  const bellRef = useRef<HTMLButtonElement>(null);
  const listRef = useRef<HTMLDivElement>(null);
  const actionsRef = useRef<ItemActions>({ onClick: () => {}, onAction: () => {} });

  const loggedIn = !!user;

  // Close dropdown on outside click.
  useEffect(() => {
    function handleOutsideClick(e: MouseEvent) {
      if (
        wrapperRef.current &&
        !wrapperRef.current.contains(e.target as Node) &&
        !bellRef.current?.contains(e.target as Node)
      ) {
        setOpen(false);
      }
    }
    document.addEventListener('click', handleOutsideClick);
    return () => document.removeEventListener('click', handleOutsideClick);
  }, []);

  const refreshUnread = useCallback(async () => {
    if (!loggedIn) return;
    try {
      const data = await getUnreadNotificationCount();
      setUnread(data?.count ?? 0);
    } catch (err) {
      console.warn('[NotificationBell] unread-count refresh failed:', err);
    }
  }, [loggedIn]);

  // Initial badge count (fallback when the stream has not pushed one yet).
  useEffect(() => {
    if (!loggedIn) return;
    getUnreadNotificationCount()
      .then((data) => setUnread(data?.count ?? 0))
      .catch((err) => console.warn('[NotificationBell] unread-count fetch failed:', err));
  }, [loggedIn]);

  // Live stream: render new items and remove deleted ones by id.
  useEffect(() => {
    if (!loggedIn) return;

    const unsubNotification = notificationStream.onNotification((n) => {
      const list = listRef.current;
      if (!list) return;

      if (n.deleted) {
        const item = list.querySelector(`[data-notification-id="${n.id}"]`);
        if (item?.classList.contains('unread')) {
          setUnread((prev) => Math.max(0, prev - 1));
        }
        item?.remove();
        if (list.children.length === 0) renderEmpty(list);
        // Re-sync with the server count (covers items that were never rendered).
        refreshUnread();
        return;
      }

      // Drop duplicates that already arrived via a previous stream event.
      if (list.querySelector(`[data-notification-id="${n.id}"]`)) return;

      list.querySelector('.notification-empty')?.remove();
      list.prepend(buildItem(n, actionsRef.current));
      setUnread((prev) => prev + (n.is_read ? 0 : 1));
    });

    const unsubUnread = notificationStream.onUnreadCount((count) => setUnread(count));
    const unsubConnected = notificationStream.onConnected(refreshUnread);

    notificationStream.connect();

    return () => {
      unsubNotification();
      unsubUnread();
      unsubConnected();
      notificationStream.disconnect();
    };
  }, [loggedIn, refreshUnread]);

  async function loadNotifications() {
    if (loaded) return;
    const list = listRef.current;
    if (!list) return;

    list.innerHTML = '';
    renderEmpty(list, 'Loading…');

    try {
      const data = await getNotifications();
      const notifications = data?.notifications ?? [];
      list.innerHTML = '';

      if (notifications.length === 0) {
        renderEmpty(list);
        return;
      }
      notifications.forEach((n) => list.append(buildItem(n, actionsRef.current)));
    } catch {
      list.innerHTML = '';
      renderEmpty(list, 'Failed to load notifications');
    } finally {
      setLoaded(true);
    }
  }

  function toggleDropdown() {
    setOpen((prev) => {
      const next = !prev;
      if (next) {
        setLoaded(false);
        setTimeout(loadNotifications, 0);
      }
      return next;
    });
  }

  async function handleClick(n: Notification) {
    if (!n.is_read) await markRead(n);

    const href = await resolveHref(n);
    if (href) {
      router.push(href);
      setOpen(false);
    }
  }

  async function resolveHref(n: Notification): Promise<string | null> {
    const href = getNotificationHref(n);
    if (href) return href;
    // Comment votes need their post resolved.
    if ((n.type === 'like' || n.type === 'dislike') && n.resource_type === 'comment') {
      try {
        const comment = await getComment(Number(n.resource_id));
        if (comment?.postId) return `/post/${comment.postId}`;
      } catch {
        return null;
      }
    }
    return null;
  }

  async function markRead(n: Notification) {
    // Optimistic UI update.
    const list = listRef.current;
    list?.querySelector(`[data-notification-id="${n.id}"]`)?.classList.remove('unread');
    setUnread((prev) => Math.max(0, prev - 1));

    try {
      await markNotificationRead(n.id);
      refreshUnread();
    } catch {
      // Failed to persist — revert optimistic count via refresh.
      refreshUnread();
    }
  }

  async function handleAction(n: Notification, action: 'accept' | 'decline') {
    try {
      if (n.type === 'follow_request') {
        await handleFollowRequest(n.actor_id, action);
      } else if (n.type === 'group_invite') {
        await respondToGroupInvitation(n.resource_id, action);
      } else if (n.type === 'group_join_request') {
        await handleJoinRequest(n.join_request_id, action);
      }
    } catch {
      return;
    }

    const list = listRef.current;
    list?.querySelector(`[data-notification-id="${n.id}"]`)?.remove();
    if (!n.is_read) setUnread((prev) => Math.max(0, prev - 1));
    refreshUnread();
    if (list && list.children.length === 0) renderEmpty(list);
  }

  async function handleMarkAllRead() {
    try {
      await markAllNotificationsRead();
    } catch {
      return;
    }

    listRef.current?.querySelectorAll('.notification-item.unread').forEach((item) => {
      item.classList.remove('unread');
    });
    setUnread(0);
  }

  // Keep the latest handlers reachable from DOM nodes built in effects.
  useEffect(() => {
    actionsRef.current = { onClick: handleClick, onAction: handleAction };
  });

  return (
    <div className="notification-wrapper" ref={wrapperRef}>
      <button
        ref={bellRef}
        type="button"
        className="notification-bell"
        aria-label="Notifications"
        aria-expanded={open}
        onClick={toggleDropdown}
      >
        <Image
          src="/images/icons/notifications-icon.svg"
          alt="Notifications"
          width={24}
          height={24}
        />
        {unread > 0 && <span className="notification-badge">{unread > 99 ? '99+' : unread}</span>}
      </button>

      <div
        className="notification-dropdown"
        style={{ display: open ? 'flex' : 'none' }}
        onClick={(e) => e.stopPropagation()}
      >
        <div className="notification-header">
          <h3>Notifications</h3>
          <button type="button" className="mark-all-read-btn" onClick={handleMarkAllRead}>
            Mark all as read
          </button>
        </div>
        <div className="notification-list" ref={listRef} data-testid="notification-list" />
      </div>
    </div>
  );
}

function escapeHtml(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;');
}

function renderEmpty(list: HTMLElement, text = 'No notifications yet') {
  const p = document.createElement('p');
  p.className = 'notification-empty';
  p.textContent = text;
  list.append(p);
}

function buildItem(n: Notification, actions: ItemActions): HTMLDivElement {
  const div = document.createElement('div');
  div.className = `notification-item${n.is_read ? '' : ' unread'}`;
  div.dataset.notificationId = String(n.id);
  div.dataset.read = String(n.is_read);

  const iconKey = getNotificationIcon(n);
  const emoji = ICON_EMOJI[iconKey] ?? '🔔';
  const actionButtons = hasNotificationActions(n)
    ? `
        <div class="notification-actions">
          <button type="button" class="notification-btn accept" data-action="accept">Accept</button>
          <button type="button" class="notification-btn decline" data-action="decline">Decline</button>
        </div>`
    : '';

  div.innerHTML = `
      <div class="notification-icon ${iconKey}">
        <span class="notification-icon-emoji">${emoji}</span>
      </div>
      <div class="notification-content">
        <div class="notification-message">${escapeHtml(getNotificationMessage(n))}</div>
        <div class="notification-time">${formatNotificationTime(n.created_at)}</div>
        ${actionButtons}
      </div>
      ${n.is_read ? '' : '<div class="notification-unread-dot"></div>'}
    `;

  div.addEventListener('click', (e) => {
    const target = e.target as HTMLElement;
    const actionBtn = target.closest('[data-action]');
    if (actionBtn) {
      e.stopPropagation();
      const action = actionBtn.getAttribute('data-action') === 'accept' ? 'accept' : 'decline';
      actions.onAction(n, action);
      return;
    }
    actions.onClick(n);
  });

  return div;
}
