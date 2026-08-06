import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import React from 'react';
import GroupContent from './GroupContent';

const mockGetGroupPosts = vi.fn();
const mockGetGroupEvents = vi.fn();
const mockGetGroupPostComments = vi.fn();
const mockCreateGroupPostComment = vi.fn();

vi.mock('@/lib/api', () => ({
  getGroupPosts: (groupId: string, page?: number) => mockGetGroupPosts(groupId, page),
  getGroupEvents: (groupId: string, cursor?: string) => mockGetGroupEvents(groupId, cursor),
  getGroupPostComments: (postId: string) => mockGetGroupPostComments(postId),
  createGroupPostComment: (postId: string, content: string, image?: File | null) =>
    mockCreateGroupPostComment(postId, content, image),
  likePost: () => Promise.resolve(undefined),
  unlikePost: () => Promise.resolve(undefined),
  voteGroupPost: () => Promise.resolve(undefined),
}));

vi.mock('next/image', () => ({
  default: (props: React.ImgHTMLAttributes<HTMLImageElement>) => <img {...props} />,
}));

vi.mock('next/link', () => ({
  default: ({ children, ...props }: { children: React.ReactNode; href: string }) => (
    <a {...props}>{children}</a>
  ),
}));

const post = {
  id: 'gp1',
  userId: 'u1',
  user: { id: 'u1', username: 'alice', firstName: 'Alice', lastName: 'Doe' },
  content: 'A group post',
  privacy: 'public',
  groupId: 'g1',
  commentsCount: 1,
  likesCount: 3,
  dislikesCount: 1,
  createdAt: new Date().toISOString(),
};

describe('GroupContent PostsTab comments', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetGroupPosts.mockResolvedValue({ data: [post], totalPages: 1, totalCount: 1 });
    mockGetGroupEvents.mockResolvedValue({ events: [], nextCursor: undefined });
    mockGetGroupPostComments.mockResolvedValue({
      data: [
        {
          id: 'c1',
          postId: 'gp1',
          userId: 'u2',
          user: { id: 'u2', username: 'bob', firstName: 'Bob', lastName: 'Smith' },
          content: 'Nice post',
          createdAt: new Date().toISOString(),
        },
      ],
      totalPages: 1,
      totalCount: 1,
    });
    mockCreateGroupPostComment.mockResolvedValue({
      id: 'c2',
      postId: 'gp1',
      userId: 'u1',
      user: { id: 'u1', username: 'alice', firstName: 'Alice', lastName: 'Doe' },
      content: 'Thanks!',
      createdAt: new Date().toISOString(),
    });
  });

  it('loads and shows group post comments', async () => {
    render(<GroupContent groupId="g1" activeTab="posts" isMember />);

    fireEvent.click(await screen.findByRole('button', { name: /show comments/i }));

    await waitFor(() => {
      expect(screen.getByText('Nice post')).toBeTruthy();
    });
  });

  it('submits a group comment with optional image via createGroupPostComment', async () => {
    render(<GroupContent groupId="g1" activeTab="posts" isMember />);

    fireEvent.click(await screen.findByRole('button', { name: /show comments/i }));

    const textarea = await screen.findByPlaceholderText('Write a comment...');
    fireEvent.change(textarea, { target: { value: 'Thanks!' } });
    fireEvent.click(screen.getByRole('button', { name: 'Post Comment' }));

    await waitFor(() => {
      expect(mockCreateGroupPostComment).toHaveBeenCalledWith('gp1', 'Thanks!', null);
    });
    await waitFor(() => {
      expect(screen.getByText('Thanks!')).toBeTruthy();
    });
  });
});
