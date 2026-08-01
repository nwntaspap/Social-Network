'use client';

import { useCallback, useEffect, useState } from 'react';
import { getFeed } from '@/lib/api';
import type { Post } from '@/lib/types';
import PostCard from './PostCard';

const PAGE_SIZE = 10;

export default function Feed() {
  const [posts, setPosts] = useState<Post[]>([]);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState('');

  const loadPage = useCallback(async (nextPage: number) => {
    try {
      const response = await getFeed(nextPage, PAGE_SIZE);
      setPosts((prev) => (nextPage === 1 ? response.data : [...prev, ...response.data]));
      setTotalPages(response.totalPages || 1);
    } catch {
      setError('Failed to load posts. Please try again.');
    } finally {
      setLoading(false);
      setLoadingMore(false);
    }
  }, []);

  useEffect(() => {
    let ignore = false;
    getFeed(1, PAGE_SIZE)
      .then((response) => {
        if (ignore) return;
        setPosts(response.data);
        setTotalPages(response.totalPages || 1);
      })
      .catch(() => {
        if (!ignore) setError('Failed to load posts. Please try again.');
      })
      .finally(() => {
        if (!ignore) setLoading(false);
      });
    return () => {
      ignore = true;
    };
  }, []);

  async function handleLoadMore() {
    const nextPage = page + 1;
    setLoadingMore(true);
    setPage(nextPage);
    await loadPage(nextPage);
  }

  if (loading) {
    return (
      <section className="feed-section">
        <h2 className="section-title">Feed</h2>
        <p className="feed-empty">Loading posts...</p>
      </section>
    );
  }

  if (error) {
    return (
      <section className="feed-section">
        <h2 className="section-title">Feed</h2>
        <p className="feed-empty">{error}</p>
      </section>
    );
  }

  return (
    <section className="feed-section">
      <h2 className="section-title">Feed</h2>
      <div className="feed-posts">
        {posts.length > 0 ? (
          posts.map((post) => <PostCard key={post.id} post={post} />)
        ) : (
          <p className="feed-empty">No posts yet. Create the first one!</p>
        )}
      </div>

      {page < totalPages && (
        <div className="feed-load-more">
          <button className="feed-load-more-btn" onClick={handleLoadMore} disabled={loadingMore}>
            {loadingMore ? 'Loading...' : 'Load more'}
          </button>
        </div>
      )}
    </section>
  );
}
