/**
 * lib/notificationFormat.ts
 *
 * Pure helpers that turn a backend store.Notification into a human-readable
 * message, a navigation href, and an icon key.
 */

import type { Notification } from './types';

const ACTION_TYPES = new Set(['follow_request', 'group_invite', 'group_join_request']);

/** Human-readable message for a notification. */
export function getNotificationMessage(n: Notification): string {
  const name = n.actor_name || 'Someone';

  switch (n.type) {
    case 'follow':
      return `${name} followed you`;
    case 'follow_request':
      return `${name} wants to follow you`;
    case 'follow_accept':
      return `${name} accepted your follow request`;
    case 'follow_declined':
      return `${name} declined your follow request`;
    case 'like':
      return `${name} liked your ${n.resource_type === 'comment' ? 'comment' : 'post'}`;
    case 'dislike':
      return `${name} disliked your ${n.resource_type === 'comment' ? 'comment' : 'post'}`;
    case 'comment':
      return `${name} commented on your post`;
    case 'group_invite':
      return `${name} invited you to join a group`;
    case 'group_invite_removed':
      return `Your group invitation was removed`;
    case 'group_join_request':
      return `${name} wants to join your group`;
    case 'group_join_accept':
      return `${name} accepted your group join request`;
    case 'group_join_declined':
      return `${name} declined your group join request`;
    case 'event':
      return n.content_text || `${name} created a new event`;
    case 'post':
      return `${name} created a new post`;
    default:
      return (
        n.content_text || (n.actor_name ? `${name} sent you a notification` : 'New notification')
      );
  }
}

/** Navigation href for a notification, or null when the item has no target. */
export function getNotificationHref(n: Notification): string | null {
  switch (n.type) {
    case 'follow':
    case 'follow_accept':
    case 'follow_declined':
      return `/profile/${n.actor_id}`;
    case 'like':
    case 'dislike':
      if (n.resource_type === 'comment') return null; // resolved via getComment
      return `/post/${n.resource_id}`;
    case 'comment':
      return `/post/${n.resource_id}`;
    case 'group_invite':
    case 'group_join_request':
    case 'group_join_accept':
    case 'group_join_declined':
    case 'event':
      return n.resource_id ? `/groups/${n.resource_id}` : null;
    default:
      return n.resource_type === 'post' && n.resource_id ? `/post/${n.resource_id}` : null;
  }
}

/** True when the notification carries accept/decline actions. */
export function hasNotificationActions(n: Notification): boolean {
  return ACTION_TYPES.has(n.type);
}

/** Icon key mapped to a CSS class (see navbar.css .notification-icon.*). */
export function getNotificationIcon(n: Notification): string {
  switch (n.type) {
    case 'like':
      return 'like';
    case 'dislike':
      return 'dislike';
    case 'follow':
    case 'follow_request':
    case 'follow_accept':
      return 'follow';
    case 'group_invite':
    case 'group_invite_removed':
    case 'group_join_request':
    case 'group_join_accept':
    case 'group_join_declined':
      return 'group';
    case 'event':
      return 'event';
    case 'comment':
      return 'comment';
    default:
      return 'mention';
  }
}
