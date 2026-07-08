'use client';

/**
 * components/features/groups/GroupSearchBar.tsx
 *
 * Group search bar built on the generic SearchDropdown.
 *
 * When the backend is ready, replace mock data with the onSearch callback.
 */

import { searchResults as mockGroups } from '@/mocks/groups';
import type { Group } from '@/lib/types';
import SearchDropdown from '@/components/ui/SearchDropdown';

export default function GroupSearchBar() {
  return (
    <SearchDropdown<Group>
      placeholder="Search groups..."
      items={mockGroups}
      filterFn={(group, q) =>
        group.title.toLowerCase().includes(q.toLowerCase()) ||
        group.description.toLowerCase().includes(q.toLowerCase())
      }
      renderItem={(group) => (
        <>
          <div className="search-result-avatar group-icon">{group.title[0].toUpperCase()}</div>
          <div className="search-result-info">
            <span className="search-result-name">{group.title}</span>
            <span className="search-result-username">{group.membersCount} members</span>
          </div>
        </>
      )}
      getItemKey={(group) => group.id}
      getItemHref={(group) => `/groups/${group.id}`}
      emptyMessage='No groups found for "{query}"'
    />
  );
}
