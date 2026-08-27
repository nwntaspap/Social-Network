import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { AuthProvider } from '@/context/AuthContext';

vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: vi.fn(),
    replace: vi.fn(),
    prefetch: vi.fn(),
    back: vi.fn(),
    forward: vi.fn(),
    refresh: vi.fn(),
  }),
}));

vi.mock('@/lib/api', () => ({
  api: {
    get: vi.fn().mockResolvedValue({
      id: '1',
      email: 'test@test.com',
      username: 'testuser',
      firstName: 'Test',
      lastName: 'User',
      dateOfBirth: '1990-01-01',
      isPublic: true,
      createdAt: '2024-01-01',
    }),
  },
  getCurrentUser: vi.fn().mockResolvedValue({
    id: '1',
    email: 'test@test.com',
    username: 'testuser',
    firstName: 'Test',
    lastName: 'User',
    dateOfBirth: '1990-01-01',
    isPublic: true,
    createdAt: '2024-01-01',
  }),
  getFeed: vi.fn().mockResolvedValue({ data: [], totalPages: 1 }),
  searchUsers: vi.fn().mockResolvedValue({ data: [] }),
  ApiError: class ApiError extends Error {
    status: number;
    constructor(status: number, message: string) {
      super(message);
      this.status = status;
    }
    get isUnauthorized() {
      return this.status === 401;
    }
    get isForbidden() {
      return this.status === 403;
    }
    get isNotFound() {
      return this.status === 404;
    }
    get isTooManyRequests() {
      return this.status === 429;
    }
  },
}));

import HomePage from './page';

describe('HomePage', () => {
  it('renders without crashing', async () => {
    render(
      <AuthProvider>
        <HomePage />
      </AuthProvider>
    );
    const heading = await screen.findByRole('heading', { name: /Feed/i });
    expect(heading).toBeInTheDocument();
  });

  it('displays the welcome message', async () => {
    render(
      <AuthProvider>
        <HomePage />
      </AuthProvider>
    );
    expect(await screen.findByText(/Suggested Users/i)).toBeInTheDocument();
  });
});
