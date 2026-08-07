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
  private subscribers: WsSubscribers = new Map();
  private listeners: Array<() => void> = [];
  private shouldReconnect = false;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private reconnectAttempts = 0;

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

  connect(): void {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) return;
    this.shouldReconnect = true;

    const socket = new WebSocket(wsUrl());

    socket.onopen = () => {
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
      this.connected = false;
      this.socket = null;
      this.notifyConnection();
      if (this.shouldReconnect) {
        this.scheduleReconnect();
      }
    };

    socket.onerror = () => {
      socket.close();
    };

    this.socket = socket;
  }

  disconnect(): void {
    this.shouldReconnect = false;
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    if (this.socket) {
      this.socket.close();
      this.socket = null;
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
      this.connect();
    }, backoff);
  }

  private notifyConnection(): void {
    this.listeners.forEach((listener) => listener());
  }
}

/** Shared singleton used across the app. */
export const chatSocket = new ChatSocket();
