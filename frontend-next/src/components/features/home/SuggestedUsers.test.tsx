import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import React from 'react';
import SuggestedUsers from './SuggestedUsers';

const mockSearchUsers = vi.fn();
const mockSendFollowRequest = vi.fn();

vi.mock('@/lib/api', () => ({
  searchUsers: (...args: unknown[]) => mockSearchUsers(...args),
  sendFollowRequest: (userId: string) => mockSendFollowRequest(userId),
}));

const wsHandlers: Record<string, (payload: unknown) => void> = {};
const wsConnectionListeners: Array<() => void> = [];

const mockChatSocket = vi.hoisted(() => ({
  on: vi.fn((type: string, handler: (payload: unknown) => void) => {
    wsHandlers[type] = handler;
    return () => {
      delete wsHandlers[type];
    };
  }),
  onConnection: vi.fn((listener: () => void) => {
    wsConnectionListeners.push(listener);
    return () => {
      const i = wsConnectionListeners.indexOf(listener);
      if (i >= 0) wsConnectionListeners.splice(i, 1);
    };
  }),
  isConnected: vi.fn(() => true),
  connect: vi.fn(),
  disconnect: vi.fn(),
}));

vi.mock('@/lib/ws', () => ({
  chatSocket: mockChatSocket,
}));

vi.mock('@/context/AuthContext', () => ({
  useAuth: () => ({
    user: {
      id: 'me',
      email: 'me@example.com',
      username: 'me',
      firstName: 'Me',
      lastName: 'Self',
      dateOfBirth: '1990-01-01',
      isPublic: true,
      createdAt: '2024-01-01T00:00:00Z',
    },
    loading: false,
  }),
}));

vi.mock('next/image', () => ({
  default: (props: React.ImgHTMLAttributes<HTMLImageElement>) => <img {...props} />,
}));

vi.mock('next/link', () => ({
  default: ({
    href,
    children,
    ...rest
  }: React.AnchorHTMLAttributes<HTMLAnchorElement> & { href: string }) => (
    <a href={href} {...rest}>
      {children}
    </a>
  ),
}));

const alice = {
  id: 'u2',
  email: 'alice@example.com',
  username: 'alice',
  firstName: 'Alice',
  lastName: 'Smith',
  dateOfBirth: '1990-01-01',
  isPublic: true,
  createdAt: '2024-01-01T00:00:00Z',
  isOnline: true,
};
const bob = {
  id: 'u3',
  email: 'bob@example.com',
  username: 'bob',
  firstName: 'Bob',
  lastName: 'Jones',
  dateOfBirth: '1991-01-01',
  isPublic: true,
  createdAt: '2024-01-01T00:00:00Z',
  isOnline: false,
};
const viewer = {
  id: 'me',
  email: 'me@example.com',
  username: 'me',
  firstName: 'Me',
  lastName: 'Self',
  dateOfBirth: '1990-01-01',
  isPublic: true,
  createdAt: '2024-01-01T00:00:00Z',
};

beforeEach(() => {
  vi.clearAllMocks();
  delete wsHandlers['isOnlineStatus.update'];
  wsConnectionListeners.length = 0;
  mockChatSocket.isConnected.mockReturnValue(true);
});

describe('SuggestedUsers', () => {
  it('renders users fetched from the API (not hardcoded)', async () => {
    mockSearchUsers.mockResolvedValue({ data: [alice, bob, viewer], totalPages: 1 });
    render(<SuggestedUsers />);

    await waitFor(() => {
      expect(mockSearchUsers).toHaveBeenCalled();
    });
    expect(await screen.findByText('Alice Smith')).toBeInTheDocument();
    expect(screen.getByText('Bob Jones')).toBeInTheDocument();
  });

  it('never shows the current user in the suggestions', async () => {
    mockSearchUsers.mockResolvedValue({ data: [viewer, alice], totalPages: 1 });
    render(<SuggestedUsers />);

    await screen.findByText('Alice Smith');
    expect(screen.queryByText('Me Self')).not.toBeInTheDocument();
  });

  it('shows the online status from the snapshot', async () => {
    mockSearchUsers.mockResolvedValue({ data: [alice, bob], totalPages: 1 });
    render(<SuggestedUsers />);

    expect(await screen.findByText('Online')).toBeInTheDocument();
    expect(screen.getByText('Offline')).toBeInTheDocument();
  });

  it('updates a user status live on isOnlineStatus.update broadcasts', async () => {
    mockSearchUsers.mockResolvedValue({ data: [alice, bob], totalPages: 1 });
    render(<SuggestedUsers />);

    await screen.findByText('Online');
    expect(screen.getAllByText('Offline').length).toBe(1);

    wsHandlers['isOnlineStatus.update']!({ user_id: 'u3', isOnline: true });

    await waitFor(() => {
      expect(screen.getAllByText('Online').length).toBe(2);
    });
  });

  it('refetches the snapshot on reconnect', async () => {
    mockSearchUsers.mockResolvedValue({ data: [alice], totalPages: 1 });
    render(<SuggestedUsers />);

    await screen.findByText('Alice Smith');
    expect(mockSearchUsers).toHaveBeenCalledTimes(1);

    mockChatSocket.isConnected.mockReturnValue(true);
    wsConnectionListeners.forEach((listener) => listener());

    await waitFor(() => {
      expect(mockSearchUsers).toHaveBeenCalledTimes(2);
    });
  });
});
