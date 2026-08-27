import AuthGuard from '@/components/AuthGuard';
import ProfileContent from '@/components/features/profile/ProfileContent';

export default function ProfilePage() {
  return (
    <AuthGuard>
      <ProfileContent />
    </AuthGuard>
  );
}
