import { chatState } from './chat.js';

export const CHAT_HISTORY_PAGE_SIZE = 20;
export const CHAT_HISTORY_LOAD_DELAY_MS = 2000;

export async function loadChatHistory(chatId, beforeMessageId = 0) {
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

export function getHistoryState(chatId) {
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

export function showHistoryLoader(show) {
  const loader = document.getElementById('chat-history-loader');
  if (!loader) return;

  loader.style.display = show ? 'flex' : 'none';
}

/**
 * Disconnects History Observer in certain actions, when we cleanupChat(),
 * when we closeChatWindow() and when we closeChatModal()
 */
export function disconnectHistoryObserver() {
  if (chatState.historyObserver) {
    chatState.historyObserver.disconnect();
    chatState.historyObserver = null;
  }
}

/**
 * Clears History State in certain actions, when we closeChatWindow() and when we closeChatModal()
 * when calling handleChatHistory() which is called after the backend has sent an array of messages
 */
export function clearPendingHistoryLoad(chatId) {
  const state = chatState.historyState[chatId];
  if (!state) return;

  if (state.pendingTimeoutId) {
    clearTimeout(state.pendingTimeoutId);
    state.pendingTimeoutId = null;
  }

  state.delayPending = false;
}

export function clearAllPendingHistoryLoads() {
  Object.keys(chatState.historyState).forEach((chatId) => {
    clearPendingHistoryLoad(chatId);
  });
}

export function restoreScrollAfterPrepend(previousScrollHeight, previousScrollTop) {
  const container = document.getElementById('chat-messages');
  if (!container) return;

  requestAnimationFrame(() => {
    container.scrollTop = container.scrollHeight - previousScrollHeight + previousScrollTop;
    observeHistorySentinel();
  });
}

export function observeHistorySentinel() {
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

export function getOldestPersistedMessageId(chatId) {
  const persistedMessages = (chatState.messageBuffer[chatId] || []).filter((msg) => msg.id > 0);

  if (persistedMessages.length === 0) return 0;

  return Math.min(...persistedMessages.map((msg) => msg.id));
}
