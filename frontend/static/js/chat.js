/**
 * chat.js
 *
 * Standalone chat module that manages:
 * - Chat widget (button to open chat)
 * - Chat modal (user list)
 * - Chat window (message interface)
 * - WebSocket connection for real-time messages
 *
 * Initialize with: initChat(user)
 * Call from anywhere to open a specific chat: openChatWithUser(userId, userName)
 */

import { fetchChatUsers, initializeChat, ApiError } from './api.js';
import { escapeHTML, formatMessageTime, throttle } from './helpers.js';
import {
  renderChatWindow,
  renderChatMessages,
  renderChatModal,
  renderChatUsersList,
  renderChatWidget,
  renderChatStatus,
  renderActiveChatTypingState,
} from './chat.render.js';

// ─── State ────────────────────────────────────────────────────────────────────

export const chatState = {
  ws: null,
  currentUser: null,
  currentChat: null,
  chatUsers: [],
  messageBuffer: {}, // Stores messages per chat_id
  unreadCounts: {}, // Stores unread counts per chat_id
  userChatMap: {}, // Maps user_id -> chat_id for quick lookup
  activeChatUserId: null,
  historyObserver: null,
  historyState: {}, // Stores pagination state per chat_id
  pendingHistoryRequests: {}, // Stores in-flight history request metadata by request_id
  typingTimeouts: {}, // Stores typing timeout per chat_id:user_id
  isConnecting: false,
  reconnectTimeout: null,
  chatInitialized: false,
};
export const CHAT_HISTORY_PAGE_SIZE = 20;
export const CHAT_HISTORY_LOAD_DELAY_MS = 2000;
export const CHAT_TYPING_THROTTLE_MS = 1000;

/**
 * Normalize message properties from backend PascalCase to our snake_case format
 */
function normalizeMessage(msg) {
  return {
    id: msg.ID !== undefined ? msg.ID : msg.id,
    chat_id: msg.ChatID !== undefined ? msg.ChatID : msg.chat_id,
    sender_id: msg.SenderID !== undefined ? msg.SenderID : msg.sender_id,
    content: msg.Content !== undefined ? msg.Content : msg.content,
    created_at: msg.CreatedAt !== undefined ? msg.CreatedAt : msg.created_at,
    client_message_id:
      msg.ClientMessageID !== undefined ? msg.ClientMessageID : msg.client_message_id,
  };
}

function getHistoryState(chatId) {
  if (!chatState.historyState[chatId]) {
    chatState.historyState[chatId] = {
      hasMore: true,
      delayPending: false,
      loadingOlder: false,
      pendingTimeoutId: null,
    };
  }

  return chatState.historyState[chatId];
}

export function disconnectHistoryObserver() {
  if (chatState.historyObserver) {
    chatState.historyObserver.disconnect();
    chatState.historyObserver = null;
  }
}

function clearPendingHistoryLoad(chatId) {
  const state = chatState.historyState[chatId];
  if (!state) return;

  if (state.pendingTimeoutId) {
    clearTimeout(state.pendingTimeoutId);
    state.pendingTimeoutId = null;
  }

  state.delayPending = false;
}

function showHistoryLoader(show) {
  const loader = document.getElementById('chat-history-loader');
  if (!loader) return;

  loader.style.display = show ? 'flex' : 'none';
}

export function sendTypingEvent(chatId) {
  if (!chatId || !chatState.ws || chatState.ws.readyState !== WebSocket.OPEN) {
    return;
  }

  chatState.ws.send(
    JSON.stringify({
      type: 'chat.typing',
      payload: {
        chat_id: chatId,
      },
    })
  );
}

export function sendChatViewEvent(type, chatId) {
  if (!chatId || !chatState.ws || chatState.ws.readyState !== WebSocket.OPEN) {
    return;
  }

  chatState.ws.send(
    JSON.stringify({
      type,
      payload: {
        chat_id: chatId,
      },
    })
  );
}

export function isActiveChatTyping() {
  if (!chatState.currentChat || !chatState.activeChatUserId) {
    return false;
  }

  return Boolean(
    chatState.typingTimeouts[`${chatState.currentChat.id}:${chatState.activeChatUserId}`]
  );
}

function handleChatTypingStatus(userId, chatId) {
  if (
    !chatState.currentChat ||
    chatState.currentChat.id !== chatId ||
    chatState.activeChatUserId !== userId
  ) {
    return;
  }

  const typingKey = `${chatId}:${userId}`;
  if (chatState.typingTimeouts[typingKey]) {
    clearTimeout(chatState.typingTimeouts[typingKey]);
  }

  chatState.typingTimeouts[typingKey] = setTimeout(() => {
    delete chatState.typingTimeouts[typingKey];
    renderActiveChatTypingState();
  }, 2500);

  renderActiveChatTypingState();
}

function clearAllPendingHistoryLoads() {
  Object.keys(chatState.historyState).forEach((chatId) => {
    clearPendingHistoryLoad(chatId);
  });
}

function getOldestPersistedMessageId(chatId) {
  const persistedMessages = (chatState.messageBuffer[chatId] || []).filter((msg) => msg.id > 0);

  if (persistedMessages.length === 0) return 0;

  return Math.min(...persistedMessages.map((msg) => msg.id));
}

function observeHistorySentinel() {
  disconnectHistoryObserver();

  if (!chatState.currentChat) return;

  const chatId = chatState.currentChat.id;
  const state = getHistoryState(chatId);
  if (!state.hasMore) return;

  const container = document.getElementById('chat-messages');
  const sentinel = document.getElementById('chat-history-sentinel');
  if (!container || !sentinel) return;

  chatState.historyObserver = new IntersectionObserver(
    (entries) => {
      if (!entries.some((entry) => entry.isIntersecting)) return;

      if (!chatState.currentChat || chatState.currentChat.id !== chatId) return;

      if (state.delayPending || state.loadingOlder) return;

      const oldestMessageId = getOldestPersistedMessageId(chatId);
      if (!oldestMessageId) return;

      state.delayPending = true;
      showHistoryLoader(true);
      state.pendingTimeoutId = setTimeout(() => {
        state.pendingTimeoutId = null;
        state.delayPending = false;

        if (!chatState.currentChat || chatState.currentChat.id !== chatId) {
          showHistoryLoader(false);
          return;
        }

        loadChatHistory(chatId, oldestMessageId);
      }, CHAT_HISTORY_LOAD_DELAY_MS);
    },
    {
      root: container,
      threshold: 0.1,
    }
  );

  chatState.historyObserver.observe(sentinel);
}

function restoreScrollAfterPrepend(previousScrollHeight, previousScrollTop) {
  const container = document.getElementById('chat-messages');
  if (!container) return;

  requestAnimationFrame(() => {
    container.scrollTop = container.scrollHeight - previousScrollHeight + previousScrollTop;
    observeHistorySentinel();
  });
}

// ─── Public API ────────────────────────────────────────────────────────────────

/**
 * Initialize the chat system. Call once on app startup.
 */
export function initChat(user) {
  if (!user?.id) return;

  if (chatState.chatInitialized) {
    console.log('Chat already initialized');
    return;
  }

  console.log('Initializing chat for:', user.username || user.email);

  chatState.chatInitialized = true;

  chatState.currentUser = user;

  renderChatWidget();
  connectWebSocket();
  loadChatUsers();
}

export function cleanupChat() {
  console.log('Cleaning up chat');

  chatState.chatInitialized = false;

  // stop reconnect loop
  if (chatState.reconnectTimeout) {
    clearTimeout(chatState.reconnectTimeout);
    chatState.reconnectTimeout = null;
  }

  // close websocket
  if (chatState.ws) {
    chatState.ws.onclose = null;
    chatState.ws.close();
    chatState.ws = null;
  }

  // reset state
  clearAllPendingHistoryLoads();
  chatState.currentUser = null;
  chatState.currentChat = null;
  chatState.chatUsers = [];
  chatState.messageBuffer = {};
  chatState.unreadCounts = {};
  chatState.userChatMap = {};
  chatState.activeChatUserId = null;
  chatState.historyState = {};
  chatState.pendingHistoryRequests = {};
  Object.values(chatState.typingTimeouts).forEach((timeoutId) => clearTimeout(timeoutId));
  chatState.typingTimeouts = {};
  chatState.isConnecting = false;
  disconnectHistoryObserver();

  // remove widget
  const widget = document.getElementById('chat-widget-button');
  if (widget) {
    widget.remove();
  }

  // remove modal
  const modal = document.getElementById('chat-modal');
  if (modal) {
    modal.remove();
  }

  // remove global handlers
  delete window.chatModule;
}

/**
 * Open chat with a specific user.
 * Called when user clicks on someone in the user list.
 */
export async function openChatWithUser(userId, userName) {
  try {
    console.log('Opening chat with user:', userName);
    const chat = await initializeChat(userId);
    if (!chat.id) throw new Error('Chat initialization failed');

    chatState.currentChat = chat;
    chatState.userChatMap[userId] = chat.id;
    renderChatWindow(userId, userName);

    const existingMessages = chatState.messageBuffer[chat.id];
    if (existingMessages && existingMessages.length > 0) {
      renderChatMessages();
      scrollToBottom();
    }
    loadChatHistory(chat.id);
  } catch (err) {
    console.error('Failed to open chat:', err);
    alert('Failed to open chat. Please try again.');
  }
}

/**
 * Close the current chat window and go back to user list.
 */
export function closeChatWindow() {
  if (chatState.currentChat) {
    sendChatViewEvent('chat.close', chatState.currentChat.id);
    clearPendingHistoryLoad(chatState.currentChat.id);
  }
  if (chatState.currentChat && chatState.activeChatUserId) {
    const typingKey = `${chatState.currentChat.id}:${chatState.activeChatUserId}`;
    if (chatState.typingTimeouts[typingKey]) {
      clearTimeout(chatState.typingTimeouts[typingKey]);
      delete chatState.typingTimeouts[typingKey];
    }
  }
  disconnectHistoryObserver();
  chatState.currentChat = null;
  chatState.activeChatUserId = null;
  renderChatModal();
}

/**
 * Close the entire chat modal.
 */
export function closeChatModal() {
  if (chatState.currentChat) {
    sendChatViewEvent('chat.close', chatState.currentChat.id);
    clearPendingHistoryLoad(chatState.currentChat.id);
  }
  if (chatState.currentChat && chatState.activeChatUserId) {
    const typingKey = `${chatState.currentChat.id}:${chatState.activeChatUserId}`;
    if (chatState.typingTimeouts[typingKey]) {
      clearTimeout(chatState.typingTimeouts[typingKey]);
      delete chatState.typingTimeouts[typingKey];
    }
  }
  disconnectHistoryObserver();
  chatState.activeChatUserId = null;
  const modal = document.getElementById('chat-modal');
  if (modal) {
    modal.style.display = 'none';
    modal.innerHTML = '';
  }
}

/**
 * Send a message in the current chat.
 */
export async function sendMessage(content) {
  if (!chatState.currentChat || !content.trim()) return;

  const trimmedContent = content.trim();
  const clientMessageId = `${chatState.currentUser.id}-${Date.now()}`;
  const chatId = chatState.currentChat.id;

  const optimisticMessage = {
    id: 0,
    chat_id: chatId,
    sender_id: chatState.currentUser.id,
    content: trimmedContent,
    created_at: new Date().toISOString(),
    client_message_id: clientMessageId,
    isOptimistic: true,
  };

  if (!chatState.messageBuffer[chatId]) {
    chatState.messageBuffer[chatId] = [];
  }
  chatState.messageBuffer[chatId].push(optimisticMessage);
  renderChatMessages();
  scrollToBottom();

  const message = {
    type: 'chat.send',
    request_id: `msg-${Date.now()}`,
    payload: {
      chat_id: chatId,
      content: trimmedContent,
      client_message_id: clientMessageId,
    },
  };

  if (chatState.ws && chatState.ws.readyState === WebSocket.OPEN) {
    chatState.ws.send(JSON.stringify(message));
  } else {
    console.error('WebSocket not ready');
    alert('Unable to send message. WebSocket is not connected.');
  }
}

// ─── WebSocket Management ─────────────────────────────────────────────────────

function connectWebSocket() {
  if (chatState.isConnecting || (chatState.ws && chatState.ws.readyState === WebSocket.OPEN))
    return;
  chatState.isConnecting = true;

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const wsUrl = `${protocol}//${window.location.host}/api/v1/ws`;

  try {
    chatState.ws = new WebSocket(wsUrl);
  } catch (err) {
    console.error('Failed to create WebSocket:', err);
    chatState.isConnecting = false;
    return;
  }

  chatState.ws.onopen = () => {
    console.log('✓ WebSocket connected');
    chatState.isConnecting = false;
  };

  chatState.ws.onmessage = (event) => {
    try {
      const envelope = JSON.parse(event.data);
      handleWebSocketMessage(envelope);
    } catch (err) {
      console.error('Failed to parse WebSocket message:', err, 'Raw data:', event.data);
    }
  };

  chatState.ws.onerror = (err) => {
    console.error('✗ WebSocket error:', err);
  };

  chatState.ws.onclose = (event) => {
    console.log('✗ WebSocket disconnected', event.code, event.reason);
    chatState.isConnecting = false;
    if (chatState.chatInitialized) {
      chatState.reconnectTimeout = setTimeout(() => {
        connectWebSocket();
      }, 3000);
    }
  };
}

function handleWebSocketMessage(envelope) {
  const { type, request_id, payload } = envelope;

  switch (type) {
    case 'isOnlineStatus.update':
      console.log('a user activity has changed', payload);
      handleChatUserOnlineStatus(payload.user_id, payload.isOnline);
      break;
    case 'chat.message':
      handleChatMessage(payload);
      break;
    case 'chat.is_typing':
      handleChatTypingStatus(payload.user_id, payload.chat_id);
      break;
    case 'chat.history_result':
      handleChatHistory(payload, request_id);
      break;
    case 'error':
      console.error('Server error:', payload?.message);
      break;
    case 'pong':
      // Keep-alive pong, no action needed
      break;
    default:
      console.warn('Unknown message type:', type);
  }
}

function handleChatUserOnlineStatus(userId, isOnline) {
  const user = chatState.chatUsers.find((u) => u.user_id === userId);
  if (!user) return;

  user.is_online = isOnline;
  const statusEl = document.querySelector(`[data-chat-user-status="${userId}"]`);
  if (statusEl) {
    const lastMessageTime = user.last_message_at ? new Date(user.last_message_at) : null;
    const timeAgo = lastMessageTime ? getTimeAgo(lastMessageTime) : 'Never';

    statusEl.className = `chat-user-status ${isOnline ? 'online' : 'offline'}`;
    statusEl.textContent = isOnline ? 'Online' : `Last seen ${timeAgo}`;
  }

  if (chatState.activeChatUserId === userId) {
    renderActiveChatTypingState();
  }
}

function handleChatMessage(message) {
  const msg = normalizeMessage(message);
  const { chat_id, sender_id, content, created_at, id } = msg;

  // Store message
  if (!chatState.messageBuffer[chat_id]) {
    chatState.messageBuffer[chat_id] = [];
  }

  if (id > 0 && chatState.messageBuffer[chat_id].some((m) => m.id === id)) {
    console.log('⏭️ Skipped duplicate message');
    return;
  }

  // Replace the optimistic client copy with the persisted message.
  if (msg.client_message_id) {
    const beforeFilter = chatState.messageBuffer[chat_id].length;
    chatState.messageBuffer[chat_id] = chatState.messageBuffer[chat_id].filter(
      (m) => m.client_message_id !== msg.client_message_id
    );
  }

  chatState.messageBuffer[chat_id].push({
    id,
    chat_id,
    sender_id,
    content,
    created_at,
    client_message_id: msg.client_message_id,
  });

  const senderInList = chatState.chatUsers.find((u) => u.user_id === sender_id);
  if (senderInList) {
    chatState.userChatMap[sender_id] = chat_id;
    senderInList.last_message_at = created_at;
    chatState.chatUsers.sort((a, b) => {
      const timeA = a.last_message_at ? new Date(a.last_message_at) : new Date(0);
      const timeB = b.last_message_at ? new Date(b.last_message_at) : new Date(0);
      return timeB - timeA;
    });
  }

  const chatIsVisible =
    chatState.currentChat &&
    chatState.currentChat.id === chat_id &&
    document.getElementById('chat-messages') !== null;

  if (chatIsVisible) {
    if (chatState.activeChatUserId === sender_id) {
      const typingKey = `${chat_id}:${sender_id}`;
      if (chatState.typingTimeouts[typingKey]) {
        clearTimeout(chatState.typingTimeouts[typingKey]);
        delete chatState.typingTimeouts[typingKey];
      }
    }
    renderChatMessages();
    scrollToBottom();
    markAsRead();
  } else {
    chatState.unreadCounts[chat_id] = (chatState.unreadCounts[chat_id] || 0) + 1;
    updateUnreadBadges();
  }
}

function handleChatHistory(messages, requestId) {
  const requestMeta = requestId ? chatState.pendingHistoryRequests[requestId] : null;
  if (requestId) {
    delete chatState.pendingHistoryRequests[requestId];
  }

  const chat_id = requestMeta?.chatId || chatState.currentChat?.id;
  if (!chat_id) return;

  const state = getHistoryState(chat_id);
  clearPendingHistoryLoad(chat_id);
  state.loadingOlder = false;
  showHistoryLoader(false);

  if (!chatState.currentChat || chatState.currentChat.id !== chat_id) return;

  if (!messages || !Array.isArray(messages)) messages = [];

  const normalizedMessages = messages.map((msg) => normalizeMessage(msg));

  if (!chatState.messageBuffer[chat_id]) {
    chatState.messageBuffer[chat_id] = [];
  }

  const incomingClientMessageIds = new Set(
    normalizedMessages.map((msg) => msg.client_message_id).filter(Boolean)
  );

  if (incomingClientMessageIds.size > 0) {
    chatState.messageBuffer[chat_id] = chatState.messageBuffer[chat_id].filter(
      (msg) =>
        !(
          msg.id === 0 &&
          msg.client_message_id &&
          incomingClientMessageIds.has(msg.client_message_id)
        )
    );
  }

  const existingIds = new Set(
    chatState.messageBuffer[chat_id].filter((msg) => msg.id > 0).map((msg) => msg.id)
  );

  const newMessages = normalizedMessages.filter((m) => !existingIds.has(m.id));
  chatState.messageBuffer[chat_id] = [...newMessages, ...chatState.messageBuffer[chat_id]];
  state.hasMore = normalizedMessages.length === CHAT_HISTORY_PAGE_SIZE;

  renderChatMessages();

  if (requestMeta?.beforeMessageId > 0) {
    restoreScrollAfterPrepend(requestMeta.previousScrollHeight, requestMeta.previousScrollTop);
    return;
  }

  scrollToBottom();
  markAsRead();
}

export function openChatModal() {
  const modal = document.getElementById('chat-modal');
  if (!modal) return;

  if (modal.innerHTML.includes('chat-window')) {
    // Chat window is open, close it
    closeChatWindow();
  } else {
    // Show modal if hidden or re-render it
    modal.style.display = 'block';
    renderChatModal();
  }
}

function updateUnreadBadges() {
  const totalUnread = Object.values(chatState.unreadCounts).reduce((a, b) => a + b, 0);

  const badge = document.getElementById('chat-unread-total');
  if (badge) {
    if (totalUnread > 0) {
      badge.textContent = totalUnread;
      badge.style.display = 'inline-flex';
    } else {
      badge.style.display = 'none';
    }
  }

  // Ανανέωσε και τη λίστα χρηστών αν είναι ανοιχτή
  const usersList = document.getElementById('chat-users-list');
  if (usersList) {
    renderChatUsersList();
  }
}

// ─── Data Loading ─────────────────────────────────────────────────────────────

export async function loadChatUsers() {
  try {
    const users = await fetchChatUsers();
    chatState.chatUsers = users || [];
    chatState.userChatMap = {};
    chatState.unreadCounts = {};

    chatState.chatUsers.forEach((user) => {
      if (user.chat_id) {
        chatState.userChatMap[user.user_id] = user.chat_id;
        chatState.unreadCounts[user.chat_id] = user.unread_count || 0;
      }
    });
    updateUnreadBadges();

    // Sort by last_message_at (most recent first)
    chatState.chatUsers.sort((a, b) => {
      const timeA = a.last_message_at ? new Date(a.last_message_at) : new Date(0);
      const timeB = b.last_message_at ? new Date(b.last_message_at) : new Date(0);
      return timeB - timeA;
    });

    renderChatUsersList();
  } catch (err) {
    console.error('Failed to load chat users:', err);
    const container = document.getElementById('chat-users-list');
    if (container) {
      container.innerHTML = `
        <div class="chat-empty-state">
          <div class="chat-empty-state-icon">⚠️</div>
          <div class="chat-empty-state-text">Failed to load users</div>
        </div>
      `;
    }
  }
}

async function loadChatHistory(chatId, beforeMessageId = 0) {
  if (!chatId) {
    console.error('No chatId provided to loadChatHistory');
    return;
  }

  const state = getHistoryState(chatId);
  const isLoadingOlder = beforeMessageId > 0;
  if (isLoadingOlder && (state.loadingOlder || state.delayPending || !state.hasMore)) {
    return;
  }

  if (!chatState.ws || chatState.ws.readyState !== WebSocket.OPEN) {
    setTimeout(() => loadChatHistory(chatId, beforeMessageId), 500);
    return;
  }

  if (isLoadingOlder) {
    state.loadingOlder = true;
    showHistoryLoader(true);
  }

  const requestId = `history-${Date.now()}-${beforeMessageId}`;
  const container = document.getElementById('chat-messages');
  chatState.pendingHistoryRequests[requestId] = {
    chatId,
    beforeMessageId,
    previousScrollHeight: isLoadingOlder && container ? container.scrollHeight : 0,
    previousScrollTop: isLoadingOlder && container ? container.scrollTop : 0,
  };

  const message = {
    type: 'chat.history',
    request_id: requestId,
    payload: {
      chat_id: chatId,
      before_message_id: beforeMessageId,
      limit: CHAT_HISTORY_PAGE_SIZE,
    },
  };

  chatState.ws.send(JSON.stringify(message));
}

export function markAsRead() {
  if (!chatState.currentChat || !chatState.ws || chatState.ws.readyState !== WebSocket.OPEN) return;

  const chatId = chatState.currentChat.id;
  const messages = chatState.messageBuffer[chatId] || [];
  if (messages.length === 0) return;

  const lastMessageId = Math.max(...messages.map((m) => m.id).filter((id) => id > 0));
  if (!lastMessageId) return;

  const message = {
    type: 'chat.mark_read',
    payload: {
      chat_id: chatId,
      up_to_message_id: lastMessageId,
    },
  };

  chatState.ws.send(JSON.stringify(message));
  chatState.unreadCounts[chatId] = 0;
  updateUnreadBadges();
}

// ─── Utility Functions ─────────────────────────────────────────────────────────

export function scrollToBottom() {
  const container = document.getElementById('chat-messages');
  if (container) {
    requestAnimationFrame(() => {
      container.scrollTop = container.scrollHeight;
      observeHistorySentinel();
    });
  }
}

export function getTimeAgo(date) {
  const seconds = Math.floor((new Date() - date) / 1000);

  if (seconds < 60) return 'just now';
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
  if (seconds < 604800) return `${Math.floor(seconds / 86400)}d ago`;

  return date.toLocaleDateString();
}
