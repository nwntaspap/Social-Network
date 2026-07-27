import type { Comment } from '@/lib/types';

export const mockComments: Record<string, Comment[]> = {
  // Comments for post '1' (feed)
  '1': [
    {
      id: 'c1',
      postId: '1',
      userId: '3',
      user: {
        id: '3',
        email: 'bob@example.com',
        username: 'bobsmith',
        firstName: 'Bob',
        lastName: 'Smith',
        dateOfBirth: '1994-11-08',
        avatarUrl: '/images/user-avatar.png',
        isPublic: false,
        createdAt: '2024-01-20T14:45:00Z',
      },
      content: 'This looks amazing! What stack did you use for the backend?',
      createdAt: new Date(Date.now() - 1800000).toISOString(),
    },
    {
      id: 'c2',
      postId: '1',
      userId: '5',
      user: {
        id: '5',
        email: 'mike@example.com',
        username: 'mikew',
        firstName: 'Mike',
        lastName: 'Wilson',
        dateOfBirth: '1993-12-14',
        avatarUrl: '/images/user-avatar.png',
        isPublic: true,
        createdAt: '2024-02-15T16:30:00Z',
      },
      content: 'The performance gains are incredible. Did you use React Server Components?',
      createdAt: new Date(Date.now() - 900000).toISOString(),
    },
    {
      id: 'c3',
      postId: '1',
      userId: '2',
      user: {
        id: '2',
        email: 'jane@example.com',
        username: 'janedoe',
        firstName: 'Jane',
        lastName: 'Doe',
        dateOfBirth: '1996-07-22',
        avatarUrl: '/images/user-avatar.png',
        isPublic: true,
        createdAt: '2024-02-01T08:15:00Z',
      },
      content: 'Thanks! Backend is Node.js with tRPC. Works like a charm!',
      createdAt: new Date(Date.now() - 600000).toISOString(),
    },
  ],

  // Comments for post '3' (feed)
  '3': [
    {
      id: 'c4',
      postId: '3',
      userId: '2',
      user: {
        id: '2',
        email: 'jane@example.com',
        username: 'janedoe',
        firstName: 'Jane',
        lastName: 'Doe',
        dateOfBirth: '1996-07-22',
        avatarUrl: '/images/user-avatar.png',
        isPublic: true,
        createdAt: '2024-02-01T08:15:00Z',
      },
      content:
        "100% agree! TypeScript has saved me from so many bugs. The type system is just *chef's kiss*",
      createdAt: new Date(Date.now() - 7200000).toISOString(),
    },
    {
      id: 'c5',
      postId: '3',
      userId: '4',
      user: {
        id: '4',
        email: 'alice@example.com',
        username: 'alicej',
        firstName: 'Alice',
        lastName: 'Johnson',
        dateOfBirth: '1997-04-30',
        avatarUrl: '/images/user-avatar.png',
        isPublic: true,
        createdAt: '2024-03-10T09:00:00Z',
      },
      content: "I was resistant at first but now I can't imagine going back to plain JS.",
      createdAt: new Date(Date.now() - 5400000).toISOString(),
    },
  ],

  // Comments for group post 'gp1'
  gp1: [
    {
      id: 'c6',
      postId: 'gp1',
      userId: '1',
      user: {
        id: '1',
        email: 'john@example.com',
        username: 'johndoe',
        firstName: 'John',
        lastName: 'Doe',
        dateOfBirth: '1995-03-15',
        avatarUrl: '/images/user-avatar.png',
        isPublic: true,
        createdAt: '2024-01-15T10:30:00Z',
      },
      content: 'Just checked it out! The accessibility features are top-notch. Great work!',
      createdAt: new Date(Date.now() - 86400000).toISOString(),
    },
    {
      id: 'c7',
      postId: 'gp1',
      userId: '5',
      user: {
        id: '5',
        email: 'mike@example.com',
        username: 'mikew',
        firstName: 'Mike',
        lastName: 'Wilson',
        dateOfBirth: '1993-12-14',
        avatarUrl: '/images/user-avatar.png',
        isPublic: true,
        createdAt: '2024-02-15T16:30:00Z',
      },
      content: 'Would love to contribute! Are you accepting PRs?',
      createdAt: new Date(Date.now() - 43200000).toISOString(),
    },
  ],

  // Comments for group post 'gp2'
  gp2: [
    {
      id: 'c8',
      postId: 'gp2',
      userId: '1',
      user: {
        id: '1',
        email: 'john@example.com',
        username: 'johndoe',
        firstName: 'John',
        lastName: 'Doe',
        dateOfBirth: '1995-03-15',
        avatarUrl: '/images/user-avatar.png',
        isPublic: true,
        createdAt: '2024-01-15T10:30:00Z',
      },
      content: 'Zustand is great! Been using it for 6 months now. Way less boilerplate than Redux.',
      createdAt: new Date(Date.now() - 172800000).toISOString(),
    },
  ],
};

// Default comments for posts without specific mock data
export const defaultComments: Comment[] = [
  {
    id: 'c-default-1',
    postId: 'default',
    userId: '6',
    user: {
      id: '6',
      email: 'sarah@example.com',
      username: 'sarahc',
      firstName: 'Sarah',
      lastName: 'Connor',
      dateOfBirth: '1998-09-03',
      avatarUrl: '/images/user-avatar.png',
      isPublic: false,
      createdAt: '2024-03-01T11:20:00Z',
    },
    content: 'Great post! Thanks for sharing.',
    createdAt: new Date(Date.now() - 86400000).toISOString(),
  },
];
