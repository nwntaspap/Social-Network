'use client';

import Image from 'next/image';
import { searchUsers } from '@/lib/api';
import { getDisplayName, getFileUrl } from '@/lib/helpers';
import type { User } from '@/lib/types';
import SearchDropdown from '@/components/ui/SearchDropdown';

export default function UserSearchBar() {
  return (
    <SearchDropdown<User>
      placeholder="Search users..."
      onSearch={(query) => searchUsers(query).then((res) => res.data)}
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
