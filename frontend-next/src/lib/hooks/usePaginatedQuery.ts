/**
 * lib/hooks/usePaginatedQuery.ts
 *
 * Thin wrapper over TanStack Query's useInfiniteQuery, shaped specifically
 * around our PaginatedResponse<T> type (page/pageSize/totalCount/totalPages).
 *
 * Every paginated feature (Feed, GroupPosts, GroupMembers, Notifications,
 * Followers/Following, ...) calls this instead of useInfiniteQuery directly.
 * That means:
 *   - "page 1 by default, Load More advances it" lives in exactly one place
 *   - hasMore / loadedCount / totalCount are computed the same way everywhere
 *   - LoadMoreButton props are always assembled the same way
 *
 * Usage:
 *   const feed = usePaginatedQuery({
 *     queryKey: queryKeys.feed(),
 *     fetchPage: (page) => getFeed(page), // or mockGetFeed(page) for now
 *   });
 *
 *   <LoadMoreButton
 *     onLoadMore={feed.loadMore}
 *     isLoading={feed.isFetchingMore}
 *     hasMore={feed.hasMore}
 *     loadedCount={feed.loadedCount}
 *     totalCount={feed.totalCount}
 *   />
 */

import { useInfiniteQuery } from '@tanstack/react-query';
import type { PaginatedResponse } from '@/lib/types';

interface UsePaginatedQueryOptions<T> {
  queryKey: readonly unknown[];
  fetchPage: (page: number, signal?: AbortSignal) => Promise<PaginatedResponse<T>>;
  /** Set false to skip fetching (e.g. waiting on a groupId param) */
  enabled?: boolean;
}

export function usePaginatedQuery<T>({
  queryKey,
  fetchPage,
  enabled = true,
}: UsePaginatedQueryOptions<T>) {
  const query = useInfiniteQuery({
    queryKey,
    queryFn: ({ pageParam, signal }) => fetchPage(pageParam, signal),
    initialPageParam: 1,
    getNextPageParam: (lastPage) =>
      lastPage.page < lastPage.totalPages ? lastPage.page + 1 : undefined,
    enabled,
  });

  const items = query.data?.pages.flatMap((p) => p.data) ?? [];
  const totalCount = query.data?.pages[0]?.totalCount ?? 0;

  return {
    items,
    isLoading: query.isLoading,
    isError: query.isError,
    error: query.error,
    hasMore: !!query.hasNextPage,
    isFetchingMore: query.isFetchingNextPage,
    loadMore: () => query.fetchNextPage(),
    loadedCount: items.length,
    totalCount,
    refetch: query.refetch,
  };
}
