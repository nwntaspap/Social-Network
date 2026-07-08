'use client';

import { useState } from 'react';
import Link from 'next/link';
import { formatRelativeDate } from '@/lib/helpers';
import type { Group, MembershipStatus } from '@/lib/types';
import InviteDropdown from './InviteDropdown';
import CreateEventForm from './CreateEventForm';

interface GroupHeaderProps {
  group: Group;
  isCreator: boolean;
  isMember: boolean;
}

export default function GroupHeader({ group, isCreator, isMember }: GroupHeaderProps) {
  const [showInvite, setShowInvite] = useState(false);
  const [showEventForm, setShowEventForm] = useState(false);
  // TODO: Wire to real API when backend is ready
  // const [status, setStatus] = useState<MembershipStatus>(group.membershipStatus || 'none');
  // const [isLoading, setIsLoading] = useState(false);

  const status: MembershipStatus = group.membershipStatus || 'none';

  const buttonConfig = {
    none: { label: 'Join', className: 'group-btn--join', disabled: false },
    pending: { label: 'Pending', className: 'group-btn--pending', disabled: true },
    member: { label: 'Leave', className: 'group-btn--leave', disabled: false },
  };

  const { label, className, disabled } = buttonConfig[status];

  return (
    <div className="group-detail-header">
      <div className="group-detail-header-top">
        <div className="group-detail-icon">{group.title[0].toUpperCase()}</div>
        <div className="group-detail-info">
          <h1 className="group-detail-title">{group.title}</h1>
          <p className="group-detail-meta">
            {group.membersCount} members · Created {formatRelativeDate(group.createdAt)}
          </p>
          <p className="group-detail-desc">{group.description}</p>
        </div>
      </div>

      <div className="group-detail-actions">
        {!isMember && (
          <button className={`group-action-btn ${className}`} disabled={disabled}>
            {label}
          </button>
        )}

        {isMember && (
          <>
            <Link href={`/create?groupId=${group.id}`} className="group-action-btn">
              Create Post
            </Link>

            <div className="group-action-wrapper">
              <button
                className="group-action-btn"
                onClick={() => {
                  setShowInvite(!showInvite);
                  setShowEventForm(false);
                }}
              >
                Invite User
              </button>
              {showInvite && (
                <InviteDropdown groupId={group.id} onClose={() => setShowInvite(false)} />
              )}
            </div>

            <div className="group-action-wrapper">
              <button
                className="group-action-btn"
                onClick={() => {
                  setShowEventForm(!showEventForm);
                  setShowInvite(false);
                }}
              >
                Create Event
              </button>
              {showEventForm && (
                <CreateEventForm groupId={group.id} onClose={() => setShowEventForm(false)} />
              )}
            </div>
          </>
        )}

        {isCreator && (
          <>
            <button className="group-action-btn group-action-btn--danger">Edit Group</button>
            <button className="group-action-btn group-action-btn--delete">Delete Group</button>
          </>
        )}
      </div>
    </div>
  );
}
