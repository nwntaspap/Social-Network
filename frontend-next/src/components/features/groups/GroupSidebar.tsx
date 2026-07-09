import { useState } from 'react';

interface GroupSidebarProps {
  groupId: string;
  isCreator: boolean;
}
// eslint-disable-next-line @typescript-eslint/no-unused-vars
export default function GroupSidebar({ groupId, isCreator }: GroupSidebarProps) {
  const [showPending, setShowPending] = useState(false);

  if (!isCreator) return null;

  return (
    <aside className="group-sidebar">
      <button className="group-sidebar-btn" onClick={() => setShowPending(!showPending)}>
        Pending Requests
      </button>

      {showPending && (
        <div className="group-pending-dropdown">
          {/* TODO: Fetch and display pending join requests */}
          <p className="group-pending-empty">No pending requests</p>
        </div>
      )}
    </aside>
  );
}
