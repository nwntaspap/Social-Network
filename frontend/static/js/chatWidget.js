const CHAT_WIDGET_HTML = /* html */ `
<div id="chat-widget" class="chat-widget" aria-label="Open chat widget">
  <button id="chat-toggle-btn" class="chat-toggle-btn" aria-label="Open Messages">
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M20 2H4C2.9 2 2 2.9 2 4V22L6 18H20C21.1 18 22 17.1 22 16V4C22 2.9 21.1 2 20 2ZM20 16H5.17L4 17.17V4H20V16Z" fill="currentColor"/>
    </svg>
    <span class="chat-label">Chat</span>
    <span id="chat-unread-badge" class="chat-unread-badge" hidden>0</span>
  </button>
</div>

<div id="chat-modal" class="chat-modal" aria-hidden="true" style="display: none;">
  <div class="chat-modal-content">
    <div class="chat-modal-header">
      <div>
        <h3>Messages</h3>
        <p class="chat-modal-subtitle">Your conversations are separate from the navbar.</p>
      </div>
      <button id="chat-close-btn" class="chat-close-btn" aria-label="Close Messages">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z" fill="currentColor"/>
        </svg>
      </button>
    </div>
    <div class="chat-modal-body">
      <div class="chat-empty-state">
        <p class="chat-empty-title">Your chat panel is ready.</p>
        <p class="chat-empty-copy">This widget is separate from the navbar and will hold your live messages.</p>
      </div>
    </div>
    <div class="chat-modal-footer">
      <button id="chat-close-footer-btn" class="chat-start-btn" type="button">Close</button>
    </div>
  </div>
</div>
`;

function openChatModal() {
  const modal = document.getElementById('chat-modal');
  if (!modal) return;
  modal.style.display = 'grid';
  modal.setAttribute('aria-hidden', 'false');
  document.body.classList.add('chat-modal-open');
}

function closeChatModal() {
  const modal = document.getElementById('chat-modal');
  if (!modal) return;
  modal.style.display = 'none';
  modal.setAttribute('aria-hidden', 'true');
  document.body.classList.remove('chat-modal-open');
}

function onDocumentKeydown(event) {
  if (event.key === 'Escape') {
    closeChatModal();
  }
}

export function initChatWidget() {
  if (document.getElementById('chat-widget') || document.getElementById('chat-modal')) {
    return;
  }

  document.body.insertAdjacentHTML('beforeend', CHAT_WIDGET_HTML);

  const toggleButton = document.getElementById('chat-toggle-btn');
  const closeButton = document.getElementById('chat-close-btn');
  const closeFooterButton = document.getElementById('chat-close-footer-btn');
  const modal = document.getElementById('chat-modal');

  if (toggleButton) {
    toggleButton.addEventListener('click', openChatModal);
  }

  if (closeButton) {
    closeButton.addEventListener('click', closeChatModal);
  }

  if (closeFooterButton) {
    closeFooterButton.addEventListener('click', closeChatModal);
  }

  if (modal) {
    modal.addEventListener('click', (event) => {
      if (event.target === modal) {
        closeChatModal();
      }
    });
  }

  window.addEventListener('keydown', onDocumentKeydown);
}

export function setChatUnreadCount(count) {
  const badge = document.getElementById('chat-unread-badge');
  if (!badge) return;
  if (count > 0) {
    badge.hidden = false;
    badge.textContent = `${count}`;
  } else {
    badge.hidden = true;
  }
}
