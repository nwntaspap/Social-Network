import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, waitFor } from '@testing-library/react';
import { AuthProvider } from '@/context/AuthContext';

const { replace, params } = vi.hoisted(() => ({
  replace: vi.fn(),
  params: { value: new URLSearchParams() },
}));

vi.mock('next/navigation', () => ({
  useRouter: () => ({
    replace,
    push: vi.fn(),
    prefetch: vi.fn(),
    back: vi.fn(),
    forward: vi.fn(),
    refresh: vi.fn(),
  }),
  useSearchParams: () => params.value,
}));

const { getCurrentUser } = vi.hoisted(() => ({
  getCurrentUser: vi.fn(),
}));

vi.mock('@/lib/api', () => ({
  getCurrentUser,
  ApiError: class ApiError extends Error {
    status: number;
    constructor(status: number, message: string) {
      super(message);
      this.status = status;
    }
  },
}));

import OAuthCallbackPage from './page';

describe('OAuthCallbackPage', () => {
  beforeEach(() => {
    replace.mockClear();
  });

  it('re-fetches the current user and navigates home on success', async () => {
    getCurrentUser.mockResolvedValue({
      id: '1',
      username: 'alice',
      email: 'alice@example.com',
    });
    params.value = new URLSearchParams('flow=login&success=ok&provider=github');

    render(
      <AuthProvider>
        <OAuthCallbackPage />
      </AuthProvider>
    );

    await waitFor(() => expect(replace).toHaveBeenCalledWith('/'));
  });

  it('navigates to login with the error code on failure', async () => {
    getCurrentUser.mockRejectedValue({ status: 401 });
    params.value = new URLSearchParams('flow=login&error=email_exists&provider=google');

    render(
      <AuthProvider>
        <OAuthCallbackPage />
      </AuthProvider>
    );

    await waitFor(() =>
      expect(replace).toHaveBeenCalledWith('/login?error=email_exists&provider=google')
    );
  });

  it('navigates to login when no callback params are present', async () => {
    getCurrentUser.mockRejectedValue({ status: 401 });
    params.value = new URLSearchParams();

    render(
      <AuthProvider>
        <OAuthCallbackPage />
      </AuthProvider>
    );

    await waitFor(() => expect(replace).toHaveBeenCalledWith('/login'));
  });
});
