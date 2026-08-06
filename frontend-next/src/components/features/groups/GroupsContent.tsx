'use client';

import { useState } from 'react';
import GroupSearchBar from './GroupSearchBar';
import GroupList from './GroupList';
import CreateGroupSection from './CreateGroupSection';
import GroupInvitations from './GroupInvitations';

export default function GroupsContent() {
  const [refreshKey, setRefreshKey] = useState(0);
  const [query, setQuery] = useState('');

  return (
    <div className="groups-container">
      <h1 className="groups-title">Groups</h1>
      <GroupSearchBar query={query} onQueryChange={setQuery} />
      <GroupInvitations />
      <CreateGroupSection onCreated={() => setRefreshKey((k) => k + 1)} />
      <GroupList refreshKey={refreshKey} query={query} />
    </div>
  );
}
