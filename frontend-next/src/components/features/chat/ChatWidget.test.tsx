import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import React from 'react';
import ChatWidget from './ChatWidget';

const mockGetChats = vi.fn();
const mockStartChat = vi.fn();
const mockGetChatMessages = vi.fn();
const mockGetMyGroups = vi.fn();
const mockGetGroupPresence = vi.fn();
const mockGetGroupChatHistory = vi.fn();
const mockGetGroupMembers = vi.fn();

vi.mock('@/lib/api', () => ({
  getChats: () => mockGetChats(),
  startChat: (userId: string) => mockStartChat(userId),
  getChatMessages: (chatId: string) => mockGetChatMessages(chatId),
  getMyGroups: () => mockGetMyGroups(),
  getGroupPresence: (groupId: string) => mockGetGroupPresence(groupId),
  getGroupChatHistory: (groupId: string) => mockGetGroupChatHistory(groupId),
  getGroupMembers: (groupId: string) => mockGetGroupMembers(groupId),
}));

const wsHandlers: Record<string, (payload: unknown) => void> = {};
const mockChatSocket = vi.hoisted(() => ({
  on: vi.fn((type: string, handler: (payload: unknown) => void) => {
    wsHandlers[type] = handler;
    return () => {
      delete wsHandlers[type];
    };
  }),
  send: vi.fn(),
  connect: vi.fn(),
  disconnect: vi.fn(),
  isConnected: vi.fn(() => true),
  onConnection: vi.fn(() => () => {}),
}));

vi.mock('@/lib/ws', () => ({
  chatSocket: mockChatSocket,
}));

const authValue = {
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
};

vi.mock('@/context/AuthContext', () => ({
  useAuth: () => authValue,
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

const group = {
  id: 'g1',
  title: 'Go Meetup',
  description: 'Gophers',
  creatorId: 'u1',
  creator: { id: 'u1', username: 'alice', firstName: 'Alice', lastName: 'Smith' },
  membersCount: 3,
  membershipStatus: 'member' as const,
  createdAt: '2024-01-01T00:00:00Z',
};

function openWidget() {
  render(<ChatWidget />);
  fireEvent.click(screen.getByRole('button', { name: 'Open chat' }));
}

describe('ChatWidget', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetChats.mockResolvedValue([]);
    mockStartChat.mockResolvedValue({ id: 'c1' });
    mockGetChatMessages.mockResolvedValue([]);
    mockGetMyGroups.mockResolvedValue({ data: [group], totalCount: 1 });
    mockGetGroupPresence.mockResolvedValue({
      groupId: 'g1',
      total: 3,
      online: 2,
      members: [
        { id: 'u1', isOnline: true },
        { id: 'u2', isOnline: true },
        { id: 'u3', isOnline: false },
      ],
    });
    mockGetGroupChatHistory.mockResolvedValue([]);
    mockGetGroupMembers.mockResolvedValue({ data: [], totalCount: 0 });
    for (const key of Object.keys(wsHandlers)) delete wsHandlers[key];
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

  it('switches to the Groups tab and lists the user groups', async () => {
    openWidget();

    fireEvent.click(screen.getByRole('button', { name: 'Groups' }));

    await waitFor(() => {
      expect(screen.getByText('Go Meetup')).toBeTruthy();
    });
    expect(mockGetMyGroups).toHaveBeenCalled();
  });

  it('opens a group chat room with the live online indicator', async () => {
    openWidget();

    fireEvent.click(screen.getByRole('button', { name: 'Groups' }));
    fireEvent.click(await screen.findByText('Go Meetup'));

    await waitFor(() => {
      expect(mockGetGroupPresence).toHaveBeenCalledWith('g1');
    });
    expect(await screen.findByText('2 online')).toBeTruthy();
  });

  it('recomputes the online count on isOnlineStatus.update for a group member', async () => {
    openWidget();

    fireEvent.click(screen.getByRole('button', { name: 'Groups' }));
    fireEvent.click(await screen.findByText('Go Meetup'));
    await screen.findByText('2 online');

    // u3 comes online -> 3 online.
    wsHandlers['isOnlineStatus.update']({ user_id: 'u3', isOnline: true });

    expect(await screen.findByText('3 online')).toBeTruthy();

    // u1 goes offline -> 2 online.
    wsHandlers['isOnlineStatus.update']({ user_id: 'u1', isOnline: false });

    expect(await screen.findByText('2 online')).toBeTruthy();
  });

  it('back button returns from a group room to the Groups list', async () => {
    openWidget();

    fireEvent.click(screen.getByRole('button', { name: 'Groups' }));
    fireEvent.click(await screen.findByText('Go Meetup'));
    await screen.findByText('2 online');

    fireEvent.click(screen.getByRole('button', { name: 'Back to groups' }));

    expect(await screen.findByText('Go Meetup')).toBeTruthy();
    expect(screen.getByText('3 members')).toBeTruthy();
  });

  it('accumulates group unread badges and clears them when the group opens', async () => {
    // The message is persisted server-side, so the groups refresh reports it.
    mockGetMyGroups.mockResolvedValue({ data: [{ ...group, unreadCount: 1 }], totalCount: 1 });
    const { container } = render(<ChatWidget />);
    fireEvent.click(screen.getByRole('button', { name: 'Open chat' }));

    wsHandlers['group_chat.message']({
      id: 'm1',
      group_id: 'g1',
      sender_id: 'u2',
      content: 'hello team',
      created_at: '2024-01-01T00:00:01Z',
    });

    const widgetButton = container.querySelector('.chat-widget-button');
    await waitFor(() => {
      expect(widgetButton).toHaveAttribute('data-unread', '1');
    });

    fireEvent.click(screen.getByRole('button', { name: 'Groups' }));
    expect(await screen.findByText('1')).toBeTruthy();

    fireEvent.click(await screen.findByText('Go Meetup'));

    await waitFor(() => {
      expect(widgetButton).not.toHaveAttribute('data-unread');
    });
  });

  it('bumps unread for a message in a conversation missing from the list', async () => {
    // Regression: a first-ever message creates the conversation server-side,
    // but no local entry exists yet — the old code silently dropped the event.
    // The widget must refresh the list, whose snapshot carries the unread.
    const newConversation = {
      ...conversation,
      id: 'c-new',
      participants: [{ id: 'u3', username: 'carol', isOnline: true }],
      unreadCount: 1,
    };
    mockGetChats.mockResolvedValueOnce([]).mockResolvedValueOnce([newConversation]);
    const { container } = render(<ChatWidget />);
    await waitFor(() => {
      expect(mockGetChats).toHaveBeenCalledTimes(1);
    });

    wsHandlers['chat.message']({
      id: 'm1',
      chat_id: 'c-new',
      sender_id: 'u3',
      content: 'first hello',
      created_at: '2024-01-01T00:00:01Z',
    });

    await waitFor(() => {
      expect(mockGetChats).toHaveBeenCalledTimes(2);
    });
    const widgetButton = container.querySelector('.chat-widget-button');
    await waitFor(() => {
      expect(widgetButton).toHaveAttribute('data-unread', '1');
    });
  });

  it('does not count group messages sent by the current user', async () => {
    const { container } = render(<ChatWidget />);
    fireEvent.click(screen.getByRole('button', { name: 'Open chat' }));

    wsHandlers['group_chat.message']({
      id: 'm1',
      group_id: 'g1',
      sender_id: 'u1',
      content: 'my own message',
      created_at: '2024-01-01T00:00:01Z',
    });

    const widgetButton = container.querySelector('.chat-widget-button');
    await waitFor(() => {
      expect(widgetButton).not.toHaveAttribute('data-unread');
    });
  });

  it('seeds group unread badges from the server snapshot', async () => {
    mockGetMyGroups.mockResolvedValue({ data: [{ ...group, unreadCount: 4 }], totalCount: 1 });
    const { container } = render(<ChatWidget />);
    fireEvent.click(screen.getByRole('button', { name: 'Open chat' }));

    const widgetButton = container.querySelector('.chat-widget-button');
    await waitFor(() => {
      expect(widgetButton).toHaveAttribute('data-unread', '4');
    });
  });
});
