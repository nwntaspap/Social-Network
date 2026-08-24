/**
 * lib/notificationStream.ts
 *
 * Minimal singleton EventSource client for the notifications service SSE
 * stream. The stream pushes:
 *   - event "connected"        → data {"type":"connected"}
 *   - (unnamed) message        → data {"type":"unread_count","count":N}
 *   - event "notification"     → data {... store.Notification JSON ...}
 *
 * Unlike the chat socket, callers do NOT keep a list of notifications here —
 * the stream only forwards each event to subscribers. Rendering and removal
 * (when deleted=true) is the responsibility of the subscriber.
 *
 * connect()/disconnect() are reference-counted so the navbar bell and any
 * page-level consumers can share one stream without tearing it down.
 */

import type { Notification } from './types';

// Proxied through Next.js (see next.config.ts rewrites) so the session
// cookie stays first-party and no CORS setup is required.
const NOTIFICATIONS_BASE = '/notifications-api/api/v1';

type NotificationHandler = (notification: Notification) => void;
type UnreadCountHandler = (count: number) => void;
type ConnectedHandler = () => void;

const RETRY_DELAY_MS = 5000;

class NotificationStream {
  private es: EventSource | null = null;
  private connectCount = 0;
  private shouldReconnect = false;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;

  private notificationHandlers = new Set<NotificationHandler>();
  private unreadHandlers = new Set<UnreadCountHandler>();
  private connectedHandlers = new Set<ConnectedHandler>();

  /** Subscribe to live notifications. Returns an unsubscribe function. */
  onNotification(handler: NotificationHandler): () => void {
    this.notificationHandlers.add(handler);
    return () => this.notificationHandlers.delete(handler);
  }

  /** Subscribe to unread-count updates. Returns an unsubscribe function. */
  onUnreadCount(handler: UnreadCountHandler): () => void {
    this.unreadHandlers.add(handler);
    return () => this.unreadHandlers.delete(handler);
  }

  /** Subscribe to stream-ready events. Returns an unsubscribe function. */
  onConnected(handler: ConnectedHandler): () => void {
    this.connectedHandlers.add(handler);
    return () => this.connectedHandlers.delete(handler);
  }

  /** Open the stream. Safe to call from multiple components (ref-counted). */
  connect(): void {
    this.connectCount += 1;
    if (this.es) return;
    this.open();
  }

  /** Release a connect() call. The stream closes when the count reaches zero. */
  disconnect(): void {
    this.connectCount = Math.max(0, this.connectCount - 1);
    if (this.connectCount > 0) return;
    this.teardown();
  }

  private open(): void {
    this.shouldReconnect = true;

    const es = new EventSource(`${NOTIFICATIONS_BASE}/notifications/stream`, {
      withCredentials: true,
    });
    this.es = es;

    es.addEventListener('connected', () => {
      this.connectedHandlers.forEach((fn) => fn());
    });

    // The backend emits the initial unread count as a bare data frame.
    es.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data as string);
        if (data?.type === 'unread_count' && typeof data.count === 'number') {
          this.unreadHandlers.forEach((fn) => fn(data.count));
        }
      } catch {
        // malformed frame — ignore
      }
    };

    es.addEventListener('notification', (event) => {
      let notification: Notification;
      try {
        notification = JSON.parse(event.data as string);
      } catch {
        return;
      }
      if (!notification || typeof notification.id === 'undefined') return;
      this.notificationHandlers.forEach((fn) => fn(notification));
    });

    es.onerror = () => {
      // Surface connection failures — silent retries hide broken wiring.
      console.warn(
        `[notificationStream] error (readyState=${es.readyState}); ` +
          (this.connectCount > 0 ? 'retrying shortly' : 'stream closed')
      );
      es.close();
      if (this.es === es) this.es = null;

      if (this.shouldReconnect && this.connectCount > 0) {
        if (this.reconnectTimer) clearTimeout(this.reconnectTimer);
        this.reconnectTimer = setTimeout(() => {
          if (this.connectCount > 0) this.open();
        }, RETRY_DELAY_MS);
      }
    };
  }

  private teardown(): void {
    this.shouldReconnect = false;
    if (this.reconnectTimer) clearTimeout(this.reconnectTimer);
    this.reconnectTimer = null;

    if (this.es) {
      this.es.close();
      this.es = null;
    }
  }
}

export const notificationStream = new NotificationStream();
