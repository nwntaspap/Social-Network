'use client';

import { useState, useRef, useEffect } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import Image from 'next/image';
import { createPost, createGroupPost, getFollowers } from '@/lib/api';
import { getDisplayName, getFileUrl } from '@/lib/helpers';
import { useAuth } from '@/context/AuthContext';
import type { PostPrivacy, User } from '@/lib/types';

export default function CreatePostForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { user } = useAuth();

  const groupId = searchParams.get('groupId');
  const isGroupPost = !!groupId;

  // Form state
  const [content, setContent] = useState('');
  const [image, setImage] = useState<File | null>(null);
  const [imagePreview, setImagePreview] = useState<string | null>(null);
  const [privacy, setPrivacy] = useState<PostPrivacy>('public');
  const [allowedUsers, setAllowedUsers] = useState<string[]>([]);
  const [followers, setFollowers] = useState<User[] | null>(null);
  const [followersError, setFollowersError] = useState('');
  const [error, setError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const fileInputRef = useRef<HTMLInputElement>(null);

  // Character limit
  const MAX_CONTENT_LENGTH = 1000;
  const charsRemaining = MAX_CONTENT_LENGTH - content.length;

  useEffect(() => {
    if (!user || isGroupPost || privacy !== 'private' || followers !== null) {
      return;
    }
    let cancelled = false;
    getFollowers(user.id)
      .then((data) => {
        if (!cancelled) {
          setFollowers(data);
          setFollowersError('');
        }
      })
      .catch(() => {
        if (!cancelled) {
          setFollowers([]);
          setFollowersError('Failed to load your followers.');
        }
      });
    return () => {
      cancelled = true;
    };
  }, [user, isGroupPost, privacy, followers]);

  function toggleAllowedUser(userId: string) {
    setAllowedUsers((prev) =>
      prev.includes(userId) ? prev.filter((id) => id !== userId) : [...prev, userId]
    );
    setError('');
  }

  function toggleAllFollowers() {
    const total = followers?.length ?? 0;
    setAllowedUsers((prev) => (prev.length === total ? [] : (followers ?? []).map((f) => f.id)));
    setError('');
  }

  function handleImageSelect(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;

    if (!file.type.startsWith('image/')) {
      setError('Please select an image file (JPG, PNG, or GIF).');
      return;
    }

    if (file.size > 5 * 1024 * 1024) {
      setError('Image size must be smaller than 5MB');
      return;
    }

    if (imagePreview) {
      URL.revokeObjectURL(imagePreview);
    }

    setImage(file);
    setImagePreview(URL.createObjectURL(file));
    setError('');
  }

  function removeImage() {
    if (imagePreview) {
      URL.revokeObjectURL(imagePreview);
    }
    setImage(null);
    setImagePreview(null);
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  }

  function resetForm() {
    setContent('');
    setPrivacy('public');
    setAllowedUsers([]);
    setError('');
    removeImage();
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError('');

    if (!content.trim()) {
      setError('Post content is required.');
      return;
    }

    if (content.length > MAX_CONTENT_LENGTH) {
      setError(`Content must be ${MAX_CONTENT_LENGTH} characters or less.`);
      return;
    }

    if (privacy === 'private' && allowedUsers.length === 0) {
      setError('Please select at least one user for private posts.');
      return;
    }

    setIsSubmitting(true);

    try {
      const formData = new FormData();
      formData.append('title', content);
      formData.append('content', content);
      formData.append('privacy', privacy);
      if (image) formData.append('image', image);
      if (groupId) formData.append('groupId', groupId);
      if (privacy === 'private' && allowedUsers.length > 0) {
        formData.append('allowedUserIds', JSON.stringify(allowedUsers));
      }

      await (groupId ? createGroupPost(formData, groupId) : createPost(formData));
      resetForm();
      router.push(groupId ? `/groups/${groupId}` : '/');
    } catch {
      setError('Failed to create post. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  }

  if (!user) {
    return (
      <div className="create-post-container">
        <p>Please log in to create a post.</p>
      </div>
    );
  }

  return (
    <div className="create-post-container">
      <h2 className="create-post-title">{isGroupPost ? 'Create Group Post' : 'Create Post'}</h2>

      {isGroupPost && (
        <p className="create-post-group-info">
          Posting to group. Only group members will be able to see this post.
        </p>
      )}

      <form className="create-post-form" onSubmit={handleSubmit}>
        {/* Content */}
        <div className="create-post-field">
          <textarea
            className="form-textarea"
            placeholder="What's on your mind?"
            value={content}
            onChange={(e) => {
              setContent(e.target.value);
              setError('');
            }}
            rows={6}
            maxLength={MAX_CONTENT_LENGTH}
          />
          <div className={`char-count ${charsRemaining < 50 ? 'char-count--warning' : ''}`}>
            {charsRemaining} characters remaining
          </div>
        </div>

        {/* Image Preview */}
        {imagePreview && (
          <div className="create-post-image-preview">
            <Image src={imagePreview} alt="Preview" width={400} height={300} />
            <button type="button" className="create-post-image-remove" onClick={removeImage}>
              ✕
            </button>
          </div>
        )}

        {/* Privacy Settings - only for non-group posts */}
        {!isGroupPost && (
          <div className="create-post-field">
            <label className="create-post-label">Privacy</label>
            <div className="privacy-options">
              <label
                className={`privacy-option ${privacy === 'public' ? 'privacy-option--active' : ''}`}
              >
                <input
                  type="radio"
                  name="privacy"
                  value="public"
                  checked={privacy === 'public'}
                  onChange={() => setPrivacy('public')}
                />
                <span className="privacy-option-content">
                  <span className="privacy-option-icon">🌍</span>
                  <span className="privacy-option-text">
                    <strong>Public</strong>
                    <small>Everyone can see this post</small>
                  </span>
                </span>
              </label>

              <label
                className={`privacy-option ${privacy === 'followers' ? 'privacy-option--active' : ''}`}
              >
                <input
                  type="radio"
                  name="privacy"
                  value="followers"
                  checked={privacy === 'followers'}
                  onChange={() => setPrivacy('followers')}
                />
                <span className="privacy-option-content">
                  <span className="privacy-option-icon">👥</span>
                  <span className="privacy-option-text">
                    <strong>Followers Only</strong>
                    <small>Only your followers can see this post</small>
                  </span>
                </span>
              </label>

              <label
                className={`privacy-option ${privacy === 'private' ? 'privacy-option--active' : ''}`}
              >
                <input
                  type="radio"
                  name="privacy"
                  value="private"
                  checked={privacy === 'private'}
                  onChange={() => setPrivacy('private')}
                />
                <span className="privacy-option-content">
                  <span className="privacy-option-icon">🔒</span>
                  <span className="privacy-option-text">
                    <strong>Private</strong>
                    <small>Only selected followers can see this post</small>
                  </span>
                </span>
              </label>
            </div>
          </div>
        )}

        {/* Allowed Users - only for private posts */}
        {privacy === 'private' && !isGroupPost && (
          <div className="create-post-field">
            <label className="create-post-label">Select Followers Who Can See This Post</label>
            <div className="allowed-users-toolbar">
              <label className="allowed-users-toggle-all">
                <input
                  type="checkbox"
                  checked={
                    (followers?.length ?? 0) > 0 && allowedUsers.length === followers?.length
                  }
                  onChange={toggleAllFollowers}
                />
                <span>Select all</span>
              </label>
              <span className="allowed-users-count">
                {allowedUsers.length} of {followers?.length ?? 0} selected
              </span>
            </div>

            {followers === null && <p className="create-post-hint">Loading followers...</p>}
            {followersError && <p className="create-post-error">{followersError}</p>}
            {followers !== null && !followersError && followers.length === 0 && (
              <p className="create-post-hint">
                You have no followers to select. Share your profile to gain followers.
              </p>
            )}

            {followers !== null && !followersError && followers.length > 0 && (
              <div className="allowed-users-list">
                {followers.map((follower) => (
                  <label key={follower.id} className="allowed-user-row">
                    <input
                      type="checkbox"
                      checked={allowedUsers.includes(follower.id)}
                      onChange={() => toggleAllowedUser(follower.id)}
                    />
                    <Image
                      src={getFileUrl(follower.avatarUrl)}
                      alt={getDisplayName(follower)}
                      width={36}
                      height={36}
                      className="allowed-user-avatar"
                    />
                    <span className="allowed-user-name">{getDisplayName(follower)}</span>
                    <span className="allowed-user-username">
                      @{follower.username || follower.nickname}
                    </span>
                  </label>
                ))}
              </div>
            )}
          </div>
        )}

        {/* Error Message */}
        {error && <p className="create-post-error">{error}</p>}

        {/* Actions */}
        <div className="create-post-actions">
          <div className="create-post-actions-left">
            <label className="create-post-upload-btn">
              <Image src="/images/icons/upload-icon.png" alt="Upload" width={20} height={20} />
              <span>Add Image</span>
              <input
                ref={fileInputRef}
                type="file"
                accept="image/*"
                onChange={handleImageSelect}
                style={{ display: 'none' }}
              />
            </label>
          </div>

          <div className="create-post-actions-right">
            <button type="button" className="create-post-cancel" onClick={() => router.back()}>
              Cancel
            </button>
            <button type="submit" className="create-post-submit" disabled={isSubmitting}>
              {isSubmitting ? 'Posting...' : 'Post'}
            </button>
          </div>
        </div>
      </form>
    </div>
  );
}
