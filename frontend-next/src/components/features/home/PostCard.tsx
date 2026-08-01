'use client';

/**
 * components/features/home/PostCard.tsx
 *
 * Individual post card with optimistic like toggle.
 * Extracted from app/page.tsx.
 *
 * When the backend is ready, wire handleLike to:
 *   liked ? unlikePost(post.id) : likePost(post.id)
 * The optimistic update pattern is already in place — just add the API call
 * and revert state on error.
 */

import { useState } from 'react';
import Image from 'next/image';
import Link from 'next/link';
import { likePost, unlikePost } from '@/lib/api';
import { getDisplayName, getFileUrl, formatRelativeDate } from '@/lib/helpers';
import type { Post } from '@/lib/types';

export default function PostCard({ post }: { post: Post }) {
  const [liked, setLiked] = useState(post.isLiked ?? false);
  const [likesCount, setLikesCount] = useState(post.likesCount);

  const canLike = !post.groupId;

  async function handleLike() {
    // Optimistic update — flip immediately, revert on error
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
      // Revert on failure
      setLiked(wasLiked);
      setLikesCount((prev) => (wasLiked ? prev + 1 : prev - 1));
    }
  }

  const privacyIcon = post.privacy === 'public' ? '🌍' : post.privacy === 'followers' ? '👥' : '🔒';

  return (
    <div className="post-card">
      {/* Header */}
      <div className="post-header">
        <Link href={`/profile/${post.userId}`} className="post-user-link">
          <Image
            src={getFileUrl(post.user.avatarUrl)}
            alt={getDisplayName(post.user)}
            width={40}
            height={40}
            className="post-avatar"
          />
          <div className="post-user-info">
            <span className="post-user-name">{getDisplayName(post.user)}</span>
            <span className="post-meta">
              @{post.user.username} · {formatRelativeDate(post.createdAt)} ·{' '}
              <span title={post.privacy}>{privacyIcon}</span>
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
              width={600}
              height={400}
              className="post-image"
              style={{ objectFit: 'cover' }}
              loading="eager"
            />
          </div>
        )}
      </div>

      {/* Actions */}
      <div className="post-actions">
        {canLike && (
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
        )}
        <Link href={`/post/${post.id}`} className="post-action-btn">
          <Image src="/images/icons/icon-comments.png" alt="Comments" width={20} height={20} />
          <span>{post.commentsCount}</span>
        </Link>
      </div>
    </div>
  );
}
