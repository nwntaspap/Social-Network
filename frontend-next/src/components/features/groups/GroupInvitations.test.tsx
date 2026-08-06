import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import React from 'react';
import GroupInvitations from './GroupInvitations';

const mockGetMyGroupInvitations = vi.fn();
const mockRespondToGroupInvitation = vi.fn();

vi.mock('@/lib/api', () => ({
  getMyGroupInvitations: () => mockGetMyGroupInvitations(),
  respondToGroupInvitation: (groupId: string, action: string) =>
    mockRespondToGroupInvitation(groupId, action),
}));

vi.mock('next/link', () => ({
  default: ({ children, ...props }: { children: React.ReactNode; href: string }) => (
    <a {...props}>{children}</a>
  ),
}));

const invitation = {
  id: 'i1',
  groupId: 'g1',
  group: { id: 'g1', title: 'Go Meetup' },
  inviterId: 'u2',
  inviter: { id: 'u2', username: 'bob' },
  inviteeId: 'u1',
  invitee: { id: 'u1', username: 'alice' },
  createdAt: new Date().toISOString(),
};

describe('GroupInvitations', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders nothing when there are no pending invitations', async () => {
    mockGetMyGroupInvitations.mockResolvedValue([]);

    render(<GroupInvitations />);

    await waitFor(() => {
      expect(screen.queryByText('Group Invitations')).not.toBeInTheDocument();
    });
  });

  it('renders pending invitations with accept and decline buttons', async () => {
    mockGetMyGroupInvitations.mockResolvedValue([invitation]);

    render(<GroupInvitations />);

    expect(await screen.findByText('Go Meetup')).toBeInTheDocument();
    expect(screen.getByText(/Invited by bob/)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Accept' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Decline' })).toBeInTheDocument();
  });

  it('accepts an invitation and removes it from the list', async () => {
    mockGetMyGroupInvitations.mockResolvedValue([invitation]);
    mockRespondToGroupInvitation.mockResolvedValue(undefined);

    render(<GroupInvitations />);

    const accept = await screen.findByRole('button', { name: 'Accept' });
    fireEvent.click(accept);

    await waitFor(() => {
      expect(mockRespondToGroupInvitation).toHaveBeenCalledWith('g1', 'accept');
    });
    await waitFor(() => {
      expect(screen.queryByText('Go Meetup')).not.toBeInTheDocument();
    });
  });

  it('declines an invitation and removes it from the list', async () => {
    mockGetMyGroupInvitations.mockResolvedValue([invitation]);
    mockRespondToGroupInvitation.mockResolvedValue(undefined);

    render(<GroupInvitations />);

    const decline = await screen.findByRole('button', { name: 'Decline' });
    fireEvent.click(decline);

    await waitFor(() => {
      expect(mockRespondToGroupInvitation).toHaveBeenCalledWith('g1', 'decline');
    });
    await waitFor(() => {
      expect(screen.queryByText('Go Meetup')).not.toBeInTheDocument();
    });
  });
});
