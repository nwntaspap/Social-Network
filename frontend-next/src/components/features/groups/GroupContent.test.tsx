import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import React from 'react';
import GroupContent from './GroupContent';

const mockGetGroupPosts = vi.fn();
const mockGetGroupEvents = vi.fn();
const mockGetGroupPostComments = vi.fn();
const mockCreateGroupPostComment = vi.fn();
const mockRespondToEvent = vi.fn();
const mockGetEventRSVPs = vi.fn();

vi.mock('@/lib/api', () => ({
  getGroupPosts: (groupId: string, page?: number) => mockGetGroupPosts(groupId, page),
  getGroupEvents: (groupId: string, cursor?: string) => mockGetGroupEvents(groupId, cursor),
  getGroupPostComments: (postId: string) => mockGetGroupPostComments(postId),
  createGroupPostComment: (postId: string, content: string, image?: File | null) =>
    mockCreateGroupPostComment(postId, content, image),
  respondToEvent: (eventId: string, optionId: string) => mockRespondToEvent(eventId, optionId),
  getEventRSVPs: (eventId: string) => mockGetEventRSVPs(eventId),
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
    render(
      <GroupContent groupId="g1" activeTab="posts" isMember isCreator={false} eventRefreshKey={0} />
    );

    fireEvent.click(await screen.findByRole('button', { name: /show comments/i }));

    await waitFor(() => {
      expect(screen.getByText('Nice post')).toBeTruthy();
    });
  });

  it('submits a group comment with optional image via createGroupPostComment', async () => {
    render(
      <GroupContent groupId="g1" activeTab="posts" isMember isCreator={false} eventRefreshKey={0} />
    );

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

describe('GroupContent EventsTab', () => {
  const event = {
    id: 'evt1',
    groupId: 'g1',
    creatorId: 'u1',
    creator: {
      id: 'u1',
      email: 'alice@example.com',
      username: 'alice',
      firstName: 'Alice',
      lastName: 'Doe',
      dateOfBirth: '1995-01-01',
      isPublic: true,
      createdAt: new Date().toISOString(),
    },
    title: 'React meetup',
    description: 'A meetup about React',
    eventDate: new Date(Date.now() + 86400000 * 7).toISOString(),
    createdAt: new Date().toISOString(),
    options: [
      { id: 'opt1', label: 'Going', tally: 2 },
      { id: 'opt2', label: 'Not going', tally: 1 },
    ],
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockGetGroupPosts.mockResolvedValue({ data: [], totalPages: 1, totalCount: 0 });
    mockGetGroupEvents.mockResolvedValue({ events: [event], nextCursor: undefined });
    mockGetEventRSVPs.mockResolvedValue({
      options: [
        {
          optionId: 'opt1',
          optionLabel: 'Going',
          users: [
            {
              id: 'u1',
              username: 'alice',
              firstName: 'Alice',
              lastName: 'Doe',
              email: 'alice@example.com',
              dateOfBirth: '1995-01-01',
              isPublic: true,
              createdAt: new Date().toISOString(),
            },
          ],
        },
        { optionId: 'opt2', optionLabel: 'Not going', users: [] },
      ],
    });
    mockRespondToEvent.mockResolvedValue(undefined);
  });

  it('renders event options with tallies', async () => {
    render(
      <GroupContent
        groupId="g1"
        activeTab="events"
        isMember
        isCreator={false}
        eventRefreshKey={0}
      />
    );

    expect(await screen.findByText('React meetup')).toBeTruthy();
    expect(screen.getByText('Going')).toBeTruthy();
    expect(screen.getByText('Not going')).toBeTruthy();
    expect(screen.getByText('2')).toBeTruthy();
  });

  it('sends an RSVP when an option is clicked', async () => {
    render(
      <GroupContent
        groupId="g1"
        activeTab="events"
        isMember
        isCreator={false}
        eventRefreshKey={0}
      />
    );

    fireEvent.click(await screen.findByRole('button', { name: /^going/i }));

    await waitFor(() => {
      expect(mockRespondToEvent).toHaveBeenCalledWith('evt1', 'opt1');
    });
  });

  it('loads and shows attendee lists per option', async () => {
    render(
      <GroupContent
        groupId="g1"
        activeTab="events"
        isMember
        isCreator={false}
        eventRefreshKey={0}
      />
    );

    fireEvent.click(await screen.findByRole('button', { name: /view attendees/i }));

    await waitFor(() => {
      expect(screen.getByText('Alice Doe')).toBeTruthy();
    });
    expect(screen.getByText('No one yet')).toBeTruthy();
  });

  it('shows the edit button only for the group creator', async () => {
    render(<GroupContent groupId="g1" activeTab="events" isMember isCreator eventRefreshKey={0} />);

    await screen.findByText('React meetup');

    expect(screen.getByRole('button', { name: /edit/i })).toBeTruthy();
  });

  it('does not show the edit button to non-creator members', async () => {
    render(
      <GroupContent
        groupId="g1"
        activeTab="events"
        isMember
        isCreator={false}
        eventRefreshKey={0}
      />
    );

    await screen.findByText('React meetup');

    expect(screen.queryByRole('button', { name: /edit/i })).toBeNull();
  });
});
