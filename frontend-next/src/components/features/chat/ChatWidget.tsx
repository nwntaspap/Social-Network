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

import { useCallback, useEffect, useRef, useState } from 'react';
import { getChats, getMyGroups, startChat } from '@/lib/api';
import { chatSocket } from '@/lib/ws';
import { useGroupPresence } from '@/lib/useGroupPresence';
import { OPEN_CHAT_EVENT } from '@/lib/chatWidget';
import { useAuth } from '@/context/AuthContext';
import ConversationList from './ConversationList';
import GroupChatList from './GroupChatList';
import ChatWindow from './ChatWindow';
import GroupChatRoom from '@/components/features/groups/GroupChatRoom';
import type { Chat, Group, GroupChatMessageWire } from '@/lib/types';

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
  // Stable identity for effect deps: AuthContext may hand back a new object
  // per render, and the socket/fetch setup must not restart when it does.
  const userId = user?.id;
  const [open, setOpen] = useState(false);
  const [tab, setTab] = useState<WidgetTab>('chats');
  const [conversations, setConversations] = useState<Chat[]>([]);
  const [myGroups, setMyGroups] = useState<Group[]>([]);
  const [groupUnread, setGroupUnread] = useState<Record<string, number>>({});
  const [activeChat, setActiveChat] = useState<Chat | null>(null);
  const [activeGroup, setActiveGroup] = useState<Group | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [startingChat, setStartingChat] = useState(false);
  // Bumped on every fresh conversation fetch so the list re-seeds its online
  // status from the (server-authoritative) snapshot instead of stale WS state.
  const [convVersion, setConvVersion] = useState(0);
  // Last live increment time per group, so a server snapshot taken BEFORE a
  // live message is never allowed to overwrite the fresher live count.
  const liveSeenRef = useRef<Record<string, number>>({});
  // Mirror of the conversation list for WS handlers, plus the last time a
  // refresh was scheduled for an unknown chat id (debounce).
  const conversationsRef = useRef<Chat[]>([]);
  const newChatRefreshAtRef = useRef(0);

  const loadChats = useCallback(async () => {
    try {
      const chats = await getChats();
      setConversations(chats);
      setError('');
      setConvVersion((v) => v + 1);
      return chats;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load chats.');
      return [];
    }
  }, []);

  useEffect(() => {
    conversationsRef.current = conversations;
  }, [conversations]);

  const loadMyGroups = useCallback(async () => {
    try {
      const requestedAt = Date.now();
      const response = await getMyGroups();
      setMyGroups(response.data);
      setGroupUnread((prev) => {
        const next: Record<string, number> = {};
        for (const g of response.data) {
          // A live increment newer than this fetch beats the stale snapshot.
          next[g.id] =
            (liveSeenRef.current[g.id] ?? 0) >= requestedAt
              ? (prev[g.id] ?? 0)
              : (g.unreadCount ?? 0);
        }
        return next;
      });
      setError('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load groups.');
    }
  }, []);

  // Own the socket + load the conversation list while the user is signed in.
  // On socket reconnect, refetch chats and groups so status/unread missed
  // during the outage are recovered (broadcasts are not replayed).
  useEffect(() => {
    if (!userId) return;
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
        if (!ignore) void loadMyGroups();
      });
    const unsubscribeConnection = chatSocket.onConnection(() => {
      if (chatSocket.isConnected()) {
        void loadChats();
        void loadMyGroups();
      }
    });
    return () => {
      ignore = true;
      unsubscribeConnection();
      chatSocket.disconnect();
    };
  }, [userId, loadChats, loadMyGroups]);

  const startChatWithUser = useCallback(
    async (userId: string) => {
      setError('');
      setStartingChat(true);
      try {
        await startChat(userId);
        const chats = await loadChats();
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
    },
    [loadChats]
  );

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
      const known = conversationsRef.current.some((c) => c.id === msg.chat_id);
      if (!known) {
        // A first-ever message creates the conversation server-side, so it is
        // missing from the local list. Pull the snapshot, which carries the
        // new conversation and its authoritative unread count.
        const now = Date.now();
        if (now - newChatRefreshAtRef.current >= 2000) {
          newChatRefreshAtRef.current = now;
          void loadChats();
        }
        return;
      }
      setConversations((prev) =>
        prev.map((c) => {
          if (c.id !== msg.chat_id) return c;
          if (c.id === activeChat?.id) return { ...c, unreadCount: 0 };
          return { ...c, unreadCount: c.unreadCount + 1 };
        })
      );
    });
  }, [activeChat, loadChats]);

  // Track unread group messages. The server broadcasts group_chat.message to
  // every member (including the sender), so own messages are skipped, and a
  // message for the open room is not counted.
  useEffect(() => {
    if (!user) return;
    return chatSocket.on('group_chat.message', (payload) => {
      const msg = payload as GroupChatMessageWire;
      if (!msg || typeof msg.group_id !== 'string') return;
      if (msg.sender_id === user.id) return;
      if (msg.group_id === activeGroup?.id) return;
      liveSeenRef.current = { ...liveSeenRef.current, [msg.group_id]: Date.now() };
      setGroupUnread((prev) => ({ ...prev, [msg.group_id]: (prev[msg.group_id] ?? 0) + 1 }));
    });
  }, [activeGroup, user]);

  if (!user) return null;

  const unreadTotal =
    conversations.reduce((sum, c) => sum + c.unreadCount, 0) +
    Object.values(groupUnread).reduce((sum, n) => sum + n, 0);
  const inConversation = activeChat !== null || activeGroup !== null;

  function selectConversation(chat: Chat) {
    setActiveChat(chat);
    setConversations((prev) => prev.map((c) => (c.id === chat.id ? { ...c, unreadCount: 0 } : c)));
  }

  function selectGroup(group: Group) {
    setActiveGroup(group);
    setGroupUnread((prev) => ({ ...prev, [group.id]: 0 }));
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
    // Always refresh so unread counts and online status missed while the
    // widget was closed (or lost to a socket reconnect) are recovered.
    void loadChats();
    void loadMyGroups();
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
              <GroupChatList groups={myGroups} unreadByGroup={groupUnread} onSelect={selectGroup} />
            ) : (
              <ConversationList
                key={convVersion}
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
