'use client';

import { useEffect, useState } from 'react';
import Image from 'next/image';
import Link from 'next/link';
import { getPendingFollowRequests, handleFollowRequest } from '@/lib/api';
import { getDisplayName, getFileUrl } from '@/lib/helpers';
import type { FollowRequest } from '@/lib/types';

export default function FollowRequestsSection() {
  const [requests, setRequests] = useState<FollowRequest[]>([]);
  const [loading, setLoading] = useState(true);
  const [busyId, setBusyId] = useState<string | null>(null);
  const [error, setError] = useState('');

  useEffect(() => {
    let cancelled = false;
    getPendingFollowRequests()
      .then((data) => {
        if (!cancelled) setRequests(data);
      })
      .catch(() => {
        if (!cancelled) setError('Failed to load follow requests.');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  async function handleResponse(request: FollowRequest, action: 'accept' | 'decline') {
    setBusyId(request.id);
    setError('');
    try {
      await handleFollowRequest(request.requesterId, action);
      setRequests((prev) => prev.filter((item) => item.id !== request.id));
    } catch {
      setError('Failed to update the request. Please try again.');
    } finally {
      setBusyId(null);
    }
  }

  if (loading || (requests.length === 0 && !error)) {
    return null;
  }

  return (
    <section className="follow-requests">
      <h2 className="follow-requests-title">Follow Requests</h2>
      {error && <p className="follow-requests-error">{error}</p>}
      {requests.map((request) => (
        <div key={request.id} className="follow-request-row">
          <Link href={`/profile/${request.requesterId}`} className="follow-request-user">
            <Image
              src={getFileUrl(request.requester?.avatarUrl)}
              alt={getDisplayName(request.requester)}
              width={40}
              height={40}
              className="follow-request-avatar"
            />
            <div className="follow-request-info">
              <span className="follow-request-name">{getDisplayName(request.requester)}</span>
              <span className="follow-request-username">
                @{request.requester?.username || request.requester?.nickname}
              </span>
            </div>
          </Link>
          <div className="follow-request-actions">
            <button
              type="button"
              className="follow-request-accept"
              disabled={busyId === request.id}
              onClick={() => handleResponse(request, 'accept')}
            >
              Accept
            </button>
            <button
              type="button"
              className="follow-request-decline"
              disabled={busyId === request.id}
              onClick={() => handleResponse(request, 'decline')}
            >
              Decline
            </button>
          </div>
        </div>
      ))}
    </section>
  );
}
