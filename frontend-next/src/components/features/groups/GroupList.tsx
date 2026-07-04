'use client';

/**
 * components/features/groups/GroupList.tsx
 *
 * Grid of group cards with join/leave/pending button.
 *
 * When the backend is ready:
 *   - Replace mockGroups with API call
 *   - Wire buttons to requestToJoinGroup / leaveGroup
 *   - Check user's membership status for each group
 */

import { mockGroups } from '@/mocks/groups';
import GroupCard from './GroupCard';

export default function GroupList() {
  // TODO: fetch groups from API when backend is ready
  // const [groups, setGroups] = useState<Group[]>([]);
  // useEffect(() => { browseGroups().then(res => setGroups(res.data)) }, []);

  return (
    <div className="groups-grid">
      {mockGroups.map((group) => (
        <GroupCard key={group.id} group={group} />
      ))}
    </div>
  );
}
