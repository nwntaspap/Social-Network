'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import Image from 'next/image';
import Link from 'next/link';
import {
  getGroupPosts,
  getGroupEvents,
  getGroupPostComments,
  createGroupPostComment,
  respondToEvent,
  getEventRSVPs,
  getGroupMembers,
} from '@/lib/api';
import PostCard from '@/components/features/home/PostCard';
import EditEventForm from './EditEventForm';
import GroupChatPanel from './GroupChatPanel';
import { formatRelativeDate, getDisplayName, getFileUrl } from '@/lib/helpers';
import type { Comment, Event, EventRSVPOption, GroupMember, Post } from '@/lib/types';
import { tabView } from './GroupDetail';

interface GroupContentProps {
  groupId: string;
  activeTab: tabView;
  isMember: boolean;
  isCreator: boolean;
  eventRefreshKey: number;
}

const PAGE_SIZE = 10;

export default function GroupContent({
  groupId,
  activeTab,
  isMember,
  isCreator,
  eventRefreshKey,
}: GroupContentProps) {
  return (
    <div className="group-content-area">
      {activeTab === 'posts' ? (
        <PostsTab groupId={groupId} />
      ) : activeTab === 'chat' ? (
        <ChatTab groupId={groupId} />
      ) : activeTab === 'members' ? (
        <MembersTab groupId={groupId} />
      ) : (
        <EventsTab
          groupId={groupId}
          isCreator={isCreator}
          isMember={isMember}
          refreshKey={eventRefreshKey}
        />
      )}
    </div>
  );
}

function ChatTab({ groupId }: { groupId: string }) {
  return <GroupChatPanel groupId={groupId} />;
}

function MembersTab({ groupId }: { groupId: string }) {
  const [members, setMembers] = useState<GroupMember[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    let ignore = false;
    getGroupMembers(groupId, 1, 100)
      .then((response) => {
        if (!ignore) setMembers(response.data);
      })
      .catch(() => {
        if (!ignore) setError('Failed to load members.');
      })
      .finally(() => {
        if (!ignore) setLoading(false);
      });
    return () => {
      ignore = true;
    };
  }, [groupId]);

  if (loading) return <p className="group-empty-state">Loading members...</p>;
  if (error) return <p className="group-empty-state">{error}</p>;
  if (members.length === 0) return <p className="group-empty-state">No members yet.</p>;

  const creators = members.filter((m) => m.role === 'creator');
  const admins = members.filter((m) => m.role === 'admin');
  const regular = members.filter((m) => m.role === 'member');

  function renderRow(m: GroupMember) {
    return (
      <div key={m.userId} className="group-member-row">
        <Link href={`/profile/${m.userId}`} className="group-member-link">
          <Image
            src={getFileUrl(m.user?.avatarUrl)}
            alt={getDisplayName(m.user)}
            width={40}
            height={40}
            className="group-member-avatar"
          />
          <div className="group-member-info">
            <span className="group-member-name">{getDisplayName(m.user)}</span>
            <span className="group-member-username">@{m.user?.username}</span>
          </div>
        </Link>
        {m.role !== 'member' && <span className="group-member-role">{m.role}</span>}
      </div>
    );
  }

  return (
    <div className="group-members">
      {creators.length > 0 && (
        <>
          <h4 className="group-members-section-title">Creator</h4>
          {creators.map(renderRow)}
        </>
      )}
      {admins.length > 0 && (
        <>
          <h4 className="group-members-section-title">Admins</h4>
          {admins.map(renderRow)}
        </>
      )}
      {regular.length > 0 && (
        <>
          <h4 className="group-members-section-title">Members</h4>
          {regular.map(renderRow)}
        </>
      )}
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
  const [expandedComments, setExpandedComments] = useState<Set<string>>(new Set());

  const toggleComments = useCallback((postId: string) => {
    setExpandedComments((prev) => {
      const next = new Set(prev);
      if (next.has(postId)) {
        next.delete(postId);
      } else {
        next.add(postId);
      }
      return next;
    });
  }, []);

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
      {posts.map((post) => {
        const expanded = expandedComments.has(post.id);
        return (
          <div key={post.id}>
            <PostCard post={post} commentsExpanded={expanded} onToggleComments={toggleComments} />
            <PostComments postId={post.id} expanded={expanded} />
          </div>
        );
      })}
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

function PostComments({ postId, expanded }: { postId: string; expanded: boolean }) {
  const [comments, setComments] = useState<Comment[]>([]);
  const [content, setContent] = useState('');
  const [image, setImage] = useState<File | null>(null);
  const [imagePreview, setImagePreview] = useState('');
  const [loaded, setLoaded] = useState(false);
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const imageInputRef = useRef<HTMLInputElement>(null);

  const loading = expanded && !loaded && !error;

  useEffect(() => {
    if (!expanded || loaded) return;
    let ignore = false;
    getGroupPostComments(postId)
      .then((response) => {
        if (ignore) return;
        setComments(response.data);
        setLoaded(true);
      })
      .catch(() => {
        if (!ignore) setError('Failed to load comments.');
      });
    return () => {
      ignore = true;
    };
  }, [expanded, loaded, postId]);

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
      {expanded && (
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

interface EventsTabProps {
  groupId: string;
  isMember: boolean;
  isCreator: boolean;
  refreshKey: number;
}

function EventsTab({ groupId, isMember, isCreator, refreshKey }: EventsTabProps) {
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
  }, [groupId, refreshKey]);

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
          <EventCard
            key={event.id}
            event={event}
            groupId={groupId}
            isMember={isMember}
            isCreator={isCreator}
            onRefresh={() => load()}
          />
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

interface EventCardProps {
  event: Event;
  groupId: string;
  isMember: boolean;
  isCreator: boolean;
  onRefresh: () => void;
}

function EventCard({ event, groupId, isMember, isCreator, onRefresh }: EventCardProps) {
  const [showAttendees, setShowAttendees] = useState(false);
  const [attendees, setAttendees] = useState<EventRSVPOption[] | null>(null);
  const [attendeeError, setAttendeeError] = useState('');
  const [showEdit, setShowEdit] = useState(false);
  const [rsvpError, setRsvpError] = useState('');

  async function handleRSVP(optionId: string) {
    setRsvpError('');
    try {
      await respondToEvent(event.id, optionId);
      onRefresh();
    } catch {
      setRsvpError('Failed to save your response. Please try again.');
    }
  }

  async function toggleAttendees() {
    setShowAttendees((prev) => !prev);
    if (attendees === null) {
      try {
        const response = await getEventRSVPs(event.id);
        setAttendees(response.options);
        setAttendeeError('');
      } catch {
        setAttendeeError('Failed to load attendees.');
      }
    }
  }

  return (
    <div className="group-event-card">
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
          <span className="event-creator">Created by {getDisplayName(event.creator)}</span>
          <span className="event-time">{formatRelativeDate(event.eventDate)}</span>
        </div>

        <div className="event-options">
          {event.options.map((option) => (
            <button
              key={option.id}
              type="button"
              className="event-option"
              disabled={!isMember}
              onClick={() => handleRSVP(option.id)}
              title="Click to select this option"
            >
              <span className="event-option-label">{option.label}</span>
              <span className="event-option-tally">{option.tally}</span>
            </button>
          ))}
        </div>

        {rsvpError && <p className="comment-error">{rsvpError}</p>}

        <div className="event-actions">
          <button type="button" className="event-attendees-toggle" onClick={toggleAttendees}>
            {showAttendees ? 'Hide attendees' : 'View attendees'}
          </button>
          {isCreator && (
            <button
              type="button"
              className="event-edit-btn"
              onClick={() => setShowEdit((prev) => !prev)}
            >
              Edit
            </button>
          )}
        </div>

        {showAttendees && (
          <div className="event-attendees">
            {attendeeError && <p className="comment-error">{attendeeError}</p>}
            {attendees === null && <p className="group-empty-state">Loading attendees...</p>}
            {attendees !== null &&
              attendees.map((option) => (
                <div key={option.optionId} className="event-attendees-option">
                  <span className="event-attendees-label">{option.optionLabel}:</span>
                  {option.users.length > 0 ? (
                    <ul className="event-attendees-list">
                      {option.users.map((user) => (
                        <li key={user.id} className="event-attendee">
                          <Link href={`/profile/${user.id}`} className="event-attendee-link">
                            {getDisplayName(user)}
                          </Link>
                        </li>
                      ))}
                    </ul>
                  ) : (
                    <span className="event-attendees-empty">No one yet</span>
                  )}
                </div>
              ))}
          </div>
        )}

        {showEdit && (
          <EditEventForm
            groupId={groupId}
            event={event}
            onClose={() => setShowEdit(false)}
            onUpdated={onRefresh}
          />
        )}
      </div>
    </div>
  );
}
