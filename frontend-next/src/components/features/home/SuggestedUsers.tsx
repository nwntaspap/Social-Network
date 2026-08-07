'use client';

import { useEffect, useState } from 'react';
import Image from 'next/image';
import Link from 'next/link';
import { searchUsers, sendFollowRequest } from '@/lib/api';
import { useAuth } from '@/context/AuthContext';
import { getDisplayName, getFileUrl, truncateText } from '@/lib/helpers';
import type { User } from '@/lib/types';

export default function SuggestedUsers() {
  const { user } = useAuth();
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function load() {
      try {
        const response = await searchUsers('', 1);
        setUsers(response.data.filter((u) => u.id !== user?.id).slice(0, 6));
      } catch (err) {
        console.error('Failed to load suggested users:', err);
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [user?.id]);

  if (loading) {
    return (
      <section className="suggested-section">
        <h2 className="section-title">Suggested Users</h2>
        <p className="suggested-empty">Loading...</p>
      </section>
    );
  }

  return (
    <section className="suggested-section">
      <h2 className="section-title">Suggested Users</h2>
      <div className="suggested-users-grid">
        {users.length > 0 ? (
          users.map((u) => <SuggestedUserCard key={u.id} user={u} />)
        ) : (
          <p className="suggested-empty">No suggested users found.</p>
        )}
      </div>
    </section>
  );
}

function SuggestedUserCard({ user }: { user: User }) {
  const [status, setStatus] = useState<'none' | 'pending' | 'following'>('none');
  const [busy, setBusy] = useState(false);

  async function handleFollow() {
    setBusy(true);
    try {
      const response = await sendFollowRequest(user.id);
      setStatus(response.status === 'pending' ? 'pending' : 'following');
    } catch (err) {
      console.error('Failed to send follow request:', err);
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="suggested-user-card">
      <Link href={`/profile/${user.id}`} className="suggested-user-link">
        <Image
          src={getFileUrl(user.avatarUrl)}
          alt={getDisplayName(user)}
          width={40}
          height={40}
          className="suggested-user-avatar"
        />
        <div className="suggested-user-info">
          <span className="suggested-user-name">{getDisplayName(user)}</span>
          <span className="suggested-user-username">@{user.username}</span>
          {user.aboutMe && <p className="suggested-user-bio">{truncateText(user.aboutMe, 60)}</p>}
        </div>
      </Link>
      <button className="follow-btn" onClick={handleFollow} disabled={busy || status !== 'none'}>
        {status === 'following' ? 'Following' : status === 'pending' ? 'Requested' : 'Follow'}
      </button>
    </div>
  );
}
