'use client';

/**
 * components/features/groups/GroupChatRoom.tsx
 *
 * Group chat room UI: loads history + members over HTTP, then lives over the
 * WebSocket. Listens for group_chat.message on the active group and sends
 * group_chat.send envelopes. Pure UI — it never opens/closes the socket; the
 * hosting component (GroupChatPanel on the group page, or the chat widget's
 * Groups tab) owns the connection. A live member-online indicator is rendered
 * when a presence snapshot is injected via props.
 */

import { useEffect, useRef, useState } from 'react';
import Image from 'next/image';
import { getGroupChatHistory, getGroupMembers } from '@/lib/api';
import { chatSocket } from '@/lib/ws';
import { useAuth } from '@/context/AuthContext';
import { formatMessageTime, getDisplayName } from '@/lib/helpers';
import MessageInput from '@/components/features/chat/MessageInput';
import type { GroupChatIsTypingPayload, GroupChatMessageWire, GroupMember } from '@/lib/types';
import type { GroupPresenceState } from '@/lib/useGroupPresence';

interface GroupChatRoomProps {
  groupId: string;
  /** Title shown in the header; defaults to "Group Chat". */
  groupTitle?: string;
  /** Live online-members snapshot to render in the header. */
  presence?: GroupPresenceState;
  /** Render inside a modal/container instead of the full-height page panel. */
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

function toDisplayMessage(msg: GroupChatMessageWire): DisplayMessage {
  return {
    id: `ws-${msg.id}`,
    senderId: msg.sender_id,
    content: msg.content,
    createdAt: msg.created_at,
  };
}

export default function GroupChatRoom({
  groupId,
  groupTitle,
  presence,
  embedded,
  onBack,
  onClose,
}: GroupChatRoomProps) {
  const { user: currentUser } = useAuth();
  const [messages, setMessages] = useState<DisplayMessage[]>([]);
  const [members, setMembers] = useState<Map<string, GroupMember>>(new Map());
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showPresence, setShowPresence] = useState(false);
  const [typers, setTypers] = useState<Set<string>>(new Set());
  const bottomRef = useRef<HTMLDivElement>(null);

  // Load history + members over HTTP on mount / group change. Members are
  // fetched with a generous page size so the online-users popover can resolve
  // every member's name and avatar.
  useEffect(() => {
    let ignore = false;
    Promise.all([getGroupChatHistory(groupId), getGroupMembers(groupId, 1, 200)])
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

  // Mark the group read when the room opens so the unread badge clears.
  useEffect(() => {
    chatSocket.send('group_chat.mark_read', { group_id: groupId });
  }, [groupId]);

  // Close the online-users popover when clicking anywhere outside it.
  useEffect(() => {
    if (!showPresence) return;
    function onDocClick(e: MouseEvent) {
      if (!(e.target as HTMLElement).closest('.group-chat-presence-wrap')) {
        setShowPresence(false);
      }
    }
    document.addEventListener('mousedown', onDocClick);
    return () => document.removeEventListener('mousedown', onDocClick);
  }, [showPresence]);

  // Subscribe to incoming group messages + typing over the WS (the socket owner
  // is upstream). Typing events carry who is typing and are auto-cleared after
  // 3s or when that member's message arrives.
  useEffect(() => {
    const timers = new Map<string, ReturnType<typeof setTimeout>>();

    const clearTyping = (userId: string) => {
      const timer = timers.get(userId);
      if (timer) {
        clearTimeout(timer);
        timers.delete(userId);
      }
      setTypers((prev) => {
        if (!prev.has(userId)) return prev;
        const next = new Set(prev);
        next.delete(userId);
        return next;
      });
    };

    const unsubscribeMessage = chatSocket.on('group_chat.message', (payload) => {
      const msg = payload as GroupChatMessageWire;
      if (msg.group_id !== groupId) return;
      const display = toDisplayMessage(msg);
      setMessages((prev) => (prev.some((m) => m.id === display.id) ? prev : [...prev, display]));
      clearTyping(msg.sender_id);
      chatSocket.send('group_chat.mark_read', { group_id: groupId });
    });

    const unsubscribeTyping = chatSocket.on('group_chat.is_typing', (payload) => {
      const p = payload as GroupChatIsTypingPayload;
      if (p.group_id !== groupId || p.user_id === currentUser?.id) return;
      setTypers((prev) => (prev.has(p.user_id) ? prev : new Set(prev).add(p.user_id)));
      const existing = timers.get(p.user_id);
      if (existing) clearTimeout(existing);
      timers.set(
        p.user_id,
        setTimeout(() => clearTyping(p.user_id), 3000)
      );
    });

    return () => {
      unsubscribeMessage();
      unsubscribeTyping();
      timers.forEach((timer) => clearTimeout(timer));
      timers.clear();
    };
  }, [groupId, currentUser?.id]);

  // Auto-scroll on new messages and typing changes.
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages.length, typers.size]);

  function handleSend(content: string) {
    chatSocket.send('group_chat.send', { group_id: groupId, content });
  }

  function handleTyping() {
    chatSocket.send('group_chat.typing', { group_id: groupId });
  }

  function typingLabel(userIDs: string[]): string {
    const names = userIDs.map((id) => {
      const gm = members.get(id);
      return gm ? getDisplayName(gm.user) : 'Someone';
    });
    if (names.length === 1) return `${names[0]} is typing`;
    if (names.length === 2) return `${names[0]} and ${names[1]} are typing`;
    return `${names.slice(0, 2).join(', ')} and ${names.length - 2} more are typing`;
  }

  const onlineCount = presence ? `${presence.online} online` : null;
  const onlineMembers = presence?.members?.filter((m) => m.isOnline) ?? [];

  return (
    <div className={`group-chat-panel${embedded ? ' group-chat-panel--embedded' : ''}`}>
      <div className="group-chat-panel-header">
        {onBack && (
          <button
            type="button"
            className="group-chat-panel-back"
            aria-label="Back to groups"
            onClick={onBack}
          >
            ←
          </button>
        )}
        <div className="group-chat-panel-header-title">
          <h3>{groupTitle ?? 'Group Chat'}</h3>
          {presence && !presence.loading && (
            <div className="group-chat-presence-wrap">
              <button
                type="button"
                className="group-chat-presence"
                aria-expanded={showPresence}
                aria-label={`${presence.online} members online`}
                onClick={() => setShowPresence((v) => !v)}
              >
                <span className="group-chat-presence-dot" />
                {onlineCount}
              </button>
              {showPresence && (
                <div className="group-chat-presence-popover" role="menu">
                  <div className="group-chat-presence-popover-title">Online now</div>
                  {onlineMembers.length === 0 ? (
                    <div className="group-chat-presence-popover-empty">
                      No one is online right now
                    </div>
                  ) : (
                    onlineMembers.map((m) => {
                      const gm = members.get(m.id);
                      const name = gm ? getDisplayName(gm.user) : 'Member';
                      const avatar = gm?.user.avatarUrl || '/images/user-avatar.png';
                      return (
                        <div key={m.id} className="group-chat-presence-popover-item">
                          <Image
                            src={avatar}
                            alt={name}
                            width={28}
                            height={28}
                            className="group-chat-presence-popover-avatar"
                            onError={(e) => {
                              (e.target as HTMLImageElement).src = '/images/user-avatar.png';
                            }}
                          />
                          <span className="group-chat-presence-popover-name">{name}</span>
                          <span className="group-chat-presence-popover-dot" />
                        </div>
                      );
                    })
                  )}
                </div>
              )}
            </div>
          )}
        </div>
        {onClose && (
          <button
            type="button"
            className="group-chat-panel-close"
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
        {typers.size > 0 && (
          <div className="chat-message chat-message-typing">
            <div className="group-chat-message-inner">
              <div className="group-chat-typing-row">
                <span className="group-chat-typing-name">{typingLabel([...typers])}</span>
              </div>
              <div className="chat-message-bubble chat-message-bubble-typing">
                <span className="chat-typing-dots">
                  <span />
                  <span />
                  <span />
                </span>
              </div>
            </div>
          </div>
        )}
        <div ref={bottomRef} />
      </div>

      <MessageInput onSend={handleSend} onTyping={handleTyping} />
    </div>
  );
}
