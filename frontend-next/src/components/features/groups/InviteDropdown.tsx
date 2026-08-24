'use client';

import { useEffect, useState, useCallback } from 'react';
import { useAuth } from '@/context/AuthContext';
import { getGroupMembers, getFollowers, inviteToGroup } from '@/lib/api';
import type { User } from '@/lib/types';
import Image from 'next/image';
import { getDisplayName, getFileUrl } from '@/lib/helpers';

interface InviteDropdownProps {
  groupId: string;
  onClose: () => void;
}

/**
 * Invite flow: only your followers can be invited into a group.
 * The backend enforces "invitee must be a follower of the inviter",
 * so the candidate list is the current user's followers.
 */
export default function InviteDropdown({ groupId, onClose }: InviteDropdownProps) {
  const { user } = useAuth();
  const [candidates, setCandidates] = useState<User[]>([]);
  const [memberIds, setMemberIds] = useState<Set<string>>(new Set());
  const [pendingInviteIds, setPendingInviteIds] = useState<Set<string>>(new Set());
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    let ignore = false;
    async function fetchData() {
      try {
        const [membersResponse, followers] = await Promise.all([
          getGroupMembers(groupId),
          user ? getFollowers(user.id) : Promise.resolve<User[]>([]),
        ]);
        if (ignore) return;
        setMemberIds(new Set(membersResponse.data.map((m) => m.userId)));
        const memberSet = new Set(membersResponse.data.map((m) => m.userId));
        setCandidates(followers.filter((f) => f.id !== user?.id && !memberSet.has(f.id)));
      } catch (err) {
        console.error('Failed to fetch invite candidates:', err);
      } finally {
        if (!ignore) setIsLoading(false);
      }
    }

    fetchData();
    return () => {
      ignore = true;
    };
  }, [groupId, user]);

  const getButtonConfig = useCallback(
    (userId: string) => {
      if (pendingInviteIds.has(userId)) {
        return { label: 'Sent', className: 'invite-btn--pending', disabled: true };
      }
      return { label: 'Invite', className: 'invite-btn--invite', disabled: false };
    },
    [pendingInviteIds]
  );

  async function handleInvite(userId: string) {
    try {
      await inviteToGroup(groupId, userId);
      setPendingInviteIds((prev) => new Set([...prev, userId]));
    } catch (err) {
      console.error('Failed to invite user:', err);
    }
  }

  return (
    <div className="group-dropdown details-user">
      <div className="group-dropdown-header">
        <span>Invite Followers</span>
        <button className="group-dropdown-close" onClick={onClose}>
          ✕
        </button>
      </div>

      {isLoading ? (
        <div className="search-no-results">Loading...</div>
      ) : candidates.length === 0 ? (
        <div className="search-no-results">
          No followers to invite. Only your followers can be invited.
        </div>
      ) : (
        <ul className="group-member-list">
          {candidates.map((candidate) => {
            const { label, className, disabled } = getButtonConfig(candidate.id);
            return (
              <li key={candidate.id} className="group-member-row">
                <Image
                  src={getFileUrl(candidate.avatarUrl)}
                  alt={getDisplayName(candidate)}
                  width={36}
                  height={36}
                  className="search-result-avatar"
                />
                <div className="search-result-info">
                  <span className="search-result-name">{getDisplayName(candidate)}</span>
                  <span className="search-result-username">@{candidate.username}</span>
                </div>
                <button
                  className={`invite-btn ${className}`}
                  disabled={disabled}
                  onClick={(e) => {
                    e.preventDefault();
                    e.stopPropagation();
                    if (!disabled) handleInvite(candidate.id);
                  }}
                >
                  {label}
                </button>
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
