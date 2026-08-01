'use client';

import { useCallback, useEffect, useState } from 'react';
import { getGroupPosts, getGroupEvents } from '@/lib/api';
import PostCard from '@/components/features/home/PostCard';
import { formatRelativeDate } from '@/lib/helpers';
import type { Event, Post } from '@/lib/types';
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
        <PostCard key={post.id} post={post} />
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
