import { tabView } from './GroupDetail';

interface GroupContentProps {
  groupId: string;
  activeTab: tabView;
  isMember: boolean;
}

export default function GroupContent({ groupId, activeTab, isMember }: GroupContentProps) {
  if (!isMember) {
    return (
      <div className="group-not-member">
        <p>You must join the group in order to see the posts and events.</p>
      </div>
    );
  }

  return (
    <div className="group-content-area">
      {activeTab === 'posts' ? (
        <div className="group-posts">
          {/* TODO: Paginated posts */}
          <p>Posts will appear here</p>
        </div>
      ) : (
        <div className="group-events">
          {/* TODO: Paginated events */}
          <p>Events will appear here</p>
        </div>
      )}
    </div>
  );
}
