import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { AuthProvider } from '@/context/AuthContext';

const { push, loginEmail, currentParams } = vi.hoisted(() => ({
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
  currentParams: { value: new URLSearchParams() },
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
  useSearchParams: () => currentParams.value,
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
    currentParams.value = new URLSearchParams();
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

  it('points OAuth buttons at the OAuth init routes', async () => {
    render(
      <AuthProvider>
        <LoginPage />
      </AuthProvider>
    );

    const google = await screen.findByRole('link', { name: /continue with google/i });
    expect(google.getAttribute('href')).toBe('/api/v1/auth/oauth/google/init');

    const github = screen.getByRole('link', { name: /continue with github/i });
    expect(github.getAttribute('href')).toBe('/api/v1/auth/oauth/github/init');
  });

  it('shows a message when the OAuth callback redirects with an error', async () => {
    currentParams.value = new URLSearchParams('error=email_exists&provider=google');

    render(
      <AuthProvider>
        <LoginPage />
      </AuthProvider>
    );

    expect(await screen.findByRole('alert')).toHaveTextContent(/google/i);
    expect(screen.getByRole('alert')).toHaveTextContent(/already exists/i);
  });
});
