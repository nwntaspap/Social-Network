'use client';

/**
 * components/features/home/PostCard.tsx
 *
 * Individual post card with optimistic like/dislike toggles.
 * Extracted from app/page.tsx.
 */

import { useState } from 'react';
import Image from 'next/image';
import Link from 'next/link';
import { likePost, dislikePost, removePostVote, voteGroupPost } from '@/lib/api';
import { getDisplayName, getFileUrl, formatRelativeDate } from '@/lib/helpers';
import type { Post } from '@/lib/types';

interface PostCardProps {
  post: Post;
  commentsExpanded?: boolean;
  onToggleComments?: (postId: string) => void;
}

function initialVote(post: Post): number | null {
  if (post.userVote !== undefined && post.userVote !== null) return post.userVote;
  if (typeof post.isLiked === 'number') return post.isLiked;
  return post.isLiked ? 1 : null;
}

export default function PostCard({ post, commentsExpanded, onToggleComments }: PostCardProps) {
  const isGroupPost = !!post.groupId;

  const [userVote, setUserVote] = useState<number | null>(initialVote(post));
  const [likesCount, setLikesCount] = useState(post.likesCount);
  const [dislikesCount, setDislikesCount] = useState(post.dislikesCount ?? 0);

  async function handleVote(reaction: 1 | -1, commit: (nextVote: number | null) => Promise<void>) {
    const prevVote = userVote;
    const nextVote = prevVote === reaction ? null : reaction;

    const deltaOf = (r: 1 | -1): number => {
      if (prevVote === r) return -1;
      if (nextVote === r) return 1;
      return 0;
    };

    setUserVote(nextVote);
    setLikesCount((prev) => prev + deltaOf(1));
    setDislikesCount((prev) => prev + deltaOf(-1));

    try {
      await commit(nextVote);
    } catch {
      setUserVote(prevVote);
      setLikesCount((prev) => prev - deltaOf(1));
      setDislikesCount((prev) => prev - deltaOf(-1));
    }
  }

  async function handleTopicVote(reaction: 1 | -1) {
    await handleVote(reaction, async (nextVote) => {
      if (nextVote === null) {
        await removePostVote(Number(post.id));
      } else if (reaction === 1) {
        await likePost(Number(post.id));
      } else {
        await dislikePost(Number(post.id));
      }
    });
  }

  async function handleGroupVote(reaction: 1 | -1) {
    await handleVote(reaction, () => voteGroupPost(post.id, reaction));
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
        <button
          type="button"
          aria-label="Like"
          className="post-action-btn"
          onClick={() => (isGroupPost ? handleGroupVote(1) : handleTopicVote(1))}
        >
          <Image
            src="/images/icons/icon-like.png"
            alt="Like"
            width={20}
            height={20}
            className={userVote === 1 ? 'icon-liked' : 'icon-not-liked'}
          />
          <span>{likesCount}</span>
        </button>
        <button
          type="button"
          aria-label="Dislike"
          className="post-action-btn"
          onClick={() => (isGroupPost ? handleGroupVote(-1) : handleTopicVote(-1))}
        >
          <Image
            src="/images/icons/icon-dislike.png"
            alt="Dislike"
            width={20}
            height={20}
            className={userVote === -1 ? 'icon-liked' : 'icon-not-liked'}
          />
          <span>{dislikesCount}</span>
        </button>
        {onToggleComments ? (
          <button
            type="button"
            aria-label="Comments"
            aria-expanded={commentsExpanded}
            className="post-action-btn"
            onClick={() => onToggleComments(post.id)}
          >
            <Image src="/images/icons/icon-comments.png" alt="Comments" width={20} height={20} />
            <span>{post.commentsCount}</span>
          </button>
        ) : (
          <Link href={`/post/${post.id}`} className="post-action-btn">
            <Image src="/images/icons/icon-comments.png" alt="Comments" width={20} height={20} />
            <span>{post.commentsCount}</span>
          </Link>
        )}
      </div>
    </div>
  );
}
