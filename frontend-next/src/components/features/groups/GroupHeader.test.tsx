import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';

const mockPush = vi.fn();

vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: mockPush,
    replace: vi.fn(),
    prefetch: vi.fn(),
    back: vi.fn(),
    forward: vi.fn(),
    refresh: vi.fn(),
  }),
}));

vi.mock('@/lib/api', () => ({
  requestToJoinGroup: vi.fn().mockResolvedValue(undefined),
  leaveGroup: vi.fn().mockResolvedValue(undefined),
  deleteGroup: vi.fn().mockResolvedValue(undefined),
}));

vi.mock('./EditGroupForm', () => ({
  default: () => <div>Edit form shown</div>,
}));

vi.mock('./InviteDropdown', () => ({
  default: () => <div>Invite dropdown shown</div>,
}));

vi.mock('./CreateEventForm', () => ({
  default: () => <div>Event form shown</div>,
}));

import { deleteGroup } from '@/lib/api';
import GroupHeader from './GroupHeader';
import type { Group } from '@/lib/types';

const creatorGroup: Group = {
  id: 'g1',
  title: 'Go Meetup',
  description: 'Gophers hang out',
  creatorId: 'u1',
  creator: {} as never,
  membersCount: 2,
  membershipStatus: 'member',
  createdAt: '2026-01-01T00:00:00Z',
};

describe('GroupHeader', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('shows Edit Group and Delete Group buttons for the creator', () => {
    render(<GroupHeader group={creatorGroup} isCreator isMember={false} />);

    expect(screen.getByRole('button', { name: /Edit Group/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Delete Group/i })).toBeInTheDocument();
  });

  it('opens the edit form when Edit Group is clicked', () => {
    render(<GroupHeader group={creatorGroup} isCreator isMember={false} />);

    fireEvent.click(screen.getByRole('button', { name: /Edit Group/i }));

    expect(screen.getByText('Edit form shown')).toBeInTheDocument();
  });

  it('deletes the group after the two-step confirmation', async () => {
    render(<GroupHeader group={creatorGroup} isCreator isMember={false} />);

    fireEvent.click(screen.getByRole('button', { name: /Delete Group/i }));
    expect(screen.getByRole('button', { name: /Confirm delete/i })).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: /Confirm delete/i }));

    await waitFor(() => {
      expect(deleteGroup).toHaveBeenCalledWith('g1');
    });
    expect(mockPush).toHaveBeenCalledWith('/groups');
  });

  it('does not delete the group if the user cancels', () => {
    render(<GroupHeader group={creatorGroup} isCreator isMember={false} />);

    fireEvent.click(screen.getByRole('button', { name: /Delete Group/i }));
    fireEvent.click(screen.getByRole('button', { name: /Cancel/i }));

    expect(screen.getByRole('button', { name: /Delete Group/i })).toBeInTheDocument();
    expect(deleteGroup).not.toHaveBeenCalled();
  });
});
