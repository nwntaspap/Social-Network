'use client';

import { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';
import Image from 'next/image';
import Link from 'next/link';
import { getPost, getComments, createComment, voteComment, likePost, unlikePost } from '@/lib/api';
import { getDisplayName, getFileUrl, formatRelativeDate } from '@/lib/helpers';
import type { Comment, Post } from '@/lib/types';

export default function PostDetail() {
  const { id: postId } = useParams<{ id: string }>();

  const [post, setPost] = useState<Post | null>(null);
  const [comments, setComments] = useState<Comment[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [liked, setLiked] = useState(false);
  const [likesCount, setLikesCount] = useState(0);

  // Comment form
  const [showCommentForm, setShowCommentForm] = useState(false);
  const [commentContent, setCommentContent] = useState('');
  const [commentError, setCommentError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    let ignore = false;
    Promise.all([getPost(Number(postId)), getComments(Number(postId))])
      .then(([postData, commentsData]) => {
        if (ignore) return;
        setPost(postData);
        setLiked(!!postData.isLiked);
        setLikesCount(postData.likesCount ?? 0);
        setComments(commentsData);
      })
      .catch(() => {
        if (!ignore) setError('Post not found.');
      })
      .finally(() => {
        if (!ignore) setLoading(false);
      });
    return () => {
      ignore = true;
    };
  }, [postId]);

  async function handleLike() {
    if (!post) return;
    const wasLiked = liked;
    setLiked(!wasLiked);
    setLikesCount((prev) => (wasLiked ? prev - 1 : prev + 1));

    try {
      if (wasLiked) {
        await unlikePost(Number(post.id));
      } else {
        await likePost(Number(post.id));
      }
    } catch {
      setLiked(wasLiked);
      setLikesCount((prev) => (wasLiked ? prev + 1 : prev - 1));
    }
  }

  function resetForm() {
    setShowCommentForm(false);
    setCommentContent('');
    setCommentError('');
  }

  async function handleCommentVote(comment: Comment, reactionType: 1 | -1) {
    const prevUserVote = comment.userVote ?? null;
    const prevUp = comment.upvoteCount ?? 0;
    const prevDown = comment.downvoteCount ?? 0;

    let newUserVote: number | null = reactionType;
    let deltaUp = 0;
    let deltaDown = 0;
    if (prevUserVote === reactionType) {
      newUserVote = null;
      if (reactionType === 1) deltaUp = -1;
      else deltaDown = -1;
    } else if (prevUserVote === -reactionType) {
      if (reactionType === 1) {
        deltaUp = 1;
        deltaDown = -1;
      } else {
        deltaDown = 1;
        deltaUp = -1;
      }
    } else if (reactionType === 1) {
      deltaUp = 1;
    } else {
      deltaDown = 1;
    }

    setComments((prev) =>
      prev.map((c) =>
        c.id === comment.id
          ? {
              ...c,
              userVote: newUserVote,
              upvoteCount: Math.max(0, prevUp + deltaUp),
              downvoteCount: Math.max(0, prevDown + deltaDown),
              voteScore: prevUp + deltaUp - (prevDown + deltaDown),
            }
          : c
      )
    );

    try {
      await voteComment(Number(comment.id), reactionType);
    } catch {
      setComments((prev) =>
        prev.map((c) =>
          c.id === comment.id
            ? {
                ...c,
                userVote: prevUserVote,
                upvoteCount: prevUp,
                downvoteCount: prevDown,
                voteScore: prevUp - prevDown,
              }
            : c
        )
      );
    }
  }

  async function handleSubmitComment(e: React.FormEvent) {
    e.preventDefault();
    setCommentError('');

    if (!commentContent.trim()) {
      setCommentError('Comment content is required.');
      return;
    }

    setIsSubmitting(true);

    try {
      const newComment = await createComment(Number(postId), commentContent);
      setComments((prev) => [newComment, ...prev]);
      resetForm();
    } catch {
      setCommentError('Failed to post comment. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  }

  if (loading) {
    return (
      <div className="post-detail-container">
        <p className="post-detail-loading">Loading post...</p>
      </div>
    );
  }

  if (error || !post) {
    return (
      <div className="post-detail-container">
        <p className="post-detail-not-found">{error || 'Post not found.'}</p>
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
              src={getFileUrl(post.user?.avatarUrl)}
              alt={getDisplayName(post.user)}
              width={48}
              height={48}
              className="post-avatar"
            />
            <div className="post-user-info">
              <span className="post-user-name">{getDisplayName(post.user)}</span>
              <span className="post-meta">
                @{post.user?.username} · {formatRelativeDate(post.createdAt)} ·{' '}
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
        {!post.groupId && (
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
        )}
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

            {commentError && <p className="comment-error">{commentError}</p>}

            <div className="create-comment-actions">
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
                    src={getFileUrl(comment.user?.avatarUrl)}
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
                      <span className="comment-username">@{comment.user?.username}</span>
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
                  <div className="comment-votes">
                    <button
                      type="button"
                      className={`comment-vote-btn${comment.userVote === 1 ? ' active' : ''}`}
                      onClick={() => handleCommentVote(comment, 1)}
                    >
                      <Image
                        src="/images/icons/icon-like.png"
                        alt="Upvote"
                        width={16}
                        height={16}
                      />
                      <span>{comment.upvoteCount ?? 0}</span>
                    </button>
                    <button
                      type="button"
                      className={`comment-vote-btn${comment.userVote === -1 ? ' active' : ''}`}
                      onClick={() => handleCommentVote(comment, -1)}
                    >
                      <Image
                        src="/images/icons/icon-dislike.png"
                        alt="Downvote"
                        width={16}
                        height={16}
                      />
                      <span>{comment.downvoteCount ?? 0}</span>
                    </button>
                  </div>
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
