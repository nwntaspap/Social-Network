'use client';

/**
 * components/features/chat/ChatPage.tsx
 *
 * Full-page chat: conversation list on the left, active conversation on the
 * right. Supports starting a new chat via the `?chat=<userId>` query param.
 */

import { useCallback, useEffect, useRef, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { getChats, startChat } from '@/lib/api';
import { chatSocket } from '@/lib/ws';
import { useAuth } from '@/context/AuthContext';
import ConversationList from './ConversationList';
import ChatWindow from './ChatWindow';
import type { Chat } from '@/lib/types';

export default function ChatPage() {
  const { user: currentUser } = useAuth();
  const router = useRouter();
  const searchParams = useSearchParams();
  const targetUserId = searchParams.get('chat');

  const [conversations, setConversations] = useState<Chat[]>([]);
  const [activeChat, setActiveChat] = useState<Chat | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [startingChat, setStartingChat] = useState(false);
  const handledTargetRef = useRef(false);

  const selectConversation = useCallback((chat: Chat) => {
    setActiveChat(chat);
    setConversations((prev) => prev.map((c) => (c.id === chat.id ? { ...c, unreadCount: 0 } : c)));
  }, []);

  // Load the conversation list.
  useEffect(() => {
    if (!currentUser) return;
    let ignore = false;
    getChats()
      .then((chats) => {
        if (ignore) return;
        setConversations(chats);
        if (chats.length > 0) {
          setActiveChat((prev) => prev ?? chats[0]);
        }
      })
      .catch((err) => {
        if (!ignore) setError(err instanceof Error ? err.message : 'Failed to load chats.');
      })
      .finally(() => {
        if (!ignore) setLoading(false);
      });
    return () => {
      ignore = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentUser]);

  // Start a chat when ?chat=<userId> is present.
  useEffect(() => {
    if (!currentUser || !targetUserId || handledTargetRef.current) return;
    handledTargetRef.current = true;
    setStartingChat(true);
    startChat(targetUserId)
      .catch((err) => {
        if (err instanceof Error) setError(err.message);
      })
      .finally(() => {
        getChats()
          .then((chats) => {
            const existing = chats.find((c) => c.participants.some((p) => p.id === targetUserId));
            setConversations(chats);
            if (existing) setActiveChat(existing);
            else if (chats.length > 0) setActiveChat(chats[0]);
          })
          .catch(() => {
            // Keep current list.
          })
          .finally(() => {
            setStartingChat(false);
            router.replace('/chat');
          });
      });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentUser, targetUserId]);

  // Keep the connection open while the page is mounted.
  useEffect(() => {
    chatSocket.connect();
    return () => chatSocket.disconnect();
  }, []);

  // Increment unread counts for messages outside the active chat.
  useEffect(() => {
    return chatSocket.on('chat.message', (payload) => {
      const msg = payload as { chat_id: string };
      if (!msg || typeof msg.chat_id !== 'string') return;
      setConversations((prev) =>
        prev.map((c) => {
          if (c.id === msg.chat_id && c.id !== activeChat?.id) {
            return { ...c, unreadCount: c.unreadCount + 1 };
          }
          return c;
        })
      );
    });
  }, [activeChat]);

  if (loading || startingChat) {
    return (
      <div className="chat-page">
        <div className="chat-page-main">
          <div className="chat-loading">
            <div className="chat-spinner" />
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="chat-page">
      <div className="chat-page-sidebar">
        <div className="chat-page-sidebar-header">
          <h2>Chats</h2>
        </div>
        <ConversationList
          conversations={conversations}
          currentUserId={currentUser?.id ?? ''}
          activeChatId={activeChat?.id ?? null}
          onSelect={selectConversation}
        />
        {error && <div className="chat-empty-state-text">{error}</div>}
      </div>

      <div className="chat-page-main">
        {activeChat ? (
          <ChatWindow
            key={activeChat.id}
            chatId={activeChat.id}
            currentUserId={currentUser?.id ?? ''}
            otherUser={
              activeChat.participants.find((p) => p.id !== currentUser?.id) ??
              activeChat.participants[0]
            }
          />
        ) : (
          <div className="chat-page-placeholder">
            <span className="chat-page-placeholder-icon">💬</span>
            <span>Select a conversation to start chatting</span>
          </div>
        )}
      </div>
    </div>
  );
}
