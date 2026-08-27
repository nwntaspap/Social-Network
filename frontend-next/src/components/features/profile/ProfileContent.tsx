'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import Image from 'next/image';
import {
  getUserProfile,
  getUserPosts,
  sendFollowRequest,
  unfollowUser,
  toggleProfilePrivacy,
  updateProfile,
} from '@/lib/api';
import { getDisplayName, getFileUrl } from '@/lib/helpers';
import { openChatWithUser } from '@/lib/chatWidget';
import { useAuth } from '@/context/AuthContext';
import PostCard from '@/components/features/home/PostCard';
import FollowRequestsSection from './FollowRequestsSection';
import FollowersFollowingModal from './FollowersFollowingModal';
import type { Post, Profile } from '@/lib/types';

const GENDER_LABELS: Record<string, string> = {
  male: 'Male',
  female: 'Female',
  other: 'Other',
  prefer_not_to_say: 'Prefer not to say',
};

// Zero-valued Go time.Time serializes as "0001-01-01" — treat it as absent
// (locked profiles never receive a real date of birth).
function hasDateOfBirth(user?: Partial<Profile> | null): boolean {
  return Boolean(user?.dateOfBirth && !user.dateOfBirth.startsWith('0001'));
}

export default function ProfileContent() {
  const { id } = useParams<{ id: string }>();
  const { user: currentUser } = useAuth();

  const [profile, setProfile] = useState<Profile | null>(null);
  const [posts, setPosts] = useState<Post[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [followState, setFollowState] = useState<'none' | 'following' | 'pending'>('none');
  const [followBusy, setFollowBusy] = useState(false);
  const [statModal, setStatModal] = useState<'followers' | 'following' | null>(null);
  const [privacyBusy, setPrivacyBusy] = useState(false);

  // Posts pagination
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [loadingMore, setLoadingMore] = useState(false);

  // About Me editor (own profile only)
  const [editingAbout, setEditingAbout] = useState(false);
  const [aboutDraft, setAboutDraft] = useState('');
  const [dobDraft, setDobDraft] = useState('');
  const [genderDraft, setGenderDraft] = useState('');
  const [savingAbout, setSavingAbout] = useState(false);

  const isOwnProfile = currentUser?.id === id;

  useEffect(() => {
    let ignore = false;
    Promise.all([getUserProfile(id), getUserPosts(id, 1, 10)])
      .then(([profileData, postsData]) => {
        if (ignore) return;
        setProfile(profileData);
        setPosts(postsData.data ?? []);
        setTotalPages(postsData.totalPages ?? 1);
        setPage(1);
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

  async function handleLoadMore() {
    if (loadingMore) return;
    setLoadingMore(true);
    try {
      const next = page + 1;
      const postsData = await getUserPosts(id, next, 10);
      setPosts((prev) => [...prev, ...(postsData.data ?? [])]);
      setTotalPages(postsData.totalPages ?? next);
      setPage(next);
    } catch (err) {
      console.error('Load more posts failed:', err);
    } finally {
      setLoadingMore(false);
    }
  }

  async function handleFollow() {
    if (!profile || followBusy) return;
    if (followState === 'following') {
      const confirmed = window.confirm(
        `Unfollow ${getDisplayName(profile)}? You will stop seeing their posts in your feed.`
      );
      if (!confirmed) return;
    }
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

  function handleEditAbout() {
    setAboutDraft(profile?.aboutMe ?? '');
    setDobDraft(hasDateOfBirth(profile) ? (profile?.dateOfBirth ?? '').slice(0, 10) : '');
    setGenderDraft(profile?.gender ?? '');
    setEditingAbout(true);
  }

  async function handleSaveAbout() {
    if (!profile || savingAbout) return;
    setSavingAbout(true);
    try {
      // The update endpoint replaces all profile fields — resend current values.
      const updated = await updateProfile({
        firstName: profile.firstName,
        lastName: profile.lastName,
        nickname: profile.nickname ?? '',
        aboutMe: aboutDraft,
        dateOfBirth: dobDraft,
        gender: genderDraft,
      });
      setProfile({ ...profile, ...updated });
      setEditingAbout(false);
    } catch (err) {
      console.error('About Me update failed:', err);
    } finally {
      setSavingAbout(false);
    }
  }

  async function handleTogglePrivacy() {
    if (!profile || privacyBusy) return;
    const nextIsPrivate = profile.isPublic;
    const confirmed = window.confirm(
      nextIsPrivate
        ? 'Make your profile private? Only your followers will see your profile details and posts.'
        : 'Make your profile public? Everyone will be able to see your profile details and public posts.'
    );
    if (!confirmed) return;

    setPrivacyBusy(true);
    try {
      await toggleProfilePrivacy(nextIsPrivate);
      setProfile({ ...profile, isPublic: !nextIsPrivate });
    } catch (err) {
      console.error('Privacy toggle failed:', err);
    } finally {
      setPrivacyBusy(false);
    }
  }

  if (loading) return <p className="activity-page-title">Loading...</p>;
  if (error || !profile) return <p className="activity-page-title">{error}</p>;

  // Locked = viewing someone else's private profile without following them.
  const isLockedProfile = !isOwnProfile && profile.isPublic === false && !profile.isFollowing;

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

            {(isOwnProfile || !isLockedProfile) && (
              <div className="profile-contact-info">
                <p className="profile-bio-text">Email: {profile.email || 'Not set yet'}</p>
                <p className="profile-bio-text">
                  Date of birth:{' '}
                  {hasDateOfBirth(profile) ? profile.dateOfBirth?.slice(0, 10) : 'Not set yet'}
                </p>
                <p className="profile-bio-text">
                  Gender:{' '}
                  {profile.gender
                    ? (GENDER_LABELS[profile.gender] ?? profile.gender)
                    : 'Not set yet'}
                </p>
              </div>
            )}

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

            <div className="profile-bio">
              {isOwnProfile && editingAbout ? (
                <>
                  <textarea
                    className="form-input"
                    placeholder="Tell us about yourself"
                    rows={4}
                    value={aboutDraft}
                    onChange={(e) => setAboutDraft(e.target.value)}
                  />
                  <div style={{ display: 'flex', gap: '0.5rem' }}>
                    <input
                      className="form-input"
                      type="date"
                      aria-label="Date of birth"
                      value={dobDraft}
                      onChange={(e) => setDobDraft(e.target.value)}
                    />
                    <select
                      className="form-input"
                      aria-label="Gender"
                      value={genderDraft}
                      onChange={(e) => setGenderDraft(e.target.value)}
                    >
                      <option value="">Not specified</option>
                      {Object.entries(GENDER_LABELS).map(([value, label]) => (
                        <option key={value} value={value}>
                          {label}
                        </option>
                      ))}
                    </select>
                  </div>
                  <div style={{ display: 'flex', gap: '0.5rem' }}>
                    <button
                      type="button"
                      className="group-action-btn"
                      onClick={handleSaveAbout}
                      disabled={savingAbout}
                    >
                      {savingAbout ? 'Saving…' : 'Save'}
                    </button>
                    <button
                      type="button"
                      className="group-action-btn"
                      onClick={() => setEditingAbout(false)}
                      disabled={savingAbout}
                    >
                      Cancel
                    </button>
                  </div>
                </>
              ) : (
                <>
                  {profile.aboutMe ? (
                    <p className="profile-bio-text">{profile.aboutMe}</p>
                  ) : (
                    <p className="profile-bio-text">About me: Not set yet</p>
                  )}
                  {isOwnProfile && (
                    <button type="button" className="group-action-btn" onClick={handleEditAbout}>
                      {profile.aboutMe ? 'Edit About Me' : 'Add About Me'}
                    </button>
                  )}
                </>
              )}
            </div>

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
                onClick={() => openChatWithUser(profile.id)}
              >
                Message
              </button>
            )}

            {isOwnProfile && (
              <button
                type="button"
                className="group-action-btn"
                onClick={handleTogglePrivacy}
                disabled={privacyBusy}
              >
                {privacyBusy
                  ? 'Updating…'
                  : profile.isPublic
                    ? 'Make Profile Private'
                    : 'Make Profile Public'}
              </button>
            )}
          </div>
        </div>
      </div>

      {isOwnProfile && <FollowRequestsSection />}

      {isLockedProfile ? (
        <p className="activity-section">This profile is private. Follow to see posts.</p>
      ) : (
        <>
          <h2 className="activity-page-title">Posts</h2>
          {posts.length === 0 ? (
            <p className="activity-section">No posts yet.</p>
          ) : (
            posts.map((post) => <PostCard key={post.id} post={post} />)
          )}
          {page < totalPages && (
            <button
              type="button"
              className="group-action-btn"
              onClick={handleLoadMore}
              disabled={loadingMore}
            >
              {loadingMore ? 'Loading…' : 'Load More Posts'}
            </button>
          )}
        </>
      )}

      {statModal && (
        <FollowersFollowingModal userId={id} mode={statModal} onClose={() => setStatModal(null)} />
      )}
    </div>
  );
}
