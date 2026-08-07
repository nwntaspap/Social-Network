'use client';

/**
 * components/features/chat/ChatWidget.tsx
 *
 * Floating chat popup (bottom-right) shown app-wide for authenticated users.
 * The button opens a modal with the conversation list; selecting a
 * conversation shows an embedded ChatWindow. Unread counts accumulate across
 * the app and are shown on the button.
 *
 * Other components can open a conversation with a specific user by dispatching
 * the `chat:open` CustomEvent (see lib/chatWidget.ts).
 */

import { useCallback, useEffect, useState } from 'react';
import { getChats, startChat } from '@/lib/api';
import { chatSocket } from '@/lib/ws';
import { OPEN_CHAT_EVENT } from '@/lib/chatWidget';
import { useAuth } from '@/context/AuthContext';
import ConversationList from './ConversationList';
import ChatWindow from './ChatWindow';
import type { Chat } from '@/lib/types';

export default function ChatWidget() {
  const { user } = useAuth();
  const [open, setOpen] = useState(false);
  const [conversations, setConversations] = useState<Chat[]>([]);
  const [activeChat, setActiveChat] = useState<Chat | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [startingChat, setStartingChat] = useState(false);

  const loadChats = useCallback(async () => {
    try {
      const chats = await getChats();
      setConversations(chats);
      setError('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load chats.');
    }
  }, []);

  // Own the socket + load the conversation list while the user is signed in.
  useEffect(() => {
    if (!user) return;
    let ignore = false;
    chatSocket.connect();
    getChats()
      .then((chats) => {
        if (ignore) return;
        setConversations(chats);
      })
      .catch((err) => {
        if (!ignore) setError(err instanceof Error ? err.message : 'Failed to load chats.');
      })
      .finally(() => {
        if (!ignore) setLoading(false);
      });
    return () => {
      ignore = true;
      chatSocket.disconnect();
    };
  }, [user]);

  const startChatWithUser = useCallback(async (userId: string) => {
    setError('');
    setStartingChat(true);
    try {
      const chat = await startChat(userId);
      const chats = await getChats();
      setConversations(chats);
      setActiveChat(chat);
      setOpen(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to start chat.');
      setOpen(true);
    } finally {
      setStartingChat(false);
    }
  }, []);

  // Allow other components to open a chat with a specific user.
  useEffect(() => {
    function handleOpenChat(e: Event) {
      const detail = (e as CustomEvent<{ userId: string }>).detail;
      if (detail?.userId) {
        void startChatWithUser(detail.userId);
      }
    }
    window.addEventListener(OPEN_CHAT_EVENT, handleOpenChat);
    return () => window.removeEventListener(OPEN_CHAT_EVENT, handleOpenChat);
  }, [startChatWithUser]);

  // Track unread counts for messages that arrive outside the active chat.
  useEffect(() => {
    return chatSocket.on('chat.message', (payload) => {
      const msg = payload as { chat_id: string };
      if (!msg || typeof msg.chat_id !== 'string') return;
      setConversations((prev) =>
        prev.map((c) => {
          if (c.id !== msg.chat_id) return c;
          if (c.id === activeChat?.id) return { ...c, unreadCount: 0 };
          return { ...c, unreadCount: c.unreadCount + 1 };
        })
      );
    });
  }, [activeChat]);

  if (!user) return null;

  const unreadTotal = conversations.reduce((sum, c) => sum + c.unreadCount, 0);

  function selectConversation(chat: Chat) {
    setActiveChat(chat);
    setConversations((prev) => prev.map((c) => (c.id === chat.id ? { ...c, unreadCount: 0 } : c)));
  }

  function openModal() {
    setOpen(true);
    if (conversations.length === 0) {
      void loadChats();
    }
  }

  function closeModal() {
    setOpen(false);
    setActiveChat(null);
  }

  function backToList() {
    setActiveChat(null);
  }

  const otherUser = activeChat
    ? (activeChat.participants.find((p) => p.id !== user.id) ?? activeChat.participants[0])
    : null;

  return (
    <>
      <button
        type="button"
        className={`chat-widget-button${unreadTotal > 0 ? ' has-unread' : ''}`}
        data-unread={unreadTotal > 0 ? unreadTotal : undefined}
        aria-label={open ? 'Close chat' : 'Open chat'}
        onClick={() => (open ? closeModal() : openModal())}
      >
        💬
      </button>

      {open && (
        <div className={`chat-modal${activeChat ? ' chat-modal--conversation' : ''}`}>
          {!activeChat && (
            <div className="chat-modal-header">
              <h2>Chats</h2>
              <button
                type="button"
                className="chat-modal-close"
                aria-label="Close chat"
                onClick={closeModal}
              >
                ×
              </button>
            </div>
          )}
          <div className="chat-modal-content">
            {loading || startingChat ? (
              <div className="chat-loading">
                <div className="chat-spinner" />
              </div>
            ) : error ? (
              <div className="chat-empty-state">
                <span className="chat-empty-state-text">{error}</span>
              </div>
            ) : activeChat && otherUser ? (
              <ChatWindow
                chatId={activeChat.id}
                currentUserId={user.id}
                otherUser={otherUser}
                embedded
                onBack={backToList}
                onClose={closeModal}
              />
            ) : (
              <ConversationList
                conversations={conversations}
                currentUserId={user.id}
                activeChatId={activeChat?.id ?? null}
                onSelect={selectConversation}
              />
            )}
          </div>
        </div>
      )}
    </>
  );
}
