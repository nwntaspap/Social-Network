import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import React from 'react';
import FollowRequestsSection from './FollowRequestsSection';

const mockGetPendingFollowRequests = vi.fn();
const mockHandleFollowRequest = vi.fn();

vi.mock('@/lib/api', () => ({
  getPendingFollowRequests: () => mockGetPendingFollowRequests(),
  handleFollowRequest: (followerId: string, action: string) =>
    mockHandleFollowRequest(followerId, action),
}));

vi.mock('next/image', () => ({
  default: (props: React.ImgHTMLAttributes<HTMLImageElement>) => <img {...props} />,
}));

vi.mock('next/link', () => ({
  default: ({ children, ...props }: { children: React.ReactNode; href: string }) => (
    <a {...props}>{children}</a>
  ),
}));

const request = {
  id: 'r1',
  requesterId: 'u2',
  requester: { id: 'u2', username: 'bob', firstName: 'Bob', lastName: 'Smith' },
  targetId: 'u1',
  target: { id: 'u1', username: 'alice', firstName: 'Alice', lastName: 'Doe' },
  status: 'pending',
  createdAt: new Date().toISOString(),
};

describe('FollowRequestsSection', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders nothing when there are no pending requests', async () => {
    mockGetPendingFollowRequests.mockResolvedValue([]);

    render(<FollowRequestsSection />);

    await waitFor(() => {
      expect(screen.queryByText('Follow Requests')).not.toBeInTheDocument();
    });
  });

  it('renders pending requests with accept and decline buttons', async () => {
    mockGetPendingFollowRequests.mockResolvedValue([request]);

    render(<FollowRequestsSection />);

    expect(await screen.findByText('Follow Requests')).toBeInTheDocument();
    expect(screen.getByText('Bob Smith')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Accept' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Decline' })).toBeInTheDocument();
  });

  it('accepts a request and removes it from the list', async () => {
    mockGetPendingFollowRequests.mockResolvedValue([request]);
    mockHandleFollowRequest.mockResolvedValue(undefined);

    render(<FollowRequestsSection />);

    const accept = await screen.findByRole('button', { name: 'Accept' });
    fireEvent.click(accept);

    await waitFor(() => {
      expect(mockHandleFollowRequest).toHaveBeenCalledWith('u2', 'accept');
    });
    await waitFor(() => {
      expect(screen.queryByText('Bob Smith')).not.toBeInTheDocument();
    });
  });

  it('declines a request and removes it from the list', async () => {
    mockGetPendingFollowRequests.mockResolvedValue([request]);
    mockHandleFollowRequest.mockResolvedValue(undefined);

    render(<FollowRequestsSection />);

    const decline = await screen.findByRole('button', { name: 'Decline' });
    fireEvent.click(decline);

    await waitFor(() => {
      expect(mockHandleFollowRequest).toHaveBeenCalledWith('u2', 'decline');
    });
    await waitFor(() => {
      expect(screen.queryByText('Bob Smith')).not.toBeInTheDocument();
    });
  });
});
