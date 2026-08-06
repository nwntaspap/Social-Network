'use client';

/**
 * components/ui/LoadMoreButton.tsx
 *
 * Generic pagination control. Pairs directly with the object returned by
 * usePaginatedQuery — pass loadMore/isFetchingMore/hasMore/loadedCount/
 * totalCount straight through, no mapping needed.
 */

interface LoadMoreButtonProps {
  onLoadMore: () => void;
  isLoading: boolean;
  hasMore: boolean;
  label?: string;
  loadingLabel?: string;
  loadedCount?: number;
  totalCount?: number;
}

export default function LoadMoreButton({
  onLoadMore,
  isLoading,
  hasMore,
  label = 'Load More',
  loadingLabel = 'Loading...',
  loadedCount,
  totalCount,
}: LoadMoreButtonProps) {
  if (!hasMore) {
    return null;
  }

  return (
    <div className="load-more-container">
      {loadedCount !== undefined && totalCount !== undefined && (
        <p className="load-more-count">
          Showing {loadedCount} of {totalCount}
        </p>
      )}
      <button
        className="load-more-button"
        onClick={onLoadMore}
        disabled={isLoading}
        aria-busy={isLoading}
      >
        {isLoading ? (
          <>
            <span className="spinner" aria-hidden="true" />
            {loadingLabel}
          </>
        ) : (
          label
        )}
      </button>
    </div>
  );
}
