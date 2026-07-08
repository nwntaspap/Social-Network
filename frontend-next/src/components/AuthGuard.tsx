'use client';

/**
 * components/AuthGuard.tsx
 *
 * Replicates the `protected: true` route flag from router.js.
 *
 * Old JS:
 *   if (route.protected && !user) { navigate('/login'); return; }
 *
 * Here:
 *   Wrap any protected page in <AuthGuard> — it waits for auth to resolve,
 *   then either renders children (user present) or redirects to /login.
 *
 * Usage in a page:
 *   export default function HomePage() {
 *     return <AuthGuard><HomeContent /></AuthGuard>;
 *   }
 */

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/context/AuthContext';

interface AuthGuardProps {
  children: React.ReactNode;
  /** Where to redirect unauthenticated users. Defaults to /login */
  redirectTo?: string;
}

export default function AuthGuard({ children, redirectTo = '/login' }: AuthGuardProps) {
  const { user, loading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    // Wait until auth is resolved before making any decision
    if (loading) return;
    if (!user) {
      router.replace(redirectTo);
    }
  }, [user, loading, router, redirectTo]);

  // Still resolving — show the same spinner the old SPA showed
  if (loading || user === undefined) {
    return (
      <div className="page-loading">
        <div className="page-loading-spinner" />
      </div>
    );
  }

  // Not logged in — redirect is already happening via useEffect
  if (!user) {
    return (
      <div className="page-loading">
        <div className="page-loading-spinner" />
      </div>
    );
  }

  return <>{children}</>;
}
