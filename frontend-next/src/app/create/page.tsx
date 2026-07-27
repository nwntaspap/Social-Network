import AuthGuard from '@/components/AuthGuard';
import CreatePostForm from '@/components/features/post/CreatePostForm';

export default function createPostPage() {
  return (
    <AuthGuard>
      <CreatePostForm />
    </AuthGuard>
  );
}
