'use client';

/**
 * components/features/home/SuggestedUsers.tsx
 *
 * Suggested users grid with follow button.
 * Extracted from app/page.tsx HomeContent.
 *
 * When the backend is ready, replace mock data with:
 *   const users = await getSuggestedUsers();
 * and wire the follow button to sendFollowRequest(user.id).
 */

import Image from 'next/image';
import Link from 'next/link';
import { getDisplayName, getFileUrl, truncateText } from '@/lib/helpers';
import { suggestedUsers } from '@/mocks/users';
import type { User } from '@/lib/types';

export default function SuggestedUsers() {
  return (
    <section className="suggested-section">
      <h2 className="section-title">Suggested Users</h2>
      <div className="suggested-users-grid">
        {suggestedUsers.map((user) => (
          <SuggestedUserCard key={user.id} user={user} />
        ))}
      </div>
    </section>
  );
}

function SuggestedUserCard({ user }: { user: User }) {
  // TODO: wire to sendFollowRequest(user.id) — optimistic toggle like PostCard likes
  function handleFollow() {
    console.log('follow', user.id);
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
      <button className="follow-btn" onClick={handleFollow}>
        Follow
      </button>
    </div>
  );
}
