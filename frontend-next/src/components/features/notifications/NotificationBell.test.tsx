import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent, act } from '@testing-library/react';
import React from 'react';
import NotificationBell from './NotificationBell';

const mockGetNotifications = vi.fn();
const mockGetUnreadNotificationCount = vi.fn();
const mockGetComment = vi.fn();
const mockMarkNotificationRead = vi.fn();
const mockMarkAllNotificationsRead = vi.fn();
const mockHandleFollowRequest = vi.fn();
const mockRespondToGroupInvitation = vi.fn();
const mockHandleJoinRequest = vi.fn();

vi.mock('@/lib/api', () => ({
  getNotifications: (...args: unknown[]) => mockGetNotifications(...args),
  getUnreadNotificationCount: (...args: unknown[]) => mockGetUnreadNotificationCount(...args),
  getComment: (...args: unknown[]) => mockGetComment(...args),
  markNotificationRead: (...args: unknown[]) => mockMarkNotificationRead(...args),
  markAllNotificationsRead: (...args: unknown[]) => mockMarkAllNotificationsRead(...args),
  handleFollowRequest: (...args: unknown[]) => mockHandleFollowRequest(...args),
  respondToGroupInvitation: (...args: unknown[]) => mockRespondToGroupInvitation(...args),
  handleJoinRequest: (...args: unknown[]) => mockHandleJoinRequest(...args),
}));

type NotificationHandler = (n: Record<string, unknown>) => void;
type UnreadHandler = (count: number) => void;
type ConnectedHandler = () => void;

const streamHandlers: {
  notification: Set<NotificationHandler>;
  unread: Set<UnreadHandler>;
  connected: Set<ConnectedHandler>;
} = { notification: new Set(), unread: new Set(), connected: new Set() };

vi.mock('@/lib/notificationStream', () => ({
  notificationStream: {
    onNotification: vi.fn((handler: NotificationHandler) => {
      streamHandlers.notification.add(handler);
      return () => streamHandlers.notification.delete(handler);
    }),
    onUnreadCount: vi.fn((handler: UnreadHandler) => {
      streamHandlers.unread.add(handler);
      return () => streamHandlers.unread.delete(handler);
    }),
    onConnected: vi.fn((handler: ConnectedHandler) => {
      streamHandlers.connected.add(handler);
      return () => streamHandlers.connected.delete(handler);
    }),
    connect: vi.fn(),
    disconnect: vi.fn(),
  },
}));

const mockPush = vi.fn();
vi.mock('next/navigation', () => ({
  useRouter: () => ({ push: mockPush }),
}));

vi.mock('next/image', () => ({
  default: (props: React.ImgHTMLAttributes<HTMLImageElement>) => <img {...props} />,
}));

vi.mock('@/context/AuthContext', () => ({
  useAuth: () => ({
    user: {
      id: 'u2',
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

function makeNotification(overrides: Record<string, unknown> = {}) {
  return {
    id: 1,
    recipient_id: 'u2',
    type: 'like',
    resource_type: 'post',
    resource_id: '42',
    actor_id: 'u1',
    actor_name: 'Alice',
    actor_avatar: '',
    content_text: '',
    image_url: '',
    join_request_id: '',
    event_id: '',
    is_read: false,
    created_at: '2026-08-14T10:00:00Z',
    deleted: false,
    ...overrides,
  };
}

describe('NotificationBell', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    streamHandlers.notification.clear();
    streamHandlers.unread.clear();
    streamHandlers.connected.clear();
    mockGetUnreadNotificationCount.mockResolvedValue({ count: 0 });
    mockGetNotifications.mockResolvedValue({ notifications: [], total: 0 });
    mockGetComment.mockResolvedValue({ id: '5', postId: '99' });
    mockMarkNotificationRead.mockResolvedValue(undefined);
    mockMarkAllNotificationsRead.mockResolvedValue(undefined);
  });

  it('shows the unread badge from the initial count', async () => {
    mockGetUnreadNotificationCount.mockResolvedValue({ count: 3 });
    render(<NotificationBell />);
    await waitFor(() => expect(screen.getByText('3')).toBeInTheDocument());
  });

  it('loads notifications when the dropdown is opened', async () => {
    mockGetNotifications.mockResolvedValue({
      notifications: [makeNotification({ id: 10, type: 'like', is_read: true })],
      total: 1,
    });

    render(<NotificationBell />);
    fireEvent.click(screen.getByRole('button', { name: /notifications/i }));

    await waitFor(() => {
      expect(screen.getByText('Alice liked your post')).toBeInTheDocument();
    });
  });

  it('shows the empty state when there are no notifications', async () => {
    mockGetNotifications.mockResolvedValue({ notifications: [], total: 0 });
    render(<NotificationBell />);
    fireEvent.click(screen.getByRole('button', { name: /notifications/i }));

    await waitFor(() => {
      expect(screen.getByText('No notifications yet')).toBeInTheDocument();
    });
  });

  it('renders live notifications from the stream and marks them unread', async () => {
    render(<NotificationBell />);

    // Let the initial unread-count fetch settle (resolves to 0) before pushing
    // a live notification, so the optimistic +1 is not overwritten by the fetch.
    await waitFor(() => expect(mockGetUnreadNotificationCount).toHaveBeenCalled());

    act(() => {
      streamHandlers.notification.forEach((handler) =>
        handler(makeNotification({ id: 7, type: 'follow' }))
      );
    });

    await waitFor(() => {
      expect(screen.getByText('Alice followed you')).toBeInTheDocument();
    });
    await waitFor(() => expect(screen.getByText('1')).toBeInTheDocument());
  });

  it('removes the unread badge when a live unread notification is deleted', async () => {
    render(<NotificationBell />);
    await waitFor(() => expect(mockGetUnreadNotificationCount).toHaveBeenCalled());

    act(() => {
      streamHandlers.notification.forEach((handler) =>
        handler(makeNotification({ id: 7, type: 'like' }))
      );
    });
    await waitFor(() => expect(screen.getByText('1')).toBeInTheDocument());
    await waitFor(() => expect(screen.getByText('Alice liked your post')).toBeInTheDocument());

    act(() => {
      streamHandlers.notification.forEach((handler) =>
        handler(makeNotification({ id: 7, type: 'like', deleted: true }))
      );
    });

    await waitFor(() => {
      expect(screen.queryByText('Alice liked your post')).not.toBeInTheDocument();
    });
    await waitFor(() => {
      expect(screen.queryByText('1')).not.toBeInTheDocument();
    });
  });

  it('removes a notification from the DOM when the stream reports deleted', async () => {
    mockGetNotifications.mockResolvedValue({
      notifications: [makeNotification({ id: 7, type: 'follow', is_read: true })],
      total: 1,
    });

    render(<NotificationBell />);
    fireEvent.click(screen.getByRole('button', { name: /notifications/i }));
    await waitFor(() => expect(screen.getByText('Alice followed you')).toBeInTheDocument());

    act(() => {
      streamHandlers.notification.forEach((handler) =>
        handler(makeNotification({ id: 7, deleted: true }))
      );
    });

    await waitFor(() => {
      expect(screen.queryByText('Alice followed you')).not.toBeInTheDocument();
    });
  });

  it('marks a notification read and navigates on click', async () => {
    mockGetNotifications.mockResolvedValue({
      notifications: [makeNotification({ id: 10, type: 'like', is_read: false })],
      total: 1,
    });

    render(<NotificationBell />);
    fireEvent.click(screen.getByRole('button', { name: /notifications/i }));
    const item = await screen.findByText('Alice liked your post');

    fireEvent.click(item);

    await waitFor(() => expect(mockMarkNotificationRead).toHaveBeenCalledWith(10));
    await waitFor(() => expect(mockPush).toHaveBeenCalledWith('/post/42'));
  });

  it('resolves comment likes to their post via getComment', async () => {
    mockGetNotifications.mockResolvedValue({
      notifications: [
        makeNotification({
          id: 11,
          type: 'like',
          resource_type: 'comment',
          resource_id: '5',
          is_read: true,
        }),
      ],
      total: 1,
    });

    render(<NotificationBell />);
    fireEvent.click(screen.getByRole('button', { name: /notifications/i }));
    const item = await screen.findByText('Alice liked your comment');

    fireEvent.click(item);

    await waitFor(() => expect(mockGetComment).toHaveBeenCalledWith(5));
    await waitFor(() => expect(mockPush).toHaveBeenCalledWith('/post/99'));
  });

  it('handles accept/decline actions on follow requests', async () => {
    mockGetNotifications.mockResolvedValue({
      notifications: [
        makeNotification({
          id: 12,
          type: 'follow_request',
          resource_type: 'user',
          resource_id: 'u1',
          is_read: true,
        }),
      ],
      total: 1,
    });

    render(<NotificationBell />);
    fireEvent.click(screen.getByRole('button', { name: /notifications/i }));
    await screen.findByText('Alice wants to follow you');

    const acceptBtn = screen.getAllByRole('button', { name: /accept/i })[0];
    fireEvent.click(acceptBtn);

    await waitFor(() => expect(mockHandleFollowRequest).toHaveBeenCalledWith('u1', 'accept'));
    await waitFor(() => {
      expect(screen.queryByText('Alice wants to follow you')).not.toBeInTheDocument();
    });
  });
});
