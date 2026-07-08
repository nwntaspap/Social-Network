'use client';

/**
 * components/features/groups/InviteDropdown.tsx
 *
 * Dropdown to search users and invite them to the group.
 * Uses the generic SearchDropdown for user search.
 *
 * TODO: Wire to inviteToGroup API when backend is ready.
 */

import SearchDropdown from '@/components/ui/SearchDropdown';
import { searchResults as mockUsers } from '@/mocks/users';
import type { User } from '@/lib/types';
import Image from 'next/image';
import { getDisplayName, getFileUrl } from '@/lib/helpers';

interface InviteDropdownProps {
  groupId: string;
  onClose: () => void;
}

export default function InviteDropdown({ groupId, onClose }: InviteDropdownProps) {
  function handleInvite(user: User) {
    // TODO: await inviteToGroup(groupId, user.id);
    console.log('invite', user.id, 'to group', groupId);
  }

  return (
    <div className="group-dropdown details-user">
      <div className="group-dropdown-header">
        <span>Invite User</span>
        <button className="group-dropdown-close" onClick={onClose}>
          ✕
        </button>
      </div>

      <SearchDropdown<User>
        placeholder="Search users to invite..."
        items={mockUsers}
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
              width={36}
              height={36}
              className="search-result-avatar"
            />
            <div className="search-result-info">
              <span className="search-result-name">{getDisplayName(user)}</span>
              <span className="search-result-username">@{user.username}</span>
            </div>
            <button
              className="invite-btn"
              onClick={(e) => {
                e.preventDefault();
                handleInvite(user);
              }}
            >
              Invite
            </button>
          </>
        )}
        getItemKey={(user) => user.id}
        getItemHref={() => '#'} // We handle click with the button instead
        emptyMessage="No users found"
      />
    </div>
  );
}
