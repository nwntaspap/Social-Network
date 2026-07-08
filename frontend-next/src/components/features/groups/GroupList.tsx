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
  // const [loading, setLoading] = useState(true);

  // useEffect(() => {
  //   async function fetchGroups() {
  //     try {
  //       const response = await browseGroups();
  //       setGroups(response.data);
  //     } catch (error) {
  //       console.error('Failed to fetch groups:', error);
  //     } finally {
  //       setLoading(false);
  //     }
  //   }

  //   fetchGroups();
  // }, []);

  // const handleStatusChange = (groupId: string, newStatus: MembershipStatus) => {
  //   setGroups((prevGroups) =>
  //     prevGroups.map((group) =>
  //       group.id === groupId ? { ...group, membershipStatus: newStatus } : group
  //     )
  //   );
  // };

  // if (loading) {
  //   return <div>Loading groups...</div>;
  // }

  return (
    <div className="groups-grid">
      {mockGroups.map((group) => (
        // onStatusChange={handleStatusChange} later
        <GroupCard key={group.id} group={group} />
      ))}
    </div>
  );
}
