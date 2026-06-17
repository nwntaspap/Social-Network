'use client';

/**
 * context/AuthContext.tsx
 *
 * Replaces the module-level auth state in auth.js.
 *
 * The old JS kept _currentUser in a module variable and called /me inside
 * authMiddleware() on every route change. Here we:
 *   1. Call /me once on mount inside AuthProvider (same effect — runs when
 *      the app first loads / hydrates)
 *   2. Expose user, loading, setUser, clearUser via context
 *   3. Wrap the whole app in AuthProvider inside layout.tsx so every page
 *      can read auth state without prop drilling
 *
 * Route protection (route.protected) is handled by the AuthGuard component.
 */

import { createContext, useCallback, useContext, useEffect, useRef, useState } from 'react';
import { api, ApiError } from '@/lib/api';
import { User } from '@/lib/types';

// ─── Types ────────────────────────────────────────────────────────────────────

interface AuthContextValue {
  /** The logged-in user, or null if unauthenticated. Undefined = not yet resolved. */
  user: User | null | undefined;
  /** True while the first /me call is in flight */
  loading: boolean;
  /** Call after a successful login to update state without another /me round-trip */
  setUser: (user: User) => void;
  /** Call after logout to clear state */
  clearUser: () => void;
  /** Re-fetch /me (used after OAuth callback) */
  refreshUser: () => Promise<void>;
}

// ─── Context ──────────────────────────────────────────────────────────────────

const AuthContext = createContext<AuthContextValue | null>(null);

// ─── Provider ─────────────────────────────────────────────────────────────────

export function AuthProvider({ children }: { children: React.ReactNode }) {
  // undefined = not yet resolved, null = resolved but not logged in
  const [user, setUserState] = useState<User | null | undefined>(undefined);
  const [loading, setLoading] = useState(true);
  const inflightRef = useRef<Promise<void> | null>(null);

  const resolve = useCallback(async () => {
    // Deduplicate concurrent calls (mirrors _authCallInProgress pattern)
    if (inflightRef.current) return inflightRef.current;

    inflightRef.current = (async () => {
      try {
        const me = await api.get<User>('/me');
        setUserState(me);
      } catch (err) {
        if (err instanceof ApiError && err.isUnauthorized) {
          setUserState(null);
        } else {
          console.warn('Could not resolve current user:', err);
          setUserState(null);
        }
      } finally {
        setLoading(false);
        inflightRef.current = null;
      }
    })();

    return inflightRef.current;
  }, []);

  // Resolve once on mount — mirrors authMiddleware() called at app boot
  useEffect(() => {
    resolve();
  }, [resolve]);

  const setUser = useCallback((u: User) => {
    setUserState(u);
    setLoading(false);
  }, []);

  const clearUser = useCallback(() => {
    setUserState(null);
    setLoading(false);
  }, []);

  const refreshUser = useCallback(async () => {
    setLoading(true);
    await resolve();
  }, [resolve]);

  return (
    <AuthContext.Provider value={{ user, loading, setUser, clearUser, refreshUser }}>
      {children}
    </AuthContext.Provider>
  );
}

// ─── Hook ─────────────────────────────────────────────────────────────────────

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used inside <AuthProvider>');
  return ctx;
}
