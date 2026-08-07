import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import React from 'react';
import FollowersFollowingModal from './FollowersFollowingModal';

const mockGetFollowers = vi.fn();
const mockGetFollowing = vi.fn();

vi.mock('@/lib/api', () => ({
  getFollowers: (userId: string) => mockGetFollowers(userId),
  getFollowing: (userId: string) => mockGetFollowing(userId),
}));

vi.mock('next/image', () => ({
  default: (props: React.ImgHTMLAttributes<HTMLImageElement>) => <img {...props} />,
}));

vi.mock('next/link', () => ({
  default: ({ children, ...props }: { children: React.ReactNode; href: string }) => (
    <a {...props}>{children}</a>
  ),
}));

const users = [
  {
    id: 'u2',
    username: 'bob',
    firstName: 'Bob',
    lastName: 'Smith',
    avatarUrl: 'avatars/bob.jpg',
  },
];

describe('FollowersFollowingModal', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders followers list with links', async () => {
    mockGetFollowers.mockResolvedValue(users);

    render(<FollowersFollowingModal userId="u1" mode="followers" onClose={() => {}} />);

    expect(await screen.findByText('Followers')).toBeInTheDocument();
    expect(screen.getByText('Bob Smith')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /Bob Smith/ })).toHaveAttribute('href', '/profile/u2');
  });

  it('renders following list when mode is following', async () => {
    mockGetFollowing.mockResolvedValue(users);

    render(<FollowersFollowingModal userId="u1" mode="following" onClose={() => {}} />);

    expect(await screen.findByText('Following')).toBeInTheDocument();
    expect(screen.getByText('Bob Smith')).toBeInTheDocument();
  });

  it('shows empty message when there are no users', async () => {
    mockGetFollowers.mockResolvedValue([]);

    render(<FollowersFollowingModal userId="u1" mode="followers" onClose={() => {}} />);

    expect(await screen.findByText('No followers yet.')).toBeInTheDocument();
  });

  it('calls onClose when the close button is clicked', async () => {
    mockGetFollowers.mockResolvedValue(users);
    const onClose = vi.fn();

    render(<FollowersFollowingModal userId="u1" mode="followers" onClose={onClose} />);

    await screen.findByText('Followers');
    fireEvent.click(screen.getByRole('button', { name: 'Close' }));

    expect(onClose).toHaveBeenCalled();
  });
});
