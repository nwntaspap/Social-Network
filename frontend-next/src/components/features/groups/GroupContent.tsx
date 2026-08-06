'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import Image from 'next/image';
import Link from 'next/link';
import {
  getGroupPosts,
  getGroupEvents,
  getGroupPostComments,
  createGroupPostComment,
} from '@/lib/api';
import PostCard from '@/components/features/home/PostCard';
import { formatRelativeDate, getDisplayName, getFileUrl } from '@/lib/helpers';
import type { Comment, Event, Post } from '@/lib/types';
import { tabView } from './GroupDetail';

interface GroupContentProps {
  groupId: string;
  activeTab: tabView;
  isMember: boolean;
}

const PAGE_SIZE = 10;

export default function GroupContent({ groupId, activeTab }: GroupContentProps) {
  return (
    <div className="group-content-area">
      {activeTab === 'posts' ? <PostsTab groupId={groupId} /> : <EventsTab groupId={groupId} />}
    </div>
  );
}

function PostsTab({ groupId }: { groupId: string }) {
  const [posts, setPosts] = useState<Post[]>([]);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState('');

  const loadPage = useCallback(
    async (nextPage: number) => {
      try {
        const response = await getGroupPosts(groupId, nextPage, PAGE_SIZE);
        setPosts((prev) => (nextPage === 1 ? response.data : [...prev, ...response.data]));
        setTotalPages(response.totalPages || 1);
      } catch {
        setError('Failed to load posts. Please try again.');
      } finally {
        setLoading(false);
        setLoadingMore(false);
      }
    },
    [groupId]
  );

  useEffect(() => {
    let ignore = false;
    getGroupPosts(groupId, 1, PAGE_SIZE)
      .then((response) => {
        if (ignore) return;
        setPosts(response.data);
        setTotalPages(response.totalPages || 1);
      })
      .catch(() => {
        if (!ignore) setError('Failed to load posts. Please try again.');
      })
      .finally(() => {
        if (!ignore) setLoading(false);
      });
    return () => {
      ignore = true;
    };
  }, [groupId]);

  if (loading) {
    return <p className="group-empty-state">Loading posts...</p>;
  }

  if (error) {
    return <p className="group-empty-state">{error}</p>;
  }

  if (posts.length === 0) {
    return (
      <div className="group-posts">
        <p className="group-empty-state">No posts yet. Be the first to create one!</p>
      </div>
    );
  }

  return (
    <div className="group-posts">
      {posts.map((post) => (
        <div key={post.id}>
          <PostCard post={post} />
          <PostComments postId={post.id} />
        </div>
      ))}
      {page < totalPages && (
        <div className="feed-load-more">
          <button
            className="feed-load-more-btn"
            disabled={loadingMore}
            onClick={async () => {
              const nextPage = page + 1;
              setLoadingMore(true);
              setPage(nextPage);
              await loadPage(nextPage);
            }}
          >
            {loadingMore ? 'Loading...' : 'Load more'}
          </button>
        </div>
      )}
    </div>
  );
}

function PostComments({ postId }: { postId: string }) {
  const [comments, setComments] = useState<Comment[]>([]);
  const [content, setContent] = useState('');
  const [image, setImage] = useState<File | null>(null);
  const [imagePreview, setImagePreview] = useState('');
  const [collapsed, setCollapsed] = useState(true);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const imageInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    let ignore = false;
    getGroupPostComments(postId)
      .then((response) => {
        if (ignore) return;
        setComments(response.data);
      })
      .catch(() => {
        if (!ignore) setError('Failed to load comments.');
      })
      .finally(() => {
        if (!ignore) setLoading(false);
      });
    return () => {
      ignore = true;
    };
  }, [postId]);

  function handleImageChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    setImage(file);
    setImagePreview(URL.createObjectURL(file));
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError('');
    if (!content.trim()) {
      setError('Comment content is required.');
      return;
    }
    setSubmitting(true);
    try {
      const newComment = await createGroupPostComment(postId, content, image);
      setComments((prev) => [newComment, ...prev]);
      setContent('');
      setImage(null);
      setImagePreview('');
      if (imageInputRef.current) imageInputRef.current.value = '';
    } catch {
      setError('Failed to post comment. Please try again.');
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="group-post-comments">
      <button
        type="button"
        className="group-comments-toggle"
        onClick={() => setCollapsed((prev) => !prev)}
      >
        {collapsed ? `Show comments (${comments.length})` : 'Hide comments'}
      </button>

      {!collapsed && (
        <div className="group-comments-body">
          <form className="create-comment-form" onSubmit={handleSubmit}>
            <div className="create-comment-field">
              <textarea
                className="form-textarea"
                placeholder="Write a comment..."
                value={content}
                onChange={(e) => setContent(e.target.value)}
                rows={2}
              />
            </div>

            {imagePreview && (
              <div className="comment-image-preview">
                <Image
                  src={imagePreview}
                  alt="Comment image preview"
                  width={120}
                  height={90}
                  style={{ objectFit: 'cover', borderRadius: '8px' }}
                />
                <button
                  type="button"
                  className="comment-image-remove"
                  onClick={() => {
                    setImage(null);
                    setImagePreview('');
                    if (imageInputRef.current) imageInputRef.current.value = '';
                  }}
                >
                  Remove
                </button>
              </div>
            )}

            {error && <p className="comment-error">{error}</p>}

            <div className="create-comment-actions">
              <div className="create-comment-buttons">
                <input
                  ref={imageInputRef}
                  type="file"
                  accept="image/*"
                  onChange={handleImageChange}
                  className="comment-image-input"
                />
                <button type="submit" className="create-comment-submit" disabled={submitting}>
                  {submitting ? 'Posting...' : 'Post Comment'}
                </button>
              </div>
            </div>
          </form>

          {loading ? (
            <p className="group-empty-state">Loading comments...</p>
          ) : comments.length > 0 ? (
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
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <p className="post-comments-empty">No comments yet. Be the first to comment!</p>
          )}
        </div>
      )}
    </div>
  );
}

function EventsTab({ groupId }: { groupId: string }) {
  const [events, setEvents] = useState<Event[]>([]);
  const [nextCursor, setNextCursor] = useState<string | undefined>(undefined);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState('');

  const load = useCallback(
    async (cursor?: string) => {
      try {
        const response = await getGroupEvents(groupId, cursor);
        setEvents((prev) => (cursor ? [...prev, ...response.events] : response.events));
        setNextCursor(response.nextCursor);
      } catch {
        setError('Failed to load events. Please try again.');
      } finally {
        setLoading(false);
        setLoadingMore(false);
      }
    },
    [groupId]
  );

  useEffect(() => {
    let ignore = false;
    getGroupEvents(groupId)
      .then((response) => {
        if (ignore) return;
        setEvents(response.events);
        setNextCursor(response.nextCursor);
      })
      .catch(() => {
        if (!ignore) setError('Failed to load events. Please try again.');
      })
      .finally(() => {
        if (!ignore) setLoading(false);
      });
    return () => {
      ignore = true;
    };
  }, [groupId]);

  if (loading) {
    return <p className="group-empty-state">Loading events...</p>;
  }

  if (error) {
    return <p className="group-empty-state">{error}</p>;
  }

  return (
    <div className="group-events">
      {events.length > 0 ? (
        events.map((event) => (
          <div key={event.id} className="group-event-card">
            <div className="event-date-badge">
              <span className="event-month">
                {new Date(event.eventDate).toLocaleString('en-US', { month: 'short' })}
              </span>
              <span className="event-day">{new Date(event.eventDate).getDate()}</span>
            </div>
            <div className="event-details">
              <h3 className="event-title">{event.title}</h3>
              <p className="event-desc">{event.description}</p>
              <div className="event-meta">
                <span className="event-creator">
                  Created by {event.creator?.firstName ?? ''} {event.creator?.lastName ?? ''}
                </span>
                <span className="event-time">{formatRelativeDate(event.eventDate)}</span>
              </div>
            </div>
          </div>
        ))
      ) : (
        <p className="group-empty-state">No events scheduled yet.</p>
      )}

      {nextCursor && (
        <div className="feed-load-more">
          <button
            className="feed-load-more-btn"
            disabled={loadingMore}
            onClick={async () => {
              setLoadingMore(true);
              await load(nextCursor);
            }}
          >
            {loadingMore ? 'Loading...' : 'Load more'}
          </button>
        </div>
      )}
    </div>
  );
}
