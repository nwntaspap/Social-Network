import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import React from 'react';
import ChatWindow from './ChatWindow';
import type { ChatUser } from '@/lib/types';

const mockGetChatMessages = vi.fn();

type WsHandler = (payload: unknown) => void;

const wsHandlers = new Map<string, WsHandler>();
const wsSends: Array<{ type: string; payload: unknown }> = [];

vi.mock('@/lib/api', () => ({
  getChatMessages: (chatId: string) => mockGetChatMessages(chatId),
}));

vi.mock('@/lib/ws', () => ({
  chatSocket: {
    on: vi.fn((type: string, handler: WsHandler) => {
      wsHandlers.set(type, handler);
      return () => {
        wsHandlers.delete(type);
      };
    }),
    send: vi.fn((type: string, payload?: unknown) => {
      wsSends.push({ type, payload });
    }),
    connect: vi.fn(),
    disconnect: vi.fn(),
  },
}));

vi.mock('next/image', () => ({
  default: (props: React.ImgHTMLAttributes<HTMLImageElement>) => <img {...props} />,
}));

const otherUser: ChatUser = {
  id: 'u2',
  username: 'bob',
  avatarUrl: undefined,
  isOnline: true,
};

describe('ChatWindow', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    wsHandlers.clear();
    wsSends.length = 0;
    mockGetChatMessages.mockResolvedValue([]);
  });

  it('renders history in the order returned (oldest first)', async () => {
    mockGetChatMessages.mockResolvedValue([
      {
        id: '1',
        senderId: 'u2',
        content: 'first message',
        createdAt: '2024-01-01T10:00:00Z',
      },
      {
        id: '2',
        senderId: 'u1',
        content: 'second message',
        createdAt: '2024-01-01T10:01:00Z',
      },
    ]);

    const { container } = render(
      <ChatWindow chatId="c1" currentUserId="u1" otherUser={otherUser} embedded />
    );

    await screen.findByText('first message');

    const bubbles = Array.from(container.querySelectorAll('.chat-message-bubble'));
    expect(bubbles[0].textContent).toContain('first message');
    expect(bubbles[1].textContent).toContain('second message');
  });

  it('shows the typing indicator on chat.is_typing from the other user', async () => {
    const { container } = render(
      <ChatWindow chatId="c1" currentUserId="u1" otherUser={otherUser} embedded />
    );

    await waitFor(() => {
      expect(wsHandlers.has('chat.is_typing')).toBe(true);
    });

    wsHandlers.get('chat.is_typing')!({ chat_id: 'c1', user_id: 'u2' });

    await waitFor(() => {
      expect(container.querySelector('.chat-message-typing')).not.toBeNull();
    });
  });

  it('clears the typing indicator immediately when the other user sends a message', async () => {
    const { container } = render(
      <ChatWindow chatId="c1" currentUserId="u1" otherUser={otherUser} embedded />
    );

    await waitFor(() => {
      expect(wsHandlers.has('chat.message')).toBe(true);
    });

    wsHandlers.get('chat.is_typing')!({ chat_id: 'c1', user_id: 'u2' });
    await waitFor(() => {
      expect(container.querySelector('.chat-message-typing')).not.toBeNull();
    });

    wsHandlers.get('chat.message')!({
      id: 1,
      chat_id: 'c1',
      sender_id: 'u2',
      content: 'hi bob',
      created_at: '2024-01-01T10:02:00Z',
    });

    await waitFor(() => {
      expect(container.querySelector('.chat-message-typing')).toBeNull();
    });
    expect(container.querySelectorAll('.chat-message').length).toBe(1);
  });

  it('updates the header online status from isOnlineStatus.update broadcasts', async () => {
    const offlineUser = { ...otherUser, isOnline: false };
    render(<ChatWindow chatId="c1" currentUserId="u1" otherUser={offlineUser} embedded />);

    await waitFor(() => {
      expect(wsHandlers.has('isOnlineStatus.update')).toBe(true);
    });

    expect(screen.getByText('Offline')).toBeTruthy();

    wsHandlers.get('isOnlineStatus.update')!({ user_id: 'u2', isOnline: true });

    await waitFor(() => {
      expect(screen.getByText('Online')).toBeTruthy();
    });
  });
});
