'use client';

import { useEffect, useState } from 'react';
import Image from 'next/image';
import Link from 'next/link';
import { getFollowers, getFollowing } from '@/lib/api';
import { getDisplayName, getFileUrl } from '@/lib/helpers';
import type { User } from '@/lib/types';

interface FollowersFollowingModalProps {
  userId: string;
  mode: 'followers' | 'following';
  onClose: () => void;
}

export default function FollowersFollowingModal({
  userId,
  mode,
  onClose,
}: FollowersFollowingModalProps) {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    let cancelled = false;
    const fetchUsers = mode === 'followers' ? getFollowers : getFollowing;
    fetchUsers(userId)
      .then((data) => {
        if (!cancelled) setUsers(data);
      })
      .catch(() => {
        if (!cancelled) setError(`Failed to load ${mode}.`);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [userId, mode]);

  return (
    <div className="followers-modal-backdrop" onClick={onClose}>
      <div className="followers-modal" onClick={(e) => e.stopPropagation()}>
        <div className="followers-modal-header">
          <h2>{mode === 'followers' ? 'Followers' : 'Following'}</h2>
          <button
            type="button"
            className="followers-modal-close"
            onClick={onClose}
            aria-label="Close"
          >
            ×
          </button>
        </div>
        <div className="followers-modal-content">
          {loading && <p className="followers-modal-empty">Loading...</p>}
          {error && <p className="followers-modal-empty">{error}</p>}
          {!loading && !error && users.length === 0 && (
            <p className="followers-modal-empty">
              {mode === 'followers' ? 'No followers yet.' : 'Not following anyone yet.'}
            </p>
          )}
          {!loading &&
            !error &&
            users.map((user) => (
              <Link key={user.id} href={`/profile/${user.id}`} className="followers-modal-row">
                <Image
                  src={getFileUrl(user.avatarUrl)}
                  alt={getDisplayName(user)}
                  width={40}
                  height={40}
                  className="followers-modal-avatar"
                />
                <div className="followers-modal-info">
                  <span className="followers-modal-name">{getDisplayName(user)}</span>
                  <span className="followers-modal-username">
                    @{user.username || user.nickname}
                  </span>
                </div>
              </Link>
            ))}
        </div>
      </div>
    </div>
  );
}
