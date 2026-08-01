'use client';

import { browseGroups } from '@/lib/api';
import type { Group } from '@/lib/types';
import SearchDropdown from '@/components/ui/SearchDropdown';

export default function GroupSearchBar() {
  return (
    <SearchDropdown<Group>
      placeholder="Search groups..."
      onSearch={(query) => browseGroups(query).then((res) => res.data)}
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
