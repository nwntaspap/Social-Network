'use client';

/**
 * context/QueryProvider.tsx
 *
 * Wraps the app in a TanStack Query QueryClientProvider.
 *
 * The QueryClient is created inside useState (not at module scope) so that
 * each request/browser session gets its own instance — important in Next.js
 * to avoid leaking cached data across users during SSR.
 *
 * Default options here are tuned for a social feed:
 *   - staleTime: data is considered fresh for 30s, so navigating back to a
 *     page you just visited won't trigger an instant refetch.
 *   - retry: 1 retry on failure (network hiccup), not the default 3 — we
 *     don't want a dead backend to hang the UI for a long time.
 *   - refetchOnWindowFocus: on for now, mirrors "check for new stuff when
 *     you come back to the tab" behavior common in social apps. Can be
 *     turned off per-query later if it gets noisy.
 */

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useState } from 'react';

export function QueryProvider({ children }: { children: React.ReactNode }) {
  const [client] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: 30_000,
            retry: 1,
            refetchOnWindowFocus: true,
          },
          mutations: {
            retry: 0,
          },
        },
      })
  );

  return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}
