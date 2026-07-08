import AuthGuard from '@/components/AuthGuard';
import HomeContent from '@/components/features/home/HomeContent';

export default function HomePage() {
  return (
    <AuthGuard>
      <HomeContent />
    </AuthGuard>
  );
}
