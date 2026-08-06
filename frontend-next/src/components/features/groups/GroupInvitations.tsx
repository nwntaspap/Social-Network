'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { getMyGroupInvitations, respondToGroupInvitation } from '@/lib/api';
import { formatRelativeDate } from '@/lib/helpers';
import type { GroupInvitation } from '@/lib/types';

export default function GroupInvitations() {
  const [invitations, setInvitations] = useState<GroupInvitation[]>([]);
  const [loading, setLoading] = useState(true);
  const [busyId, setBusyId] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    getMyGroupInvitations()
      .then((data) => {
        if (!cancelled) setInvitations(data);
      })
      .catch((error) => console.error('Failed to load invitations:', error))
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  async function handleResponse(invitation: GroupInvitation, action: 'accept' | 'decline') {
    setBusyId(invitation.id);
    try {
      await respondToGroupInvitation(invitation.groupId, action);
      setInvitations((prev) => prev.filter((item) => item.id !== invitation.id));
    } catch (error) {
      console.error('Failed to respond to invitation:', error);
    } finally {
      setBusyId(null);
    }
  }

  if (loading || invitations.length === 0) {
    return null;
  }

  return (
    <section className="group-invitations">
      <h2 className="group-invitations-title">Group Invitations</h2>
      {invitations.map((invitation) => (
        <div key={invitation.id} className="group-card">
          <div className="group-card-header">
            <Link href={`/groups/${invitation.groupId}`} className="group-card-icon">
              {invitation.group.title[0].toUpperCase()}
            </Link>
            <div className="group-card-title-area">
              <Link href={`/groups/${invitation.groupId}`} className="group-card-title-link">
                <h3 className="group-card-title">{invitation.group.title}</h3>
              </Link>
              <span className="group-card-meta">
                Invited by {invitation.inviter.username}
                {' · '}
                {formatRelativeDate(invitation.createdAt)}
              </span>
            </div>
          </div>
          <div className="group-card-footer group-invitation-actions">
            <button
              className="group-btn group-btn--join"
              disabled={busyId === invitation.id}
              onClick={() => handleResponse(invitation, 'accept')}
            >
              Accept
            </button>
            <button
              className="group-btn group-btn--leave"
              disabled={busyId === invitation.id}
              onClick={() => handleResponse(invitation, 'decline')}
            >
              Decline
            </button>
          </div>
        </div>
      ))}
    </section>
  );
}
