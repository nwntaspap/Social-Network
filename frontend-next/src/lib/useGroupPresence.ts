'use client';

/**
 * lib/useGroupPresence.ts
 *
 * Live "how many group members are online" indicator. Fetches the presence
 * snapshot over HTTP, then recomputes the online count locally as
 * isOnlineStatus.update broadcasts arrive, so no extra requests are needed
 * while the room is open. The presence endpoint returns the member id list so
 * updates for non-members are simply ignored.
 *
 * The hook is meant to be mounted while a group room is open (i.e. with a
 * stable, non-null groupId); host components remount it per group.
 */

import { useEffect, useState } from 'react';
import { getGroupPresence } from './api';
import { chatSocket } from './ws';
import type { GroupPresence, GroupPresenceMember, IsOnlineStatusPayload } from './types';

export interface GroupPresenceState {
  total: number;
  online: number;
  loading: boolean;
  members?: GroupPresenceMember[];
}

const EMPTY: GroupPresenceState = { total: 0, online: 0, loading: true, members: [] };

export function useGroupPresence(groupId: string): GroupPresenceState {
  const [state, setState] = useState<GroupPresenceState>(EMPTY);

  useEffect(() => {
    let ignore = false;

    const onlineByMember = new Map<string, boolean>();
    let memberIds: string[] = [];

    const recompute = () => {
      if (ignore) return;
      const members = memberIds.map((id) => ({ id, isOnline: onlineByMember.get(id) === true }));
      const online = members.reduce((count, m) => count + (m.isOnline ? 1 : 0), 0);
      setState({ total: memberIds.length, online, loading: false, members });
    };

    const load = () => {
      getGroupPresence(groupId)
        .then((presence: GroupPresence) => {
          if (ignore) return;
          memberIds = presence.members.map((m) => m.id);
          presence.members.forEach((m) => onlineByMember.set(m.id, m.isOnline));
          recompute();
        })
        .catch(() => {
          if (!ignore) setState({ total: 0, online: 0, loading: false, members: [] });
        });
    };

    // Fetch the snapshot once connected. If a reconnect happens later, missed
    // broadcasts are recovered by refetching the snapshot on the next open.
    if (chatSocket.isConnected()) {
      load();
    }
    const unsubscribeConnection = chatSocket.onConnection(() => {
      if (chatSocket.isConnected()) load();
    });

    const unsubscribe = chatSocket.on('isOnlineStatus.update', (payload) => {
      const p = payload as IsOnlineStatusPayload;
      if (!p || typeof p.user_id !== 'string') return;
      if (memberIds.includes(p.user_id)) {
        onlineByMember.set(p.user_id, p.isOnline);
        recompute();
      }
    });

    return () => {
      ignore = true;
      unsubscribeConnection();
      unsubscribe();
    };
  }, [groupId]);

  return state;
}
