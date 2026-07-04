'use client';

/**
 * components/features/groups/GroupCard.tsx
 *
 * Single group card with:
 *   - group name (link to /groups/:id)
 *   - description
 *   - member count
 *   - join / pending / leave button
 *
 * Button states:
 *   Join     → not a member, not requested
 *   Pending  → request sent, waiting for approval (disabled)
 *   Leave    → already a member
 */

import { useState } from 'react';
import Link from 'next/link';
import { formatRelativeDate, truncateText } from '@/lib/helpers';
import type { Group } from '@/lib/types';

type MembershipState = 'join' | 'pending' | 'leave';

export default function GroupCard({ group }: { group: Group }) {
  // TODO: determine initial state from backend (user's membership status)
  const [membership, setMembership] = useState<MembershipState>('join');
  const [loading, setLoading] = useState(false);

  function handleClick() {
    if (membership === 'leave') {
      handleLeave();
    } else if (membership === 'join') {
      handleJoin();
    }
  }

  // TODO: replace with real API calls when backend is ready
  function handleJoin() {
    setLoading(true);
    // Simulate API call
    setTimeout(() => {
      setMembership('pending');
      setLoading(false);
    }, 300);
  }

  function handleLeave() {
    setLoading(true);
    setTimeout(() => {
      setMembership('join');
      setLoading(false);
    }, 300);
  }

  const buttonLabel =
    membership === 'leave' ? 'Leave' : membership === 'pending' ? 'Pending' : 'Join';

  const isDisabled = membership === 'pending' || loading;

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
          className={`group-btn group-btn--${membership}`}
          onClick={handleClick}
          disabled={isDisabled}
        >
          {loading ? '...' : buttonLabel}
        </button>
      </div>
    </div>
  );
}
