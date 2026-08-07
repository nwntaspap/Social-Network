'use client';

/**
 * components/features/chat/ConversationList.tsx
 *
 * Sidebar listing the current user's conversations, with online status and
 * unread counts. Clicking a conversation selects it.
 */

import Image from 'next/image';
import type { Chat, ChatUser, IsOnlineStatusPayload } from '@/lib/types';
import { useEffect, useState } from 'react';
import { chatSocket } from '@/lib/ws';

interface ConversationListProps {
  conversations: Chat[];
  currentUserId: string;
  activeChatId: string | null;
  onSelect: (chat: Chat) => void;
}

function otherParticipant(chat: Chat, currentUserId: string): ChatUser {
  return chat.participants.find((p) => p.id !== currentUserId) ?? chat.participants[0];
}

export default function ConversationList({
  conversations,
  currentUserId,
  activeChatId,
  onSelect,
}: ConversationListProps) {
  const [onlineMap, setOnlineMap] = useState<Record<string, boolean>>({});

  useEffect(() => {
    return chatSocket.on('isOnlineStatus.update', (payload) => {
      const p = payload as IsOnlineStatusPayload;
      setOnlineMap((prev) => ({ ...prev, [p.user_id]: p.isOnline }));
    });
  }, []);

  if (conversations.length === 0) {
    return (
      <div className="chat-users-list">
        <div className="chat-empty-state">
          <span className="chat-empty-state-icon">💬</span>
          <span className="chat-empty-state-text">No conversations yet</span>
        </div>
      </div>
    );
  }

  return (
    <div className="chat-users-list">
      {conversations.map((chat) => {
        const other = otherParticipant(chat, currentUserId);
        const isOnline = onlineMap[other.id] ?? other.isOnline;
        return (
          <button
            key={chat.id}
            type="button"
            className={`chat-user-item${chat.id === activeChatId ? ' active' : ''}`}
            onClick={() => onSelect(chat)}
          >
            <span className="chat-user-avatar">
              <Image
                src={other.avatarUrl || '/images/user-avatar.png'}
                alt={other.username}
                width={40}
                height={40}
                style={{ borderRadius: '50%', objectFit: 'cover' }}
                onError={(e) => {
                  (e.target as HTMLImageElement).src = '/images/user-avatar.png';
                }}
              />
            </span>
            <span className="chat-user-info">
              <span className="chat-user-name">{other.username}</span>
              <span className={`chat-user-status ${isOnline ? 'online' : 'offline'}`}>
                {isOnline ? 'Online' : 'Offline'}
              </span>
            </span>
            {chat.unreadCount > 0 && <span className="chat-user-unread">{chat.unreadCount}</span>}
          </button>
        );
      })}
    </div>
  );
}
