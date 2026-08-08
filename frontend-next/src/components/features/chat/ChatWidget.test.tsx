import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import React from 'react';
import ChatWidget from './ChatWidget';

const mockGetChats = vi.fn();
const mockStartChat = vi.fn();
const mockGetChatMessages = vi.fn();

vi.mock('@/lib/api', () => ({
  getChats: () => mockGetChats(),
  startChat: (userId: string) => mockStartChat(userId),
  getChatMessages: (chatId: string) => mockGetChatMessages(chatId),
}));

vi.mock('@/lib/ws', () => ({
  chatSocket: {
    on: vi.fn(() => () => {}),
    send: vi.fn(),
    connect: vi.fn(),
    disconnect: vi.fn(),
  },
}));

vi.mock('@/context/AuthContext', () => ({
  useAuth: () => ({
    user: {
      id: 'u1',
      email: 'alice@example.com',
      username: 'alice',
      firstName: 'Alice',
      lastName: 'Smith',
      dateOfBirth: '1990-01-01',
      isPublic: true,
      createdAt: '2024-01-01T00:00:00Z',
    },
    loading: false,
    setUser: vi.fn(),
    clearUser: vi.fn(),
    refreshUser: vi.fn(),
  }),
}));

vi.mock('next/image', () => ({
  default: (props: React.ImgHTMLAttributes<HTMLImageElement>) => <img {...props} />,
}));

const conversation = {
  id: 'c1',
  participants: [{ id: 'u2', username: 'bob', isOnline: true }],
  unreadCount: 0,
  createdAt: '2024-01-01T00:00:00Z',
};

describe('ChatWidget', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetChats.mockResolvedValue([]);
    mockStartChat.mockResolvedValue({ id: 'c1' });
    mockGetChatMessages.mockResolvedValue([]);
  });

  it('opens an existing chat when /chat/start returns a chat without participants', async () => {
    // /chat/start returns the raw chat wire shape ({id, user_one_id, ...}) which
    // has no `participants`. The active chat must come from the conversation
    // list, otherwise rendering otherUser crashes on undefined.find.
    mockGetChats.mockResolvedValueOnce([]).mockResolvedValueOnce([conversation]);

    render(<ChatWidget />);

    window.dispatchEvent(new CustomEvent('chat:open', { detail: { userId: 'u2' } }));

    await waitFor(() => {
      expect(screen.getByText('bob')).toBeTruthy();
    });
  });

  it('clears the unread badge when a chat with unread messages is opened', async () => {
    const unreadConversation = { ...conversation, unreadCount: 3 };
    mockGetChats.mockResolvedValueOnce([]).mockResolvedValueOnce([unreadConversation]);

    const { container } = render(<ChatWidget />);

    window.dispatchEvent(new CustomEvent('chat:open', { detail: { userId: 'u2' } }));

    await waitFor(() => {
      expect(screen.getByText('bob')).toBeTruthy();
    });
    // Opening the chat marks it read, so the floating button must no longer
    // advertise unread messages.
    const widgetButton = container.querySelector('.chat-widget-button');
    expect(widgetButton).toBeTruthy();
    expect(widgetButton).not.toHaveAttribute('data-unread');
  });
});
