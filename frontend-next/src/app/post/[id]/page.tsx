<<<<<<< HEAD
import AuthGuard from '@/components/AuthGuard';
import PostDetail from '@/components/features/post/PostDetail';

export default function PostPage() {
  return (
    <AuthGuard>
      <PostDetail />
    </AuthGuard>
  );
=======
export default function PostPage({ params }: { params: Promise<{ id: string }> }) {
  return <div>Post</div>;
>>>>>>> 709106ca (fix(platform): add stub pages for /create and /post/[id] to fix tsc gate)
}
