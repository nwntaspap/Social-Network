'use client';

import { useEffect, useState, useCallback } from 'react';
import { useAuth } from '@/context/AuthContext';
import SearchDropdown from '@/components/ui/SearchDropdown';
import { getGroupMembers, inviteToGroup, searchUsers } from '@/lib/api';
import type { User } from '@/lib/types';
import Image from 'next/image';
import { getDisplayName, getFileUrl } from '@/lib/helpers';

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
        const membersResponse = await getGroupMembers(groupId);
        setMemberIds(new Set(membersResponse.data.map((m) => m.userId)));
      } catch (err) {
        console.error('Failed to fetch group members:', err);
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
      await inviteToGroup(groupId, userId);
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
        onSearch={(query) => searchUsers(query).then((res) => res.data)}
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
                  if (!disabled && searchUser.id !== user?.id) {
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
