import type { Post } from '@/lib/types';

export const mockGroupPosts: Record<string, Post[]> = {
  // Posts for group 1 (React Developers)
  '1': [
    {
      id: 'gp1',
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
        'Just released a new React component library! Check it out: react-ui-kit.dev. Would love some feedback from the community.',
      privacy: 'public',
      groupId: '1',
      commentsCount: 23,
      likesCount: 156,
      isLiked: true,
      createdAt: new Date(Date.now() - 86400000 * 2).toISOString(),
    },
    {
      id: 'gp2',
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
      content:
        "What are you all using for state management in 2024? I've been using Zustand lately and loving it. Anyone else have experience with it vs Redux Toolkit?",
      privacy: 'public',
      groupId: '1',
      commentsCount: 45,
      likesCount: 89,
      isLiked: false,
      createdAt: new Date(Date.now() - 86400000 * 5).toISOString(),
    },
    {
      id: 'gp3',
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
      content:
        'Sharing my latest blog post on React performance optimization techniques. I cover memo, useMemo, useCallback, and code splitting. Let me know what you think! 🔥',
      imageUrl: '/images/uploads/5eef947c-0c8a-4c0e-8a90-acb49153734d.png',
      privacy: 'public',
      groupId: '1',
      commentsCount: 12,
      likesCount: 203,
      isLiked: true,
      createdAt: new Date(Date.now() - 86400000 * 7).toISOString(),
    },
  ],

  // Posts for group 3 (TypeScript Tips & Tricks)
  '3': [
    {
      id: 'gp4',
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
      content:
        "Just discovered conditional types in TypeScript. Game changer! Here's a quick example:\n\ntype IsString<T> = T extends string ? true : false;\n\nSo powerful for building type-safe libraries!",
      privacy: 'public',
      groupId: '3',
      commentsCount: 34,
      likesCount: 278,
      isLiked: true,
      createdAt: new Date(Date.now() - 86400000 * 3).toISOString(),
    },
    {
      id: 'gp5',
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
        'Pro tip: Use template literal types for type-safe routing in Next.js apps. No more runtime errors for wrong route params!',
      privacy: 'public',
      groupId: '3',
      commentsCount: 18,
      likesCount: 145,
      isLiked: false,
      createdAt: new Date(Date.now() - 86400000 * 10).toISOString(),
    },
  ],

  // Posts for group 5 (Open Source Projects)
  '5': [
    {
      id: 'gp6',
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
      content:
        'Excited to announce v2.0 of our open source CLI tool! Added plugin support, better error handling, and 50% faster execution. Contributors welcome! 🚀\n\nGitHub: github.com/awesome-cli',
      privacy: 'public',
      groupId: '5',
      commentsCount: 56,
      likesCount: 312,
      isLiked: false,
      createdAt: new Date(Date.now() - 86400000 * 15).toISOString(),
    },
  ],
};

// Default posts for groups without specific mock data
export const defaultGroupPosts: Post[] = [
  {
    id: 'gp-default-1',
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
    content:
      "Welcome to the group! Feel free to introduce yourself and share what you're working on.",
    privacy: 'public',
    commentsCount: 5,
    likesCount: 42,
    isLiked: false,
    createdAt: new Date(Date.now() - 86400000 * 30).toISOString(),
  },
  {
    id: 'gp-default-2',
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
      'Hey everyone! Excited to be part of this group. Looking forward to collaborating with you all!',
    privacy: 'public',
    commentsCount: 3,
    likesCount: 28,
    isLiked: true,
    createdAt: new Date(Date.now() - 86400000 * 20).toISOString(),
  },
];
