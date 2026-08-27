import type { Event } from '@/lib/types';

const eventOptions = [
  { id: 'opt-going', label: 'Going', tally: 0 },
  { id: 'opt-not-going', label: 'Not going', tally: 0 },
];

export const mockGroupEvents: Record<string, Event[]> = {
  // Events for group 1 (React Developers)
  '1': [
    {
      id: 'evt1',
      groupId: '1',
      creatorId: '1',
      creator: {
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
      title: 'React 19 Feature Preview & Discussion',
      description:
        "Join us for an exciting session exploring the upcoming React 19 features! We'll cover the new React Compiler, Server Components improvements, and more. Bring your questions and ideas!",
      eventDate: new Date(Date.now() + 86400000 * 14).toISOString(), // 2 weeks from now
      createdAt: new Date(Date.now() - 86400000 * 5).toISOString(),
      options: eventOptions,
    },
    {
      id: 'evt2',
      groupId: '1',
      creatorId: '3',
      creator: {
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
      title: 'Monthly Code Review Session',
      description:
        "Bring your React code for peer review! We'll break into small groups and provide constructive feedback on each other's projects. All skill levels welcome.",
      eventDate: new Date(Date.now() + 86400000 * 7).toISOString(), // 1 week from now
      createdAt: new Date(Date.now() - 86400000 * 3).toISOString(),
      options: eventOptions,
    },
    {
      id: 'evt3',
      groupId: '1',
      creatorId: '1',
      creator: {
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
      title: 'Building a Full-Stack App with Next.js 14',
      description:
        "Hands-on workshop where we'll build a complete full-stack application using Next.js 14, React Server Components, and Prisma. Prerequisites: Basic React and TypeScript knowledge.",
      eventDate: new Date(Date.now() + 86400000 * 21).toISOString(), // 3 weeks from now
      createdAt: new Date(Date.now() - 86400000 * 2).toISOString(),
      options: eventOptions,
    },
  ],

  // Events for group 5 (Open Source Projects)
  '5': [
    {
      id: 'evt4',
      groupId: '5',
      creatorId: '3',
      creator: {
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
      title: 'Open Source Contributors Meetup',
      description:
        'Virtual meetup for open source contributors. Share your projects, find collaborators, and discuss best practices for maintaining healthy open source communities.',
      eventDate: new Date(Date.now() + 86400000 * 10).toISOString(),
      createdAt: new Date(Date.now() - 86400000 * 6).toISOString(),
      options: eventOptions,
    },
    {
      id: 'evt5',
      groupId: '5',
      creatorId: '6',
      creator: {
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
      title: 'Hacktoberfest Planning Session',
      description:
        "Planning session for Hacktoberfest 2024. We'll identify projects to contribute to, set up mentoring pairs, and create contribution guidelines for newcomers.",
      eventDate: new Date(Date.now() + 86400000 * 30).toISOString(),
      createdAt: new Date(Date.now() - 86400000 * 1).toISOString(),
      options: eventOptions,
    },
  ],

  // Events for group 3 (TypeScript Tips & Tricks)
  '3': [
    {
      id: 'evt6',
      groupId: '3',
      creatorId: '5',
      creator: {
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
      title: 'Advanced TypeScript Patterns Workshop',
      description:
        'Deep dive into advanced TypeScript patterns including conditional types, mapped types, template literals, and type inference. Bring your challenging type problems!',
      eventDate: new Date(Date.now() + 86400000 * 12).toISOString(),
      createdAt: new Date(Date.now() - 86400000 * 4).toISOString(),
      options: eventOptions,
    },
  ],
};

// Default events for groups without specific mock data
export const defaultGroupEvents: Event[] = [
  {
    id: 'evt-default-1',
    groupId: 'default',
    creatorId: '1',
    creator: {
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
    title: 'Welcome & Introduction Session',
    description:
      'New to the group? Join our welcome session to meet other members, learn about group activities, and share what you hope to get out of the community.',
    eventDate: new Date(Date.now() + 86400000 * 5).toISOString(),
    createdAt: new Date(Date.now() - 86400000 * 10).toISOString(),
    options: eventOptions,
  },
  {
    id: 'evt-default-2',
    groupId: 'default',
    creatorId: '2',
    creator: {
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
    title: 'Monthly Community Call',
    description:
      'Regular monthly catch-up where we discuss group updates, upcoming events, and any topics members want to bring up. Everyone is welcome!',
    eventDate: new Date(Date.now() + 86400000 * 15).toISOString(),
    createdAt: new Date(Date.now() - 86400000 * 8).toISOString(),
    options: eventOptions,
  },
];
