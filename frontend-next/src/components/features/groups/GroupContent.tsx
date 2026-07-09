import { mockGroupPosts, defaultGroupPosts } from '@/mocks/group-posts';
import { mockGroupEvents, defaultGroupEvents } from '@/mocks/group-events';
import PostCard from '@/components/features/home/PostCard';
import { formatRelativeDate } from '@/lib/helpers';
import { tabView } from './GroupDetail';

interface GroupContentProps {
  groupId: string;
  activeTab: tabView;
  isMember: boolean;
}

export default function GroupContent({ groupId, activeTab }: GroupContentProps) {
  if (activeTab === 'posts') {
    const posts = mockGroupPosts[groupId] || defaultGroupPosts;

    return (
      <div className="group-content-area">
        {posts.length > 0 ? (
          <div className="group-posts">
            {posts.map((post) => (
              <PostCard key={post.id} post={post} />
            ))}
          </div>
        ) : (
          <div className="group-posts">
            <p className="group-empty-state">No posts yet. Be the first to create one!</p>
          </div>
        )}
      </div>
    );
  }

  // Events tab
  const events = mockGroupEvents[groupId] || defaultGroupEvents;

  return (
    <div className="group-content-area">
      {events.length > 0 ? (
        <div className="group-events">
          {events.map((event) => (
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
                    Created by {event.creator.firstName} {event.creator.lastName}
                  </span>
                  <span className="event-time">{formatRelativeDate(event.eventDate)}</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div className="group-events">
          <p className="group-empty-state">No events scheduled yet.</p>
        </div>
      )}
    </div>
  );
}
