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
import { connectWebSocket, sendChatViewEvent } from './chat.websocket.js';
import {
  loadChatHistory,
  disconnectHistoryObserver,
  clearPendingHistoryLoad,
  clearAllPendingHistoryLoads,
  observeHistorySentinel,
} from './chat.history.js';
import {
  renderChatWindow,
  renderChatMessages,
  renderChatModal,
  renderChatUsersList,
  renderChatWidget,
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

export function updateUnreadBadges() {
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

/**
 * Normalize message properties from backend PascalCase to our snake_case format
 */
export function normalizeMessage(msg) {
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
