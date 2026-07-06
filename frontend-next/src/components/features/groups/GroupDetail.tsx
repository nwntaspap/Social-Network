'use client';

import { useState } from 'react';
import { useParams } from 'next/navigation';
import GroupHeader from './GroupHeader';
import GroupContent from './GroupContent';
import GroupSidebar from './GroupSidebar';
import { mockGroups } from '@/mocks/groups';

export type tabView = 'posts' | 'events';

export default function GroupDetail() {
  const { id } = useParams<{ id: string }>();
  const [activeTab, setActiveTab] = useState<tabView>('posts');

  // TODO: Fetch group from API when backend is ready
  const group = mockGroups.find((g) => g.id === id);

  if (!group) {
    return (
      <div className="groups-container">
        <p>Group not found.</p>
      </div>
    );
  }

  const isMember = group.membershipStatus === 'member';
  const isCreator = group.creatorId === '1'; // TODO: compare with actual current user ID

  return (
    <div className="group-detail-container">
      <GroupHeader group={group} isCreator={isCreator} isMember={isMember} />

      <div className="group-detail-layout">
        <GroupSidebar groupId={group.id} isCreator={isCreator} />

        <div className="group-main-content">
          <div className="group-tabs">
            <button
              className={`group-tab ${activeTab === 'posts' ? 'group-tab--active' : ''}`}
              onClick={() => setActiveTab('posts')}
            >
              Posts
            </button>
            <button
              className={`group-tab ${activeTab === 'events' ? 'group-tab--active' : ''}`}
              onClick={() => setActiveTab('events')}
            >
              Events
            </button>
          </div>

          <GroupContent groupId={group.id} activeTab={activeTab} isMember={isMember} />
        </div>
      </div>
    </div>
  );
}
