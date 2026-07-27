import type { GroupInvitation } from '@/lib/types';
import { mockGroups } from './groups';
import { searchResults } from './users';

export const mockGroupInvitations: Record<string, GroupInvitation[]> = {
  '1': [
    {
      id: 'inv1',
      groupId: '1',
      group: mockGroups[0],
      inviterId: '1',
      inviter: mockGroups[0].creator,
      inviteeId: '7', // Chris Pratt
      invitee: searchResults[0],
      status: 'pending',
      createdAt: '2026-07-20T10:00:00Z',
    },
    {
      id: 'inv2',
      groupId: '1',
      group: mockGroups[0],
      inviterId: '1',
      inviter: mockGroups[0].creator,
      inviteeId: '8', // Emma Watson
      invitee: searchResults[1],
      status: 'pending',
      createdAt: '2026-07-20T11:00:00Z',
    },
  ],
  '2': [],
  '3': [],
  '4': [],
  '5': [
    {
      id: 'inv3',
      groupId: '5',
      group: mockGroups[4],
      inviterId: '3',
      inviter: mockGroups[4].creator,
      inviteeId: '7', // Chris Pratt
      invitee: searchResults[0],
      status: 'pending',
      createdAt: '2026-07-19T09:00:00Z',
    },
  ],
  '6': [],
};
