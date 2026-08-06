'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import GroupHeader from './GroupHeader';
import GroupContent from './GroupContent';
import GroupSidebar from './GroupSidebar';
import { getGroup } from '@/lib/api';
import { useAuth } from '@/context/AuthContext';
import type { Group } from '@/lib/types';

export type tabView = 'posts' | 'events';

export default function GroupDetail() {
  const { id } = useParams<{ id: string }>();
  const { user } = useAuth();
  const [activeTab, setActiveTab] = useState<tabView>('posts');
  const [group, setGroup] = useState<Group | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [eventRefreshKey, setEventRefreshKey] = useState(0);

  useEffect(() => {
    let cancelled = false;
    getGroup(id)
      .then((g) => {
        if (!cancelled) setGroup(g);
      })
      .catch(() => {
        if (!cancelled) setError('Group not found.');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [id]);

  function handleGroupUpdated() {
    getGroup(id)
      .then(setGroup)
      .catch(() => setError('Failed to refresh group.'));
  }

  function handleEventCreated() {
    setEventRefreshKey((key) => key + 1);
  }

  if (loading) {
    return (
      <div className="groups-container">
        <p>Loading group...</p>
      </div>
    );
  }

  if (error || !group) {
    return (
      <div className="groups-container">
        <p>{error || 'Group not found.'}</p>
      </div>
    );
  }

  const isMember = group.membershipStatus === 'member';
  const isCreator = user?.id === group.creatorId;

  if (!isMember) {
    return (
      <div className="group-detail-container">
        <GroupHeader
          group={group}
          isCreator={isCreator}
          isMember={isMember}
          onGroupUpdated={handleGroupUpdated}
        />
        <div className="group-not-member">
          <p>You must join the group in order to see the posts and events.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="group-detail-container">
      <GroupHeader
        group={group}
        isCreator={isCreator}
        isMember={isMember}
        onGroupUpdated={handleGroupUpdated}
        onEventCreated={handleEventCreated}
      />

      {isCreator ? (
        // Creator: sidebar + main content with tabs
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
            <GroupContent
              groupId={group.id}
              activeTab={activeTab}
              isMember={isMember}
              isCreator={isCreator}
              eventRefreshKey={eventRefreshKey}
            />
          </div>
        </div>
      ) : (
        // Non-creator member: just centered content with tabs
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
          <GroupContent
            groupId={group.id}
            activeTab={activeTab}
            isMember={isMember}
            isCreator={isCreator}
            eventRefreshKey={eventRefreshKey}
          />
        </div>
      )}
    </div>
  );
}
