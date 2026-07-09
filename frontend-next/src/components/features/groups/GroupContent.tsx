import { tabView } from './GroupDetail';

interface GroupContentProps {
  groupId: string;
  activeTab: tabView;
  isMember: boolean;
}
// eslint-disable-next-line @typescript-eslint/no-unused-vars
export default function GroupContent({ groupId, activeTab }: GroupContentProps) {
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
