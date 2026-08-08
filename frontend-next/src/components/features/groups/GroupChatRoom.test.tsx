import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import React from 'react';
import GroupChatRoom from './GroupChatRoom';

const mockGetGroupChatHistory = vi.fn();
const mockGetGroupMembers = vi.fn();

vi.mock('@/lib/api', () => ({
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
}));

vi.mock('@/lib/ws', () => ({
  chatSocket: mockChatSocket,
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
  }),
}));

vi.mock('next/image', () => ({
  default: (props: React.ImgHTMLAttributes<HTMLImageElement>) => <img {...props} />,
}));

const member = {
  groupId: 'g1',
  userId: 'u2',
  user: { id: 'u2', username: 'bob', firstName: 'Bob', lastName: 'Jones' },
  role: 'member' as const,
  joinedAt: '2024-01-01T00:00:00Z',
};

describe('GroupChatRoom', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetGroupChatHistory.mockResolvedValue([]);
    mockGetGroupMembers.mockResolvedValue({ data: [member], totalCount: 1 });
    for (const key of Object.keys(wsHandlers)) delete wsHandlers[key];
  });

  it('loads and renders the chat history', async () => {
    mockGetGroupChatHistory.mockResolvedValue([
      {
        id: 'm1',
        group_id: 'g1',
        sender_id: 'u2',
        content: 'Hello team',
        created_at: '2024-01-01T00:00:00Z',
      },
    ]);

    render(<GroupChatRoom groupId="g1" />);

    await waitFor(() => {
      expect(screen.getByText('Hello team')).toBeTruthy();
    });
    expect(screen.getByText('Bob Jones')).toBeTruthy();
  });

  it('appends live group_chat.message broadcasts for the active group', async () => {
    render(<GroupChatRoom groupId="g1" />);
    await waitFor(() => {
      expect(screen.getByText('No messages yet. Start the conversation!')).toBeTruthy();
    });

    wsHandlers['group_chat.message']({
      id: 'm2',
      group_id: 'g1',
      sender_id: 'u2',
      content: 'Live message',
      created_at: '2024-01-01T00:00:01Z',
    });

    expect(await screen.findByText('Live message')).toBeTruthy();
  });

  it('ignores messages for other groups', async () => {
    render(<GroupChatRoom groupId="g1" />);
    await waitFor(() => {
      expect(screen.getByText('No messages yet. Start the conversation!')).toBeTruthy();
    });

    wsHandlers['group_chat.message']({
      id: 'm3',
      group_id: 'g2',
      sender_id: 'u2',
      content: 'Wrong group',
      created_at: '2024-01-01T00:00:01Z',
    });

    expect(screen.queryByText('Wrong group')).toBeNull();
  });

  it('sends group_chat.send with the group id', async () => {
    render(<GroupChatRoom groupId="g1" />);
    await waitFor(() => {
      expect(screen.getByText('No messages yet. Start the conversation!')).toBeTruthy();
    });

    const input = screen.getByPlaceholderText('Type a message...');
    fireEvent.change(input, { target: { value: 'Hey!' } });
    fireEvent.click(screen.getByRole('button', { name: '➤' }));

    expect(mockChatSocket.send).toHaveBeenCalledWith('group_chat.send', {
      group_id: 'g1',
      content: 'Hey!',
    });
  });

  it('shows the live online-members indicator when presence is provided', async () => {
    render(<GroupChatRoom groupId="g1" presence={{ total: 3, online: 2, loading: false }} />);

    await waitFor(() => {
      expect(screen.getByText('2 online')).toBeTruthy();
    });
  });

  it('sends group_chat.mark_read when the room opens', async () => {
    render(<GroupChatRoom groupId="g1" />);

    await waitFor(() => {
      expect(mockChatSocket.send).toHaveBeenCalledWith('group_chat.mark_read', { group_id: 'g1' });
    });
  });

  it('re-marks the group read when a new message arrives', async () => {
    render(<GroupChatRoom groupId="g1" />);
    await waitFor(() => {
      expect(screen.getByText('No messages yet. Start the conversation!')).toBeTruthy();
    });

    mockChatSocket.send.mockClear();

    wsHandlers['group_chat.message']({
      id: 'm2',
      group_id: 'g1',
      sender_id: 'u2',
      content: 'Live message',
      created_at: '2024-01-01T00:00:01Z',
    });

    await waitFor(() => {
      expect(mockChatSocket.send).toHaveBeenCalledWith('group_chat.mark_read', { group_id: 'g1' });
    });
  });

  it('opens a popover listing the online members', async () => {
    render(
      <GroupChatRoom
        groupId="g1"
        presence={{
          total: 3,
          online: 2,
          loading: false,
          members: [
            { id: 'u2', isOnline: true },
            { id: 'u3', isOnline: true },
          ],
        }}
      />
    );
    await waitFor(() => {
      expect(screen.getByText('2 online')).toBeTruthy();
    });

    fireEvent.click(screen.getByRole('button', { name: '2 members online' }));

    expect(await screen.findByText('Online now')).toBeTruthy();
    // u2's name resolves from the members list fetched over HTTP.
    expect(await screen.findByText('Bob Jones')).toBeTruthy();
    // u3 has no member record and falls back to a generic label.
    expect(screen.getByText('Member')).toBeTruthy();
  });

  it('closes the online-members popover on outside click', async () => {
    render(
      <GroupChatRoom
        groupId="g1"
        presence={{
          total: 1,
          online: 1,
          loading: false,
          members: [{ id: 'u2', isOnline: true }],
        }}
      />
    );
    await waitFor(() => {
      expect(screen.getByText('1 online')).toBeTruthy();
    });

    fireEvent.click(screen.getByRole('button', { name: '1 members online' }));
    expect(await screen.findByText('Online now')).toBeTruthy();

    fireEvent.mouseDown(document.body);
    await waitFor(() => {
      expect(screen.queryByText('Online now')).toBeNull();
    });
  });

  it('calls onBack and onClose handlers', async () => {
    const onBack = vi.fn();
    const onClose = vi.fn();
    render(<GroupChatRoom groupId="g1" groupTitle="Go" onBack={onBack} onClose={onClose} />);

    await waitFor(() => {
      expect(screen.getByText('Go')).toBeTruthy();
    });
    fireEvent.click(screen.getByRole('button', { name: 'Back to groups' }));
    fireEvent.click(screen.getByRole('button', { name: 'Close chat' }));
    expect(onBack).toHaveBeenCalledTimes(1);
    expect(onClose).toHaveBeenCalledTimes(1);
  });
});
