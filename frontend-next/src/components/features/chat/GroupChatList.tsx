'use client';

/**
 * components/features/chat/GroupChatList.tsx
 *
 * Sidebar listing the groups the current user belongs to, for the chat
 * widget's Groups tab. Clicking a group opens its chat room.
 */

import type { Group } from '@/lib/types';

interface GroupChatListProps {
  groups: Group[];
  /** Live unread counts keyed by group id. */
  unreadByGroup?: Record<string, number>;
  onSelect: (group: Group) => void;
}

export default function GroupChatList({ groups, unreadByGroup, onSelect }: GroupChatListProps) {
  if (groups.length === 0) {
    return (
      <div className="chat-users-list">
        <div className="chat-empty-state">
          <span className="chat-empty-state-icon">👥</span>
          <span className="chat-empty-state-text">You are not in any groups yet</span>
        </div>
      </div>
    );
  }

  return (
    <div className="chat-users-list">
      {groups.map((group) => {
        const unread = unreadByGroup?.[group.id] ?? group.unreadCount ?? 0;
        return (
          <button
            key={group.id}
            type="button"
            className={`chat-user-item${unread > 0 ? ' has-unread' : ''}`}
            onClick={() => onSelect(group)}
          >
            <span className="chat-user-avatar">👥</span>
            <span className="chat-user-info">
              <span className="chat-user-name">{group.title}</span>
              <span className="chat-user-status">
                {group.membersCount} {group.membersCount === 1 ? 'member' : 'members'}
              </span>
            </span>
            {unread > 0 && <span className="chat-user-unread">{unread}</span>}
          </button>
        );
      })}
    </div>
  );
}
