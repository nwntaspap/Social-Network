'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import Image from 'next/image';
import { useRouter } from 'next/navigation';
import { getUserProfile, getUserPosts, sendFollowRequest, unfollowUser } from '@/lib/api';
import { getDisplayName, getFileUrl } from '@/lib/helpers';
import { useAuth } from '@/context/AuthContext';
import PostCard from '@/components/features/home/PostCard';
import FollowRequestsSection from './FollowRequestsSection';
import FollowersFollowingModal from './FollowersFollowingModal';
import type { Post, Profile } from '@/lib/types';
export default function ProfileContent() {
  const { id } = useParams<{ id: string }>();
  const { user: currentUser } = useAuth();
  const router = useRouter();

  const [profile, setProfile] = useState<Profile | null>(null);
  const [posts, setPosts] = useState<Post[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [followState, setFollowState] = useState<'none' | 'following' | 'pending'>('none');
  const [followBusy, setFollowBusy] = useState(false);
  const [statModal, setStatModal] = useState<'followers' | 'following' | null>(null);

  const isOwnProfile = currentUser?.id === id;

  useEffect(() => {
    let ignore = false;
    Promise.all([getUserProfile(id), getUserPosts(id, 1, 10)])
      .then(([profileData, postsData]) => {
        if (ignore) return;
        setProfile(profileData);
        setPosts(postsData.data ?? []);
        setFollowState(profileData.isFollowing ? 'following' : 'none');
      })
      .catch(() => {
        if (!ignore) setError('Profile not found.');
      })
      .finally(() => {
        if (!ignore) setLoading(false);
      });
    return () => {
      ignore = true;
    };
  }, [id]);

  async function handleFollow() {
    if (!profile || followBusy) return;
    setFollowBusy(true);
    try {
      if (followState === 'following') {
        await unfollowUser(profile.id);
        setFollowState('none');
      } else if (followState === 'none') {
        const response = await sendFollowRequest(profile.id);
        setFollowState(response.status === 'pending' ? 'pending' : 'following');
      }
    } catch (err) {
      console.error('Follow action failed:', err);
    } finally {
      setFollowBusy(false);
    }
  }

  if (loading) return <p className="activity-page-title">Loading...</p>;
  if (error || !profile) return <p className="activity-page-title">{error}</p>;

  return (
    <div className="activity-container">
      <div className="profile-card">
        <div className="profile-card-top">
          <div className="profile-avatar-shell">
            <div className="profile-avatar">
              <Image
                src={getFileUrl(profile.avatarUrl)}
                alt={getDisplayName(profile)}
                width={300}
                height={300}
                onError={(e) => {
                  (e.target as HTMLImageElement).src = '/images/user-avatar.png';
                }}
              />
            </div>
          </div>

          <div className="profile-main">
            <h1 className="profile-name">{getDisplayName(profile)}</h1>
            <p className="profile-bio-text">@{profile.username || profile.nickname}</p>

            <div className="profile-stats">
              {!isOwnProfile && profile.isPublic === false && !profile.isFollowing ? (
                <>
                  <div className="profile-stat">
                    <span className="profile-stat-value">{profile.followersCount ?? 0}</span>
                    <span className="profile-stat-label">Followers</span>
                  </div>
                  <div className="profile-stat">
                    <span className="profile-stat-value">{profile.followingCount ?? 0}</span>
                    <span className="profile-stat-label">Following</span>
                  </div>
                </>
              ) : (
                <>
                  <button
                    type="button"
                    className="profile-stat"
                    onClick={() => setStatModal('followers')}
                  >
                    <span className="profile-stat-value">{profile.followersCount ?? 0}</span>
                    <span className="profile-stat-label">Followers</span>
                  </button>
                  <button
                    type="button"
                    className="profile-stat"
                    onClick={() => setStatModal('following')}
                  >
                    <span className="profile-stat-value">{profile.followingCount ?? 0}</span>
                    <span className="profile-stat-label">Following</span>
                  </button>
                </>
              )}
            </div>

            {profile.aboutMe && (
              <div className="profile-bio">
                <p className="profile-bio-text">{profile.aboutMe}</p>
              </div>
            )}

            {!isOwnProfile && (
              <button
                type="button"
                className="group-action-btn"
                onClick={handleFollow}
                disabled={followBusy}
              >
                {followState === 'following'
                  ? 'Unfollow'
                  : followState === 'pending'
                    ? 'Request Pending'
                    : 'Follow'}
              </button>
            )}

            {!isOwnProfile && (
              <button
                type="button"
                className="group-action-btn"
                onClick={() => router.push(`/chat?chat=${profile.id}`)}
              >
                Message
              </button>
            )}
          </div>
        </div>
      </div>

      {isOwnProfile && <FollowRequestsSection />}

      <h2 className="activity-page-title">Posts</h2>
      {posts.length === 0 ? (
        <p className="activity-section">No posts yet.</p>
      ) : (
        posts.map((post) => <PostCard key={post.id} post={post} />)
      )}

      {statModal && (
        <FollowersFollowingModal userId={id} mode={statModal} onClose={() => setStatModal(null)} />
      )}
    </div>
  );
}
