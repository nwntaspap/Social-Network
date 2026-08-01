'use client';

import { useState } from 'react';
import Link from 'next/link';
import { requestToJoinGroup, leaveGroup } from '@/lib/api';
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
    } catch (error) {
      console.error('Failed to leave group:', error);
    } finally {
      setIsLoading(false);
    }
  }

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
        {!isCreator && (
          <button
            className={`group-btn ${className}`}
            disabled={disabled || isLoading}
            onClick={status === 'member' ? handleLeave : handleJoin}
          >
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
