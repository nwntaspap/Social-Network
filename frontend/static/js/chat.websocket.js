import {
  chatState,
  markAsRead,
  scrollToBottom,
  updateUnreadBadges,
  getTimeAgo,
  normalizeMessage,
} from './chat.js';
import {
  renderChatMessages,
  renderChatUsersList,
  renderActiveChatTypingState,
} from './chat.render.js';
import {
  loadChatHistory,
  getHistoryState,
  clearPendingHistoryLoad,
  restoreScrollAfterPrepend,
  showHistoryLoader,
  CHAT_HISTORY_PAGE_SIZE,
} from './chat.history.js';

// ─── WebSocket Management ─────────────────────────────────────────────────────

export function connectWebSocket() {
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

/**
 * Handles Websocket Events based on Type
 * @param {Object} envelope - The incoming WebSocket message wrapper.
 * @param {string} envelope.type - The type of the event (e.g., 'chat.message').
 * @param {string} [envelope.request_id] - Optional ID to track specific requests/responses.
 * @param {Object} envelope.payload - The actual data associated with the event.
 */
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
