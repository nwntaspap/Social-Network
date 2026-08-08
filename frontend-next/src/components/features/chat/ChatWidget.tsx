'use client';

/**
 * components/features/chat/ChatWidget.tsx
 *
 * Floating chat popup (bottom-right) shown app-wide for authenticated users.
 * The button opens a modal with two tabs: Chats (private conversations) and
 * Groups (the user's group chats). Selecting a conversation shows an embedded
 * ChatWindow; selecting a group shows the group chat room with a live
 * "members online" indicator. Unread counts accumulate across the app and are
 * shown on the button.
 *
 * Other components can open a conversation with a specific user by dispatching
 * the `chat:open` CustomEvent (see lib/chatWidget.ts).
 */

import { useCallback, useEffect, useState } from 'react';
import { getChats, getMyGroups, startChat } from '@/lib/api';
import { chatSocket } from '@/lib/ws';
import { useGroupPresence } from '@/lib/useGroupPresence';
import { OPEN_CHAT_EVENT } from '@/lib/chatWidget';
import { useAuth } from '@/context/AuthContext';
import ConversationList from './ConversationList';
import GroupChatList from './GroupChatList';
import ChatWindow from './ChatWindow';
import GroupChatRoom from '@/components/features/groups/GroupChatRoom';
import type { Chat, Group } from '@/lib/types';

type WidgetTab = 'chats' | 'groups';

interface ActiveGroupRoomProps {
  group: Group;
  onBack: () => void;
  onClose: () => void;
}

function ActiveGroupRoom({ group, onBack, onClose }: ActiveGroupRoomProps) {
  const presence = useGroupPresence(group.id);
  return (
    <GroupChatRoom
      groupId={group.id}
      groupTitle={group.title}
      presence={presence}
      embedded
      onBack={onBack}
      onClose={onClose}
    />
  );
}

export default function ChatWidget() {
  const { user } = useAuth();
  const [open, setOpen] = useState(false);
  const [tab, setTab] = useState<WidgetTab>('chats');
  const [conversations, setConversations] = useState<Chat[]>([]);
  const [myGroups, setMyGroups] = useState<Group[]>([]);
  const [activeChat, setActiveChat] = useState<Chat | null>(null);
  const [activeGroup, setActiveGroup] = useState<Group | null>(null);
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

  const loadMyGroups = useCallback(async () => {
    try {
      const response = await getMyGroups();
      setMyGroups(response.data);
      setError('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load groups.');
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
      await startChat(userId);
      const chats = await getChats();
      setConversations(chats);
      // /chat/start returns the raw chat row without participants, so the active
      // conversation must come from the (refreshed) conversation list.
      const active = chats.find((c) => c.participants.some((p) => p.id === userId));
      if (active) {
        setActiveChat(active);
        setConversations((prev) =>
          prev.map((c) => (c.id === active.id ? { ...c, unreadCount: 0 } : c))
        );
      }
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
  const inConversation = activeChat !== null || activeGroup !== null;

  function selectConversation(chat: Chat) {
    setActiveChat(chat);
    setConversations((prev) => prev.map((c) => (c.id === chat.id ? { ...c, unreadCount: 0 } : c)));
  }

  function selectGroup(group: Group) {
    setActiveGroup(group);
  }

  function switchTab(next: WidgetTab) {
    setTab(next);
    setActiveChat(null);
    setActiveGroup(null);
    if (next === 'groups' && myGroups.length === 0) {
      void loadMyGroups();
    }
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
    setActiveGroup(null);
  }

  function backToList() {
    setActiveChat(null);
    setActiveGroup(null);
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
        <div className={`chat-modal${inConversation ? ' chat-modal--conversation' : ''}`}>
          {!inConversation && (
            <div className="chat-modal-header">
              <h2>{tab === 'chats' ? 'Chats' : 'Groups'}</h2>
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
          {!inConversation && (
            <div className="chat-tabs">
              <button
                type="button"
                className={`chat-tab${tab === 'chats' ? ' active' : ''}`}
                onClick={() => switchTab('chats')}
              >
                Chats
              </button>
              <button
                type="button"
                className={`chat-tab${tab === 'groups' ? ' active' : ''}`}
                onClick={() => switchTab('groups')}
              >
                Groups
              </button>
            </div>
          )}
          <div className="chat-modal-content">
            {loading || startingChat ? (
              <div className="chat-loading">
                <div className="chat-spinner" />
              </div>
            ) : error && !inConversation ? (
              <div className="chat-empty-state">
                <span className="chat-empty-state-text">{error}</span>
              </div>
            ) : activeChat && otherUser ? (
              <ChatWindow
                key={activeChat.id}
                chatId={activeChat.id}
                currentUserId={user.id}
                otherUser={otherUser}
                embedded
                onBack={backToList}
                onClose={closeModal}
              />
            ) : activeGroup ? (
              <ActiveGroupRoom
                key={activeGroup.id}
                group={activeGroup}
                onBack={backToList}
                onClose={closeModal}
              />
            ) : tab === 'groups' ? (
              <GroupChatList groups={myGroups} onSelect={selectGroup} />
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
