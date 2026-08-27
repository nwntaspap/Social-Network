/**
 * lib/ws.ts
 *
 * Minimal singleton WebSocket client for the realtime API.
 *
 * Wire format mirrors the backend realtime.Envelope:
 *   { type: string, request_id?: string, payload: any }
 *
 * Auth is cookie-based (credentials travel with the upgrade request since
 * the socket is opened against the same origin /api/v1/ws proxy).
 *
 * connect()/disconnect() are reference-counted so a global chat widget and
 * page-level components can share one socket without tearing it down while
 * another consumer still needs it. Only a single WebSocket connection is ever
 * created; redundant connect() calls while a socket is OPEN or CONNECTING are
 * no-ops, which prevents duplicate deliveries.
 */

export type WsMessageHandler = (payload: unknown, requestId?: string) => void;

type WsSubscribers = Map<string, Set<WsMessageHandler>>;

const WS_PATH = '/api/v1/ws';

function wsUrl(): string {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  return `${protocol}//${window.location.host}${WS_PATH}`;
}

class ChatSocket {
  private socket: WebSocket | null = null;
  private connecting = false;
  private subscribers: WsSubscribers = new Map();
  private listeners: Array<() => void> = [];
  private shouldReconnect = false;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private reconnectAttempts = 0;
  private connectCount = 0;

  private connected = false;

  /** Subscribe to a message type. Returns an unsubscribe function. */
  on(type: string, handler: WsMessageHandler): () => void {
    if (!this.subscribers.has(type)) {
      this.subscribers.set(type, new Set());
    }
    this.subscribers.get(type)!.add(handler);
    return () => {
      this.subscribers.get(type)?.delete(handler);
    };
  }

  /** Subscribe to connection state changes (open/close). Returns unsubscribe. */
  onConnection(listener: () => void): () => void {
    this.listeners.push(listener);
    return () => {
      this.listeners = this.listeners.filter((l) => l !== listener);
    };
  }

  isConnected(): boolean {
    return this.connected;
  }

  /** Open the socket. Safe to call from multiple components (ref-counted). */
  connect(): void {
    this.connectCount += 1;
    if (this.socket) return;
    this.open();
  }

  /** Release a connect() call. The socket closes when the count reaches zero. */
  disconnect(): void {
    this.connectCount = Math.max(0, this.connectCount - 1);
    if (this.connectCount > 0) return;
    this.teardown();
  }

  private open(): void {
    this.shouldReconnect = true;
    this.connecting = true;

    const socket = new WebSocket(wsUrl());

    socket.onopen = () => {
      if (this.socket !== socket) return;
      this.connecting = false;
      this.connected = true;
      this.reconnectAttempts = 0;
      this.notifyConnection();
    };

    socket.onmessage = (event) => {
      let envelope: { type: string; request_id?: string; payload: unknown };
      try {
        envelope = JSON.parse(event.data as string);
      } catch {
        return;
      }
      if (!envelope || typeof envelope.type !== 'string') return;
      const handlers = this.subscribers.get(envelope.type);
      if (handlers) {
        handlers.forEach((handler) => handler(envelope.payload, envelope.request_id));
      }
    };

    socket.onclose = () => {
      if (this.socket === socket) this.socket = null;
      this.connecting = false;
      this.connected = false;
      this.notifyConnection();
      if (this.connectCount > 0 && this.shouldReconnect) {
        this.scheduleReconnect();
      }
    };

    socket.onerror = () => {
      socket.close();
    };

    this.socket = socket;
  }

  private teardown(): void {
    this.shouldReconnect = false;
    this.connecting = false;
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    if (this.socket) {
      const closing = this.socket;
      this.socket = null;
      closing.close();
    }
    this.connected = false;
  }

  /** Send an envelope to the server. Drops the message if the socket is closed. */
  send(type: string, payload?: unknown, requestId?: string): void {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) return;
    const envelope: Record<string, unknown> = { type };
    if (requestId) envelope.request_id = requestId;
    envelope.payload = payload ?? {};
    this.socket.send(JSON.stringify(envelope));
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer) return;
    const backoff = Math.min(1000 * 2 ** this.reconnectAttempts, 15000);
    this.reconnectAttempts += 1;
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      if (this.connectCount > 0) {
        this.open();
      }
    }, backoff);
  }

  private notifyConnection(): void {
    this.listeners.forEach((listener) => listener());
  }
}

/** Shared singleton used across the app. */
export const chatSocket = new ChatSocket();
