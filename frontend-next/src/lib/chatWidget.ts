/**
 * lib/chatWidget.ts
 *
 * Tiny bridge so any component (e.g. a profile "Message" button) can ask the
 * global ChatWidget to open and start a conversation with a specific user.
 * The widget listens for the `chat:open` event.
 */

export interface OpenChatEventDetail {
  userId: string;
}

export const OPEN_CHAT_EVENT = 'chat:open';

/** Ask the chat popup to open a conversation with the given user. */
export function openChatWithUser(userId: string): void {
  if (typeof window === 'undefined') return;
  window.dispatchEvent(
    new CustomEvent<OpenChatEventDetail>(OPEN_CHAT_EVENT, { detail: { userId } })
  );
}
