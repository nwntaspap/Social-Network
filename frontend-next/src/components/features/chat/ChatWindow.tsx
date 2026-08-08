'use client';

/**
 * components/features/chat/ChatWindow.tsx
 *
 * Private 1:1 conversation: loads history over HTTP, then lives over the
 * WebSocket. Listens for chat.message on the active chat, marks messages
 * read, and sends chat.send envelopes.
 *
 * Incoming messages are deduplicated by id so a single WS delivery can never
 * produce duplicate rows (and React duplicate-key warnings).
 */

import { useEffect, useRef, useState } from 'react';
import Image from 'next/image';
import { getChatMessages } from '@/lib/api';
import { chatSocket } from '@/lib/ws';
import { formatMessageTime } from '@/lib/helpers';
import MessageInput from './MessageInput';
import type { ChatUser, PrivateWsMessage } from '@/lib/types';

interface ChatWindowProps {
  chatId: string;
  currentUserId: string;
  otherUser: ChatUser;
  /** Render inside a modal/container instead of the fixed floating window. */
  embedded?: boolean;
  onBack?: () => void;
  onClose?: () => void;
}

interface DisplayMessage {
  id: string;
  senderId: string;
  content: string;
  createdAt: string;
}

function toDisplayMessage(msg: PrivateWsMessage): DisplayMessage {
  return {
    id: `ws-${msg.id}`,
    senderId: msg.sender_id,
    content: msg.content,
    createdAt: msg.created_at,
  };
}

export default function ChatWindow({
  chatId,
  currentUserId,
  otherUser,
  embedded,
  onBack,
  onClose,
}: ChatWindowProps) {
  const [messages, setMessages] = useState<DisplayMessage[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [isTyping, setIsTyping] = useState(false);
  const bottomRef = useRef<HTMLDivElement>(null);

  // Load history over HTTP on mount / chat change.
  useEffect(() => {
    let ignore = false;
    getChatMessages(chatId)
      .then((history) => {
        if (ignore) return;
        setMessages(
          history.map((m) => ({
            id: `http-${m.id}`,
            senderId: m.senderId,
            content: m.content,
            createdAt: m.createdAt,
          }))
        );
      })
      .catch((err) => {
        if (!ignore) setError(err instanceof Error ? err.message : 'Failed to load messages.');
      })
      .finally(() => {
        if (!ignore) setLoading(false);
      });
    return () => {
      ignore = true;
    };
  }, [chatId]);

  // Incoming messages + typing + open/close bookkeeping on the WS.
  useEffect(() => {
    let isTypingTimer: ReturnType<typeof setTimeout> | null = null;

    const unsubscribeMessage = chatSocket.on('chat.message', (payload) => {
      const msg = payload as PrivateWsMessage;
      if (msg.chat_id !== chatId) return;
      const display = toDisplayMessage(msg);
      setMessages((prev) => (prev.some((m) => m.id === display.id) ? prev : [...prev, display]));
      setIsTyping(false);
      if (isTypingTimer) {
        clearTimeout(isTypingTimer);
        isTypingTimer = null;
      }
      chatSocket.send('chat.mark_read', { chat_id: chatId, up_to_message_id: msg.id });
    });

    const unsubscribeTyping = chatSocket.on('chat.is_typing', (payload) => {
      const p = payload as { chat_id: string; user_id: string };
      if (p.chat_id !== chatId || p.user_id === currentUserId) return;
      setIsTyping(true);
      if (isTypingTimer) clearTimeout(isTypingTimer);
      isTypingTimer = setTimeout(() => setIsTyping(false), 3000);
    });

    chatSocket.send('chat.open', { chat_id: chatId });

    return () => {
      unsubscribeMessage();
      unsubscribeTyping();
      if (isTypingTimer) clearTimeout(isTypingTimer);
      chatSocket.send('chat.close', {});
    };
  }, [chatId, currentUserId]);

  // Auto-scroll on new messages.
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages.length, isTyping]);

  function handleSend(content: string) {
    chatSocket.send('chat.send', { chat_id: chatId, content });
  }

  function handleTyping() {
    chatSocket.send('chat.typing', { chat_id: chatId });
  }

  return (
    <div className={`chat-window${embedded ? ' chat-window--embedded' : ''}`}>
      <div className="chat-window-header">
        {onBack && (
          <button
            type="button"
            className="chat-window-back"
            aria-label="Back to conversations"
            onClick={onBack}
          >
            ←
          </button>
        )}
        <div className="chat-window-header-user">
          <span className="chat-user-avatar">
            <Image
              src={otherUser.avatarUrl || '/images/user-avatar.png'}
              alt={otherUser.username}
              width={40}
              height={40}
              style={{ borderRadius: '50%', objectFit: 'cover' }}
              onError={(e) => {
                (e.target as HTMLImageElement).src = '/images/user-avatar.png';
              }}
            />
          </span>
          <div className="chat-window-header-title">
            <h3>{otherUser.username}</h3>
            <span className={`chat-user-status ${otherUser.isOnline ? 'online' : 'offline'}`}>
              {otherUser.isOnline ? 'Online' : 'Offline'}
            </span>
          </div>
        </div>
        {onClose && (
          <button
            type="button"
            className="chat-window-close"
            aria-label="Close chat"
            onClick={onClose}
          >
            ×
          </button>
        )}
      </div>

      <div className="chat-messages-container">
        {loading ? (
          <div className="chat-loading">
            <div className="chat-spinner" />
          </div>
        ) : error ? (
          <div className="chat-empty-state">
            <span className="chat-empty-state-text">{error}</span>
          </div>
        ) : messages.length === 0 ? (
          <div className="chat-empty-state">
            <span className="chat-empty-state-icon">💬</span>
            <span className="chat-empty-state-text">Say hello to {otherUser.username}!</span>
          </div>
        ) : (
          messages.map((msg) => (
            <div
              key={msg.id}
              className={`chat-message${msg.senderId === currentUserId ? ' own' : ''}`}
            >
              <div className="chat-message-bubble">
                {msg.content}
                <div className="chat-message-time">{formatMessageTime(msg.createdAt)}</div>
              </div>
            </div>
          ))
        )}
        {isTyping && (
          <div className="chat-message chat-message-typing">
            <div className="chat-message-bubble chat-message-bubble-typing">
              <span className="chat-typing-dots">
                <span />
                <span />
                <span />
              </span>
            </div>
          </div>
        )}
        <div ref={bottomRef} />
      </div>

      <MessageInput onSend={handleSend} onTyping={handleTyping} />
    </div>
  );
}
