'use client';

/**
 * components/ui/SearchDropdown.tsx
 *
 * Generic search input with debounced dropdown results.
 * Works with local mock data now, async API calls later.
 *
 * For local filtering: provide items + filterFn
 * When backend is ready, replace with:
 *   - Remove: items and filterFn props
 *   - Add: onSearch={handleSearch} where handleSearch calls API
 */

import { useEffect, useMemo, useRef, useState } from 'react';
import Image from 'next/image';
import Link from 'next/link';
import { debounce } from '@/lib/helpers';

type SearchDropdownProps<T> = {
  placeholder?: string;
  items?: T[];
  filterFn?: (item: T, query: string) => boolean;
  renderItem: (item: T) => React.ReactNode;
  getItemKey: (item: T) => string;
  getItemHref: (item: T) => string;
  onSearch?: (query: string) => Promise<T[]>;
  emptyMessage?: string;
};

export default function SearchDropdown<T>({
  placeholder = 'Search...',
  items = [],
  filterFn,
  renderItem,
  getItemKey,
  getItemHref,
  onSearch,
  emptyMessage = 'No results found',
}: SearchDropdownProps<T>) {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<T[]>([]);
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
      debounce(async (q: string) => {
        if (!q.trim()) {
          setResults([]);
          setShowResults(false);
          setIsSearching(false);
          return;
        }

        setIsSearching(true);

        // Use async callback if provided, otherwise filter locally
        let searchResults: T[];
        if (onSearch) {
          searchResults = await onSearch(q);
        } else if (filterFn) {
          searchResults = items.filter((item) => filterFn(item, q));
        } else {
          searchResults = [];
        }

        setResults(searchResults);
        setShowResults(true);
        setIsSearching(false);
      }, 300),
    [items, filterFn, onSearch]
  );

  function handleChange(e: React.ChangeEvent<HTMLInputElement>) {
    const val = e.target.value;
    setQuery(val);
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
          placeholder={placeholder}
          value={query}
          onChange={handleChange}
          onFocus={() => query && results.length > 0 && setShowResults(true)}
        />
        {isSearching && <div className="search-spinner" />}
      </div>

      {showResults && (
        <div className="search-results-dropdown">
          {results.length > 0 ? (
            results.map((item) => (
              <Link
                key={getItemKey(item)}
                href={getItemHref(item)}
                className="search-result-item"
                onClick={() => setShowResults(false)}
              >
                {renderItem(item)}
              </Link>
            ))
          ) : (
            <div className="search-no-results">{emptyMessage.replace('{query}', query)}</div>
          )}
        </div>
      )}
    </section>
  );
}
