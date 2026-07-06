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

import Link from 'next/link';
import { formatRelativeDate, truncateText } from '@/lib/helpers';
import type { Group, MembershipStatus } from '@/lib/types';

export default function GroupCard({ group }: { group: Group }) {
  // TODO: Wire to real API when backend is ready
  // const [status, setStatus] = useState(group.membershipStatus);
  //
  // async function handleJoin() {
  //   await requestToJoinGroup(group.id);
  //   setStatus('pending');
  // }
  //
  // async function handleLeave() {
  //   await leaveGroup(group.id);
  //   setStatus('none');
  // }

  const status: MembershipStatus = group.membershipStatus || 'none';

  const buttonConfig = {
    none: { label: 'Join', className: 'group-btn--join', disabled: false },
    pending: { label: 'Pending', className: 'group-btn--pending', disabled: true },
    member: { label: 'Leave', className: 'group-btn--leave', disabled: false },
  };

  const { label, className, disabled } = buttonConfig[status];

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
        {/* TODO: Add onClick handler when backend is ready */}
        <button className={`group-btn ${className}`} disabled={disabled}>
          {label}
        </button>
      </div>
    </div>
  );
}
