import { escapeHTML, formatMessageTime, throttle } from './helpers.js';
import {
  chatState,
  markAsRead,
  scrollToBottom,
  openChatModal,
  closeChatWindow,
  closeChatModal,
  sendMessage,
  openChatWithUser,
  loadChatUsers,
  getTimeAgo,
} from './chat.js';
import {
  disconnectHistoryObserver,
} from './chat.history.js';
import { sendChatViewEvent,sendTypingEvent } from './chat.websocket.js';

// ─── Chat Window (Message Interface) ─────────────────────────────────────────
const CHAT_TYPING_THROTTLE_MS = 1000;

export function renderChatWindow(userId, userName) {
  const root = document.getElementById('chat-modal');
  if (!root || !chatState.currentChat) return;
  chatState.activeChatUserId = userId;

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

  sendChatViewEvent('chat.open', chatState.currentChat.id);
  renderChatStatus();

  // Focus input
  setTimeout(() => {
    const input = document.getElementById('chat-message-input');
    if (input) {
      const sendTypingThrottled = throttle(() => {
        if (!chatState.currentChat || !input.value.trim()) {
          return;
        }

        sendTypingEvent(chatState.currentChat.id);
      }, CHAT_TYPING_THROTTLE_MS);

      input.addEventListener('input', () => {
        sendTypingThrottled();
      });

      input.focus();
    }
    markAsRead();
  }, 100);
}

export function renderChatMessages() {
  if (!chatState.currentChat) return;

  const chatId = chatState.currentChat.id;
  const container = document.getElementById('chat-messages');
  if (!container) return;

  const messages = chatState.messageBuffer[chatId] || [];

  // Display messages sorted by time (oldest first for display)
  const displayMessages = [...messages].sort(
    (a, b) => new Date(a.created_at) - new Date(b.created_at)
  );

  if (displayMessages.length === 0) {
    disconnectHistoryObserver();
    container.innerHTML = `
      <div class="chat-empty-state">
        <div class="chat-empty-state-icon">💬</div>
        <div class="chat-empty-state-text">No messages yet. Start the conversation!</div>
      </div>
    `;
    return;
  }

  container.innerHTML = `
    <div id="chat-history-loader" class="chat-loading" style="display: none; min-height: auto; padding: 8px 0;">
      <div class="chat-spinner"></div>
    </div>
    <div id="chat-history-sentinel" style="height: 1px; width: 100%;"></div>
    ${displayMessages
      .map((msg) => {
        const isOwn = msg.sender_id === chatState.currentUser.id;

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
      .join('')}
  `;

  ensureTypingBubble(isActiveChatTyping());
}

// ─── Chat Modal (User List) ──────────────────────────────────────────────────

export function renderChatModal() {
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

export function renderChatUsersList() {
  const container = document.getElementById('chat-users-list');
  if (!container) return;

  if (chatState.chatUsers.length === 0) {
    container.innerHTML = `
      <div class="chat-empty-state">
        <div class="chat-empty-state-icon">👥</div>
        <div class="chat-empty-state-text">No users online right now</div>
      </div>
    `;
    return;
  }

  container.innerHTML = chatState.chatUsers
    .map((user) => {
      const initials = (user.nickname || 'U')
        .split(' ')
        .map((word) => word[0].toUpperCase())
        .join('')
        .slice(0, 2);

      // Get unread count for this user by looking up their chat_id
      const chatId = chatState.userChatMap[user.user_id];
      const unreadCount = chatId ? chatState.unreadCounts[chatId] || 0 : 0;

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
            <div class="chat-user-status ${user.is_online ? 'online' : 'offline'}" data-chat-user-status="${user.user_id}">
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

export function renderChatWidget() {
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

export function renderChatStatus() {
  const statusEl = document.getElementById('chat-status');
  if (!statusEl || !chatState.activeChatUserId) return;

  const activeUser = chatState.chatUsers.find(
    (user) => user.user_id === chatState.activeChatUserId
  );
  if (!activeUser) {
    statusEl.className = 'chat-window-header-subtitle';
    statusEl.textContent = 'Offline';
    return;
  }

  const lastMessageTime = activeUser.last_message_at ? new Date(activeUser.last_message_at) : null;
  const timeAgo = lastMessageTime ? getTimeAgo(lastMessageTime) : 'Never';
  statusEl.className = 'chat-window-header-subtitle';
  statusEl.textContent = activeUser.is_online ? 'Online' : `Last seen ${timeAgo}`;
}

export function renderActiveChatTypingState() {
  renderChatStatus();

  if (chatState.currentChat && document.getElementById('chat-messages')) {
    ensureTypingBubble(isActiveChatTyping());
    scrollToBottom();
  }
}

function ensureTypingBubble(isTyping) {
  const container = document.getElementById('chat-messages');
  if (!container) {
    return;
  }
  const existingBubble = document.getElementById('chat-typing-indicator');
  if (!isTyping) {
    existingBubble?.remove();
    return;
  }
  if (existingBubble) {
    return;
  }
  const bubble = document.createElement('div');
  bubble.id = 'chat-typing-indicator';
  bubble.className = 'chat-message chat-message-typing';
  bubble.innerHTML = `
    <div>
      <div class="chat-message-bubble chat-message-bubble-typing">
        <span class="chat-typing-dots">
          <span></span>
          <span></span>
          <span></span>
        </span>
      </div>
    </div>
  `;
  container.appendChild(bubble);
}

function isActiveChatTyping() {
  if (!chatState.currentChat || !chatState.activeChatUserId) {
    return false;
  }

  return Boolean(
    chatState.typingTimeouts[`${chatState.currentChat.id}:${chatState.activeChatUserId}`]
  );
}
