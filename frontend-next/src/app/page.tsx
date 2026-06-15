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
import { prepareCategories, type Category } from '@/lib/helpers';

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
  const [openDetails, setOpenDetails] = useState(false);
  const [selectedCategory, setSelectedCategory] = useState<Category | null>(null);

  useEffect(() => {
    async function load() {
      try {
        const data = await fetchCategories();
        const raw = Array.isArray(data)
          ? data
          : ((data as Record<string, unknown>)?.categories ??
            (data as Record<string, unknown>)?.Categories ??
            []);
        setCategories(prepareCategories(raw as Category[]));
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
    if (!openDetails) return;
    function handleClick() {
      setOpenDetails(false);
    }
    document.addEventListener('click', handleClick, { passive: true });
    return () => document.removeEventListener('click', handleClick);
  }, [openDetails]);

  if (loading) return <HomeSkeleton />;
  if (error) return <HomeError message={error} />;

  return (
    <>
      <h1 className="forum-title">Welcome to Forum</h1>
      <div className="main-container">
        {/* Category details dropdown — mirrors buildCategoryDetailsHTML() */}
        <details
          className="category-details"
          open={openDetails}
          onClick={(e) => e.stopPropagation()}
        >
          <summary
            onClick={(e) => {
              e.preventDefault();
              setOpenDetails((prev) => !prev);
            }}
          >
            {selectedCategory ? selectedCategory.Name || selectedCategory.name : 'All Categories'}
          </summary>
          <div className="category-details-list">
            <button
              className="category-details-item"
              onClick={() => {
                setSelectedCategory(null);
                setOpenDetails(false);
              }}
            >
              All Categories
            </button>
            {categories.map((cat) => (
              <button
                key={cat.id || cat.ID}
                className="category-details-item"
                style={{ borderLeft: `4px solid ${cat.Color}` }}
                onClick={() => {
                  setSelectedCategory(cat);
                  setOpenDetails(false);
                }}
              >
                {cat.Name || cat.name}
              </button>
            ))}
          </div>
        </details>

        {/* Category cards — mirrors buildCategoriesListHTML() */}
        <CategoryList
          categories={
            selectedCategory
              ? categories.filter(
                  (c) => (c.id || c.ID) === (selectedCategory.id || selectedCategory.ID)
                )
              : categories
          }
        />
      </div>
    </>
  );
}

// ─── Category list ────────────────────────────────────────────────────────────

function CategoryList({ categories }: { categories: Category[] }) {
  if (!categories.length) {
    return (
      <div className="categories-container" style={{ padding: '2rem', textAlign: 'center' }}>
        <p style={{ color: 'var(--grey-color)' }}>No categories found. Check back later!</p>
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
  const imgSrc = cat.ImagePath || cat.imagePath || '/images/categories/default_category.png';
  const name = cat.Name || cat.name || '';
  const description = cat.Description || cat.description || '';

  return (
    <div className="category" style={{ borderTop: `4px solid ${cat.Color}` }}>
      <div className="category-img-box">
        {/* eslint-disable-next-line @next/next/no-img-element */}
        <img src={imgSrc} alt={name} className="category-img" />
      </div>
      <div className="category-info">
        <h3 className="category-name">{name}</h3>
        <p className="category-description">{description}</p>
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
        {[1, 2, 3].map((i) => (
          <div key={i} className="category category--skeleton">
            <div className="category-img-box">
              <div className="skeleton skeleton-img" />
            </div>
            <div className="category-info">
              <div className="skeleton skeleton-title" />
              <div className="skeleton skeleton-desc" />
            </div>
          </div>
        ))}
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
