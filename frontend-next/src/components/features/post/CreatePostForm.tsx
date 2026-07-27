'use client';

import { useState, useRef } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import Image from 'next/image';
import { useAuth } from '@/context/AuthContext';
import type { PostPrivacy } from '@/lib/types';

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
  const [error, setError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const fileInputRef = useRef<HTMLInputElement>(null);

  // Character limit
  const MAX_CONTENT_LENGTH = 1000;
  const charsRemaining = MAX_CONTENT_LENGTH - content.length;

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

    // TODO: Replace with real API call when backend is ready
    // try {
    //   const formData = new FormData();
    //   formData.append('content', content);
    //   formData.append('privacy', privacy);
    //   if (image) formData.append('image', image);
    //   if (privacy === 'private') {
    //     formData.append('allowedUsers', JSON.stringify(allowedUsers));
    //   }
    //
    //   await createPost(formData, groupId || undefined);
    //   resetForm();
    //   router.push(groupId ? `/groups/${groupId}` : '/');
    // } catch (err) {
    //   setError('Failed to create post. Please try again.');
    // } finally {
    //   setIsSubmitting(false);
    // }

    console.log('Post submitted:', {
      content,
      privacy,
      image,
      allowedUsers,
      groupId,
    });

    resetForm();
    setIsSubmitting(false);

    // Redirect after successful post
    router.push(groupId ? `/groups/${groupId}` : '/');
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
            {/* TODO: Add follower selection component here */}
            <p className="create-post-hint">
              Select followers from your follower list. (Coming soon)
            </p>
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
