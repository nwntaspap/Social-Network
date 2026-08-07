import AuthGuard from '@/components/AuthGuard';
import ChatPage from '@/components/features/chat/ChatPage';

export default function ChatRoute() {
  return (
    <AuthGuard>
      <ChatPage />
    </AuthGuard>
  );
}
