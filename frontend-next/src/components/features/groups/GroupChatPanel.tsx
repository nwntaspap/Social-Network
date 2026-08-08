'use client';

/**
 * components/features/groups/GroupChatPanel.tsx
 *
 * Group page chat tab. Thin wrapper that owns the WebSocket connection and
 * delegates the UI to the reusable GroupChatRoom (also used by the chat
 * widget's Groups tab).
 */

import { useEffect } from 'react';
import { chatSocket } from '@/lib/ws';
import { useGroupPresence } from '@/lib/useGroupPresence';
import GroupChatRoom from './GroupChatRoom';

interface GroupChatPanelProps {
  groupId: string;
}

export default function GroupChatPanel({ groupId }: GroupChatPanelProps) {
  const presence = useGroupPresence(groupId);

  useEffect(() => {
    chatSocket.connect();
    return () => chatSocket.disconnect();
  }, []);

  return <GroupChatRoom groupId={groupId} presence={presence} />;
}
