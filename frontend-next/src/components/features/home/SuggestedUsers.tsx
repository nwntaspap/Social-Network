'use client';

import { useEffect, useState } from 'react';
import Image from 'next/image';
import Link from 'next/link';
import { getSuggestedUsers, sendFollowRequest } from '@/lib/api';
import { useAuth } from '@/context/AuthContext';
import { getDisplayName, getFileUrl, truncateText } from '@/lib/helpers';
import { chatSocket } from '@/lib/ws';
import type { IsOnlineStatusPayload, User } from '@/lib/types';

const MAX_SUGGESTIONS = 6;

export default function SuggestedUsers() {
  const { user } = useAuth();
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    chatSocket.connect();

    const load = () => {
      if (!user?.id) return;
      getSuggestedUsers(user.id)
        .then((response) => {
          setUsers(response.data.filter((u) => u.id !== user?.id).slice(0, MAX_SUGGESTIONS));
        })
        .catch((err) => {
          console.error('Failed to load suggested users:', err);
        })
        .finally(() => {
          setLoading(false);
        });
    };

    // Fetch the snapshot once connected. If a reconnect happens later, missed
    // isOnlineStatus.update broadcasts are recovered by refetching the snapshot.
    if (chatSocket.isConnected()) {
      load();
    }
    const unsubscribeConnection = chatSocket.onConnection(() => {
      if (chatSocket.isConnected()) load();
    });

    const unsubscribe = chatSocket.on('isOnlineStatus.update', (payload) => {
      const p = payload as IsOnlineStatusPayload;
      if (!p || typeof p.user_id !== 'string') return;
      setUsers((prev) =>
        prev.map((u) => (u.id === p.user_id ? { ...u, isOnline: p.isOnline } : u))
      );
    });

    return () => {
      unsubscribeConnection();
      unsubscribe();
      chatSocket.disconnect();
    };
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
      <div className="suggested-users-list">
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
  const online = user.isOnline === true;

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
      <div className="suggested-user-status">
        <span className={`suggested-user-status-dot ${online ? 'online' : 'offline'}`} />
        <span className="suggested-user-status-label">{online ? 'Online' : 'Offline'}</span>
      </div>
      <button className="follow-btn" onClick={handleFollow} disabled={busy || status !== 'none'}>
        {status === 'following' ? 'Following' : status === 'pending' ? 'Requested' : 'Follow'}
      </button>
    </div>
  );
}
