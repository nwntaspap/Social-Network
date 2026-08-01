'use client';

import { useEffect, useState } from 'react';
import { getPendingJoinRequests, handleJoinRequest } from '@/lib/api';
import { getDisplayName } from '@/lib/helpers';
import type { GroupJoinRequest } from '@/lib/types';

interface GroupSidebarProps {
  groupId: string;
  isCreator: boolean;
}

export default function GroupSidebar({ groupId, isCreator }: GroupSidebarProps) {
  const [showPending, setShowPending] = useState(false);
  const [requests, setRequests] = useState<GroupJoinRequest[]>([]);

  async function handleResponse(requestId: string, action: 'accept' | 'decline') {
    try {
      await handleJoinRequest(requestId, action);
      setRequests((prev) => prev.filter((req) => req.id !== requestId));
    } catch (err) {
      console.error('Failed to respond to request:', err);
    }
  }

  useEffect(() => {
    if (!showPending || !isCreator) return;
    let ignore = false;
    getPendingJoinRequests(groupId)
      .then((pending) => {
        if (!ignore) setRequests(pending);
      })
      .catch(() => {
        if (!ignore) console.error('Failed to load pending requests.');
      });
    return () => {
      ignore = true;
    };
  }, [showPending, isCreator, groupId]);

  if (!isCreator) return null;

  return (
    <aside className="group-sidebar">
      <button className="group-sidebar-btn" onClick={() => setShowPending(!showPending)}>
        Pending Requests
      </button>

      {showPending && (
        <div className="group-pending-dropdown">
          {requests.length > 0 ? (
            requests.map((req) => (
              <div key={req.id} className="group-pending-item">
                <span className="group-pending-name">{getDisplayName(req.requester)}</span>
                <div className="group-pending-actions">
                  <button
                    className="group-pending-accept"
                    onClick={() => handleResponse(req.id, 'accept')}
                  >
                    Accept
                  </button>
                  <button
                    className="group-pending-decline"
                    onClick={() => handleResponse(req.id, 'decline')}
                  >
                    Decline
                  </button>
                </div>
              </div>
            ))
          ) : (
            <p className="group-pending-empty">No pending requests</p>
          )}
        </div>
      )}
    </aside>
  );
}
