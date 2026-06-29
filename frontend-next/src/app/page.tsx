'use client';

/**
 * app/page.tsx  →  route: /
 *
 * Mirrors renderHomePage() from pages/home.js.
 * Protected — redirects to /login if not authenticated.
 *
 * What the old handler did:
 *   1. fetchCategories()
 *   2. prepareCategories() (color normalisation)
 *   3. Render category list + detail panel
 *
 * We replicate that here with the same skeleton/error states.
 */

import { useEffect, useState } from 'react';
import AuthGuard from '@/components/AuthGuard';
import { fetchCategories } from '@/lib/api';
import { prepareCategories, formatRelativeDate } from '@/lib/helpers';
import type { Category, Topic } from '@/lib/types';
import Link from 'next/link';

export default function HomePage() {
  return (
    <AuthGuard>
      <HomeContent />
    </AuthGuard>
  );
}

function HomeContent() {
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  import SearchBar from '@/components/features/home/SearchBar';
  import SuggestedUsers from '@/components/features/home/SuggestedUsers';
  import Feed from '@/components/features/home/Feed';

  function HomeContent() {
    return (
      <div className="home-container">
        <SearchBar />
        <SuggestedUsers />
        <Feed />
      </div>
    );
  }
  useEffect(() => {
    async function load() {
      try {
        const data = await fetchCategories();
        const raw = Array.isArray(data)
          ? data
          : ((data as Record<string, unknown>)?.categories ??
            (data as Record<string, unknown>)?.Categories ??
            []);
        setCategories(prepareCategories(raw as Record<string, unknown>[]));
      } catch (err: unknown) {
        setError(err instanceof Error ? err.message : 'Failed to load categories');
      } finally {
        setLoading(false);
      }
    }
    load();
  }, []);

  // Close the category details dropdown when clicking outside
  useEffect(() => {
    const details = document.querySelector('.category-details');
    if (!details) return;

    function handleClick(e: MouseEvent) {
      if (!details!.contains(e.target as Node)) {
        details!.removeAttribute('open');
      }
    }

    document.addEventListener('click', handleClick, { passive: true });
    return () => document.removeEventListener('click', handleClick);
  }, [categories]); // Re-attach when categories load

  if (loading) return <HomeSkeleton />;
  if (error) return <HomeError message={error} />;

  return (
    <>
      <h1 className="forum-title">Welcome to SocialNet</h1>
      <div className="main-container">
        {/* Category details dropdown + nav buttons — mirrors buildCategoryDetailsHTML() */}
        <div className="nav-categories">
          <details className="category-details">
            <summary>Categories</summary>
            <div className="details-content">
              {categories.length > 0 ? (
                categories.map((cat) => {
                  const id = cat.ID ?? cat.id ?? '';
                  const name = cat.Name ?? cat.name ?? '';
                  const color = cat.Color ?? cat.color ?? '#00C6FF';
                  const topicCount = cat.TopicCount ?? cat.topic_count ?? 0;

                  return (
                    <Link
                      key={id}
                      href={`/topics?search=&category=${id}`}
                      className="details-category-link"
                    >
                      <div className="details-text-box">
                        <span
                          className="category-title-color"
                          style={{ backgroundColor: color }}
                        ></span>
                        <span className="details-category-title">{name}</span>
                      </div>
                      <span className="category-count">{topicCount}</span>
                    </Link>
                  );
                })
              ) : (
                <span className="details-category-title">No categories yet</span>
              )}
            </div>
          </details>
          <Link href="/categories" className="nav-categories-btn">
            Categories
          </Link>
          <Link href="/topics" className="nav-categories-btn">
            Topics
          </Link>
        </div>

        {/* Category cards — mirrors buildCategoriesListHTML() */}
        <CategoryList categories={categories} />
      </div>
    </>
  );
}

// ─── Category list ────────────────────────────────────────────────────────────

function CategoryList({ categories }: { categories: Category[] }) {
  if (!categories.length) {
    return (
      <div className="categories-container">
        <p className="no-topics-message">No categories found. Check back later!</p>
      </div>
    );
  }

  return (
    <div className="categories-container">
      {categories.map((cat) => (
        <CategoryCard key={cat.id || cat.ID} category={cat} />
      ))}
    </div>
  );
}

function CategoryCard({ category: cat }: { category: Category }) {
  const id = String(cat.ID ?? cat.id ?? '');
  const name = cat.Name ?? cat.name ?? '';
  const description = cat.Description ?? cat.description ?? '';
  const color = cat.Color ?? cat.color ?? '#00C6FF';
  const imagePath = cat.ImagePath ?? cat.image_path ?? '/images/categories/default_category.png';
  console.log(imagePath);

  const topics = Array.isArray(cat.Topics)
    ? cat.Topics
    : Array.isArray(cat.topics)
      ? cat.topics
      : [];

  return (
    <div className="category">
      <div className="category-wrapper">
        <div className="category-img-box">
          <Link href={`/topics?search=&category=${id}`} className="category-link">
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img className="category-img" src={imagePath} alt={name} />
          </Link>
        </div>
        <div className="category-info">
          <div className="category-info-box">
            <Link href={`/topics?search=&category=${id}`} className="category-link">
              <div className="category-title-box">
                <span className="category-title-color" style={{ backgroundColor: color }}></span>
                <span className="category-title">{name}</span>
              </div>
            </Link>
            <p className="category-description">{description}</p>
          </div>
        </div>
      </div>
      <div className="category-posts">
        {topics.length > 0 ? (
          topics.slice(0, 3).map((topic: Topic) => {
            const topicID = String(topic.ID ?? topic.id ?? '');
            const topicTitle = topic.Title ?? topic.title ?? 'Untitled';
            const topicDate = formatRelativeDate(topic.CreatedAt ?? topic.created_at ?? '');

            return (
              <div key={topicID} className="category-post">
                <Link href={`/topic/${topicID}`} className="topic-link">
                  <span className="category-post-title">
                    <span className="left-arrow">&#10147;</span> {topicTitle}
                  </span>
                </Link>
                <span className="category-post-date">{topicDate}</span>
              </div>
            );
          })
        ) : (
          <span className="category-post-date">No posts yet</span>
        )}
      </div>
    </div>
  );
}

// ─── Skeleton ─────────────────────────────────────────────────────────────────

function HomeSkeleton() {
  return (
    <>
      <h1 className="forum-title">Welcome to Forum</h1>
      <div className="main-container">
        <div className="nav-categories">
          <div
            className="skeleton skeleton-btn"
            style={{ width: 110, height: 40, borderRadius: 4 }}
          />
          <div
            className="skeleton skeleton-btn"
            style={{ width: 100, height: 40, borderRadius: 4 }}
          />
          <div
            className="skeleton skeleton-btn"
            style={{ width: 80, height: 40, borderRadius: 4 }}
          />
        </div>
        <div className="categories-container">
          {[1, 2, 3].map((i) => (
            <div key={i} className="category category--skeleton">
              <div className="category-wrapper">
                <div className="skeleton skeleton-img" />
                <div className="category-info-box">
                  <div className="skeleton skeleton-title" />
                  <div className="skeleton skeleton-desc" />
                </div>
              </div>
              <div className="category-posts">
                <div className="skeleton skeleton-post" />
                <div className="skeleton skeleton-post" />
                <div className="skeleton skeleton-post" />
              </div>
            </div>
          ))}
        </div>
      </div>
    </>
  );
}

// ─── Error state ──────────────────────────────────────────────────────────────

function HomeError({ message }: { message: string }) {
  return (
    <>
      <h1 className="forum-title">Welcome to Forum</h1>
      <div className="main-container">
        <div className="categories-container" style={{ padding: '2rem', textAlign: 'center' }}>
          <p style={{ color: '#e53e3e', fontSize: '1.1rem' }}>⚠️ {message}</p>
          <p style={{ marginTop: '1rem', color: 'var(--grey-color)' }}>
            Could not load categories. Please try refreshing the page.
          </p>
        </div>
      </div>
    </>
  );
}
