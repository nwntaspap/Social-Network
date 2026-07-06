'use client';

// import { useState } from 'react';
import { useParams } from 'next/navigation';
import { mockGroups } from '@/mocks/groups';

// type tabView = 'posts' | 'events';

export default function GroupDetail() {
  const { id } = useParams<{ id: string }>();
  // const [activeTab, setActiveTab] = useState<tabView>('posts');

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
      {isMember && <p>I am member</p>}
      {isCreator && <p>I am creator</p>}
    </div>
  );
}
