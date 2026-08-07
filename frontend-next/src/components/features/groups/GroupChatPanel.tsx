'use client';

/**
 * components/features/groups/GroupChatPanel.tsx
 *
 * Group chat room: loads history over HTTP, then lives over the WebSocket.
 * Listens for group_chat.message on the active group and sends group_chat.send
 * envelopes. Member names are resolved from the group member list.
 */

import { useEffect, useRef, useState } from 'react';
import Image from 'next/image';
import { getGroupChatHistory, getGroupMembers } from '@/lib/api';
import { chatSocket } from '@/lib/ws';
import { useAuth } from '@/context/AuthContext';
import { formatMessageTime, getDisplayName } from '@/lib/helpers';
import MessageInput from '@/components/features/chat/MessageInput';
import type { GroupChatMessageWire, GroupMember } from '@/lib/types';

interface GroupChatPanelProps {
  groupId: string;
}

interface DisplayMessage {
  id: string;
  senderId: string;
  content: string;
  createdAt: string;
}

function toDisplayMessage(msg: GroupChatMessageWire): DisplayMessage {
  return {
    id: `ws-${msg.id}`,
    senderId: msg.sender_id,
    content: msg.content,
    createdAt: msg.created_at,
  };
}

export default function GroupChatPanel({ groupId }: GroupChatPanelProps) {
  const { user: currentUser } = useAuth();
  const [messages, setMessages] = useState<DisplayMessage[]>([]);
  const [members, setMembers] = useState<Map<string, GroupMember>>(new Map());
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const bottomRef = useRef<HTMLDivElement>(null);

  // Load history + members over HTTP on mount / group change.
  useEffect(() => {
    let ignore = false;
    Promise.all([getGroupChatHistory(groupId), getGroupMembers(groupId)])
      .then(([history, memberPage]) => {
        if (ignore) return;
        setMessages(
          history.map((m) => ({
            id: `http-${m.id}`,
            senderId: m.sender_id,
            content: m.content,
            createdAt: m.created_at,
          }))
        );
        setMembers(new Map(memberPage.data.map((m) => [m.userId, m])));
      })
      .catch((err) => {
        if (!ignore) setError(err instanceof Error ? err.message : 'Failed to load chat.');
      })
      .finally(() => {
        if (!ignore) setLoading(false);
      });
    return () => {
      ignore = true;
    };
  }, [groupId]);

  // Keep the connection open and subscribe to incoming group messages.
  useEffect(() => {
    chatSocket.connect();
    const unsubscribe = chatSocket.on('group_chat.message', (payload) => {
      const msg = payload as GroupChatMessageWire;
      if (msg.group_id !== groupId) return;
      setMessages((prev) => [...prev, toDisplayMessage(msg)]);
    });
    return () => {
      unsubscribe();
      chatSocket.disconnect();
    };
  }, [groupId]);

  // Auto-scroll on new messages.
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages.length]);

  function handleSend(content: string) {
    chatSocket.send('group_chat.send', { group_id: groupId, content });
  }

  return (
    <div className="group-chat-panel">
      <div className="group-chat-panel-header">
        <h3>Group Chat</h3>
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
            <span className="chat-empty-state-text">No messages yet. Start the conversation!</span>
          </div>
        ) : (
          messages.map((msg) => {
            const own = msg.senderId === currentUser?.id;
            const sender = members.get(msg.senderId)?.user;
            return (
              <div key={msg.id} className={`chat-message${own ? ' own' : ''}`}>
                {!own ? (
                  <div className="group-chat-message-inner">
                    <div className="group-chat-sender-row">
                      <Image
                        src={sender?.avatarUrl || '/images/user-avatar.png'}
                        alt={getDisplayName(sender ?? {})}
                        width={24}
                        height={24}
                        className="group-chat-sender-avatar"
                        onError={(e) => {
                          (e.target as HTMLImageElement).src = '/images/user-avatar.png';
                        }}
                      />
                      <span className="group-chat-sender-name">{getDisplayName(sender ?? {})}</span>
                    </div>
                    <div className="chat-message-bubble">
                      {msg.content}
                      <div className="chat-message-time">{formatMessageTime(msg.createdAt)}</div>
                    </div>
                  </div>
                ) : (
                  <div className="chat-message-bubble">
                    {msg.content}
                    <div className="chat-message-time">{formatMessageTime(msg.createdAt)}</div>
                  </div>
                )}
              </div>
            );
          })
        )}
        <div ref={bottomRef} />
      </div>

      <MessageInput onSend={handleSend} />
    </div>
  );
}
