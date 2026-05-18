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
import { escapeHTML, formatMessageTime } from './helpers.js';

// ─── State ────────────────────────────────────────────────────────────────────

let ws = null;
let currentUser = null;
let currentChat = null;
let chatUsers = [];
let messageBuffer = {}; // Stores messages per chat_id
let unreadCounts = {}; // Stores unread counts per chat_id
let userChatMap = {}; // Maps user_id -> chat_id for quick lookup
let isConnecting = false;
let reconnectTimeout = null;
let chatInitialized = false;

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

// ─── Public API ────────────────────────────────────────────────────────────────

/**
 * Initialize the chat system. Call once on app startup.
 */
export function initChat(user) {
  if (!user?.id) return;

  if (chatInitialized) {
    console.log('Chat already initialized');
    return;
  }

  console.log('Initializing chat for:', user.username || user.email);

  chatInitialized = true;

  currentUser = user;

  renderChatWidget();
  connectWebSocket();
}

export function cleanupChat() {
  console.log('Cleaning up chat');

  chatInitialized = false;

  // stop reconnect loop
  if (reconnectTimeout) {
    clearTimeout(reconnectTimeout);
    reconnectTimeout = null;
  }

  // close websocket
  if (ws) {
    ws.onclose = null;
    ws.close();
    ws = null;
  }

  // reset state
  currentUser = null;
  currentChat = null;
  chatUsers = [];
  messageBuffer = {};
  unreadCounts = {};
  userChatMap = {};
  isConnecting = false;

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
    console.log('Opening chat with user:', userId);
    const chat = await initializeChat(userId);
    console.log('Full chat response from backend:', chat);
    console.log('Chat ID:', chat.id);

    if (!chat.id) {
      throw new Error('Chat initialization failed: no chat ID returned');
    }

    currentChat = chat;
    userChatMap[userId] = chat.id; // Store mapping for unread lookups
    renderChatWindow(userId, userName);
    console.log('Chat window rendered, loading history for:', chat.id);
    // Show cached Messages before loading History for Better UX
    const existingMessages = messageBuffer[chat.id];
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
  currentChat = null;
  renderChatModal();
}

/**
 * Close the entire chat modal.
 */
export function closeChatModal() {
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
  if (!currentChat || !content.trim()) return;

  const trimmedContent = content.trim();
  const clientMessageId = `${currentUser.id}-${Date.now()}`;
  const chatId = currentChat.id;

  console.log('📤 Sending message with clientMessageId:', clientMessageId);

  // Add optimistic message to UI immediately
  const optimisticMessage = {
    id: 0,
    chat_id: chatId,
    sender_id: currentUser.id,
    content: trimmedContent,
    created_at: new Date().toISOString(),
    client_message_id: clientMessageId,
    isOptimistic: true,
  };

  if (!messageBuffer[chatId]) {
    messageBuffer[chatId] = [];
  }
  messageBuffer[chatId].push(optimisticMessage);
  console.log('📝 Added optimistic message, buffer size now:', messageBuffer[chatId].length);
  renderChatMessages();
  scrollToBottom();

  // Send via WebSocket
  console.log('Sending message to chat:', chatId, 'with content:', trimmedContent);

  const message = {
    type: 'chat.send',
    request_id: `msg-${Date.now()}`,
    payload: {
      chat_id: chatId,
      content: trimmedContent,
      client_message_id: clientMessageId,
    },
  };

  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify(message));
    console.log('Message sent:', clientMessageId);
  } else {
    console.error(
      'WebSocket not ready. State:',
      ws?.readyState,
      'Connected:',
      ws?.readyState === WebSocket.OPEN
    );
    alert('Unable to send message. WebSocket is not connected. Please wait and try again.');
  }
}

// ─── WebSocket Management ─────────────────────────────────────────────────────

function connectWebSocket() {
  if (isConnecting || (ws && ws.readyState === WebSocket.OPEN)) return;
  isConnecting = true;

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const wsUrl = `${protocol}//${window.location.host}/api/v1/ws`;

  console.log('Connecting to WebSocket:', wsUrl);

  try {
    ws = new WebSocket(wsUrl);
  } catch (err) {
    console.error('Failed to create WebSocket:', err);
    isConnecting = false;
    return;
  }

  ws.onopen = () => {
    console.log('✓ WebSocket connected');
    isConnecting = false;
  };

  ws.onmessage = (event) => {
    try {
      const envelope = JSON.parse(event.data);
      handleWebSocketMessage(envelope);
    } catch (err) {
      console.error('Failed to parse WebSocket message:', err, 'Raw data:', event.data);
    }
  };

  ws.onerror = (err) => {
    console.error('✗ WebSocket error:', err);
  };

  ws.onclose = (event) => {
    console.log('✗ WebSocket disconnected', event.code, event.reason);
    isConnecting = false;
    // Attempt to reconnect after 3 seconds
    if (chatInitialized) {
      reconnectTimeout = setTimeout(() => {
        connectWebSocket();
      }, 3000);
    }
  };
}

function handleWebSocketMessage(envelope) {
  const { type, request_id, payload } = envelope;

  switch (type) {
    case 'chat.message':
      console.log('Received chat.message:', payload);
      handleChatMessage(payload);
      break;
    case 'chat.history_result':
      console.log(
        'Received chat.history_result, payload:',
        payload,
        'isArray:',
        Array.isArray(payload),
        'length:',
        payload?.length
      );
      handleChatHistory(payload);
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

function handleChatMessage(message) {
  const msg = normalizeMessage(message);
  const { chat_id, sender_id, content, created_at, id } = msg;
  console.log('Received message:', {
    id,
    chat_id,
    sender_id,
    client_message_id: msg.client_message_id,
  });

  // Store message
  if (!messageBuffer[chat_id]) {
    messageBuffer[chat_id] = [];
  }

  // Check if message already exists (dedup by ID or client_message_id)
  const messageExists = messageBuffer[chat_id].some(
    (m) => m.id === id || (msg.client_message_id && m.client_message_id === msg.client_message_id)
  );

  if (!messageExists) {
    // Remove optimistic message with same client_message_id if it exists
    if (msg.client_message_id) {
      const beforeFilter = messageBuffer[chat_id].length;
      messageBuffer[chat_id] = messageBuffer[chat_id].filter(
        (m) => m.client_message_id !== msg.client_message_id
      );
      const afterFilter = messageBuffer[chat_id].length;
      console.log('🗑️ Removed optimistic:', {
        removed: beforeFilter - afterFilter,
        before: beforeFilter,
        after: afterFilter,
      });
    }

    messageBuffer[chat_id].push({
      id,
      chat_id,
      sender_id,
      content,
      created_at,
      client_message_id: msg.client_message_id,
    });
    console.log('✅ Added message to buffer, size now:', messageBuffer[chat_id].length);
  } else {
    console.log('⏭️ Skipped duplicate message');
  }

  const senderInList = chatUsers.find((u) => u.user_id === sender_id);
  if (senderInList) {
    userChatMap[sender_id] = chat_id;
  }

  const chatIsVisible =
    currentChat && currentChat.id === chat_id && document.getElementById('chat-messages') !== null;

  if (chatIsVisible) {
    renderChatMessages();
    scrollToBottom();
    markAsRead();
  } else {
    unreadCounts[chat_id] = (unreadCounts[chat_id] || 0) + 1;
    updateUnreadBadges();
  }
}

function handleChatHistory(messages) {
  if (!currentChat) {
    return;
  }

  const chat_id = currentChat.id;

  // Handle null/undefined messages (shouldn't happen, but be safe)
  if (!messages || !Array.isArray(messages)) {
    messages = [];
  }

  // Normalize all messages to snake_case
  const normalizedMessages = messages.map((msg) => normalizeMessage(msg));

  // Store messages in reverse chronological order (newest first from server)
  if (!messageBuffer[chat_id]) {
    messageBuffer[chat_id] = [];
  }

  // Merge with existing messages, avoiding duplicates
  const existingIds = new Set(messageBuffer[chat_id].filter((m) => m.id > 0).map((m) => m.id));

  const newMessages = normalizedMessages.filter((m) => !existingIds.has(m.id));

  messageBuffer[chat_id] = [...newMessages, ...messageBuffer[chat_id]];

  renderChatMessages();
  scrollToBottom();
  markAsRead();
}

// ─── Chat Window (Message Interface) ──────────────────────────────────────────

function renderChatWindow(userId, userName) {
  const root = document.getElementById('chat-modal');
  if (!root) return;

  // Get user's initials for avatar
  const initials = (userName || 'U')
    .split(' ')
    .map((word) => word[0].toUpperCase())
    .join('')
    .slice(0, 2);

  root.innerHTML = `
    <div class="chat-window">
      <div class="chat-window-header">
        <button class="chat-window-back" onclick="window.chatModule.closeChatWindow()">
          ←
        </button>
        <div class="chat-window-header-user">
          <div class="chat-user-avatar" style="width: 36px; height: 36px; font-size: 14px;">
            ${escapeHTML(initials)}
          </div>
          <div class="chat-window-header-title">
            <h3>${escapeHTML(userName)}</h3>
            <div class="chat-window-header-subtitle" id="chat-status">Online</div>
          </div>
        </div>
        <button class="chat-window-close" onclick="window.chatModule.closeChatModal()">
          ✕
        </button>
      </div>
      <div class="chat-messages-container" id="chat-messages">
        <div class="chat-loading">
          <div class="chat-spinner"></div>
        </div>
      </div>
      <div class="chat-input-area">
        <input
          type="text"
          class="chat-input"
          id="chat-message-input"
          placeholder="Type a message..."
          onkeypress="if(event.key==='Enter') window.chatModule.sendMessageFromInput()"
        />
        <button
          class="chat-send-button"
          id="chat-send-btn"
          onclick="window.chatModule.sendMessageFromInput()"
          title="Send message"
        >
          ↗
        </button>
      </div>
    </div>
  `;

  // Focus input
  setTimeout(() => {
    const input = document.getElementById('chat-message-input');
    if (input) input.focus();
    markAsRead();
  }, 100);
}

function renderChatMessages() {
  if (!currentChat) return;

  const chatId = currentChat.id;
  const container = document.getElementById('chat-messages');
  if (!container) return;

  const messages = messageBuffer[chatId] || [];
  console.log(
    'renderChatMessages - chatId:',
    chatId,
    'messages in buffer:',
    messages.length,
    'first msg structure:',
    messages[0]
  );

  // Display messages sorted by time (oldest first for display)
  const displayMessages = [...messages].sort(
    (a, b) => new Date(a.created_at) - new Date(b.created_at)
  );

  if (displayMessages.length === 0) {
    container.innerHTML = `
      <div class="chat-empty-state">
        <div class="chat-empty-state-icon">💬</div>
        <div class="chat-empty-state-text">No messages yet. Start the conversation!</div>
      </div>
    `;
    return;
  }

  container.innerHTML = displayMessages
    .map((msg) => {
      const isOwn = msg.sender_id === currentUser.id;
      console.log('Rendering message:', {
        id: msg.id,
        content: msg.content,
        created_at: msg.created_at,
        sender_id: msg.sender_id,
        isOwn,
        currentUserId: currentUser.id,
      });
      return `
        <div class="chat-message ${isOwn ? 'own' : ''}">
          <div>
            <div class="chat-message-bubble">
              ${escapeHTML(msg.content)}
            </div>
            <div class="chat-message-time">
              ${formatMessageTime(msg.created_at)}
            </div>
          </div>
        </div>
      `;
    })
    .join('');
}

// ─── Chat Modal (User List) ──────────────────────────────────────────────────

function renderChatModal() {
  const root = document.getElementById('chat-modal');
  if (!root) return;

  root.innerHTML = `
    <div class="chat-modal">
      <div class="chat-modal-header">
        <h2>Messages</h2>
        <button class="chat-modal-close" onclick="window.chatModule.closeChatModal()">
          ✕
        </button>
      </div>
      <div class="chat-modal-content">
        <div class="chat-users-list" id="chat-users-list">
          <div class="chat-loading">
            <div class="chat-spinner"></div>
          </div>
        </div>
      </div>
    </div>
  `;

  loadChatUsers();
}

function renderChatUsersList() {
  const container = document.getElementById('chat-users-list');
  if (!container) return;

  if (chatUsers.length === 0) {
    container.innerHTML = `
      <div class="chat-empty-state">
        <div class="chat-empty-state-icon">👥</div>
        <div class="chat-empty-state-text">No users online right now</div>
      </div>
    `;
    return;
  }

  container.innerHTML = chatUsers
    .map((user) => {
      const initials = (user.nickname || 'U')
        .split(' ')
        .map((word) => word[0].toUpperCase())
        .join('')
        .slice(0, 2);

      // Get unread count for this user by looking up their chat_id
      const chatId = userChatMap[user.user_id];
      const unreadCount = chatId ? unreadCounts[chatId] || 0 : 0;

      const lastMessageTime = user.last_message_at ? new Date(user.last_message_at) : null;
      const timeAgo = lastMessageTime ? getTimeAgo(lastMessageTime) : 'Never';

      return `
        <div
          class="chat-user-item"
          onclick="window.chatModule.openChatWithUser('${user.user_id}', '${escapeHTML(user.nickname)}')"
        >
          <div class="chat-user-avatar">
            ${escapeHTML(initials)}
          </div>
          <div class="chat-user-info">
            <div class="chat-user-name">${escapeHTML(user.nickname)}</div>
            <div class="chat-user-status ${user.is_online ? 'online' : 'offline'}">
              ${user.is_online ? 'Online' : `Last seen ${timeAgo}`}
            </div>
          </div>
          ${unreadCount > 0 ? `<div class="chat-user-unread">${unreadCount}</div>` : ''}
        </div>
      `;
    })
    .join('');
}

// ─── Chat Widget (Button) ──────────────────────────────────────────────────────

function renderChatWidget() {
  // Check if widget already exists
  if (document.getElementById('chat-widget-button')) {
    return;
  }

  // Create widget button
  const button = document.createElement('button');
  button.id = 'chat-widget-button';
  button.className = 'chat-widget-button';
  button.innerHTML =
    '💬 <span class="chat-unread-badge" id="chat-unread-total" style="display:none"></span>';
  button.onclick = openChatModal;
  button.title = 'Open messages';

  document.body.appendChild(button);

  // Create modal container
  const modal = document.createElement('div');
  modal.id = 'chat-modal';
  document.body.appendChild(modal);

  // Make functions available globally for onclick handlers immediately
  window.chatModule = {
    closeChatWindow,
    closeChatModal,
    sendMessage,
    sendMessageFromInput() {
      const input = document.getElementById('chat-message-input');
      if (input) {
        const content = input.value;
        if (content.trim()) {
          sendMessage(content);
          input.value = '';
          input.focus();
        }
      }
    },
    openChatWithUser,
  };
}

function openChatModal() {
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
  const totalUnread = Object.values(unreadCounts).reduce((a, b) => a + b, 0);

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

async function loadChatUsers() {
  try {
    const users = await fetchChatUsers();
    chatUsers = users || [];

    // Sort by last_message_at (most recent first)
    chatUsers.sort((a, b) => {
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

async function loadChatHistory(chatId) {
  if (!chatId) {
    console.error('No chatId provided to loadChatHistory');
    return;
  }

  console.log('loadChatHistory called with chatId:', chatId);

  if (!ws || ws.readyState !== WebSocket.OPEN) {
    console.log('WebSocket not ready for history, retrying in 500ms. State:', ws?.readyState);
    // Wait for WebSocket to be ready
    setTimeout(() => loadChatHistory(chatId), 500);
    return;
  }

  const message = {
    type: 'chat.history',
    request_id: `history-${Date.now()}`,
    payload: {
      chat_id: chatId,
      before_message_id: 0,
      limit: 20,
    },
  };

  console.log('Requesting chat history for:', chatId, 'message:', message);
  console.log('WebSocket state before send:', ws.readyState, 'OPEN=', WebSocket.OPEN);
  ws.send(JSON.stringify(message));
  console.log('History request sent successfully');
}

function markAsRead() {
  if (!currentChat || !ws || ws.readyState !== WebSocket.OPEN) return;

  const chatId = currentChat.id;
  const messages = messageBuffer[chatId] || [];
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

  ws.send(JSON.stringify(message));
  unreadCounts[chatId] = 0;
  updateUnreadBadges();
}

// ─── Utility Functions ─────────────────────────────────────────────────────────

function scrollToBottom() {
  const container = document.getElementById('chat-messages');
  if (container) {
    setTimeout(() => {
      container.scrollTop = container.scrollHeight;
    }, 0);
  }
}

function getTimeAgo(date) {
  const seconds = Math.floor((new Date() - date) / 1000);

  if (seconds < 60) return 'just now';
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
  if (seconds < 604800) return `${Math.floor(seconds / 86400)}d ago`;

  return date.toLocaleDateString();
}
