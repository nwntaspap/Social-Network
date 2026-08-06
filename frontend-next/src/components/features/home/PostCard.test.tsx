import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import React from 'react';
import PostCard from './PostCard';

const mockLikePost = vi.fn().mockResolvedValue(undefined);
const mockUnlikePost = vi.fn().mockResolvedValue(undefined);
const mockVoteGroupPost = vi.fn().mockResolvedValue(undefined);

vi.mock('@/lib/api', () => ({
  likePost: (id: number) => mockLikePost(id),
  unlikePost: (id: number) => mockUnlikePost(id),
  voteGroupPost: (id: string, reaction: 1 | -1) => mockVoteGroupPost(id, reaction),
}));

vi.mock('next/image', () => ({
  default: (props: React.ImgHTMLAttributes<HTMLImageElement>) => <img {...props} />,
}));

vi.mock('next/link', () => ({
  default: ({ children, ...props }: { children: React.ReactNode; href: string }) => (
    <a {...props}>{children}</a>
  ),
}));

function makePost(overrides: Record<string, unknown> = {}) {
  return {
    id: 'gp1',
    userId: 'u1',
    user: { id: 'u1', username: 'alice', firstName: 'Alice', lastName: 'Doe' },
    content: 'A group post',
    privacy: 'public',
    groupId: 'g1',
    commentsCount: 2,
    likesCount: 3,
    dislikesCount: 1,
    createdAt: new Date().toISOString(),
    ...overrides,
  } as never;
}

describe('PostCard', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders like and dislike buttons for a group post and votes via voteGroupPost', async () => {
    render(<PostCard post={makePost()} />);

    const likeBtn = screen.getByRole('button', { name: 'Like' });
    const dislikeBtn = screen.getByRole('button', { name: 'Dislike' });

    fireEvent.click(likeBtn);
    await waitFor(() => {
      expect(mockVoteGroupPost).toHaveBeenCalledWith('gp1', 1);
    });

    fireEvent.click(dislikeBtn);
    await waitFor(() => {
      expect(mockVoteGroupPost).toHaveBeenCalledWith('gp1', -1);
    });
    expect(mockLikePost).not.toHaveBeenCalled();
  });

  it('renders a single like button for a topic post and uses likePost/unlikePost', async () => {
    render(<PostCard post={makePost({ groupId: undefined })} />);

    fireEvent.click(screen.getByRole('button', { name: /like/i }));
    await waitFor(() => {
      expect(mockLikePost).toHaveBeenCalledWith(Number('gp1'));
    });
    expect(mockVoteGroupPost).not.toHaveBeenCalled();
  });
});
