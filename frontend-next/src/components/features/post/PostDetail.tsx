'use client';

import { useState, useMemo, useRef } from 'react';
import { useParams } from 'next/navigation';
import Image from 'next/image';
import Link from 'next/link';
import { getDisplayName, getFileUrl, formatRelativeDate } from '@/lib/helpers';
import { feedPosts } from '@/mocks/posts';
import { mockGroupPosts, defaultGroupPosts } from '@/mocks/group-posts';
import { mockGroups } from '@/mocks/groups';
import { mockComments, defaultComments } from '@/mocks/comments';
import type { Post, Comment } from '@/lib/types';

export default function PostDetail() {
  const { id: postId } = useParams<{ id: string }>();

  const resolvedPost = useMemo(() => {
    const allGroupPosts = Object.values(mockGroupPosts).flat();
    const allPosts = [...feedPosts, ...allGroupPosts, ...defaultGroupPosts];

    let foundPost = allPosts.find((p) => p.id === postId);

    if (foundPost?.groupId && !foundPost.group) {
      const group = mockGroups.find((g) => g.id === foundPost!.groupId);
      if (group) {
        foundPost = { ...foundPost, group: { id: group.id, title: group.title } };
      }
    }

    return foundPost ?? null;
  }, [postId]);

  const [post] = useState<Post | null>(resolvedPost);
  const [comments] = useState<Comment[]>(mockComments[postId] || defaultComments);
  const [loading] = useState(false);
  const [liked, setLiked] = useState(resolvedPost?.isLiked ?? false);
  const [likesCount, setLikesCount] = useState(resolvedPost?.likesCount ?? 0);

  // Comment form
  const [showCommentForm, setShowCommentForm] = useState(false);
  const [commentContent, setCommentContent] = useState('');
  const [commentImage, setCommentImage] = useState<File | null>(null);
  const [commentImagePreview, setCommentImagePreview] = useState<string | null>(null);
  const [commentError, setCommentError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  async function handleLike() {
    const wasLiked = liked;
    setLiked(!wasLiked);
    setLikesCount((prev) => (wasLiked ? prev - 1 : prev + 1));

    try {
      // TODO: wasLiked ? await unlikePost(postId) : await likePost(postId)
    } catch {
      setLiked(wasLiked);
      setLikesCount((prev) => (wasLiked ? prev + 1 : prev - 1));
    }
  }

  function handleImageSelect(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;

    if (!file.type.startsWith('image/')) {
      setCommentError('Please select an image file (JPG, PNG, or GIF).');
      return;
    }

    if (file.size > 5 * 1024 * 1024) {
      setCommentError('Image size must be smaller than 5MB');
      return;
    }

    if (commentImagePreview) {
      URL.revokeObjectURL(commentImagePreview);
    }

    setCommentImage(file);
    setCommentImagePreview(URL.createObjectURL(file));
    setCommentError('');
  }

  function removeImage() {
    if (commentImagePreview) {
      URL.revokeObjectURL(commentImagePreview);
    }

    setCommentImage(null);
    setCommentImagePreview(null);
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  }

  function resetForm() {
    setShowCommentForm(false);
    setCommentContent('');
    setCommentError('');
    removeImage();
  }

  async function handleSubmitComment(e: React.FormEvent) {
    e.preventDefault();
    setCommentError('');

    if (!commentContent.trim()) {
      setCommentError('Comment content is required.');
      return;
    }

    setIsSubmitting(true);

    // TODO: When backend is ready, replace with:
    // try {
    //   const { data: newComment } = await createComment(postId, {
    //     content: commentContent,
    //     image: commentImage,
    //   });
    //   setComments((prev) => [newComment, ...prev]);
    //   resetForm();
    // } catch {
    //   setCommentError('Failed to post comment. Please try again.');
    // } finally {
    //   setIsSubmitting(false);
    // }

    console.log('Comment submitted:', { postId, content: commentContent, image: commentImage });
    resetForm();
    setIsSubmitting(false);
  }

  if (loading) {
    return (
      <div className="post-detail-container">
        <p className="post-detail-loading">Loading post...</p>
      </div>
    );
  }

  if (!post) {
    return (
      <div className="post-detail-container">
        <p className="post-detail-not-found">Post not found.</p>
      </div>
    );
  }

  return (
    <div className="post-detail-container">
      {/* Post */}
      <div className="post-detail-card">
        {/* Header */}
        <div className="post-header">
          <Link href={`/profile/${post.userId}`} className="post-user-link">
            <Image
              src={getFileUrl(post.user.avatarUrl)}
              alt={getDisplayName(post.user)}
              width={48}
              height={48}
              className="post-avatar"
            />
            <div className="post-user-info">
              <span className="post-user-name">{getDisplayName(post.user)}</span>
              <span className="post-meta">
                @{post.user.username} · {formatRelativeDate(post.createdAt)} ·{' '}
                {post.group ? (
                  <Link href={`/groups/${post.group.id}`} className="post-meta-group">
                    {post.group.title}
                  </Link>
                ) : (
                  <span title={post.privacy}>
                    {post.privacy === 'public' ? '🌍' : post.privacy === 'followers' ? '👥' : '🔒'}
                  </span>
                )}
              </span>
            </div>
          </Link>
        </div>

        {/* Content */}
        <div className="post-content">
          <p className="post-text">{post.content}</p>
          {post.imageUrl && (
            <div className="post-image-wrapper">
              <Image
                src={post.imageUrl}
                alt="Post image"
                width={800}
                height={500}
                className="post-image"
                loading="eager"
              />
            </div>
          )}
        </div>

        {/* Actions */}
        <div className="post-actions">
          <button className="post-action-btn" onClick={handleLike}>
            <Image
              src="/images/icons/heart.png"
              alt={liked ? 'Unlike' : 'Like'}
              width={20}
              height={20}
              className={liked ? 'icon-liked' : 'icon-not-liked'}
            />
            <span>{likesCount}</span>
          </button>
          <div className="post-action-btn">
            <Image src="/images/icons/icon-comments.png" alt="Comments" width={20} height={20} />
            <span>{post.commentsCount}</span>
          </div>
        </div>
      </div>

      {/* Create Comment */}
      <div className="create-comment-wrapper">
        {!showCommentForm ? (
          <button className="create-comment-btn" onClick={() => setShowCommentForm(true)}>
            Create Comment
          </button>
        ) : (
          <form className="create-comment-form" onSubmit={handleSubmitComment}>
            <div className="create-comment-field">
              <textarea
                className="form-textarea"
                placeholder="Write your comment..."
                value={commentContent}
                onChange={(e) => setCommentContent(e.target.value)}
                rows={3}
              />
            </div>

            {commentImagePreview && (
              <div className="comment-image-preview">
                <Image src={commentImagePreview} alt="Preview" width={200} height={150} />
                <button type="button" className="comment-image-remove" onClick={removeImage}>
                  ✕
                </button>
              </div>
            )}

            {commentError && <p className="comment-error">{commentError}</p>}

            <div className="create-comment-actions">
              <label className="comment-upload-btn">
                <Image src="/images/icons/upload-icon.png" alt="Upload" width={20} height={20} />
                <span>Image/GIF</span>
                <input
                  ref={fileInputRef}
                  type="file"
                  accept="image/*"
                  onChange={handleImageSelect}
                  style={{ display: 'none' }}
                />
              </label>

              <div className="create-comment-buttons">
                <button type="button" className="create-comment-cancel" onClick={resetForm}>
                  Cancel
                </button>
                <button type="submit" className="create-comment-submit" disabled={isSubmitting}>
                  {isSubmitting ? 'Posting...' : 'Post Comment'}
                </button>
              </div>
            </div>
          </form>
        )}
      </div>

      {/* Comments */}
      <div className="post-comments-section">
        <h3 className="post-comments-title">Comments</h3>
        {comments.length > 0 ? (
          <div className="post-comments-list">
            {comments.map((comment) => (
              <div key={comment.id} className="comment-card">
                <Link href={`/profile/${comment.userId}`} className="comment-avatar-link">
                  <Image
                    src={getFileUrl(comment.user.avatarUrl)}
                    alt={getDisplayName(comment.user)}
                    width={36}
                    height={36}
                    className="comment-avatar"
                  />
                </Link>
                <div className="comment-body">
                  <div className="comment-header">
                    <Link href={`/profile/${comment.userId}`} className="comment-user-link">
                      <span className="comment-user-name">{getDisplayName(comment.user)}</span>
                      <span className="comment-username">@{comment.user.username}</span>
                    </Link>
                    <span className="comment-time">{formatRelativeDate(comment.createdAt)}</span>
                  </div>
                  <p className="comment-content">{comment.content}</p>
                  {comment.imageUrl && (
                    <div className="comment-image">
                      <Image
                        src={comment.imageUrl}
                        alt="Comment image"
                        width={300}
                        height={200}
                        style={{ objectFit: 'cover', borderRadius: '8px', marginTop: '0.5rem' }}
                      />
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        ) : (
          <p className="post-comments-empty">No comments yet. Be the first to comment!</p>
        )}
      </div>
    </div>
  );
}
