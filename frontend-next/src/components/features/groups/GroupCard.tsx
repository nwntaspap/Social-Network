'use client';

import { useState } from 'react';
import Link from 'next/link';
import { requestToJoinGroup, leaveGroup } from '@/lib/api';
import { formatRelativeDate, truncateText } from '@/lib/helpers';
import type { Group, MembershipStatus } from '@/lib/types';

interface GroupCardProps {
  group: Group;
  onStatusChange?: (groupId: string, newStatus: MembershipStatus) => void;
}

export default function GroupCard({ group, onStatusChange }: GroupCardProps) {
  const [status, setStatus] = useState<MembershipStatus>(group.membershipStatus || 'none');
  const [isLoading, setIsLoading] = useState(false);

  const buttonConfig = {
    none: { label: 'Join', className: 'group-btn--join', disabled: false },
    pending: { label: 'Pending', className: 'group-btn--pending', disabled: true },
    member: { label: 'Leave', className: 'group-btn--leave', disabled: false },
  };

  const { label, className, disabled } = buttonConfig[status];

  async function handleJoin() {
    try {
      setIsLoading(true);
      await requestToJoinGroup(group.id);
      setStatus('pending');
      onStatusChange?.(group.id, 'pending');
    } catch (error) {
      console.error('Failed to request join:', error);
    } finally {
      setIsLoading(false);
    }
  }

  async function handleLeave() {
    try {
      setIsLoading(true);
      await leaveGroup(group.id);
      setStatus('none');
      onStatusChange?.(group.id, 'none');
    } catch (error) {
      console.error('Failed to leave group:', error);
    } finally {
      setIsLoading(false);
    }
  }

  const handleClick = status === 'member' ? handleLeave : handleJoin;

  return (
    <div className="group-card">
      <div className="group-card-header">
        <Link href={`/groups/${group.id}`} className="group-card-icon">
          {group.title[0].toUpperCase()}
        </Link>
        <div className="group-card-title-area">
          <Link href={`/groups/${group.id}`} className="group-card-title-link">
            <h3 className="group-card-title">{group.title}</h3>
          </Link>
          <span className="group-card-meta">
            {group.membersCount} members · {formatRelativeDate(group.createdAt)}
          </span>
        </div>
      </div>

      <p className="group-card-desc">{truncateText(group.description, 120)}</p>

      <div className="group-card-footer">
        <button
          className={`group-btn ${className}`}
          disabled={disabled || isLoading}
          onClick={handleClick}
        >
          {label}
        </button>
      </div>
    </div>
  );
}
