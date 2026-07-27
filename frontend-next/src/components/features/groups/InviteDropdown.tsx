'use client';

import { useEffect, useState, useCallback } from 'react';
import { useAuth } from '@/context/AuthContext';
import SearchDropdown from '@/components/ui/SearchDropdown';
// import { getGroupMembers, getPendingInvitations } from '@/lib/api';
import { searchResults, suggestedUsers } from '@/mocks/users';
import type { User } from '@/lib/types';
import Image from 'next/image';
import { getDisplayName, getFileUrl } from '@/lib/helpers';
import { mockGroupInvitations } from '@/mocks/group-invitations';
import { mockGroups } from '@/mocks/groups';

interface InviteDropdownProps {
  groupId: string;
  onClose: () => void;
}

export default function InviteDropdown({ groupId, onClose }: InviteDropdownProps) {
  const { user } = useAuth();
  const [memberIds, setMemberIds] = useState<Set<string>>(new Set());
  const [pendingInviteIds, setPendingInviteIds] = useState<Set<string>>(new Set());
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    async function fetchData() {
      try {
        // TODO: Replace with real API calls when backend is ready
        // const [membersResponse, invitations] = await Promise.all([
        //   getGroupMembers(groupId),
        //   getPendingInvitations(groupId)
        // ]);
        // setMemberIds(new Set(membersResponse.data.map(m => m.id)));
        // setPendingInviteIds(new Set(invitations.map(inv => inv.inviteeId)));

        // Mock data logic
        await new Promise((resolve) => setTimeout(resolve, 300)); // Simulate network delay

        const group = mockGroups.find((g) => g.id === groupId);
        if (!group) throw new Error('Group not found');

        // Mock members: creator + some suggested users based on group ID
        const mockMemberIds = new Set<string>([group.creatorId]);

        // Add some suggested users as members (different per group)
        if (groupId === '1') {
          mockMemberIds.add('2'); // Jane Doe
          mockMemberIds.add('3'); // Bob Smith
        } else if (groupId === '5') {
          mockMemberIds.add('2'); // Jane Doe
          mockMemberIds.add('4'); // Alice Johnson
        } else {
          // For other groups, add some random members
          mockMemberIds.add('2'); // Jane Doe
          mockMemberIds.add('6'); // Sarah Connor
        }

        setMemberIds(mockMemberIds);

        // Mock pending invitations
        const invitations = mockGroupInvitations[groupId] || [];
        setPendingInviteIds(new Set(invitations.map((inv) => inv.inviteeId)));
      } catch (err) {
        console.error('Failed to fetch group data:', err);
      } finally {
        setIsLoading(false);
      }
    }

    fetchData();
  }, [groupId]);

  const getButtonConfig = useCallback(
    (userId: string) => {
      if (memberIds.has(userId)) {
        return { label: 'Member', className: 'invite-btn--member', disabled: true };
      }
      if (pendingInviteIds.has(userId)) {
        return { label: 'Sent', className: 'invite-btn--pending', disabled: true };
      }
      return { label: 'Invite', className: 'invite-btn--invite', disabled: false };
    },
    [memberIds, pendingInviteIds]
  );

  async function handleInvite(userId: string) {
    try {
      // TODO: Replace with real API call when backend is ready
      // await inviteToGroup(groupId, userId);

      console.log('invite', userId, 'to group', groupId);

      // Optimistic update
      setPendingInviteIds((prev) => new Set([...prev, userId]));
    } catch (err) {
      console.error('Failed to invite user:', err);
    }
  }

  if (isLoading) {
    return (
      <div className="group-dropdown details-user">
        <div className="group-dropdown-header">
          <span>Invite User</span>
          <button className="group-dropdown-close" onClick={onClose}>
            ✕
          </button>
        </div>
        <div className="search-no-results">Loading...</div>
      </div>
    );
  }

  // Combine ALL users for search: suggested users + search results
  const allUsers = [...suggestedUsers, ...searchResults];

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
        items={allUsers}
        filterFn={(searchUser, q) =>
          searchUser.id !== user?.id &&
          (searchUser.username.toLowerCase().includes(q.toLowerCase()) ||
            searchUser.firstName.toLowerCase().includes(q.toLowerCase()) ||
            searchUser.lastName.toLowerCase().includes(q.toLowerCase()))
        }
        renderItem={(searchUser) => {
          const { label, className, disabled } = getButtonConfig(searchUser.id);

          return (
            <>
              <Image
                src={getFileUrl(searchUser.avatarUrl)}
                alt={getDisplayName(searchUser)}
                width={36}
                height={36}
                className="search-result-avatar"
              />
              <div className="search-result-info">
                <span className="search-result-name">{getDisplayName(searchUser)}</span>
                <span className="search-result-username">@{searchUser.username}</span>
              </div>
              <button
                className={`invite-btn ${className}`}
                disabled={disabled}
                onClick={(e) => {
                  e.preventDefault();
                  e.stopPropagation();
                  if (!disabled) {
                    handleInvite(searchUser.id);
                  }
                }}
              >
                {label}
              </button>
            </>
          );
        }}
        getItemKey={(searchUser) => searchUser.id}
        getItemHref={() => '#'}
        emptyMessage="No users found"
      />
    </div>
  );
}
