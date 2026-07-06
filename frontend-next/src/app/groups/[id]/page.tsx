import AuthGuard from '@/components/AuthGuard';
import GroupDetail from '@/components/features/groups/GroupDetail';

export default function GroupPage() {
  return (
    <AuthGuard>
      <GroupDetail />
    </AuthGuard>
  );
}
