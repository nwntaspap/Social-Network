'use client';

/**
 * components/features/home/SearchBar.tsx
 *
 * User search input with debounced dropdown results.
 * Extracted from app/page.tsx HomeContent.
 *
 * When the backend is ready, replace the mock filter with:
 *   const results = await searchUsers(query);
 */

import { useEffect, useMemo, useRef, useState } from 'react';
import Image from 'next/image';
import Link from 'next/link';
import { getDisplayName, getFileUrl, debounce } from '@/lib/helpers';
import { searchResults as mockSearchResults } from '@/mocks/users';
import type { User } from '@/lib/types';

export default function SearchBar() {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<User[]>([]);
  const [showResults, setShowResults] = useState(false);
  const [isSearching, setIsSearching] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  // Close dropdown on outside click
  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setShowResults(false);
      }
    }
    document.addEventListener('click', handleClick);
    return () => document.removeEventListener('click', handleClick);
  }, []);

  const debouncedSearch = useMemo(
    () =>
      debounce((q: string) => {
        if (!q.trim()) {
          setResults([]);
          setShowResults(false);
          setIsSearching(false);
          return;
        }

        // TODO: replace with real API call: searchUsers(q)
        const filtered = mockSearchResults.filter(
          (u) =>
            u.username.toLowerCase().includes(q.toLowerCase()) ||
            u.firstName.toLowerCase().includes(q.toLowerCase()) ||
            u.lastName.toLowerCase().includes(q.toLowerCase())
        );

        setResults(filtered);
        setShowResults(true);
        setIsSearching(false);
      }, 300),
    []
  );

  function handleChange(e: React.ChangeEvent<HTMLInputElement>) {
    const val = e.target.value;
    setQuery(val);
    setIsSearching(true);
    debouncedSearch(val);
  }

  return (
    <section className="search-section" ref={containerRef}>
      <div className="search-input-wrapper">
        <Image
          src="/images/icons/search-icon.png"
          alt="Search"
          width={20}
          height={20}
          className="search-icon"
        />
        <input
          type="text"
          className="search-input"
          placeholder="Search users..."
          value={query}
          onChange={handleChange}
          onFocus={() => query && results.length > 0 && setShowResults(true)}
        />
        {isSearching && <div className="search-spinner" />}
      </div>

      {showResults && (
        <div className="search-results-dropdown">
          {results.length > 0 ? (
            results.map((user) => (
              <Link
                key={user.id}
                href={`/profile/${user.id}`}
                className="search-result-item"
                onClick={() => setShowResults(false)}
              >
                <Image
                  src={getFileUrl(user.avatarUrl)}
                  alt={getDisplayName(user)}
                  width={40}
                  height={40}
                  className="search-result-avatar"
                />
                <div className="search-result-info">
                  <span className="search-result-name">{getDisplayName(user)}</span>
                  <span className="search-result-username">@{user.username}</span>
                </div>
              </Link>
            ))
          ) : (
            <div className="search-no-results">No users found for &quot;{query}&quot;</div>
          )}
        </div>
      )}
    </section>
  );
}
