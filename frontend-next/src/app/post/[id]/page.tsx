import AuthGuard from '@/components/AuthGuard';
import PostDetail from '@/components/features/post/PostDetail';

export default function PostPage() {
  return (
    <AuthGuard>
      <PostDetail />
    </AuthGuard>
  );
}
