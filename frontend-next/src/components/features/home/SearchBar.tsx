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

import Image from 'next/image';
import { getDisplayName, getFileUrl } from '@/lib/helpers';
import { searchResults as mockSearchResults } from '@/mocks/users';
import type { User } from '@/lib/types';
import SearchDropdown from '@/components/ui/SearchDropdown';

export default function UserSearchBar() {
  return (
    <SearchDropdown<User>
      placeholder="Search users..."
      items={mockSearchResults}
      filterFn={(user, q) =>
        user.username.toLowerCase().includes(q.toLowerCase()) ||
        user.firstName.toLowerCase().includes(q.toLowerCase()) ||
        user.lastName.toLowerCase().includes(q.toLowerCase())
      }
      renderItem={(user) => (
        <>
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
        </>
      )}
      getItemKey={(user) => user.id}
      getItemHref={(user) => `/profile/${user.id}`}
      emptyMessage='No users found for "{query}"'
    />
  );
}
