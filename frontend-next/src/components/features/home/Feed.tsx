'use client';

/**
 * components/features/home/Feed.tsx
 *
 * Feed section — list of PostCards.
 * Extracted from app/page.tsx HomeContent.
 *
 * When the backend is ready, replace mock data with:
 *   const { data: posts } = await getFeed(page);
 * and add infinite scroll or a "Load more" button using the
 * PaginatedResponse shape from types.ts.
 */

import { feedPosts } from '@/mocks/posts';
import PostCard from './PostCard';

export default function Feed() {
  return (
    <section className="feed-section">
      <h2 className="section-title">Feed</h2>
      <div className="feed-posts">
        {feedPosts.map((post) => (
          <PostCard key={post.id} post={post} />
        ))}
      </div>
    </section>
  );
}
