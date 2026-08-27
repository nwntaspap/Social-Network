'use client';

/**
 * app/auth/callback/page.tsx  →  route: /auth/callback
 *
 * Landing page after an OAuth provider redirects back to the backend. The
 * backend exchanges the code, creates a session, sets the access cookie, then
 * 307-redirects here with `?flow=login&success=ok&provider=github` (or
 * `?flow=...&error=<code>&provider=...` on failure).
 *
 * On success we re-fetch /me (the cookie is now set) and bounce to the home
 * page. On error we bounce back to /login with the error code so the login
 * page can show a message.
 */

import { Suspense, useEffect, useRef } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useAuth } from '@/context/AuthContext';

function OAuthCallback() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { user, refreshUser } = useAuth();
  const started = useRef(false);

  useEffect(() => {
    if (started.current) return;
    started.current = true;

    const success = searchParams.get('success');
    const error = searchParams.get('error');
    const provider = searchParams.get('provider');
    const providerParam = provider ? `&provider=${encodeURIComponent(provider)}` : '';

    if (success === 'ok') {
      refreshUser();
      return;
    }
    if (error) {
      router.replace(`/login?error=${encodeURIComponent(error)}${providerParam}`);
      return;
    }
    router.replace('/login');
  }, [router, searchParams, refreshUser]);

  // Once auth has re-resolved after refreshUser(), navigate based on outcome.
  useEffect(() => {
    if (!started.current || user === undefined) return;

    if (user) {
      router.replace('/');
    } else {
      const provider = searchParams.get('provider');
      router.replace(
        `/login?error=errorAtOauthLogin${provider ? `&provider=${encodeURIComponent(provider)}` : ''}`
      );
    }
  }, [user, router, searchParams]);

  return (
    <div className="page-loading">
      <div className="page-loading-spinner" />
    </div>
  );
}

export default function AuthCallbackPage() {
  return (
    <Suspense
      fallback={
        <div className="page-loading">
          <div className="page-loading-spinner" />
        </div>
      }
    >
      <OAuthCallback />
    </Suspense>
  );
}
