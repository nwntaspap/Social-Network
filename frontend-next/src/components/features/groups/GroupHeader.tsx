import { formatRelativeDate } from '@/lib/helpers';
import { Group } from '@/lib/types';
import Link from 'next/link';

interface GroupHeaderProps {
  group: Group;
  isCreator: boolean;
  isMember: boolean;
}

export default function GroupHeader({ group, isCreator, isMember }: GroupHeaderProps) {
  return (
    <div className="group-detail-header">
      <div className="group-detail-header-top">
        <div className="group-detail-icon">{group.title[0].toUpperCase()}</div>
        <div className="group-detail-info">
          <h1 className="group-detail-title">{group.title}</h1>
          <p className="group-detail-meta">
            {group.membersCount} members · Created {formatRelativeDate(group.createdAt)}
          </p>
          <p className="group-detail-desc">{group.description}</p>
        </div>
      </div>

      <div className="group-detail-actions">
        {isMember && (
          <>
            <Link href={`/create?groupId=${group.id}`} className="group-action-btn">
              Create Post
            </Link>
            <button className="group-action-btn">Invite User</button>
            <button className="group-action-btn">Create Event</button>
          </>
        )}

        {isCreator && (
          <>
            <button className="group-action-btn group-action-btn--danger">Edit Group</button>
            <button className="group-action-btn group-action-btn--delete">Delete Group</button>
          </>
        )}
      </div>
    </div>
  );
}
