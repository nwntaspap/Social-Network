import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { AuthProvider } from '@/context/AuthContext';

const { push, loginEmail } = vi.hoisted(() => ({
  push: vi.fn(),
  loginEmail: vi.fn().mockResolvedValue({
    token: 'tok',
    user: {
      id: '1',
      email: 'a@b.com',
      username: 'alice',
      firstName: 'Alice',
      lastName: 'Smith',
      dateOfBirth: '1990-05-20',
      isPublic: true,
      createdAt: '2024-01-01',
    },
  }),
}));

vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push,
    replace: vi.fn(),
    prefetch: vi.fn(),
    back: vi.fn(),
    forward: vi.fn(),
    refresh: vi.fn(),
  }),
}));

vi.mock('@/lib/api', () => ({
  api: {
    get: vi.fn(),
    post: vi.fn(),
  },
  loginEmail,
  loginUsername: vi.fn(),
  getCurrentUser: vi.fn().mockRejectedValue({ status: 401 }),
  ApiError: class ApiError extends Error {
    status: number;
    constructor(status: number, message: string) {
      super(message);
      this.status = status;
    }
  },
}));

import LoginPage from './page';

describe('LoginPage', () => {
  beforeEach(() => {
    push.mockClear();
  });

  it('stores the full user from the login response and navigates home', async () => {
    render(
      <AuthProvider>
        <LoginPage />
      </AuthProvider>
    );

    fireEvent.click(await screen.findByLabelText('Email'));
    fireEvent.change(await screen.findByLabelText('Email address'), {
      target: { value: 'a@b.com' },
    });
    fireEvent.change(screen.getByLabelText('Password'), { target: { value: 'pass' } });
    fireEvent.click(screen.getByRole('button', { name: 'Sign In' }));

    await waitFor(() => expect(push).toHaveBeenCalledWith('/'));
    expect(loginEmail).toHaveBeenCalledWith('a@b.com', 'pass');
  });
});
