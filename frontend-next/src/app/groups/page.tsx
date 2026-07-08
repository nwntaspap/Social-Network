import AuthGuard from '@/components/AuthGuard';
import GroupsContent from '@/components/features/groups/GroupsContent';

export default function GroupsPage() {
  return (
    <AuthGuard>
      <GroupsContent />
    </AuthGuard>
  );
}
